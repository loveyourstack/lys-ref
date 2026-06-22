package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys/lyserr"
)

const (
	dmMcpCallToolRequestID = 2 // arbitrary ID for the tool call request; used to match the response in dmExtractToolResult
)

// dmMcpQuery maps a limited set of NL demo queries to MCP tool calls and returns the tool result.
func (srvApp *httpServerApplication) dmMcpQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := lys.ExtractJsonBody(r, srvApp.PostOptions.MaxBodySize)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.ExtractJsonBody failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	type mcpBridgeInput struct {
		Query string `json:"query"`
	}

	inp, err := lys.DecodeJsonBody[mcpBridgeInput](body)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.DecodeJsonBody failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	mapping, err := dmMapNlToMcpQuery(inp.Query)
	if err != nil {
		lys.HandleUserError(lyserr.User{Message: err.Error(), StatusCode: http.StatusBadRequest}, w)
		return
	}

	mcpCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	toolResult, err := dmRunMcpToolCall(mcpCtx, srvApp.Config.McpServer, mapping.Tool, mapping.Args)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("dmRunMcpToolCall failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data: map[string]any{
			"query":       inp.Query,
			"mapped_tool": mapping.Tool,
			"mapped_args": mapping.Args,
			"mcp_result":  toolResult,
		},
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

type mcpQueryMapping struct {
	Tool string
	Args map[string]any
}

// dmMapNlToMcpQuery maps a natural language query to an MCP tool call (tool name + arguments).
// In a real app, this would be done by an LLM.
func dmMapNlToMcpQuery(query string) (mcpQueryMapping, error) {
	normalized := dmMcpNormalizeQuery(query)

	switch normalized {
	case "show me revenue by day for the last 7 days":
		return mcpQueryMapping{
			Tool: "get_daily_trend",
			Args: map[string]any{"days": 7},
		}, nil
	case "give me a summary of this week's performance":
		return mcpQueryMapping{
			Tool: "get_performance_summary",
			Args: map[string]any{"period": "week"},
		}, nil
	case "which campaigns had the highest roi in the last month":
		return mcpQueryMapping{
			Tool: "get_top_campaigns",
			Args: map[string]any{"period": "month", "order_by": "return_on_investment", "limit": 10},
		}, nil
	case "what are the best performing verticals today":
		return mcpQueryMapping{
			Tool: "get_vertical_performance",
			Args: map[string]any{"period": "day", "order_by": "profit_eur"},
		}, nil
	default:
		return mcpQueryMapping{}, fmt.Errorf("unsupported query: allowed demo queries only")
	}
}

func dmMcpNormalizeQuery(query string) string {
	s := strings.ToLower(strings.TrimSpace(query))
	s = strings.TrimSuffix(s, ".")
	s = strings.TrimSuffix(s, "?")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// dmRunMcpToolCall executes the mcpsrv binary with the given tool and arguments, and returns the result.
// In a real app, an MCP host, such as an LLM that supports tool calling, would be used which calls mcpsrv internally.
func dmRunMcpToolCall(ctx context.Context, mcpCfg myapp.McpServer, tool string, args map[string]any) (any, error) {

	// execute mcpsrv as a subprocess
	cmd := exec.CommandContext(ctx, mcpCfg.BinaryPath, "-config", mcpCfg.ConfigFilePath)

	// build JSON-RPC payload to call the tool with arguments
	payload, err := dmBuildMcpPayload(tool, args)
	if err != nil {
		return nil, fmt.Errorf("dmBuildMcpPayload failed: %w", err)
	}

	// set payload as stdin for the mcpsrv process
	cmd.Stdin = bytes.NewReader(payload)

	// capture stdout and stderr
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("mcpsrv failed: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("failed to execute mcpsrv: %w", err)
	}

	// parse stdout to extract the tool call result
	result, err := dmExtractToolResult(out, dmMcpCallToolRequestID)
	if err != nil {
		return nil, fmt.Errorf("dmExtractToolResult failed: %w", err)
	}

	return result, nil
}

// dmBuildMcpPayload constructs the JSON-RPC payload to call the specified tool with arguments.
func dmBuildMcpPayload(tool string, args map[string]any) ([]byte, error) {
	initMsg := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-11-25",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "refsrv-mcp-bridge",
				"version": "1.0",
			},
		},
	}

	initializedMsg := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
		"params":  map[string]any{},
	}

	callMsg := map[string]any{
		"jsonrpc": "2.0",
		"id":      dmMcpCallToolRequestID,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      tool,
			"arguments": args,
		},
	}

	msgs := []map[string]any{initMsg, initializedMsg, callMsg}

	var lines []string
	for _, msg := range msgs {
		b, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		lines = append(lines, string(b))
	}

	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

// dmExtractToolResult parses the stdout from mcpsrv to find the response for the tool call with the specified ID, and returns the result or error.
func dmExtractToolResult(stdout []byte, toolCallID int) (any, error) {
	lines := strings.SplitSeq(string(stdout), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var msg struct {
			ID     *json.RawMessage `json:"id,omitempty"`
			Result json.RawMessage  `json:"result,omitempty"`
			Error  *struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error,omitempty"`
		}

		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}

		if msg.ID == nil {
			continue
		}

		var id int
		if err := json.Unmarshal(*msg.ID, &id); err != nil {
			continue
		}

		if id != toolCallID {
			continue
		}

		if msg.Error != nil {
			return nil, fmt.Errorf("mcp error %d: %s", msg.Error.Code, msg.Error.Message)
		}

		if len(msg.Result) == 0 {
			return nil, fmt.Errorf("missing result payload")
		}

		var result any
		if err := json.Unmarshal(msg.Result, &result); err != nil {
			return nil, err
		}

		return result, nil
	}

	return nil, fmt.Errorf("tool call response not found")
}

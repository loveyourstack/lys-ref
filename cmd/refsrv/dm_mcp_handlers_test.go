package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDmMcpNormalizeQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{name: "already normalized", query: "show me revenue", want: "show me revenue"},
		{name: "upper case", query: "Show Me REVENUE", want: "show me revenue"},
		{name: "surrounding whitespace", query: "  \tshow me revenue\n ", want: "show me revenue"},
		{name: "inner whitespace collapsed", query: "show   me \t revenue", want: "show me revenue"},
		{name: "trailing question mark", query: "show me revenue?", want: "show me revenue"},
		{name: "trailing period", query: "show me revenue.", want: "show me revenue"},
		{name: "trailing period then whitespace", query: "show me revenue.  ", want: "show me revenue"},
		{name: "apostrophe kept", query: "This week's performance", want: "this week's performance"},
		{name: "empty", query: "", want: ""},
		{name: "whitespace only", query: " \t ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dmMcpNormalizeQuery(tt.query); got != tt.want {
				t.Errorf("dmMcpNormalizeQuery(%q) = %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}

func TestDmMapNlToMcpQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    mcpQueryMapping
		wantErr bool
	}{
		{
			name:  "daily trend",
			query: "Show me revenue by day for the last 7 days",
			want:  mcpQueryMapping{Tool: "get_daily_trend", Args: map[string]any{"days": 7}},
		},
		{
			name:  "performance summary",
			query: "Give me a summary of this week's performance.",
			want:  mcpQueryMapping{Tool: "get_performance_summary", Args: map[string]any{"period": "week"}},
		},
		{
			name:  "top campaigns",
			query: "  Which campaigns had the highest ROI in the last month?  ",
			want: mcpQueryMapping{
				Tool: "get_top_campaigns",
				Args: map[string]any{"period": "month", "order_by": "return_on_investment", "limit": 10},
			},
		},
		{
			name:  "vertical performance",
			query: "What are the best   performing verticals today",
			want: mcpQueryMapping{
				Tool: "get_vertical_performance",
				Args: map[string]any{"period": "day", "order_by": "profit_eur"},
			},
		},
		{name: "unsupported", query: "drop table users", wantErr: true},
		{name: "partial match", query: "show me revenue by day", wantErr: true},
		{name: "empty", query: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dmMapNlToMcpQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Fatalf("dmMapNlToMcpQuery(%q) error = %v, wantErr %v", tt.query, err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dmMapNlToMcpQuery(%q) = %+v, want %+v", tt.query, got, tt.want)
			}
		})
	}
}

func TestDmBuildMcpPayload(t *testing.T) {
	tests := []struct {
		name string
		tool string
		args map[string]any
		want map[string]any // expected params of the tools/call message
	}{
		{
			name: "with args",
			tool: "get_daily_trend",
			args: map[string]any{"days": 7},
			want: map[string]any{"name": "get_daily_trend", "arguments": map[string]any{"days": float64(7)}},
		},
		{
			name: "nil args are sent as an empty object",
			tool: "get_performance_summary",
			args: nil,
			want: map[string]any{"name": "get_performance_summary", "arguments": map[string]any{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := dmBuildMcpPayload(tt.tool, tt.args)
			if err != nil {
				t.Fatalf("dmBuildMcpPayload() error = %v", err)
			}

			// one JSON message per line, terminated by a newline
			if !strings.HasSuffix(string(payload), "\n") {
				t.Errorf("payload does not end with a newline: %q", payload)
			}
			lines := strings.Split(strings.TrimSuffix(string(payload), "\n"), "\n")
			if len(lines) != 3 {
				t.Fatalf("payload has %d lines, want 3: %q", len(lines), payload)
			}

			msgs := make([]map[string]any, len(lines))
			for i, line := range lines {
				if err := json.Unmarshal([]byte(line), &msgs[i]); err != nil {
					t.Fatalf("line %d is not valid JSON: %v: %q", i, err, line)
				}
				if msgs[i]["jsonrpc"] != "2.0" {
					t.Errorf("line %d jsonrpc = %v, want 2.0", i, msgs[i]["jsonrpc"])
				}
			}

			if msgs[0]["method"] != "initialize" || msgs[0]["id"] != float64(1) {
				t.Errorf("first message = %v, want initialize with id 1", msgs[0])
			}

			// a notification must not have an id
			if msgs[1]["method"] != "notifications/initialized" {
				t.Errorf("second message method = %v, want notifications/initialized", msgs[1]["method"])
			}
			if _, ok := msgs[1]["id"]; ok {
				t.Errorf("notification has an id: %v", msgs[1])
			}

			if msgs[2]["method"] != "tools/call" {
				t.Errorf("third message method = %v, want tools/call", msgs[2]["method"])
			}
			if msgs[2]["id"] != float64(dmMcpCallToolRequestID) {
				t.Errorf("tool call id = %v, want %d", msgs[2]["id"], dmMcpCallToolRequestID)
			}
			if !reflect.DeepEqual(msgs[2]["params"], tt.want) {
				t.Errorf("tool call params = %v, want %v", msgs[2]["params"], tt.want)
			}
		})
	}
}

func TestDmExtractToolResult(t *testing.T) {
	tests := []struct {
		name    string
		stdout  string
		id      int
		want    any
		wantErr string // substring of the error
	}{
		{
			name:   "result found",
			stdout: `{"jsonrpc":"2.0","id":2,"result":{"content":[{"type":"text","text":"ok"}]}}` + "\n",
			id:     2,
			want:   map[string]any{"content": []any{map[string]any{"type": "text", "text": "ok"}}},
		},
		{
			name: "skips other messages, blank lines and garbage",
			stdout: strings.Join([]string{
				`{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-11-25"}}`,
				``,
				`not json`,
				`{"jsonrpc":"2.0","method":"notifications/message","params":{}}`,
				`  {"jsonrpc":"2.0","id":2,"result":{"n":1}}  `,
			}, "\n"),
			id:   2,
			want: map[string]any{"n": float64(1)},
		},
		{
			name:    "mcp error",
			stdout:  `{"jsonrpc":"2.0","id":2,"error":{"code":-32602,"message":"unknown tool"}}`,
			id:      2,
			wantErr: "mcp error -32602: unknown tool",
		},
		{
			name:    "error for another id is ignored",
			stdout:  `{"jsonrpc":"2.0","id":1,"error":{"code":-1,"message":"boom"}}`,
			id:      2,
			wantErr: "tool call response not found",
		},
		{
			name:    "missing result",
			stdout:  `{"jsonrpc":"2.0","id":2}`,
			id:      2,
			wantErr: "missing result payload",
		},
		{name: "no response", stdout: "", id: 2, wantErr: "tool call response not found"},
		{name: "only invalid lines", stdout: "garbage\n{broken", id: 2, wantErr: "tool call response not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dmExtractToolResult([]byte(tt.stdout), tt.id)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("dmExtractToolResult() error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("dmExtractToolResult() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dmExtractToolResult() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

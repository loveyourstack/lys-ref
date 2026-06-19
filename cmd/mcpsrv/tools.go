package main

import (
	"context"
	"fmt"

	"github.com/loveyourstack/lys-ref/internal/enums/aggperiod"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampperf"
	"github.com/mark3labs/mcp-go/mcp"
)

// analyticsHandler holds the store used by all tool handlers.
type analyticsHandler struct {
	campPerfStore dmcampperf.Store
}

func (h *analyticsHandler) getDailyTrend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	days := req.GetInt("days", 7)

	rows, err := h.campPerfStore.SelectDailyTrend(ctx, days)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("query failed", err), nil
	}

	result, err := mcp.NewToolResultJSON(rows)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to encode result", err), nil
	}
	result.Content = append(result.Content,
		mcp.NewTextContent(fmt.Sprintf("Daily trend for the past %d days (%d data points)", days, len(rows))),
	)
	return result, nil
}

func (h *analyticsHandler) getPerformanceSummary(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	period := req.GetString("period", "week")

	summary, err := h.campPerfStore.SelectPerfSummary(ctx, aggperiod.FromString(period))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("query failed", err), nil
	}

	result, err := mcp.NewToolResultJSON(summary)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to encode result", err), nil
	}
	result.Content = append(result.Content,
		mcp.NewTextContent(fmt.Sprintf("Performance summary for period=%s: %d active campaigns, profit=%.2f EUR, ROI=%.2f%%",
			period, summary.ActiveCampaigns, summary.ProfitEur, summary.ReturnOnInvestment*100)),
	)
	return result, nil
}

func (h *analyticsHandler) getTopCampaigns(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	period := req.GetString("period", "week")
	orderBy := req.GetString("order_by", "profit_eur")
	limit := req.GetInt("limit", 10)

	rows, err := h.campPerfStore.SelectTopCampaigns(ctx, aggperiod.FromString(period), orderBy, limit)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("query failed", err), nil
	}

	result, err := mcp.NewToolResultJSON(rows)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to encode result", err), nil
	}
	result.Content = append(result.Content,
		mcp.NewTextContent(fmt.Sprintf("Top %d campaigns for period=%s, ranked by %s", len(rows), period, orderBy)),
	)
	return result, nil
}

func (h *analyticsHandler) getVerticalPerformance(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	period := req.GetString("period", "week")
	orderBy := req.GetString("order_by", "profit_eur")

	rows, err := h.campPerfStore.SelectVerticalPerformance(ctx, aggperiod.FromString(period), orderBy)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("query failed", err), nil
	}

	result, err := mcp.NewToolResultJSON(rows)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to encode result", err), nil
	}
	result.Content = append(result.Content,
		mcp.NewTextContent(fmt.Sprintf("Vertical performance for period=%s, ordered by %s (%d verticals)", period, orderBy, len(rows))),
	)
	return result, nil
}

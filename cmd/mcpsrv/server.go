package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampperf"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func runServer(db *pgxpool.Pool) error {
	h := &analyticsHandler{
		campPerfStore: dmcampperf.Store{Db: db},
	}

	s := server.NewMCPServer(
		"lys-ref-analytics",
		"1.0.0",
		server.WithInstructions(
			"Analytics tools for lys-ref digital marketing campaigns. "+
				"Use these tools to answer questions about campaign and vertical performance. "+
				"Periods: day, week (default), month, year.",
		),
	)

	/*
		Note on Tools vs Resources:
		* Tools: have side effects, but also on-demand computation, e.g. dynamic analytics queries
		* Resources: read-only, static objects, e.g. pre-computed or named reports
	*/

	s.AddTool(
		mcp.NewTool("get_daily_trend",
			mcp.WithDescription(
				"Returns per-day aggregated spend, revenue, profit, clicks, and conversions for the past N days. "+
					"Use this for questions like 'show me revenue by day for the last 7 days'."),
			mcp.WithInteger("days",
				mcp.Description("Number of days to look back (1-31). Defaults to 7."),
				mcp.DefaultNumber(7),
				mcp.Min(1),
				mcp.Max(31),
			),
		),
		h.getDailyTrend,
	)

	s.AddTool(
		mcp.NewTool("get_performance_summary",
			mcp.WithDescription(
				"Returns a single-row summary of total clicks, conversions, spend, revenue, profit, "+
					"and ROI across all campaigns for a given period. "+
					"Use this for questions like 'give me a summary of this week's performance'."),
			mcp.WithString("period",
				mcp.Description("Time period: day, week, month, or year. Defaults to week."),
				mcp.Enum("day", "week", "month", "year"),
				mcp.DefaultString("week"),
			),
		),
		h.getPerformanceSummary,
	)

	s.AddTool(
		mcp.NewTool("get_top_campaigns",
			mcp.WithDescription(
				"Returns the top N campaigns ranked by a chosen metric for a given period. "+
					"Use this to answer questions like 'which campaigns had the highest ROI this month?'"),
			mcp.WithString("period",
				mcp.Description("Time period: day, week, month, or year. Defaults to week."),
				mcp.Enum("day", "week", "month", "year"),
				mcp.DefaultString("week"),
			),
			mcp.WithString("order_by",
				mcp.Description("Metric to rank campaigns by: clicks, conversions, impressions, profit_eur (default), return_on_investment, revenue_eur, spend_eur"),
				mcp.Enum("clicks", "conversions", "impressions", "profit_eur", "return_on_investment", "revenue_eur", "spend_eur"),
				mcp.DefaultString("profit_eur"),
			),
			mcp.WithInteger("limit",
				mcp.Description("Number of campaigns to return (1-100). Defaults to 10."),
				mcp.DefaultNumber(10),
				mcp.Min(1),
				mcp.Max(100),
			),
		),
		h.getTopCampaigns,
	)

	s.AddTool(
		mcp.NewTool("get_vertical_performance",
			mcp.WithDescription(
				"Returns aggregated campaign performance grouped by marketing vertical for a given period. "+
					"Use this to answer questions like 'what are the best performing verticals this week?'"),
			mcp.WithString("period",
				mcp.Description("Time period to analyse: day, week, month, or year. Defaults to week."),
				mcp.Enum("day", "week", "month", "year"),
				mcp.DefaultString("week"),
			),
			mcp.WithString("order_by",
				mcp.Description("Metric to rank verticals by: clicks, conversions, impressions, profit_eur (default), return_on_investment, revenue_eur, spend_eur"),
				mcp.Enum("clicks", "conversions", "impressions", "profit_eur", "return_on_investment", "revenue_eur", "spend_eur"),
				mcp.DefaultString("profit_eur"),
			),
		),
		h.getVerticalPerformance,
	)

	return server.ServeStdio(s)
}

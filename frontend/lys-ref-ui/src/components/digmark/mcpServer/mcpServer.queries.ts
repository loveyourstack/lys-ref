import { type McpQueryDef, type PerfSummary, type RevenueByDay, type TopCampaign, type VerticalPerf } from '@/types/digmark'

const formatter = new Intl.NumberFormat()
const formatterDec0 = new Intl.NumberFormat(undefined , { maximumFractionDigits: 0, minimumFractionDigits: 0})
const formatterPctDec0 = new Intl.NumberFormat(undefined , { style: 'percent', maximumFractionDigits: 0, minimumFractionDigits: 0})

export const queries: McpQueryDef<any>[] = [
  {
    id: 1,
    naturalLanguage: 'Show me revenue by day for the last 7 days',
    mcpTool: 'get_daily_trend',
    params: { days: 7 },
    columns: [
      { key: 'day', label: 'Day' },
      { key: 'revenue_eur', label: 'Revenue', format: (val) => `${formatterDec0.format(val)} €`, align: 'end' },
    ],
    normalize: raw => Array.isArray(raw) ? raw as RevenueByDay[] : [],
  },
  {
    id: 2,
    naturalLanguage: "Give me a summary of this week's performance",
    mcpTool: 'get_performance_summary',
    params: { period: 'week' },
    columns: [
      { key: 'active_campaigns', label: '# Active campaigns', format: (val) => formatter.format(val) },
      { key: 'impressions', label: 'Impressions', format: (val) => formatter.format(val) },
      { key: 'clicks', label: 'Clicks', format: (val) => formatter.format(val) },
      { key: 'conversions', label: 'Conversions', format: (val) => formatter.format(val) },
      { key: 'revenue_eur', label: 'Revenue', format: (val) => `${formatterDec0.format(val)} €` },
      { key: 'spend_eur', label: 'Spend', format: (val) => `${formatterDec0.format(val)} €` },
      { key: 'profit_eur', label: 'Profit', format: (val) => `${formatterDec0.format(val)} €` },
      { key: 'return_on_investment', label: 'ROI', format: (val) => formatterPctDec0.format(val) },
    ],
    normalize: raw => typeof raw === 'object' && raw !== null ? [raw as PerfSummary] : [],
  },
  {
    id: 3,
    naturalLanguage: 'Which campaigns had the highest ROI in the last month?',
    mcpTool: 'get_top_campaigns',
    params: { period: 'month', orderBy: 'ROI', limit: 10 },
    columns: [
      { key: 'campaign', label: 'Campaign' },
      { key: 'manager', label: 'Manager' },
      { key: 'profit_eur', label: 'Profit', format: (val) => `${formatterDec0.format(val)} €`, align: 'end' },
      { key: 'return_on_investment', label: 'ROI', format: (val) => formatterPctDec0.format(val), align: 'end' },
    ],
    normalize: raw => Array.isArray(raw) ? raw as TopCampaign[] : [],
  },
  {
    id: 4,
    naturalLanguage: 'What are the best performing verticals today?',
    mcpTool: 'get_vertical_performance',
    params: { period: 'day', orderBy: 'profit' },
    columns: [
      { key: 'vertical', label: 'Vertical' },
      { key: 'profit_eur', label: 'Profit', format: (val) => `${formatterDec0.format(val)} €`, align: 'end' },
      { key: 'return_on_investment', label: 'ROI', format: (val) => formatterPctDec0.format(val), align: 'end' },
    ],
    normalize: raw => Array.isArray(raw) ? raw as VerticalPerf[] : [],
  },
]
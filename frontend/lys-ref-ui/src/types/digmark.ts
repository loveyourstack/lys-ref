
export interface BudgetByManager {
  manager: string
  total_budget: number
}
export interface BudgetByVertical {
  vertical: string
  total_budget: number
}

// ------------------------------------------------------------------------------------------------------------------------------------------

export interface CampaignInput {
  country_fk: number | undefined
  daily_budget_eur: number | undefined
  is_active: boolean
  manager: string | undefined
  name: string | undefined
  vertical_fk: number | undefined
}
export interface Campaign extends CampaignInput {
  id: number
  country: string
  country_iso2: string
  created_at: Date
  performance_range: string
  updated_at: Date
  vertical: string
}
export function NewCampaign(): Campaign {
  return  {
    country_fk: undefined,
    daily_budget_eur: undefined,
    is_active: false,
    manager: undefined,
    name: undefined,
    vertical_fk: undefined,

    id: 0,
    country: '',
    country_iso2: '',
    created_at: new Date(),
    performance_range: '',
    updated_at: new Date(),
    vertical: '',
  }
}
export function GetCampaignInputFromItem(item: Campaign): CampaignInput {
  return  {
    country_fk: item.country_fk,
    daily_budget_eur: item.daily_budget_eur,
    is_active: item.is_active,
    manager: item.manager,
    name: item.name,
    vertical_fk: item.vertical_fk,
  }
}

// ------------------------------------------------------------------------------------------------------------------------------------------

export interface CampaignOptimizer {
  id: number
  campaign_fk: number
  clicks: number
  conversions: number
  country: string
  country_fk: number
  country_iso2: string
  created_at: Date
  daily_budget_eur: number
  editing_daily_budget: boolean
  end_day: Date
  impressions: number
  is_active: boolean
  name: string
  patch_daily_budget_icon: string
  patch_is_active_icon: string
  profit_eur: number
  return_on_investment: number
  revenue_eur: number
  spend_eur: number
  start_day: Date
  trend: number
  vertical: string
  vertical_fk: number
  volatility: number
}

export interface CampaignOptimizerAggregates {
  clicks: number
  conversions: number
  daily_budget_eur: number
  impressions: number
  is_active: number
  profit_eur: number
  return_on_investment: number
  revenue_eur: number
  spend_eur: number
}

// ------------------------------------------------------------------------------------------------------------------------------------------

export interface CampaignPerformance {
  id: number
  campaign: string
  campaign_fk: number
  clicks: number
  conversions: number
  country: string
  country_iso2: string
  created_at: Date
  day_cet: Date
  impressions: number
  profit_eur: number
  return_on_investment: number
  revenue_eur: number
  spend_eur: number
  updated_at: Date
  vertical: string
}

export interface CampaignPerfLatestSummary {
  day: Date
  total_revenue: number
  total_spend: number
}

// ------------------------------------------------------------------------------------------------------------------------------------------

export interface GeneratedCampaignInput {
  body: string
  call_to_action: string
  headline: string
  image_path: string
  model: string
  product: string
}
export interface GeneratedCampaign extends GeneratedCampaignInput {
  id: number
  created_at: Date
}

// ------------------------------------------------------------------------------------------------------------------------------------------

export interface LauncherInput {
  daily_budget_eur: number | undefined
  manager: string | undefined
  name: string | undefined
}
export interface Launcher {
  id: number
  country: string
  country_fk: number
  country_iso2: string
  created_at: Date
  created_at_day: Date
  max_steps: number
  message: string
  partner: string
  status: string
  step: number
  updated_at: Date
  vertical: string
  vertical_fk: number
}

// ---------------------------------------------------------

export interface LauncherInputFb extends LauncherInput {
  fan_page: string | undefined
}
export interface LauncherFb extends LauncherInputFb, Launcher {
  fb_account_id: string
  fb_campaign_id: string
  fb_creative_id: string
}
export function NewLauncherFb(): LauncherFb {
  return {
    daily_budget_eur: undefined,
    manager: undefined,
    name: undefined,

    fan_page: undefined,

    id: 0,
    country: '',
    country_fk: 0,
    country_iso2: '',
    created_at: new Date(),
    created_at_day: new Date(),
    max_steps: 0,
    message: '',
    partner: '',
    status: '',
    step: 0,
    updated_at: new Date(),
    vertical: '',
    vertical_fk: 0,

    fb_account_id: '',
    fb_campaign_id: '',
    fb_creative_id: '',
  }
}
export function GetLauncherInputFbFromItem(item: LauncherFb): LauncherInputFb {
  return  {
    daily_budget_eur: item.daily_budget_eur,
    manager: item.manager,
    name: item.name,

    fan_page: item.fan_page,
  }
}
export interface LauncherFbImport {
  name: string
  manager: string
  fan_page: string
  daily_budget_eur: number
}
export const launcherFbImportColumns = [
  'name',
  'manager',
  'fan_page',
  'daily_budget_eur',
] as const satisfies readonly (keyof LauncherFbImport)[]

// ---------------------------------------------------------

export interface LauncherInputGAds extends LauncherInput {
}
export interface LauncherGAds extends LauncherInputGAds, Launcher {
  gads_account_id: number
  gads_ad_id: number
  gads_ad_group_id: number
  gads_campaign_id: number
}
export function NewLauncherGAds(): LauncherGAds {
  return {
    daily_budget_eur: undefined,
    manager: undefined,
    name: undefined,

    id: 0,
    country: '',
    country_fk: 0,
    country_iso2: '',
    created_at: new Date(),
    created_at_day: new Date(),
    max_steps: 0,
    message: '',
    partner: '',
    status: '',
    step: 0,
    updated_at: new Date(),
    vertical: '',
    vertical_fk: 0,

    gads_account_id: 0,
    gads_ad_id: 0,
    gads_ad_group_id: 0,
    gads_campaign_id: 0,
  }
}
export function GetLauncherInputGAdsFromItem(item: LauncherGAds): LauncherInputGAds {
  return  {
    daily_budget_eur: item.daily_budget_eur,
    manager: item.manager,
    name: item.name,
  }
}
export interface LauncherGAdsImport {
  name: string
  manager: string
  daily_budget_eur: number
}
export const launcherGAdsImportColumns = [
  'name',
  'manager',
  'daily_budget_eur',
] as const satisfies readonly (keyof LauncherGAdsImport)[]

// ------------------------------------------------------------------------------------------------------------------------------------------

export type McpColumnDef = {
  key: string
  label: string
  format?: (value: any) => string
  align?: 'start' | 'end'
}

export type McpQueryDef<T> = {
  id: number
  naturalLanguage: string
  mcpTool: string
  params: Record<string, unknown>
  columns: McpColumnDef[]
  normalize: (raw: unknown) => T[]
}

export interface PerfSummary {
  active_campaigns: number
  clicks: number
  conversions: number
  impressions: number
  profit_eur: number
  return_on_investment: number
  revenue_eur: number
  spend_eur: number
}
export interface RevenueByDay {
  day: string
  revenue_eur: number
}
export interface TopCampaign {
  campaign: string
  manager: string
  profit_eur: number
  return_on_investment: number
}
export interface VerticalPerf {
  profit_eur: number
  return_on_investment: number
  vertical: string
}

// ------------------------------------------------------------------------------------------------------------------------------------------

export interface VerticalInput {
  name: string | undefined
}
export interface Vertical extends VerticalInput {
  id: number
  campaign_count: number
  created_at: Date
  updated_at: Date
}
export function NewVertical(): Vertical {
  return  {
    name: undefined,

    id: 0,
    campaign_count: 0,
    created_at: new Date(),
    updated_at: new Date(),
  }
}
export function GetVerticalInputFromItem(item: Vertical): VerticalInput {
  return  {
    name: item.name,
  }
}

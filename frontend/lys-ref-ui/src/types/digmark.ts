
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

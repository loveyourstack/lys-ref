import type { ApiCallBase } from "@/types/app"

export interface ApiCall extends ApiCallBase {
}

export interface Currency {
  id: number
  code: string
  created_at: Date
  is_active: boolean // metadata
  metadata_id: number // metadata
  name: string
  symbol: string // metadata
  updated_at: Date
}

export interface ExchangeRate {
  id: number
  created_at: Date
  day: Date
  frequency: string
  from_currency_fk: number
  from_currency: string
  rate: number
  to_currency_fk: number
  to_currency: string
  updated_at: Date
}

export interface XrPerfNormalized {
  id: number
  created_at: Date
  day: Date
  from_currency_fk: number
  from_currency_code: string
  normalized_perf: number
  period: string
  rate: number
  to_currency_fk: number
  to_currency_code: string
}
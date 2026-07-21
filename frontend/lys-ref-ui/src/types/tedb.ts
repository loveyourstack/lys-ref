import type { ApiCallBase } from "@/types/app"

export interface ApiCall extends ApiCallBase {
}

export interface VatRateSummary {
  categories: string
  comment: string
  country: string
  rate: number
  situation_on: Date
  type: string
}
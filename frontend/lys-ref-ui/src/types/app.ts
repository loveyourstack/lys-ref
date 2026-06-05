
export interface ApiCallBase {
  id: number
  attempt: number
  created_at: Date
  created_at_date: Date
  duration_ms: number
  endpoint: string
  method: string
  page: number
  result: string
  status_code: number
}

export interface ApiError {
  method: string
  url: string
  errMsg: string
}

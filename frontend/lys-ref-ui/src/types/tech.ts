
export interface LoginAttempt {
  created_at: Date
  ip : string
  is_blocked: boolean
  num_attempts: number
}

export interface ServerRequest {
  id: number
  created_at: Date
  created_at_date: Date
  duration_ms: number
  endpoint: string
  ip: string
  method: string
  status_code: number
  user_name: string
}

export interface Session {
  allow_multiple_sessions: boolean
  created_at: Date
  expires_at: Date
  force_password_change: boolean
  geo_ip_country_iso_code: string
  geo_ip_location: string
  ip : string
  last_access_at: Date
  roles: string[]
  user_agent: string
  user_id: number
  user_name: string
}
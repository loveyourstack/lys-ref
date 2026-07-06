import type { SelectionItem } from 'lys-vue'

export interface LoginResponse {
  default_locale: string
  force_password_change: boolean
  has_aws_sg_rules: boolean
  geo_ip_country_iso_code: string
  geo_ip_location: string
  ip: string
  roles: Role[]
  token: string
  user_id: number
  user_name: string
}

export interface Notification {
  created_at: string
  id: number
  is_read: boolean
  message: string
  not_type: string
  updated_at: string
  user_fk: number
}

export enum Role {
  Standard = 'Standard',
  Viewer = 'Viewer',
  Tech = 'Tech',
}
export const WriterRoles = [Role.Standard, Role.Tech]

export interface StoreData {
  core_mandatory_enums: string[]
  core_optional_enums: string[]
  core_periods: string[]

  digmark_launcher_stati: string[]
  digmark_managers: string[]
  digmark_verticals: SelectionItem[]

  ecb_active_currencies_ex_eur: SelectionItem[]

  geo_countries: SelectionItem[]
  geo_oceans: SelectionItem[]

  process_flows: SelectionItem[]

  pub_authors: SelectionItem[]

  supp_companies: SelectionItem[]
  supp_product_categories: SelectionItem[]
}
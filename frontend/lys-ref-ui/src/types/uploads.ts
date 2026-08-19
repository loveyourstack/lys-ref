
export interface Constraint {
  mime_type: string
  path: string
  title: string
  value: number
}

export interface UploadResult {
  mime_type: string
  original_name: string
  size_bytes: number
  stored_name: string
}

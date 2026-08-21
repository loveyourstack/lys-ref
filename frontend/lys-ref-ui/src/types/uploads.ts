
export interface Constraint {
  mime_type: string
  path: string
  title: string
  value: number
}

export type SelectionChange = CustomEvent<{
  x: number
  y: number
  width: number
  height: number
}>

export interface UploadResult {
  mime_type: string
  original_name: string
  size_bytes: number
  stored_name: string
}


export interface Employee {
  children: Employee[]
  department: string
  full_name: string
  email: boolean
  honorific: string
  job_title: string
  id: number
  profile_pic: string
  reports_to: number
  reports_to_full_name: string
  reports_to_job_title: string
}

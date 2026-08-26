
export interface Employee {
  children: Employee[]
  date_of_birth: Date
  department: string
  department_fk: number
  full_name: string
  email: string
  honorific: string
  job_title: string
  join_date: Date
  id: number
  profile_pic: string
  reports_to: number
  reports_to_full_name: string
  reports_to_job_title: string
  sex: string
}

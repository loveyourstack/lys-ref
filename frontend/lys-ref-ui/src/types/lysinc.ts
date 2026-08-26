
// v-calendar's event interface
export interface CalendarEvent {
  name: string
  category: string
  start: Date
  end: Date
  color: string
  timed: boolean
  message?: string // added by us
}

// v-calendar's change event interface
export interface CalendarTimestamp {
  date: string
  year: number
  month: number
  day: number
}
export interface CalendarChangeEvent {
  start: CalendarTimestamp
  end: CalendarTimestamp
}

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

export interface Event {
  department: string
  employee_id: number
  event_date: Date
  event_type: string
  full_name: string
  job_title: string
  message: string
  sex: string
  years: number
}
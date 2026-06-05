
export interface EmployeeInput {
  company_fk: number | undefined
  email: string | undefined
  given_name: string | undefined
  family_name: string | undefined
}
export interface Employee extends EmployeeInput {
  id: number
  company: string
  created_at: Date
  name: string
  updated_at: Date
}
export function NewEmployee(): Employee {
  return  {
    company_fk: undefined,
    email: undefined,
    given_name: undefined,
    family_name: undefined,

    id: 0,
    company: '',
    created_at: new Date(),
    name: '',
    updated_at: new Date(),
  }
}
export function GetEmployeeInputFromItem(item: Employee): EmployeeInput {
  return  {
    company_fk: item.company_fk,
    email: item.email,
    given_name: item.given_name,
    family_name: item.family_name,
  }
}

// ------------------------------------------------------------------------------------------------------------------------------------------

export interface ProductInput {
  category_fk: number | undefined
  company_fk: number | undefined
  name: string | undefined
  units_on_order: number
}
export interface Product extends ProductInput {
  id: number
  category: string
  company: string
  created_at: Date
  created_by: string
  last_user_update_by: string
  updated_at: Date
}
export function NewProduct(companyId: number): Product {
  return  {
    category_fk: undefined,
    company_fk: companyId,
    name: undefined,
    units_on_order: 0,

    id: 0,
    category: '',
    company: '',
    created_at: new Date(),
    created_by: '',
    last_user_update_by: '',
    updated_at: new Date(),
  }
}
export function GetProductInputFromItem(item: Product): ProductInput {
  return  {
    category_fk: item.category_fk,
    company_fk: item.company_fk,
    name: item.name,
    units_on_order: item.units_on_order,
  }
}

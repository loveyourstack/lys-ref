
export interface ArrayTypeInput {
  c_bool: boolean[]
  c_date: string[]
  c_enum: string[]
  c_int: number[]
  c_numeric: number[]
  c_text: string[]
}
export interface ArrayType extends ArrayTypeInput {
  id: number
  created_at: Date
  updated_at: Date
}
export function NewArrayType(): ArrayType {
  return  {
    c_bool: [],
    c_date: [],
    c_enum: [],
    c_int: [],
    c_numeric: [],
    c_text: [],

    id: 0,
    created_at: new Date(),
    updated_at: new Date(),
  }
}
export function GetArrayTypeInputFromItem(item: ArrayType): ArrayTypeInput {
  return  {
    c_bool: item.c_bool,
    c_date: item.c_date,
    c_enum: item.c_enum,
    c_int: item.c_int,
    c_numeric: item.c_numeric,
    c_text: item.c_text,
  }
}

// ------------------------------------------------------------------------------------------------------------------------------------------

export interface DefaultValueInput {
  c_default_text: string | undefined
  c_suggested_text: string | undefined
}
export interface DefaultValue extends DefaultValueInput {
  id: number
  created_at: Date
  updated_at: Date
}
export function NewDefaultValue(): DefaultValue {
  return  {
    c_default_text: undefined,
    c_suggested_text: 'Suggested text', // apply suggestion here

    id: 0,
    created_at: new Date(),
    updated_at: new Date(),
  }
}
export function GetDefaultValueInputFromItem(item: DefaultValue): DefaultValueInput {
  return  {
    c_default_text: item.c_default_text,
    c_suggested_text: item.c_suggested_text,
  }
}
export interface DefaultValueImport {
  c_default_text: string
  c_suggested_text: string
}
export const defaultValueImportColumns = [
  // c_default_text omitted
  'c_suggested_text',
] as const satisfies readonly (keyof DefaultValueImport)[]

// ------------------------------------------------------------------------------------------------------------------------------------------

// mandatory values: initialize as undefined - form rules will force an entry
export interface MandatoryValueInput {
  c_bool: boolean
  c_date_cet: string | undefined // don't use Date: it sends full timestamp to API, but API expects YYYY-MM-DD
  c_enum: string | undefined
  c_int: number | undefined
  c_numeric: number | undefined
  c_table_fk: number | undefined
  c_text: string | undefined
  c_time: string | undefined
}
export interface MandatoryValue extends MandatoryValueInput {
  id: number
  c_table: string
  created_at: Date
  updated_at: Date
}
export function NewMandatoryValue(): MandatoryValue {
  return  {
    c_bool: false,
    c_date_cet: undefined,
    c_enum: undefined,
    c_int: undefined,
    c_numeric: undefined,
    c_table_fk: undefined,
    c_text: undefined,
    c_time: undefined,

    id: 0,
    c_table: '',
    created_at: new Date(),
    updated_at: new Date(),
  }
}
export function GetMandatoryValueInputFromItem(item: MandatoryValue): MandatoryValueInput {
  return  {
    c_bool: item.c_bool,
    c_date_cet: item.c_date_cet,
    c_enum: item.c_enum,
    c_int: item.c_int,
    c_numeric: item.c_numeric,
    c_table_fk: item.c_table_fk,
    c_text: item.c_text,
    c_time: item.c_time,
  }
}
// import: don't use undefined since there is no form to enforce rules
export interface MandatoryValueImport {
  c_bool: boolean
  c_date_cet: string
  c_enum: string
  c_int: number
  c_numeric: number
  c_table_name: string // foreign key: use joined table name not ID. It will be replaced by ID in backend (see coremandatoryvalue section in routes.go)
  c_text: string
  c_time: string
}
// define expected columns in order to show column list in import dialog, and to check expected number of columns on each line
export const mandatoryValueImportColumns = [
  'c_bool',
  'c_date_cet',
  'c_enum',
  'c_int',
  'c_numeric',
  'c_table_name',
  'c_text',
  'c_time',
] as const satisfies readonly (keyof MandatoryValueImport)[]

// ------------------------------------------------------------------------------------------------------------------------------------------

// optional values: initialize with zero value except for date, so that date picker is initialized to today, not 0001-01-01
export interface OptionalValueInput {
  c_bool: boolean
  c_date_cet: string | undefined
  c_enum: string
  c_int: number
  c_numeric: number
  c_table_fk: number
  c_text: string
  c_time: string
}
export interface OptionalValue extends OptionalValueInput {
  id: number
  c_table: string
  created_at: Date
  updated_at: Date
}
export function NewOptionalValue(): OptionalValue {
  return  {
    c_bool: false,
    c_date_cet: undefined, // will become 0001-01-01
    c_enum: '',
    c_int: 0,
    c_numeric: 0.0,
    c_table_fk: -1, // table contains -1 (None)
    c_text: '',
    c_time: '00:00',

    id: 0,
    c_table: '',
    created_at: new Date(),
    updated_at: new Date(),
  }
}
export function GetOptionalValueInputFromItem(item: OptionalValue): OptionalValueInput {
  return  {
    c_bool: item.c_bool,
    c_date_cet: item.c_date_cet,
    c_enum: item.c_enum,
    c_int: item.c_int,
    c_numeric: item.c_numeric,
    c_table_fk: item.c_table_fk,
    c_text: item.c_text,
    c_time: item.c_time,
  }
}
export interface OptionalValueImport {
  c_bool: boolean
  c_date_cet: string
  c_enum: string
  c_int: number
  c_numeric: number
  c_table_name: string
  c_text: string
  c_time: string
}
export const optionalValueImportColumns = [
  'c_bool',
  'c_date_cet',
  'c_enum',
  'c_int',
  'c_numeric',
  'c_table_name',
  'c_text',
  'c_time',
] as const satisfies readonly (keyof OptionalValueImport)[]

// ------------------------------------------------------------------------------------------------------------------------------------------

export interface VariantTypeInput {
  c_constrained_text: string | undefined
  c_ip: string | undefined
  c_long_text: string | undefined
  c_money_amount: number | undefined
  c_percent: number | undefined
}
export interface VariantType extends VariantTypeInput {
  id: number
  c_long_text_short: string,
  created_at: Date
  updated_at: Date
}
export function NewVariantType(): VariantType {
  return  {
    c_constrained_text: undefined,
    c_ip: undefined,
    c_long_text: undefined,
    c_money_amount: undefined,
    c_percent: undefined,

    id: 0,
    c_long_text_short: '',
    created_at: new Date(),
    updated_at: new Date(),
  }
}
export function GetVariantTypeInputFromItem(item: VariantType): VariantTypeInput {
  return  {
    c_constrained_text: item.c_constrained_text,
    c_ip: item.c_ip,
    c_long_text: item.c_long_text,
    c_money_amount: item.c_money_amount,
    c_percent: item.c_percent,
  }
}
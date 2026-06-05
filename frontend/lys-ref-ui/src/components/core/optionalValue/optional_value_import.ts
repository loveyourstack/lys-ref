import { strToBool, strToDate, strToFloat, strToInt, strToTime } from 'lys-vue'
import { type Result } from 'lys-vue'
import { type OptionalValueImport, optionalValueImportColumns } from '@/types/core'
import { useCoreStore } from '@/stores/core'
import { useGeoStore } from '@/stores/geo'

const coreStore = useCoreStore()
const geoStore = useGeoStore()

export function getOptionalValueImportItems(entryA: string[], maxItems: number): Result<OptionalValueImport[]> {

  // de-duplicate incoming entries
  entryA = [...new Set(entryA)]

  // check for max # of items
  if (entryA.length > maxItems) {
    return { ok: false, error: 'Only ' + maxItems + ' items may be imported.' }
  }

  let itemA: OptionalValueImport[] = []
  const expectedNumCols: number = optionalValueImportColumns.length

  const countryNames = geoStore.countries.map(v => v.name)

  for (let i = 0; i < entryA.length; i++) {

    const lineStr = 'Line ' + (i+1) + ': '

    const colA = entryA[i]!.split('\t')
    if (colA.length != expectedNumCols) {
      return { ok: false, error: lineStr + 'expected ' + expectedNumCols + ' columns, got ' + colA.length }
    }

    // if values are missing, initialize optional columns to their zero value as shown in core.ts
    const c_bool_str = !colA[0] ? 'false' : colA[0]
    const c_date_cet_str = !colA[1] ? '0001-01-01' : colA[1]
    const c_enum_str = !colA[2] ? '' : colA[2]
    const c_int_str = !colA[3] ? '0' : colA[3]
    const c_numeric_str = !colA[4] ? '0' : colA[4]
    const c_table_name_str = !colA[5] ? 'None' : colA[5]
    const c_text_str = !colA[6] ? '' : colA[6]
    const c_time_str = !colA[7] ? '00:00' : colA[7]

    // validate boolean
    const bRes = strToBool(c_bool_str)
    if (!bRes.ok) { return { ok: false, error: lineStr + bRes.error + ' for \'c_bool\' column' } }

    // validate date
    const dRes = strToDate(c_date_cet_str)
    if (!dRes.ok) { return { ok: false, error: lineStr + dRes.error + ' for \'c_date_cet\' column' } }

    // validate enum
    if (!coreStore.optionalEnums.includes(c_enum_str)) {
      return { ok: false, error: lineStr + 'invalid value \'' + c_enum_str + '\' for \'c_enum\' column' }
    }

    // validate int
    const iRes = strToInt(c_int_str)
    if (!iRes.ok) {
      return { ok: false, error: lineStr + iRes.error + ' for \'c_int\' column' }
    }

    // validate numeric
    const nRes = strToFloat(c_numeric_str)
    if (!nRes.ok) {
      return { ok: false, error: lineStr + nRes.error + ' for \'c_numeric\' column' }
    }

    // validate table
    if (!countryNames.includes(c_table_name_str)) {
      return { ok: false, error: lineStr + 'invalid value \'' + c_table_name_str + '\' for \'c_table_name\' column' }
    }

    // text: empty text is allowed, no validation

    // validate time
    const timeRes = strToTime(c_time_str)
    if (!timeRes.ok) { 
      return { ok: false, error: lineStr + timeRes.error + ' for \'c_time\' column' }
    }

    const item: OptionalValueImport = {
      c_bool: bRes.value,
      c_date_cet: dRes.value.toISOString().split('T')[0]!, // format as 'YYYY-MM-DD'
      c_enum: c_enum_str,
      c_int: iRes.value,
      c_numeric: nRes.value,
      c_table_name: c_table_name_str,
      c_text: c_text_str,
      c_time: timeRes.value.hours.toString().padStart(2, '0') + ':' + timeRes.value.minutes.toString().padStart(2, '0'),
    }
    itemA.push(item)
  }

  return { ok: true, value: itemA }
}
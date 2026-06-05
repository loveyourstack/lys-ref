import { type Result } from 'lys-vue'
import { type DefaultValueImport, defaultValueImportColumns } from '@/types/core'

export function getDefaultValueImportItems(entryA: string[], maxItems: number): Result<DefaultValueImport[]> {

  // de-duplicate incoming entries
  entryA = [...new Set(entryA)]

  // check for max # of items
  if (entryA.length > maxItems) {
    return { ok: false, error: 'Only ' + maxItems + ' items may be imported.' }
  }

  let itemA: DefaultValueImport[] = []
  const expectedNumCols: number = defaultValueImportColumns.length

  for (let i = 0; i < entryA.length; i++) {

    const lineStr = 'Line ' + (i+1) + ': '

    const colA = entryA[i]!.split('\t')
    if (colA.length != expectedNumCols) {
      return { ok: false, error: lineStr + 'expected ' + expectedNumCols + ' columns, got ' + colA.length }
    }

    const c_suggested_text = colA[0]!

    if (c_suggested_text == '')  {
      return { ok: false, error: lineStr + 'empty string is not allowed for \'c_suggested_text\' column' }
    }

    const item: DefaultValueImport = {
      c_default_text: '', // default value will be set in backend
      c_suggested_text: c_suggested_text,
    }
    itemA.push(item)
  }

  return { ok: true, value: itemA }
}
import { strToFloat, type Result } from 'lys-vue'
import { type LauncherGAdsImport, launcherGAdsImportColumns } from '@/types/digmark'

export function getLauncherGAdsImportItems(entryA: string[], maxItems: number): Result<LauncherGAdsImport[]> {

  // de-duplicate incoming entries
  entryA = [...new Set(entryA)]

  // check for max # of items
  if (entryA.length > maxItems) {
    return { ok: false, error: 'Only ' + maxItems + ' items may be imported.' }
  }

  let itemA: LauncherGAdsImport[] = []
  const expectedNumCols: number = launcherGAdsImportColumns.length

  for (let i = 0; i < entryA.length; i++) {

    const lineStr = 'Line ' + (i+1) + ': '

    const colA = entryA[i]!.split('\t')
    if (colA.length != expectedNumCols) {
      return { ok: false, error: lineStr + 'expected ' + expectedNumCols + ' columns, got ' + colA.length }
    }

    // define string vars here to avoid use of array indices below
    const name_str = colA[0]!
    const manager_str = colA[1]!
    const daily_budget_eur_str = colA[2]!

    if (name_str == '')  {
      return { ok: false, error: lineStr + 'empty string is not allowed for \'name\' column' }
    }
    if (manager_str == '')  {
      return { ok: false, error: lineStr + 'empty string is not allowed for \'manager\' column' }
    }

    const budgetRes = strToFloat(daily_budget_eur_str)
    if (!budgetRes.ok) {
      return { ok: false, error: lineStr + budgetRes.error + ' for \'daily_budget_eur\' column' }
    }

    const item: LauncherGAdsImport = {
      name: name_str,
      manager: manager_str,
      daily_budget_eur: budgetRes.value,
    }
    itemA.push(item)
  }

  return { ok: true, value: itemA }
}
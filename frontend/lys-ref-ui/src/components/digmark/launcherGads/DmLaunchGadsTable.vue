<template>
  <l-dialog-card v-model="showEdit">
    <dm-launch-gads-form :id="editID"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @delete="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></dm-launch-gads-form>
  </l-dialog-card>

  <l-dialog-card v-model="showImport">
    <l-text-array-entry :title="$t('launchers.import_enter')" :saving="importing" :max-items="maxImportItems" :enterDisabled="!auth.isWriter()"
      :subtitle="`The expected columns are: ${launcherGAdsImportColumns.join(', ')}`"
      sampleSheetLink="https://docs.google.com/spreadsheets/d/10klDBUMBk5ByLsXmJ5T2jMqp1aeoHJOV-hT6_7-IFrM/edit?pli=1&gid=826607128#gid=826607128"
      :enterLabel="$t('actions.enter')"
      @cancel="showImport = false"
      @enter="async (valA: string[]) => { const success = await importItems(valA); if (success) { showImport = false } }"
    ></l-text-array-entry>
  </l-dialog-card>

  <v-data-table-server
    v-model="selected"
    v-model:items-per-page="itemsPerPage"
    v-model:page="page"
    v-model:sortBy="sortBy"
    :headers="selectedHeaders"
    hover
    :items-length="totalItems"
    :items="items"
    multi-sort
    :search="search"
    show-current-page
    item-value="id"
    show-select
    @update:options="loadItems"
  >
    <template #top>
      <l-dt-top :ax="ax" :title="props.title ?? $t('launchers.title') + ' - Google Ads'" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
        <v-btn color="secondary" @click="editID = 0; showEdit = true">{{ $t('actions.add') }}</v-btn>
        <v-btn color="secondary" :loading="importing" @click="showImport = true">{{ $t('actions.import') }}</v-btn>
        <v-btn icon flat v-tooltip="$t('actions.refresh')" @click="refreshItems()">
          <v-icon icon="mdi-refresh"></v-icon>
        </v-btn>
      </l-dt-top>

      <v-row density="compact">
        <v-col>
          <dm-launch-table-filters @update="refreshItems()" @updateDebounced="refreshItemsDebounced()"
            v-model:filterCountryFKs="filterCountryFKs"
            v-model:filterDailyBudget="filterDailyBudget"
            v-model:filterManagers="filterManagers"
            v-model:filterName="filterName"
            v-model:filterVerticalFks="filterVerticalFks"
          />
        </v-col>
      </v-row>
      
      <v-row v-if="selected.length > 0" density="compact" class="mt-0">
        <v-col>
          <dm-launch-table-bulk-edit :launch-ids="selected" :base-url="baseUrl" partner="Google Ads" @update="refreshItems()" />
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.created_at_day`]="{ item }">
      <span>{{ useDateFormat(item.created_at_day, 'DD MMM YYYY') }}</span>
    </template>

     <template v-slot:[`item.daily_budget_eur`]="{ item }">
      {{ formatterDec2.format(item.daily_budget_eur!) + ' €' }}
    </template>

    <template v-slot:[`item.status`]="{ item }">
      <v-chip :color="statusColor(item.status)">{{ item.status }}</v-chip>
    </template>

    <template v-slot:[`item.step`]="{ item }">
      <dm-launch-step-indicator :step="item.step" :max-steps="item.max_steps" />
    </template>

    <template v-slot:[`item.actions`]="{ item }">
      <v-btn icon flat size="small" v-tooltip="`${$t('actions.edit')}`" @click="editID = item.id; showEdit = true">
        <v-icon color="primary" icon="mdi-square-edit-outline"></v-icon>
      </v-btn>
    </template>

    <template #bottom>
      <l-dt-bottom :itemsPerPage="itemsPerPage" :page="page" :totalItemsIsEstimate="totalItemsIsEstimate" :totalItemsEstimated="totalItemsEstimated"></l-dt-bottom>
    </template>

  </v-data-table-server>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useDateFormat } from '@vueuse/core'
import { type SortItem } from 'vuetify/lib/components/VDataTable/composables/sort.mjs'
import { type NumericFilter, getNumericFilterUrlParams, getTextFilterUrlParam } from 'lys-vue'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { statusColor } from '@/components/digmark/launcher/launcher_funcs'
import { getLauncherGAdsImportItems } from '@/components/digmark/launcherGads/launcher_gads_import'
import { type LauncherGAds, launcherGAdsImportColumns } from '@/types/digmark'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Created', key: 'created_at_day' },
  { title: 'Name', key: 'name' },
  { title: 'Manager', key: 'manager' },
  { title: 'Daily budget', key: 'daily_budget_eur', align: 'end' },
  { title: 'Status', key: 'status' },
  { title: 'Step', key: 'step' },
  { title: 'Message', key: 'message' },
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/digmark/launchers-gads'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl, getFilterStr)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems, refreshItemsDebounced
} = useTableState<LauncherGAds>({ ax, baseUrl, getFilterStr, onFetchSuccess: () => { selected.value = [] } })

const filterCountryFKs = ref<number[]>()
const filterDailyBudget = ref<NumericFilter>()
const filterManagers = ref<string[]>()
const filterName = ref<string>()
const filterVerticalFks = ref<number[]>()

const editID = ref(0)
const showEdit = ref(false)

const showImport = ref(false)
const importing = ref(false)
const maxImportItems = 10

const formatterDec2 = new Intl.NumberFormat(undefined , { maximumFractionDigits: 2, minimumFractionDigits: 2})

const selected = ref<number[]>([])

const { resetTable } = useJsonLs({
  lsKey: 'launchers_gads_dt',
  refs: {
    excludedHeaders,
    filterCountryFKs,
    filterDailyBudget,
    filterManagers,
    filterName,
    filterVerticalFks,
    itemsPerPage,
    sortBy,
  },
})

function getFilterStr(): string {
  let ret = ''

  if (filterCountryFKs.value && filterCountryFKs.value.length > 0) { ret += '&country_fk=' + filterCountryFKs.value.join('|') }
  ret += getNumericFilterUrlParams('daily_budget_eur', filterDailyBudget.value)
  if (filterManagers.value && filterManagers.value.length > 0) { ret += '&manager=' + filterManagers.value.join('|') }
  ret += getTextFilterUrlParam('name', filterName.value)
  if (filterVerticalFks.value && filterVerticalFks.value.length > 0) { ret += '&vertical_fk=' + filterVerticalFks.value.join('|') }

  return ret
}

async function importItems(entryA: string[]): Promise<boolean> {

  // validate and convert raw string array to item array
  const itemA = getLauncherGAdsImportItems(entryA, maxImportItems)
  if (!itemA.ok) { alert(itemA.error); return false }

  importing.value = true
  try {
    await ax.post(baseUrl+'/import', JSON.stringify(itemA.value))
    // add slight delay to allow listener process to prepare the new items
    await new Promise(resolve => setTimeout(resolve, 1000))
    refreshItems()
    return true
  } catch (error) {
    // handled by interceptor
    return false
  } finally {
    importing.value = false
  }
}

</script>

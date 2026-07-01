<template>
  <l-dialog-card v-model="showEdit">
    <dm-launch-fb-form :id="editID"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @delete="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></dm-launch-fb-form>
  </l-dialog-card>

  <l-dialog-card v-model="showImport">
    <l-text-array-entry :title="$t('launchers.import_enter')" :saving="importing" :max-items="maxImportItems" :enterDisabled="!auth.isWriter()"
      :subtitle="`The expected columns are: ${launcherFbImportColumns.join(', ')}`"
      sampleSheetLink="https://docs.google.com/spreadsheets/d/10klDBUMBk5ByLsXmJ5T2jMqp1aeoHJOV-hT6_7-IFrM/edit?pli=1&gid=678826924#gid=678826924"
      :enterLabel="$t('actions.enter')"
      @cancel="showImport = false"
      @enter="(valA: string[]) => { const success = importItems(valA); if (success) { showImport = false } }"
    ></l-text-array-entry>
  </l-dialog-card>

  <v-data-table-server
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
    @update:options="loadItems"
  >
    <template #top>
      <l-dt-top :ax="ax" :title="props.title ?? $t('launchers.title') + ' - Facebook'" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
        <v-btn color="secondary" @click="editID = 0; showEdit = true">{{ $t('actions.add') }}</v-btn>
        <v-btn color="secondary" :loading="importing" @click="showImport = true">{{ $t('actions.import') }}</v-btn>
      </l-dt-top>

      <v-row density="compact">
        <v-col>
          <dm-launch-fb-table-filters @update="refreshItems()" @updateDebounced="refreshItemsDebounced()"
            v-model:filterCountryFKs="filterCountryFKs"
            v-model:filterDailyBudget="filterDailyBudget"
            v-model:filterManagers="filterManagers"
            v-model:filterName="filterName"
            v-model:filterVerticalFks="filterVerticalFks"
          />
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.created_at_day`]="{ item }">
      <span>{{ useDateFormat(item.created_at_day, 'DD MMM YYYY') }}</span>
    </template>

     <template v-slot:[`item.daily_budget_eur`]="{ item }">
      {{ formatterDec2.format(item.daily_budget_eur!) + ' €' }}
    </template>

    <template v-slot:[`item.step`]="{ item }">
      <v-rating v-model="item.step" :length="item.max_steps" empty-icon="mdi-circle-outline" full-icon="mdi-circle" 
        readonly color="primary" density="compact"
      ></v-rating>
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
import { getLauncherFbImportItems } from '@/components/digmark/launcherFb/launcher_fb_import'
import { type LauncherFb, launcherFbImportColumns } from '@/types/digmark'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Created', key: 'created_at_day' },
  { title: 'Name', key: 'name' },
  { title: 'Manager', key: 'manager' },
  { title: 'Country', key: 'country' },
  { title: 'Vertical', key: 'vertical' },
  { title: 'Fan page', key: 'fan_page' },
  { title: 'Daily budget', key: 'daily_budget_eur', align: 'end' },
  { title: 'Status', key: 'status' },
  { title: 'Step', key: 'step' },
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/digmark/launchers-fb'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl, getFilterStr)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems, refreshItemsDebounced
} = useTableState<LauncherFb>({ ax, baseUrl, getFilterStr })

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

const { resetTable } = useJsonLs({
  lsKey: 'launchers_fb_dt',
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

function importItems(entryA: string[]): boolean {

  // validate and convert raw string array to item array
  const itemA = getLauncherFbImportItems(entryA, maxImportItems)
  if (!itemA.ok) { alert(itemA.error); return false }

  importing.value = true
  ax.post(baseUrl+'/import', JSON.stringify(itemA.value))
    .then(response => {
      refreshItems()
    })
    .catch() // handled by interceptor
    .finally(() => { importing.value = false })
  
  // return true to close the import dialog before the request completes. Import button will show the loading state
  return true
}

</script>

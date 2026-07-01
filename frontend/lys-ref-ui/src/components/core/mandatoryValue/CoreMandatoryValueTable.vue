<template>
  <l-dialog-card v-model="showEdit">
    <core-mandatory-value-form :id="editID"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @delete="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></core-mandatory-value-form>
  </l-dialog-card>

  <l-dialog-card v-model="showImport">
    <l-text-array-entry :title="$t('mandatory_values.import_enter')" :saving="importing" :max-items="maxImportItems" :enterDisabled="!auth.isWriter()"
      :subtitle="`The expected columns are: ${mandatoryValueImportColumns.join(', ')}`"
      sampleSheetLink="https://docs.google.com/spreadsheets/d/10klDBUMBk5ByLsXmJ5T2jMqp1aeoHJOV-hT6_7-IFrM/edit?pli=1&gid=0#gid=0"
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
      <l-dt-top :ax="ax" :title="props.title ?? $t('mandatory_values.title')" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
        <v-btn color="secondary" @click="editID = 0; showEdit = true">{{ $t('actions.add') }}</v-btn>
        <v-btn color="secondary" :loading="importing" @click="showImport = true">{{ $t('actions.import') }}</v-btn>
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('mandatory_values.p1') }}</div>
          <div class="dt-subtitle">
            {{ $t('mandatory_values.p2') }}
            <ul class="mt-1 mb-1">
              <li>{{ $t('mandatory_values.p2_list.item_1') }}</li>
              <li>{{ $t('mandatory_values.p2_list.item_2') }}</li>
              <li>{{ $t('mandatory_values.p2_list.item_3') }}</li>
              <li>{{ $t('mandatory_values.p2_list.item_4') }}</li>
              <li>{{ $t('mandatory_values.p2_list.item_5') }}</li>
            </ul>
          </div>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col>
          <core-mandatory-value-table-filters @update="refreshItems()" @updateDebounced="refreshItemsDebounced()"
            v-model:filterBool="filterBool"
            v-model:filterDateCet="filterDateCet"
            v-model:filterEnum="filterEnum"
            v-model:filterEnums="filterEnums"
            v-model:filterInt="filterInt"
            v-model:filterNumeric="filterNumeric"
            v-model:filterTableFk="filterTableFk"
            v-model:filterTableFks="filterTableFks"
            v-model:filterText="filterText"
            v-model:filterTimestamp="filterTimestamp"
          />
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.c_bool`]="{ item }">
      <v-icon v-if="item.c_bool" size="small" icon="mdi-check"></v-icon>
    </template>

    <template v-slot:[`item.c_date_cet`]="{ item }">
      <span>{{ useDateFormat(item.c_date_cet, 'DD MMM YYYY') }}</span>
    </template>

     <template v-slot:[`item.c_int`]="{ item }">
      <!-- need NaN check since 0 is allowed and omitempty is used in backend, so value might be missing in JSON response -->
      <span v-if="!isNaN(item.c_int!)" :class="item.c_int! < 0 ? 'text-error' : ''">{{ formatter.format(item.c_int!) }}</span>
      <span v-else>0</span>
    </template>

     <template v-slot:[`item.c_numeric`]="{ item }">
      <!-- need NaN check since 0 is allowed and omitempty is used in backend, so value might be missing in JSON response -->
      <span v-if="!isNaN(item.c_numeric!)" :class="item.c_numeric! < 0 ? 'text-error' : ''">{{ formatterDec2.format(item.c_numeric!) }}</span>
      <span v-else>0</span>
    </template>

    <template v-slot:[`item.updated_at`]="{ item }">
      <!-- omit timezone (z) to save space if not needed -->
      <span>{{ useDateFormat(item.updated_at, 'DD MMM YYYY HH:mm:ss z') }}</span>
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
import { type DateFilter, type NumericFilter, getDateFilterUrlParams, getNumericFilterUrlParams, getTextFilterUrlParam } from 'lys-vue'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { getMandatoryValueImportItems } from '@/components/core/mandatoryValue/mandatory_value_import'
import { type MandatoryValue, mandatoryValueImportColumns } from '@/types/core'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Bool', key: 'c_bool' },
  { title: 'Date', key: 'c_date_cet' },
  { title: 'Enum', key: 'c_enum' },
  { title: 'Int', key: 'c_int', align: 'end' },
  { title: 'Numeric', key: 'c_numeric', align: 'end' },
  { title: 'Table (join)', key: 'c_table' },
  { title: 'Text', key: 'c_text' },
  { title: 'Time', key: 'c_time' },
  { title: 'Timestamp', key: 'updated_at' },
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/core/mandatory-values'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl, getFilterStr)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems, refreshItemsDebounced
} = useTableState<MandatoryValue>({ ax, baseUrl, getFilterStr })

const filterBool = ref<boolean>()
const filterDateCet = ref<DateFilter>()
const filterEnum = ref<string>()
const filterEnums = ref<string[]>()
const filterInt = ref<NumericFilter>()
const filterNumeric = ref<NumericFilter>()
const filterTableFk = ref<number>()
const filterTableFks = ref<number[]>()
const filterText = ref<string>()
const filterTimestamp = ref<DateFilter>()

const editID = ref(0)
const showEdit = ref(false)

const showImport = ref(false)
const importing = ref(false)
const maxImportItems = 10

// undefined NumberFormat locale: uses system default
const formatter = new Intl.NumberFormat()
const formatterDec2 = new Intl.NumberFormat(undefined , { maximumFractionDigits: 2, minimumFractionDigits: 2})

const { resetTable } = useJsonLs({
  lsKey: 'mandatory_values_dt',
  refs: {
    excludedHeaders,
    filterBool,
    filterDateCet,
    filterEnum,
    filterEnums,
    filterInt,
    filterNumeric,
    filterTableFk,
    filterTableFks,
    filterText,
    filterTimestamp,
    itemsPerPage,
    sortBy,
  },
})

function getFilterStr(): string {
  let ret = ''

  // bool
  if (filterBool.value != undefined) { ret += '&c_bool=' + filterBool.value }

  // date
  ret += getDateFilterUrlParams('c_date_cet', filterDateCet.value)

  // enum
  if (filterEnum.value) { ret += '&c_enum=' + filterEnum.value }

  // enum (multi)
  if (filterEnums.value && filterEnums.value.length > 0) { ret += '&c_enum=' + filterEnums.value.join('|') }

  // int
  ret += getNumericFilterUrlParams('c_int', filterInt.value)

  // numeric
  ret += getNumericFilterUrlParams('c_numeric', filterNumeric.value)

  // table
  if (filterTableFk.value) { ret += '&c_table_fk=' + filterTableFk.value }

  // table (multi)
  if (filterTableFks.value && filterTableFks.value.length > 0) { ret += '&c_table_fk=' + filterTableFks.value.join('|') }

  // text
  ret += getTextFilterUrlParam('c_text', filterText.value)

  // timestamp
  ret += getDateFilterUrlParams('updated_at', filterTimestamp.value)

  return ret
}

function importItems(entryA: string[]): boolean {

  // validate and convert raw string array to item array
  const itemA = getMandatoryValueImportItems(entryA, maxImportItems)
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

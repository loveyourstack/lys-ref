<template>
  <l-dialog-card v-model="showEdit">
    <core-optional-value-form :id="editID"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @delete="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></core-optional-value-form>
  </l-dialog-card>

  <l-dialog-card v-model="showImport">
    <l-text-array-entry title="Enter optional values" :saving="importing" :max-items="maxImportItems" :enterDisabled="!auth.isWriter()"
      :subtitle="`The expected columns are: ${optionalValueImportColumns.join(', ')}`"
      sampleSheetLink="https://docs.google.com/spreadsheets/d/10klDBUMBk5ByLsXmJ5T2jMqp1aeoHJOV-hT6_7-IFrM/edit?pli=1&gid=1760759881#gid=1760759881"
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
      <l-dt-top :ax="ax" :title="props.title ?? $t('type_handling.optional_values.title')" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
        <v-btn color="secondary" @click="editID = 0; showEdit = true">{{ $t('actions.add') }}</v-btn>
        <v-btn color="secondary" :loading="importing" @click="showImport = true">{{ $t('actions.import') }}</v-btn>
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('type_handling.optional_values.p1') }}</div>
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.c_bool`]="{ item }">
      <v-icon v-if="item.c_bool" size="small" icon="mdi-check"></v-icon>
    </template>

    <template v-slot:[`item.c_date_cet`]="{ item }">
      <!-- optional date: hide zero value -->
      <span v-if="new Date(item.c_date_cet!) >= new Date('1900-01-01')">{{ useDateFormat(item.c_date_cet, 'DD MMM YYYY') }}</span>
    </template>

     <template v-slot:[`item.c_int`]="{ item }">
      <!-- optional: need NaN check since omitempty is used in backend, so value will be missing in JSON response -->
      <!-- optional: if zero value is invalid, hide it too -->
      <span v-if="!isNaN(item.c_int) && item.c_int != 0" :class="item.c_int < 0 ? 'text-error' : ''">{{ formatter.format(item.c_int) }}</span>
    </template>

     <template v-slot:[`item.c_numeric`]="{ item }">
      <!-- optional: need NaN check since omitempty is used in backend, so value will be missing in JSON response -->
      <!-- optional: if zero value is invalid, hide it too -->
      <span v-if="!isNaN(item.c_numeric) && item.c_numeric != 0" :class="item.c_numeric < 0 ? 'text-error' : ''">{{ formatterDec2.format(item.c_numeric) }}</span>
    </template>

     <template v-slot:[`item.c_table`]="{ item }">
      <!-- optional join: hide None value -->
      <span v-if="item.c_table_fk !== -1">{{ item.c_table }}</span>
    </template>

     <template v-slot:[`item.c_time`]="{ item }">
      <!-- optional: if zero value is invalid, hide it -->
      <span v-if="item.c_time !== '00:00'">{{ item.c_time }}</span>
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
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { getOptionalValueImportItems } from '@/components/core/optionalValue/optional_value_import'
import { type OptionalValue, optionalValueImportColumns } from '@/types/core'

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
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/core/optional-values'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems
} = useTableState<OptionalValue>({ ax, baseUrl })

const editID = ref(0)
const showEdit = ref(false)

const showImport = ref(false)
const importing = ref(false)
const maxImportItems = 10

const formatter = new Intl.NumberFormat()
const formatterDec2 = new Intl.NumberFormat(undefined , { maximumFractionDigits: 2, minimumFractionDigits: 2})

const { resetTable } = useJsonLs({
  lsKey: 'optional_values_dt',
  refs: {
    excludedHeaders,
    itemsPerPage,
    sortBy,
  },
})

function importItems(entryA: string[]): boolean {

  const itemA = getOptionalValueImportItems(entryA, maxImportItems)
  if (!itemA.ok) { alert(itemA.error); return false }

  importing.value = true
  ax.post(baseUrl+'/import', JSON.stringify(itemA.value))
    .then(response => {
      refreshItems()
    })
    .catch() // handled by interceptor
    .finally(() => { importing.value = false })
  
  return true
}

</script>

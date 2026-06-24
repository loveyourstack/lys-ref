<template>
  <l-dialog-card v-model="showEdit">
    <ecb-currency-md-form :id="editID"
      @cancel="showEdit = false"
      @update="showEdit = false; refreshItems()"
    ></ecb-currency-md-form>
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
      <l-dt-top :ax="ax" :title="props.title ?? 'External data: currencies'" :headers="headers" :excelDlUrl="excelDlUrl" v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()">
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">Currency code and name are synced from the European Central Bank API.</div>
          <div class="dt-subtitle">Active and symbol are user-maintained metadata fields, stored independently of the synced data.</div>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col>
          <ecb-currency-table-filters @update="refreshItems()" @updateDebounced="refreshItemsDebounced()"
            v-model:filterCode="filterCode"
            v-model:filterIsActive="filterIsActive"
          />
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.is_active`]="{ item }">
      <v-icon v-if="item.is_active" size="small" icon="mdi-check"></v-icon>
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
import { type SortItem } from 'vuetify/lib/components/VDataTable/composables/sort.mjs'
import { getTextFilterUrlParam } from 'lys-vue'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import { type Currency } from '@/types/ecb'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Code', key: 'code' },
  { title: 'Name', key: 'name' },
  { title: 'Active', key: 'is_active' },
  { title: 'Symbol', key: 'symbol' },
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/ecb/currencies'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl, getFilterStr)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems, refreshItemsDebounced
} = useTableState<Currency>({ ax, baseUrl, getFilterStr, mapUrl })

const filterCode = ref<string>()
const filterIsActive = ref<boolean>()

const editID = ref(0)
const showEdit = ref(false)

const { resetTable } = useJsonLs({
  lsKey: 'currencies_dt',
  refs: {
    excludedHeaders,
    filterCode,
    filterIsActive,
    itemsPerPage,
    sortBy,
  },
})

function getFilterStr(): string {
  let ret = ''

  ret += getTextFilterUrlParam('code', filterCode.value)
  if (filterIsActive.value != undefined) { ret += '&is_active=' + filterIsActive.value }

  return ret
}

function mapUrl(url: string, options: { page: number, itemsPerPage: number, sortBy: SortItem[] }): string {

  // if no sort specified: default to code
  if (!options.sortBy || options.sortBy.length == 0) { url += '&xsort=code' }

  return url
}

</script>

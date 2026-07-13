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
      <l-dt-top :ax="ax" :title="props.title ?? $t('currencies.title')" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('currencies.p1') }}</div>
          <div class="dt-subtitle">{{ $t('currencies.p2') }}</div>
          <div class="dt-subtitle">{{ $t('currencies.p3') }}</div>
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
      <span class="ml-1 opacity-70">{{ $t('external_data.last_synced') }}: {{ lastSyncAtMsg }}</span>
    </template>

  </v-data-table-server>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue'
import { useDateFormat, useTimeAgo } from '@vueuse/core'
import { type AxiosResponse } from 'axios'
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

const lastSyncAt = ref<Date | null>(null)
const lastSyncFmt = useDateFormat(
  computed(() => lastSyncAt.value ?? 0), 'DD MMM YYYY HH:mm:ss'
)
const lastSyncAgo = useTimeAgo(
  computed(() => lastSyncAt.value ?? 0)
)
const lastSyncAtMsg = computed(() => {
  if (!lastSyncAt.value) { return 'Unknown' }
  const oneYrAgo = new Date(new Date().setFullYear(new Date().getFullYear() - 1))
  if (lastSyncAt.value < oneYrAgo) { return 'Unknown' }

  return `${lastSyncFmt.value} (${lastSyncAgo.value})`
})

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems, refreshItemsDebounced
} = useTableState<Currency>({ ax, baseUrl, getFilterStr, mapUrl, onFetchSuccess: (resp: AxiosResponse) => { 
  const lastSyncHdr = resp.headers['last-sync-at']
  const lastSyncVal = Array.isArray(lastSyncHdr) ? lastSyncHdr[0] : lastSyncHdr
  if (typeof lastSyncVal === 'string') {
    lastSyncAt.value = new Date(lastSyncVal)
  }
}})

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

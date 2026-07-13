<template>
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
      <l-dt-top :ax="ax" :title="props.title ?? $t('exchange_rates.title')" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('exchange_rates.p1') }}</div>
          <div class="dt-subtitle">{{ $t('exchange_rates.p2') }}</div>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col>
          <ecb-exchange-rate-table-filters @update="refreshItems()" @updateDebounced="refreshItemsDebounced()"
            v-model:filterDay="filterDay"
            v-model:filterToCurrFk="filterToCurrFk"
          />
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.day`]="{ item }">
      <span>{{ useDateFormat(item.day, 'DD MMM YYYY') }}</span>
    </template>

     <template v-slot:[`item.rate`]="{ item }">
      {{ formatterDec4.format(item.rate) }}
    </template>

    <template #bottom>
      <l-dt-bottom :itemsPerPage="itemsPerPage" :page="page" :totalItemsIsEstimate="totalItemsIsEstimate" :totalItemsEstimated="totalItemsEstimated"></l-dt-bottom>
      <span class="ml-1 opacity-70">{{ $t('external_data.last_synced') }}: {{ lastSyncAtMsg }}</span>
    </template>

  </v-data-table-server>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useDateFormat } from '@vueuse/core'
import { type SortItem } from 'vuetify/lib/components/VDataTable/composables/sort.mjs'
import { type DateFilter, getDateFilterUrlParams } from 'lys-vue'
import { useJsonLs, useLastSyncAt, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import { type ExchangeRate } from '@/types/ecb'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Day', key: 'day' },
  { title: 'From', key: 'from_currency' },
  { title: 'To', key: 'to_currency' },
  { title: 'Rate', key: 'rate', align: 'end' },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/ecb/exchange-rates'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl, getFilterStr)

const { lastSyncAtMsg, getLastSyncAt } = useLastSyncAt()

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems, refreshItemsDebounced
} = useTableState<ExchangeRate>({ ax, baseUrl, getFilterStr, mapUrl, onFetchSuccess: getLastSyncAt })

const filterDay = ref<DateFilter>()
const filterToCurrFk = ref<number>()

const formatterDec4 = new Intl.NumberFormat(undefined , { maximumFractionDigits: 4, minimumFractionDigits: 4 })

const { resetTable } = useJsonLs({
  lsKey: 'exchange_rates_dt',
  refs: {
    excludedHeaders,
    filterDay,
    filterToCurrFk,
    itemsPerPage,
    sortBy,
  },
})

function getFilterStr(): string {
  let ret = ''

  ret += getDateFilterUrlParams('day', filterDay.value)
  if (filterToCurrFk.value) { ret += '&to_currency_fk=' + filterToCurrFk.value }

  return ret
}

function mapUrl(url: string, options: { page: number, itemsPerPage: number, sortBy: SortItem[] }): string {

  // if no sort specified: default to day DESC, from, to
  if (!options.sortBy || options.sortBy.length == 0) { url += '&xsort=-day,from_currency,to_currency' }

  return url
}

</script>

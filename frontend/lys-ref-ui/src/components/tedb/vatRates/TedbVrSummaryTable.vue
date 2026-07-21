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
      <l-dt-top :ax="ax" :title="props.title ?? $t('vat_rates.title')" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('vat_rates.p1') }}</div>
          <div class="dt-subtitle">{{ $t('vat_rates.p2') }}</div>
          <div class="dt-subtitle">{{ $t('vat_rates.p3') }}</div>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col>
          <tedb-vr-summary-table-filters @update="refreshItems()"
            v-model:filterCountryFk="filterCountryFk"
          />
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.situation_on`]="{ item }">
      <span>{{ useDateFormat(item.situation_on, 'DD MMM YYYY') }}</span>
    </template>

     <template v-slot:[`item.rate`]="{ item }">
      {{ formatterPctDec1.format(item.rate / 100) }}
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
import { useJsonLs, useLastSyncAt, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import { type VatRateSummary } from '@/types/tedb'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Country', key: 'country' },
  { title: 'Type', key: 'type' },
  { title: 'Categories', key: 'categories' },
  { title: 'Day', key: 'situation_on' },
  { title: 'Comment', key: 'comment' },
  { title: 'Rate', key: 'rate', align: 'end' },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/tedb/vat-rate-summary'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl)

const { lastSyncAtMsg, getLastSyncAt } = useLastSyncAt()

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems
} = useTableState<VatRateSummary>({ ax, baseUrl, getFilterStr, onFetchSuccess: getLastSyncAt })

const filterCountryFk = ref<number>()

const formatterPctDec1 = new Intl.NumberFormat(undefined , { style: 'percent', maximumFractionDigits: 1, minimumFractionDigits: 1 })

const { resetTable } = useJsonLs({
  lsKey: 'exchange_rates_dt',
  refs: {
    excludedHeaders,
    filterCountryFk,
    itemsPerPage,
    sortBy,
  },
})

function getFilterStr(): string {
  let ret = ''

  if (filterCountryFk.value) { ret += '&country_fk=' + filterCountryFk.value }

  return ret
}

</script>

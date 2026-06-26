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
      <l-dt-top :ax="ax" :title="props.title ?? $t('campaign_perf.title')" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('campaign_perf.p1') }}</div>
          <div class="dt-subtitle">{{ $t('campaign_perf.p2') }}</div>
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.clicks`]="{ item }">
      {{ formatter.format(item.clicks!) }}
    </template>

    <template v-slot:[`item.conversions`]="{ item }">
      {{ formatter.format(item.conversions!) }}
    </template>

    <template v-slot:[`item.day_cet`]="{ item }">
      <span>{{ useDateFormat(item.day_cet, 'DD MMM YYYY') }}</span>
    </template>

     <template v-slot:[`item.impressions`]="{ item }">
      {{ formatter.format(item.impressions!) }}
    </template>

     <template v-slot:[`item.profit_eur`]="{ item }">
      <span :class="item.profit_eur! > 0 ? 'text-success' : 'text-error'">
        {{ formatterDec2.format(item.profit_eur!) + ' €' }}
      </span>
    </template>

    <template v-slot:[`item.return_on_investment`]="{ item }">
      <span :class="item.return_on_investment! > 0 ? 'text-success' : 'text-error'">
        {{ formatterPctDec2.format(item.return_on_investment!) }}
      </span>
    </template>

     <template v-slot:[`item.revenue_eur`]="{ item }">
      {{ formatterDec2.format(item.revenue_eur!) + ' €' }}
    </template>

     <template v-slot:[`item.spend_eur`]="{ item }">
      {{ formatterDec2.format(item.spend_eur!) + ' €' }}
    </template>

    <template #bottom>
      <l-dt-bottom :itemsPerPage="itemsPerPage" :page="page" :totalItemsIsEstimate="totalItemsIsEstimate" :totalItemsEstimated="totalItemsEstimated"></l-dt-bottom>
    </template>

  </v-data-table-server>
</template>

<script lang="ts" setup>
import { useDateFormat } from '@vueuse/core'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import { type CampaignPerformance } from '@/types/digmark'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Day', key: 'day_cet' },  
  { title: 'Campaign', key: 'campaign' },
  { title: 'Country', key: 'country' },
  { title: 'Vertical', key: 'vertical' },
  { title: 'Impressions', key: 'impressions', align: 'end' },
  { title: 'Clicks', key: 'clicks', align: 'end' },
  { title: 'Conversions', key: 'conversions', align: 'end' },
  { title: 'Revenue', key: 'revenue_eur', align: 'end' },
  { title: 'Spend', key: 'spend_eur', align: 'end' },
  { title: 'Profit', key: 'profit_eur', align: 'end' },
  { title: 'ROI', key: 'return_on_investment', align: 'end' },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/digmark/campaign-performance'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems
} = useTableState<CampaignPerformance>({ ax, baseUrl })

const formatter = new Intl.NumberFormat()
const formatterDec2 = new Intl.NumberFormat(undefined , { maximumFractionDigits: 2, minimumFractionDigits: 2})
const formatterPctDec2 = new Intl.NumberFormat(undefined , { style: 'percent', maximumFractionDigits: 2, minimumFractionDigits: 2})

const { resetTable } = useJsonLs({
  lsKey: 'campaign_perf_dt',
  refs: {
    excludedHeaders,
    itemsPerPage,
    sortBy,
  },
})

</script>

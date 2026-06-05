<template>
  <tr v-if="item" class="v-data-table__tr">
    <td class="v-data-table__td border-t-lg"><!-- selection column --></td>
    <td v-if="headers.includes('Name')" class="v-data-table__td border-t-lg">
      <div class="d-flex">
        Total <v-icon class="ml-1" size="small" color="primary" v-tooltip="'Unpaged total, not just of the rows shown.'" icon="mdi-information"></v-icon>
      </div>
    </td>
    <td v-if="headers.includes('Manager')" class="v-data-table__td border-t-lg"></td>
    <td v-if="headers.includes('Country')" class="v-data-table__td border-t-lg"></td>
    <td v-if="headers.includes('Vertical')" class="v-data-table__td border-t-lg"></td>
    <td v-if="headers.includes('Active')" class="v-data-table__td border-t-lg v-data-table-column--align-center">
      {{ formatter.format(item.is_active) }}
    </td>
    <td v-if="headers.includes('Daily budget')" class="v-data-table__td border-t-lg v-data-table-column--align-end">
      <span class="text-no-wrap">
        <span>{{ formatterDec2.format(item.daily_budget_eur) }} €</span>
      </span>
    </td>
    <td v-if="headers.includes('Impressions')" class="v-data-table__td border-t-lg v-data-table-column--align-end">
      {{ formatter.format(item.impressions) }}
    </td>
    <td v-if="headers.includes('Clicks')" class="v-data-table__td border-t-lg v-data-table-column--align-end">
      {{ formatter.format(item.clicks) }}
    </td>
    <td v-if="headers.includes('Conversions')" class="v-data-table__td border-t-lg v-data-table-column--align-end">
      {{ formatter.format(item.conversions) }}
    </td>
    <td v-if="headers.includes('Revenue')" class="v-data-table__td border-t-lg v-data-table-column--align-end">
      <span class="text-no-wrap">
        <span>{{ formatterDec2.format(item.revenue_eur) }} €</span>
      </span>
    </td>
    <td v-if="headers.includes('Spend')" class="v-data-table__td border-t-lg v-data-table-column--align-end">
      <span class="text-no-wrap">
        <span>{{ formatterDec2.format(item.spend_eur) }} €</span>
      </span>
    </td>
    <td v-if="headers.includes('Profit')" class="v-data-table__td border-t-lg v-data-table-column--align-end">
      <span class="text-no-wrap" :class="item.profit_eur > 0 ? 'text-success' : 'text-error'">
        <span>{{ formatterDec2.format(item.profit_eur) }} €</span>
      </span>
    </td>
    <td v-if="headers.includes('ROI')" class="v-data-table__td border-t-lg v-data-table-column--align-end">
      <span :class="item.return_on_investment > 0 ? 'text-success' : 'text-error'">
        {{ formatterPctDec2.format(item.return_on_investment) }}
      </span>
    </td>
    <td class="v-data-table__td border-t-lg"><!-- expanded column --></td>
  </tr>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, watch } from 'vue'
import { fetchOnce } from 'lys-vue'
import ax from '@/api'
import { type CampaignOptimizerAggregates } from '@/types/digmark'

const props = defineProps<{
  baseUrl: string
  filterStr: string
  refresh: string
  selectedHeaders: any
}>()

const headers = computed(() => {
  return props.selectedHeaders.map((v: any) => v.title)
})

const formatter = new Intl.NumberFormat()
const formatterDec2 = new Intl.NumberFormat(undefined , { maximumFractionDigits: 2, minimumFractionDigits: 2})
const formatterPctDec2 = new Intl.NumberFormat(undefined , { style: 'percent', maximumFractionDigits: 2, minimumFractionDigits: 2})

const item = ref<CampaignOptimizerAggregates>()

function load() {
  let myUrl = props.baseUrl + '/aggregates'
  const filterStr = props.filterStr.replace(props.filterStr.charAt(0), '?')
  myUrl += filterStr

  fetchOnce({ ax, myUrl, result: item })
}

watch(() => props.refresh, () => {
  load()
})

</script>

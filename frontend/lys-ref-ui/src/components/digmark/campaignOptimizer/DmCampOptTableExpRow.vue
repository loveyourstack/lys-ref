<template>
  <tr v-for="(item, idx) in items" class="v-data-table__tr">
    <td class="v-data-table__td" :class="idx == items!.length -1 ? 'border-b-lg' : ''"><!-- selection column --></td>
    <td v-if="headers.includes('Name')" class="v-data-table__td" :class="idx == items!.length -1 ? 'border-b-lg' : ''"></td>
    <td v-if="headers.includes('Manager')" class="v-data-table__td" :class="idx == items!.length -1 ? 'border-b-lg' : ''"></td>
    <td v-if="headers.includes('Country')" class="v-data-table__td" :class="idx == items!.length -1 ? 'border-b-lg' : ''"></td>
    <td v-if="headers.includes('Vertical')" class="v-data-table__td" :class="idx == items!.length -1 ? 'border-b-lg' : ''"></td>
    <td v-if="headers.includes('Active')" class="v-data-table__td" :class="idx == items!.length -1 ? 'border-b-lg' : ''"></td>
    <td v-if="headers.includes('Daily budget')" class="v-data-table__td v-data-table-column--align-end" :class="idx == items!.length -1 ? 'border-b-lg' : ''">
      {{ useDateFormat(item.day_cet, 'DD MMM YYYY') }}
    </td>
    <td v-if="headers.includes('Impressions')" class="v-data-table__td v-data-table-column--align-end" :class="idx == items!.length -1 ? 'border-b-lg' : ''">
      {{ formatter.format(item.impressions) }}
    </td>
    <td v-if="headers.includes('Clicks')" class="v-data-table__td v-data-table-column--align-end" :class="idx == items!.length -1 ? 'border-b-lg' : ''">
      {{ formatter.format(item.clicks) }}
    </td>
    <td v-if="headers.includes('Conversions')" class="v-data-table__td v-data-table-column--align-end" :class="idx == items!.length -1 ? 'border-b-lg' : ''">
      {{ formatter.format(item.conversions) }}
    </td>
    <td v-if="headers.includes('Revenue')" class="v-data-table__td v-data-table-column--align-end" :class="idx == items!.length -1 ? 'border-b-lg' : ''">
      <span class="text-no-wrap">
        <span>{{ formatterDec2.format(item.revenue_eur) }} €</span>
      </span>
    </td>
    <td v-if="headers.includes('Spend')" class="v-data-table__td v-data-table-column--align-end" :class="idx == items!.length -1 ? 'border-b-lg' : ''">
      <span class="text-no-wrap">
        <span>{{ formatterDec2.format(item.spend_eur) }} €</span>
      </span>
    </td>
    <td v-if="headers.includes('Profit')" class="v-data-table__td v-data-table-column--align-end" :class="idx == items!.length -1 ? 'border-b-lg' : ''">
      <span class="text-no-wrap" :class="item.profit_eur > 0 ? 'text-success' : 'text-error'">
        <span>{{ formatterDec2.format(item.profit_eur) }} €</span>
      </span>
    </td>
    <td v-if="headers.includes('ROI')" class="v-data-table__td v-data-table-column--align-end" :class="idx == items!.length -1 ? 'border-b-lg' : ''">
      <span :class="item.return_on_investment > 0 ? 'text-success' : 'text-error'">
        {{ formatterPctDec2.format(item.return_on_investment) }}
      </span>
    </td>
    <td class="v-data-table__td" :class="idx == items!.length -1 ? 'border-b-lg' : ''"><!-- expanded column --></td>
  </tr>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue'
import { useDateFormat, useNow } from '@vueuse/core'
import { fetchOnce } from 'lys-vue'
import ax from '@/api'
import { type CampaignPerformance } from '@/types/digmark'

const props = defineProps<{
  campId: number
  selectedHeaders: any
}>()

const baseUrl = '/a/digmark/campaign-performance'

const headers = computed(() => {
  return props.selectedHeaders.map((v: any) => v.title)
})

const formatter = new Intl.NumberFormat()
const formatterDec2 = new Intl.NumberFormat(undefined , { maximumFractionDigits: 2, minimumFractionDigits: 2})
const formatterPctDec2 = new Intl.NumberFormat(undefined , { style: 'percent', maximumFractionDigits: 2, minimumFractionDigits: 2})

const items = ref<CampaignPerformance[]>([])

function load() {
  const endDay = useNow().value
  const startDay = endDay.setDate(endDay.getDate() - 6)
  let myUrl = baseUrl + '?campaign_fk=' + props.campId
  myUrl += '&day_cet=>eq' + useDateFormat(startDay, 'YYYY-MM-DD').value + '&xsort=-day_cet'

  fetchOnce({ ax, myUrl, result: items })
}

onMounted(() => {
  load()
})

</script>

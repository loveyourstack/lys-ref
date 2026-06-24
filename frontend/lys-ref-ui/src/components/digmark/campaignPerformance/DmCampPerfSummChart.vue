<template>
  <v-row>
    <v-col>
      <div class="dt-title">
        <slot name="title">{{ $t('charts.campaign_charts.spend_vs_revenue') }}</slot>
      </div>
    </v-col>
  </v-row>

  <v-row>
    <v-col>
      <div>
        <Bar v-if="!isLoading" :data="chartData" :options="chartOptions" />
      </div>
    </v-col>
  </v-row>

</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { useDateFormat } from '@vueuse/core'
import { useTheme } from 'vuetify'
import { type ChartData, type ChartOptions, Chart as ChartJS, Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale } from 'chart.js'
import { Bar } from 'vue-chartjs'
import { fetchOnce } from 'lys-vue'
import ax from '@/api'
import { type CampaignPerfLatestSummary } from '@/types/digmark'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend)

const theme = useTheme()
const chartTextColor = computed<string>(() => String(theme.global.current.value.colors['on-surface'] ?? '#000000'))

const chartOptions = computed<ChartOptions<'bar'>>(() => ({
  responsive: true,
  plugins: {
    legend: {
      labels: {
        color: chartTextColor.value,
      },
    },
  },
}))

const chartData = ref<ChartData<'bar'>>({
  labels: [],
  datasets: [],
})

const baseUrl = '/a/digmark/campaign-performance-latest-summary' // source returns data sorted by day

const items = ref<CampaignPerfLatestSummary[]>([])
const isLoading = ref(false)

function refreshItems() {
  chartData.value.labels = []
  chartData.value.datasets = []

  fetchOnce({ ax, myUrl: baseUrl, result: items, isLoading, onSuccess: () => {

    const dayLabels = items.value.map((item: CampaignPerfLatestSummary) => useDateFormat(new Date(item.day), 'DD MMM').value)

    const chartObj: ChartData<'bar'> = {
      labels: dayLabels,
      datasets: [
        {
          label: 'Spend',
          data: items.value.map((item: CampaignPerfLatestSummary) => item.total_spend),
          backgroundColor: '#ff7f0e',
        },
        {
          label: 'Revenue',
          data: items.value.map((item: CampaignPerfLatestSummary) => item.total_revenue),
          backgroundColor: '#2ca02c',
        },
      ],
    }

    chartData.value = chartObj
  }})
}

onMounted(() => {
  refreshItems()
})

</script>
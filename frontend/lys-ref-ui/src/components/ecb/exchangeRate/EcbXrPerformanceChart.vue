<template>
  <v-row>
    <v-col>
      <div class="dt-title">
        <slot name="title">{{ $t('charts.xr_performance.title') }}</slot>
      </div>
    </v-col>
  </v-row>

  <v-row density="comfortable">
    <v-col class="mb-2">
      <div class="dt-subtitle">{{ $t('charts.xr_performance.p1') }}</div>
      <div class="dt-subtitle">{{ $t('charts.xr_performance.p2') }}</div>
    </v-col>
  </v-row>

  <v-row density="comfortable">
    <v-col>
    </v-col>

    <v-col>
      <v-autocomplete label="Performance period" v-model="selectedPeriod" density="comfortable" width="300" class="float-right"
        :items="coreStore.xrPeriods"
        @update:model-value="refreshItems()"
      ></v-autocomplete>
    </v-col>
  </v-row>

  <v-row>
    <v-col>
      <div>
        <Line v-if="!isLoading" :data="chartData" :options="chartOptions" />
      </div>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import { onMounted, ref, computed } from 'vue'
import { useDateFormat } from '@vueuse/core'
import { useTheme } from 'vuetify'
import { type ChartData, type ChartOptions, Chart as ChartJS, Title, Tooltip, Legend, LineElement, CategoryScale, LinearScale, PointElement } from 'chart.js'
import { Line } from 'vue-chartjs'
import { fetchOnce, useJsonLs } from 'lys-vue'
import ax from '@/api'
import { useCoreStore } from '@/stores/core'
import { type XrPerfNormalized } from '@/types/ecb'

// extend Chart.js dataset type to include rates for tooltip display
type XrPerfDataset = ChartData<'line'>['datasets'][number] & {
  rates?: number[]
}

ChartJS.register(CategoryScale, LinearScale, LineElement, PointElement, Title, Tooltip, Legend)

const coreStore = useCoreStore()

// compute chart text colors to respect light/dark theme, defaulting to black if not found
const theme = useTheme()
const chartTextColor = computed<string>(() => String(theme.global.current.value.colors['on-surface'] ?? '#000000'))

const chartOptions = computed<ChartOptions<'line'>>(() => ({ 
  responsive: true,
  plugins: {
    legend: {
      labels: {
        color: chartTextColor.value,
      },
    },
    tooltip: {
      callbacks: {
        label(context) {
          const dataset = context.dataset as XrPerfDataset
          const rate = dataset.rates?.[context.dataIndex]

          // show rate in tooltip if available, otherwise just show performance value
          return String(context.dataset.label ?? '') + ': ' + String(rate ?? context.raw)
        },
      },
    },
  },
}))

const chartData = ref<ChartData<'line'>>({
  labels: [],
  datasets: [],
})

const baseUrl = '/a/ecb/xr-performance-normalized'

const items = ref<XrPerfNormalized[]>([])
const isLoading = ref(false)

const selectedPeriod = ref('Last 14 days')

useJsonLs({
  lsKey: 'xr_perf_chart',
  refs: {
    selectedPeriod,
  },
})

function refreshItems() {
  chartData.value.labels = []
  chartData.value.datasets = []

  let myUrl = baseUrl + '?period=' + selectedPeriod.value
  myUrl += '&from_currency_code=EUR&xsort=day,to_currency_code&xper_page=5000'

  fetchOnce({ ax, myUrl, result: items, isLoading, onSuccess: () => {

    // distinct days
    const days = Array.from(new Set(items.value.map((item: XrPerfNormalized) => String(item.day)))).sort()

    // formatted day labels
    const dayLabels = days.map(day => useDateFormat(new Date(day), 'DD MMM').value)

    // distinct to_curr codes
    const toCurrCodes = Array.from(new Set(items.value.map((item: XrPerfNormalized) => item.to_currency_code))).sort()

    // line color palette
    const palette = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728', '#9467bd', '#8c564b', '#e377c2', '#7f7f7f']

    let currDatasets: XrPerfDataset[] = []
  
    toCurrCodes.forEach((code, i) => {

      // filter data by this code. Data must be sorted by day in the API call above
      const codeData = items.value.filter((item: XrPerfNormalized) => item.to_currency_code === code)
      
      currDatasets.push({
        label: code,
        data: codeData.map((item: XrPerfNormalized) => item.normalized_perf),
        rates: codeData.map((item: XrPerfNormalized) => item.rate),
        borderColor: palette[i % palette.length], // using modulo to cycle through palette if more currencies than colors
        fill: false,
        tension: 0.1,
        spanGaps: true
      } as XrPerfDataset)
    })

    const chartObj: ChartData<'line'> = {
      labels: dayLabels,
      datasets: currDatasets
    }

    // replace chart data in a single assignment, or else reactivity doesn't work
    chartData.value = chartObj
  }})
}

onMounted(() => {
  refreshItems()
})

</script>
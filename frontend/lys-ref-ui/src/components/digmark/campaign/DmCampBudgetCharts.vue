<template>
  <v-row>
    <v-col>
      <div class="dt-title">
        <slot name="title">{{ $t('charts.campaign_charts.budget_breakdown') }}</slot>
      </div>
    </v-col>
  </v-row>

  <v-row>
    <v-col>
      <div>
        <Pie v-if="!isLoading" :style="myStyles" :data="managerChartData" :options="managerChartOptions" />
      </div>
    </v-col>

    <v-col>
      <div>
        <Pie v-if="!isLoading" :style="myStyles" :data="verticalChartData" :options="verticalChartOptions" />
      </div>
    </v-col>
  </v-row>

</template>

<script lang="ts" setup>
import { onMounted, ref, computed } from 'vue'
import { useTheme } from 'vuetify'
import { type ChartData, type ChartOptions, Chart as ChartJS, ArcElement, Tooltip, Legend, Title } from 'chart.js'
import { Pie } from 'vue-chartjs'
import ax from '@/api'
import { type BudgetByManager, type BudgetByVertical } from '@/types/digmark'

ChartJS.register(ArcElement, Tooltip, Legend, Title)

const byManagerUrl = '/a/digmark/manager-budgets'
const byVerticalUrl = '/a/digmark/vertical-budgets'

const managerItems = ref<BudgetByManager[]>([])
const verticalItems = ref<BudgetByVertical[]>([])
const isLoading = ref(false)

const palette = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728', '#9467bd', '#8c564b', '#e377c2', '#7f7f7f', '#bcbd22', '#17becf']

const managerChartData = ref<ChartData<'pie'>>({ labels: [], datasets: [] })
const verticalChartData = ref<ChartData<'pie'>>({ labels: [], datasets: [] })

const theme = useTheme()
const chartTextColor = computed<string>(() => String(theme.global.current.value.colors['on-surface'] ?? '#000000'))

// pass styles to be applied to outer div per example: https://vue-chartjs.org/guide/examples.html#chart-with-dynamic-styles
const myStyles = {
  // make chart larger than default so that the pie is not squished due to the large number of categories
  height: '450px',
  width: '450px',
  position: 'relative',
}

// compute options so that theme value is properly reactive
const baseOptions = computed<ChartOptions<'pie'>>(() => ({
  responsive: true,
  maintainAspectRatio: false, // needed for container size override
  plugins: {
    legend: { 
      position: 'right', // due to large number of verticals. If on top per default, pie is squished
      labels: {
        color: chartTextColor.value,
      },
    },
    title: {
      display: true,
      text: '',
      color: chartTextColor.value,
      font: { size: 16 },
      padding: { top: 8, bottom: 12 },
    },
  },
}))

const managerChartOptions = computed<ChartOptions<'pie'>>(() => {
  const options = structuredClone(baseOptions.value)
  options.plugins!.title!.text = 'By manager'
  return options
})

const verticalChartOptions = computed<ChartOptions<'pie'>>(() => {
  const options = structuredClone(baseOptions.value)
  options.plugins!.title!.text = 'By vertical'
  return options
})

function refreshItems() {
  isLoading.value = true
  
  Promise.all([
    ax.get(byManagerUrl),
    ax.get(byVerticalUrl),
  ]).then(([
    managersResp,
    verticalsResp
  ]) => {
    managerItems.value = managersResp.data.data
    verticalItems.value = verticalsResp.data.data

    managerChartData.value = {
      labels: managerItems.value.map((item: BudgetByManager) => item.manager),
      datasets: [{
        data: managerItems.value.map((item: BudgetByManager) => item.total_budget),
        backgroundColor: palette.slice(0, managerItems.value.length),
      }],
    }

    verticalChartData.value = {
      labels: verticalItems.value.map((item: BudgetByVertical) => item.vertical),
      datasets: [{
        data: verticalItems.value.map((item: BudgetByVertical) => item.total_budget),
        backgroundColor: palette.slice(0, verticalItems.value.length),
      }],
    }
  })
  .catch() // handled by interceptor
  .finally(() => {
    isLoading.value = false
  })
}

onMounted(() => {
  refreshItems()
})

</script>
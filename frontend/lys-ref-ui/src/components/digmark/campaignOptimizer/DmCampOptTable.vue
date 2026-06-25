<template>
  <v-data-table-server
    v-model="selected"
    v-model:expanded="expanded"
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
    show-select
    show-expand
    @update:options="loadItems"
  >
    <template #top>
      <l-dt-top :ax="ax" :title="props.title ?? $t('advanced_tables.optimizer.title')" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('advanced_tables.optimizer.p1') }}</div>
          <div class="dt-subtitle">
            {{ $t('advanced_tables.optimizer.p2') }}
            <ul class="mt-1 mb-1">
              <li>{{ $t('advanced_tables.optimizer.p2_list.item_1') }}</li>
              <li>{{ $t('advanced_tables.optimizer.p2_list.item_2') }}</li>
              <li>{{ $t('advanced_tables.optimizer.p2_list.item_3') }}</li>
              <li>{{ $t('advanced_tables.optimizer.p2_list.item_4') }}</li>
              <li>{{ $t('advanced_tables.optimizer.p2_list.item_5') }}</li>
            </ul>
          </div>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col>
          <dm-camp-opt-table-filters @update="refreshItems()" @updateDebounced="refreshItemsDebounced()"
            v-model:filterClicks="filterClicks"
            v-model:filterConversions="filterConversions"
            v-model:filterCountryFKs="filterCountryFKs"
            v-model:filterDailyBudget="filterDailyBudget"
            v-model:filterImpressions="filterImpressions"
            v-model:filterIsActive="filterIsActive"
            v-model:filterName="filterName"
            v-model:filterProfit="filterProfit"
            v-model:filterRoi="filterRoi"
            v-model:filterRevenue="filterRevenue"
            v-model:filterSpend="filterSpend"
            v-model:filterVerticalFks="filterVerticalFks"
          />
        </v-col>
      </v-row>

      <v-row density="comfortable">
        <v-col>
          <v-tabs v-model="selectedManager" @update:model-value="refreshItems()">
            <v-tab v-for="manager in digmarkStore.managers" :value="manager" :key="manager">{{ manager }}</v-tab>
          </v-tabs>
        </v-col>

        <v-col>
          <v-autocomplete label="Performance period" v-model="selectedPeriod" density="comfortable" width="300" class="float-right"
            :items="coreStore.periods"
            @update:model-value="refreshItems()"
          ></v-autocomplete>
        </v-col>
      </v-row>

      <v-row v-if="selected.length > 0" density="compact" class="mt-0">
        <v-col>
          <dm-camp-opt-table-bulk-edit :campaign-ids="selected" :patch-url="patchUrl" @update="preserveSelection = true; refreshItems()" />
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.clicks`]="{ item }">
      {{ formatter.format(item.clicks!) }}
    </template>

    <template v-slot:[`item.conversions`]="{ item }">
      {{ formatter.format(item.conversions!) }}
    </template>

    <template v-slot:[`item.daily_budget_eur`]="{ item }">
      <span v-if="item.editing_daily_budget" class="d-flex justify-content align-center">
        <v-form ref="dailyBudgetForm">
          <v-text-field v-model.number="item.daily_budget_eur" hide-details density="compact" style="min-width: 110px;"
            :append-inner-icon="item.patch_daily_budget_icon"
            autofocus @focus="(e: any) => (e.target as HTMLInputElement)?.select()"
            @keydown="onDailyBudgetKeydown(item, $event)"
            :rules="[(v: number) => v >= 0 && v <= 2000 || 'Daily budget must be between 0 and 2,000']"
          ></v-text-field>
        </v-form>
        <v-icon class="ml-4" color="secondary" icon="mdi-content-save" :disabled="!auth.isWriter()" @click="patchDailyBudget(item)"></v-icon>
        <v-icon class="ml-4" color="secondary" icon="mdi-close" @click="refreshItems()"></v-icon>
      </span>
      <span v-else class="d-flex text-no-wrap" style="justify-content: flex-end;">
        {{ formatterDec2.format(item.daily_budget_eur!) + ' €' }}
        <v-icon class="ml-3" color="secondary" icon="mdi-square-edit-outline" :disabled="!auth.isWriter()" @click="item.editing_daily_budget = true"></v-icon>
      </span>
    </template>

     <template v-slot:[`item.impressions`]="{ item }">
      {{ formatter.format(item.impressions!) }}
    </template>

    <template v-slot:[`item.is_active`]="{ item }">
      <v-switch v-model="item.is_active" :disabled="!auth.isWriter()" hide-details color="secondary" @update:modelValue="patchIsActive(item)"
        :append-icon="item.patch_is_active_icon">
      </v-switch>
    </template>

     <template v-slot:[`item.profit_eur`]="{ item }">
      <span class="d-flex text-no-wrap justify-end ga-1">
        <span :class="item.profit_eur! > 0 ? 'text-success' : 'text-error'">
          {{ formatterDec2.format(item.profit_eur!) + ' €' }}
        </span>
        <v-tooltip v-if="showProfitTooltips()" :text="getTrendInfo(item.trend).tooltip" :disabled="getTrendInfo(item.trend).tooltip === ''">
          <template v-slot:activator="{ props }">
            <v-icon v-bind="props" size="small" :color="getTrendInfo(item.trend).color" :icon="getTrendInfo(item.trend).icon"></v-icon>
          </template>
        </v-tooltip>
        <v-tooltip v-if="showProfitTooltips()":text="getVolatilityInfo(item.volatility).tooltip" :disabled="getVolatilityInfo(item.volatility).tooltip === ''">
          <template v-slot:activator="{ props }">
            <v-icon v-bind="props" size="small" :color="getVolatilityInfo(item.volatility).color" :icon="getVolatilityInfo(item.volatility).icon"></v-icon>
          </template>
        </v-tooltip>
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

    <template v-slot:[`expanded-row`]="{ item }">
      <dm-camp-opt-table-exp-row :camp-id="item.id" :selected-headers="selectedHeaders" />
    </template>

    <template v-slot:[`body.append`]="{}">
      <dm-camp-opt-table-totals-row :base-url="baseUrl" :filter-str="getFilterStr()" :refresh="refreshTotals" :selected-headers="selectedHeaders" />
    </template>

    <template #bottom>
      <l-dt-bottom :itemsPerPage="itemsPerPage" :page="page" :totalItemsIsEstimate="totalItemsIsEstimate" :totalItemsEstimated="totalItemsEstimated"></l-dt-bottom>
    </template>

  </v-data-table-server>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { type SortItem } from 'vuetify/lib/components/VDataTable/composables/sort.mjs'
import { fetchDtItems, type NumericFilter, processURIOptions, getNumericFilterUrlParams, getTextFilterUrlParam } from 'lys-vue'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { useCoreStore } from '@/stores/core'
import { useDigmarkStore } from '@/stores/digmark'
import { type CampaignOptimizer } from '@/types/digmark'

const props = defineProps<{
  title?: string
}>()

const coreStore = useCoreStore()
const digmarkStore = useDigmarkStore()

const headers = [
  { title: 'Name', key: 'name' },
  { title: 'Manager', key: 'manager' },
  { title: 'Country', key: 'country' },
  { title: 'Vertical', key: 'vertical' },
  { title: 'Active', key: 'is_active' },  
  { title: 'Daily budget', key: 'daily_budget_eur', align: 'end' },
  { title: 'Impressions', key: 'impressions', align: 'end' },
  { title: 'Clicks', key: 'clicks', align: 'end' },
  { title: 'Conversions', key: 'conversions', align: 'end' },
  { title: 'Revenue', key: 'revenue_eur', align: 'end' },
  { title: 'Spend', key: 'spend_eur', align: 'end' },
  { title: 'Profit', key: 'profit_eur', align: 'end' },
  { title: 'ROI', key: 'return_on_investment', align: 'end' },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/digmark/campaign-optimizer'
const patchUrl = '/a/digmark/campaigns'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl, getFilterStr)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  refreshItems, refreshItemsDebounced
} = useTableState<CampaignOptimizer>({ ax, baseUrl, getFilterStr })

const dailyBudgetForm = ref()

const filterClicks = ref<NumericFilter>()
const filterConversions = ref<NumericFilter>()
const filterCountryFKs = ref<number[]>()
const filterDailyBudget = ref<NumericFilter>()
const filterImpressions = ref<NumericFilter>()
const filterIsActive = ref<boolean>()
const filterName = ref<string>()
const filterProfit = ref<NumericFilter>()
const filterRoi = ref<NumericFilter>()
const filterRevenue = ref<NumericFilter>()
const filterSpend = ref<NumericFilter>()
const filterVerticalFks = ref<number[]>()

// mandatory filters: use prefix "selected" rather than "filter"
const selectedManager = ref<string>('All')
const selectedPeriod = ref<string>('Today')

const formatter = new Intl.NumberFormat()
const formatterDec2 = new Intl.NumberFormat(undefined , { maximumFractionDigits: 2, minimumFractionDigits: 2})
const formatterPctDec2 = new Intl.NumberFormat(undefined , { style: 'percent', maximumFractionDigits: 2, minimumFractionDigits: 2})

const selected = ref<number[]>([])
const preserveSelection = ref(false)
const selectedIdsForRestore = ref<number[]>([])

const expanded = ref([])

const refreshTotals = ref('')

const { resetTable } = useJsonLs({
  lsKey: 'campaign_opt_dt',
  refs: {
    excludedHeaders,
    filterClicks,
    filterConversions,
    filterCountryFKs,
    filterDailyBudget,
    filterImpressions,
    filterIsActive,
    filterName,
    filterProfit,
    filterRoi,
    filterRevenue,
    filterSpend,
    filterVerticalFks,
    itemsPerPage,
    selectedManager,
    selectedPeriod,
    sortBy,
  },
})

function getFilterStr(): string {
  let ret = ''

  // mandatory filters
  ret += '&period=' + selectedPeriod.value
  if (selectedManager.value != 'All') { ret += '&manager=' + selectedManager.value }

  // optional filters
  ret += getNumericFilterUrlParams('clicks', filterClicks.value)
  ret += getNumericFilterUrlParams('conversions', filterConversions.value)
  if (filterCountryFKs.value && filterCountryFKs.value.length > 0) { ret += '&country_fk=' + filterCountryFKs.value.join('|') }
  ret += getNumericFilterUrlParams('daily_budget_eur', filterDailyBudget.value)
  ret += getNumericFilterUrlParams('impressions', filterImpressions.value)
  if (filterIsActive.value != undefined) { ret += '&is_active=' + filterIsActive.value }
  ret += getTextFilterUrlParam('name', filterName.value)
  ret += getNumericFilterUrlParams('profit_eur', filterProfit.value)
  ret += getNumericFilterUrlParams('return_on_investment', filterRoi.value, true)
  ret += getNumericFilterUrlParams('revenue_eur', filterRevenue.value)
  ret += getNumericFilterUrlParams('spend_eur', filterSpend.value)
  if (filterVerticalFks.value && filterVerticalFks.value.length > 0) { ret += '&vertical_fk=' + filterVerticalFks.value.join('|') }

  return ret
}

function getTrendInfo(trend: number): {icon: string, color: string, tooltip: string} {
  if (trend < -100) { return { icon: 'mdi-arrow-down', color: 'error', tooltip: 'Strong downward trend' } }
  if (trend < -50) { return { icon: 'mdi-arrow-down', color: 'orange', tooltip: 'Moderate downward trend' } }
  if (trend > 50) { return { icon: 'mdi-arrow-up', color: 'yellow', tooltip: 'Moderate upward trend' } }
  if (trend > 100) { return { icon: 'mdi-arrow-up', color: 'success', tooltip: 'Strong upward trend' } }
  return { icon: 'mdi-none', color: '', tooltip: '' }
}

function getVolatilityInfo(vol: number): {icon: string, color: string, tooltip: string} {
  if (vol >= 0.09) { return { icon: 'mdi-chart-timeline-variant', color: 'error', tooltip: 'High volatility' } }
  if (vol >= 0.06) { return { icon: 'mdi-chart-timeline-variant', color: 'orange', tooltip: 'Moderate volatility' } }
  if (vol >= 0.02 && vol <= 0.028) { return { icon: 'mdi-chart-timeline-variant', color: 'yellow', tooltip: 'Low volatility' } }
  if (vol > 0 && vol < 0.02) { return { icon: 'mdi-chart-timeline-variant', color: 'success', tooltip: 'Very low volatility' } }
  return { icon: 'mdi-none', color: '', tooltip: '' }
}

function loadItems(options: { page: number, itemsPerPage: number, sortBy: SortItem[] }) {

  // if explicitly preserving selection (e.g. after bulk edit), save selection for restoring after data refresh
  if (preserveSelection.value) {
    selectedIdsForRestore.value = [...selected.value]
  } else {
    selectedIdsForRestore.value = []
  }

  // reset selected and expanded, since filtering / paging / sorting will potentially no longer show the affected rows
  selected.value = []
  expanded.value = []

  let myUrl = processURIOptions(baseUrl, options)
  myUrl += getFilterStr()
  fetchDtItems({ ax, myUrl, page: options.page, itemsPerPage: options.itemsPerPage, items, totalItems, totalItemsIsEstimate, totalItemsEstimated, onSuccess: () => {
    items.value.forEach(item => {
      item.editing_daily_budget = false
      item.patch_daily_budget_icon = 'mdi-none'
      item.patch_is_active_icon = 'mdi-none'
    })

    // if selection was preserved, restore selection of rows which are still visible after the update
    if (preserveSelection.value) {
      selected.value = selectedIdsForRestore.value.filter(id => items.value.some(item => item.id === id))
      preserveSelection.value = false
    }

    // update totals
    refreshTotals.value = String(Date.now())
  } })
}

function onDailyBudgetKeydown(item: CampaignOptimizer, event: KeyboardEvent) {
  if (event.key === 'Enter') {
    event.preventDefault()
    patchDailyBudget(item)
    return
  }

  if (event.key === 'Escape') {
    event.preventDefault()
    refreshItems()
  }
}

async function patchDailyBudget(item: CampaignOptimizer) {

  const {valid} = await dailyBudgetForm.value?.validate()
  if (!valid) {
    return
  }

  item.patch_daily_budget_icon = 'mdi-loading'

  const patchInput: { 'daily_budget_eur': number } = { 'daily_budget_eur': item.daily_budget_eur }
  const req = ax.patch(patchUrl + '/' + item.id, patchInput)

  // add artificial delay so that loading icon is visible
  await Promise.all([req, new Promise((r) => setTimeout(r, import.meta.env.VITE_FAKE_API_DELAY_MS))])
    .then(() => {
    })
    .catch() // handled by interceptor
    .finally(() => {
      item.patch_daily_budget_icon = 'mdi-none'
      refreshItems() // always refresh items, so that previous value is shown if patch fails
    })
}

async function patchIsActive(item: CampaignOptimizer) {

  item.patch_is_active_icon = 'mdi-loading'

  const patchInput: { 'is_active': boolean } = { 'is_active': item.is_active }
  const req = ax.patch(patchUrl + '/' + item.id, patchInput)

  await Promise.all([req, new Promise((r) => setTimeout(r, import.meta.env.VITE_FAKE_API_DELAY_MS))])
    .then(() => {
    })
    .catch() // handled by interceptor
    .finally(() => {
      item.patch_is_active_icon = 'mdi-none'
      refreshItems()
    })
}

function showProfitTooltips() {
  return selectedPeriod.value !== 'Today' && selectedPeriod.value !== 'Yesterday'
}

</script>

<template>
  <l-dialog-card v-model="showEdit">
    <dm-camp-form :id="editID"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @delete="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></dm-camp-form>
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
      <l-dt-top :ax="ax" :title="props.title ?? $t('entity_relationships.campaigns.title')" :headers="headers" :excelDlUrl="excelDlUrl" v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()">
        <v-btn color="secondary" @click="editID = 0; showEdit = true">{{ $t('actions.add') }}</v-btn>
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('entity_relationships.campaigns.p1') }}</div>
          <div class="dt-subtitle">{{ $t('entity_relationships.campaigns.p2') }}</div>
          <div class="dt-subtitle">{{ $t('entity_relationships.campaigns.p3') }}</div>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col>
          <dm-camp-table-filters @update="refreshItems()" @updateDebounced="refreshItemsDebounced()"
            v-model:filterCountryFKs="filterCountryFKs"
            v-model:filterDailyBudget="filterDailyBudget"
            v-model:filterIsActive="filterIsActive"
            v-model:filterManagers="filterManagers"
            v-model:filterName="filterName"
            v-model:filterVerticalFks="filterVerticalFks"
          />
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.is_active`]="{ item }">
      <v-icon v-if="item.is_active" size="small" icon="mdi-check"></v-icon>
    </template>

     <template v-slot:[`item.daily_budget_eur`]="{ item }">
      {{ formatterDec2.format(item.daily_budget_eur!) + ' €' }}
    </template>

    <template v-slot:[`item.actions`]="{ item }">
      <v-btn icon flat size="small" v-tooltip="`${$t('actions.edit')}`" @click="editID = item.id; showEdit = true">
        <v-icon color="primary" icon="mdi-square-edit-outline"></v-icon>
      </v-btn>
    </template>

    <template #bottom>
      <l-dt-bottom :itemsPerPage="itemsPerPage" :page="page" :totalItemsIsEstimate="totalItemsIsEstimate" :totalItemsEstimated="totalItemsEstimated"></l-dt-bottom>
    </template>

  </v-data-table-server>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { type SortItem } from 'vuetify/lib/components/VDataTable/composables/sort.mjs'
import { type NumericFilter, getNumericFilterUrlParams, getTextFilterUrlParam } from 'lys-vue'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import { type Campaign } from '@/types/digmark'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Name', key: 'name' },
  { title: 'Manager', key: 'manager' },
  { title: 'Country', key: 'country' },
  { title: 'Vertical', key: 'vertical' },
  { title: 'Active', key: 'is_active' },
  { title: 'Daily budget', key: 'daily_budget_eur', align: 'end' },
  { title: 'Performance range', key: 'performance_range', sortable: false },
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/digmark/campaigns'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl, getFilterStr)

const route = useRoute()
const router = useRouter()

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems, refreshItemsDebounced
} = useTableState<Campaign>({ ax, baseUrl, getFilterStr, mapUrl })

const filterCountryFKs = ref<number[]>()
const filterDailyBudget = ref<NumericFilter>()
const filterIsActive = ref<boolean>()
const filterManagers = ref<string[]>()
const filterName = ref<string>()
const filterVerticalFks = ref<number[]>()

const editID = ref(0)
const showEdit = ref(false)

const formatterDec2 = new Intl.NumberFormat(undefined , { maximumFractionDigits: 2, minimumFractionDigits: 2})

const { resetTable } = useJsonLs({
  lsKey: 'campaigns_dt',
  refs: {
    excludedHeaders,
    filterCountryFKs,
    filterDailyBudget,
    filterIsActive,
    filterManagers,
    filterName,
    filterVerticalFks,
    itemsPerPage,
    sortBy,
  },
})

function getFilterStr(): string {
  let ret = ''

  if (filterCountryFKs.value && filterCountryFKs.value.length > 0) { ret += '&country_fk=' + filterCountryFKs.value.join('|') }
  ret += getNumericFilterUrlParams('daily_budget_eur', filterDailyBudget.value)
  if (filterIsActive.value != undefined) { ret += '&is_active=' + filterIsActive.value }
  if (filterManagers.value && filterManagers.value.length > 0) { ret += '&manager=' + filterManagers.value.join('|') }
  ret += getTextFilterUrlParam('name', filterName.value)
  if (filterVerticalFks.value && filterVerticalFks.value.length > 0) { ret += '&vertical_fk=' + filterVerticalFks.value.join('|') }

  return ret
}

function mapUrl(url: string, options: { page: number, itemsPerPage: number, sortBy: SortItem[] }): string {

  // allow passing of vertical_fk via URL param
  if (route.query['vertical_fk']) {

    filterVerticalFks.value = String(route.query['vertical_fk']).split('|').map(Number)

    // change is now in LS due to watcher in composable: redirect to same page without this URL param
    const nextQuery = { ...route.query }
    delete nextQuery.vertical_fk
    router.replace({ path: route.path, query: nextQuery })
  }

  return url
}

</script>

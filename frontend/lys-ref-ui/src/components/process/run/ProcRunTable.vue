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
      <l-dt-top :ax="ax" :title="props.title ?? 'Runs'" :headers="headers" :excelDlUrl="excelDlUrl" v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()">
        <v-btn icon flat size="small" v-tooltip="'Refresh list'" @click="refreshItems()">
          <v-icon icon="mdi-refresh"></v-icon>
        </v-btn>
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">Each run is an execution of a flow step and its dependencies. Click 'View steps' to see the result of each process step.</div>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col>
          <proc-run-table-filters @update="refreshItems()" @updateDebounced="refreshItemsDebounced()"
            v-model:filterFlowFk="filterFlowFk"
          />
        </v-col>
      </v-row>
    </template>

     <template v-slot:[`item.started_at`]="{ item }">
      <span v-if="new Date(item.started_at) > new Date(2000,1,1)">{{ useDateFormat(item.started_at, 'DD MMM YYYY HH:mm:ss').value }}</span>
    </template>

     <template v-slot:[`item.finished_at`]="{ item }">
      <span v-if="new Date(item.finished_at) > new Date(2000,1,1)">{{ useDateFormat(item.finished_at, 'DD MMM YYYY HH:mm:ss').value }}</span>
    </template>

     <template v-slot:[`item.point_count`]="{ item }">
      <span>
        {{ item.point_count }}
        <v-btn icon flat size="small" v-tooltip="'View steps'" :to="{ name: 'Points', params: { id: item.id }}">
          <v-icon color="primary" icon="mdi-focus-field"></v-icon>
        </v-btn>
      </span>
    </template>

    <template #bottom>
      <l-dt-bottom :itemsPerPage="itemsPerPage" :page="page" :totalItemsIsEstimate="totalItemsIsEstimate" :totalItemsEstimated="totalItemsEstimated"></l-dt-bottom>
    </template>

  </v-data-table-server>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDateFormat } from '@vueuse/core'
import { type SortItem } from 'vuetify/lib/components/VDataTable/composables/sort.mjs'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import { type Run } from '@/types/process'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Flow', key: 'flow' },  
  { title: 'Step', key: 'step_name' },
  { title: 'Started at', key: 'started_at' },
  { title: 'Finished at', key: 'finished_at' },
  { title: 'Stati', key: 'point_stati' },
  { title: '# steps', key: 'point_count' },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/process/runs'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl, getFilterStr)

const route = useRoute()
const router = useRouter()

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems, refreshItemsDebounced
} = useTableState<Run>({ ax, baseUrl, getFilterStr, mapUrl })

const filterFlowFk = ref<number>()

const { resetTable } = useJsonLs({
  lsKey: 'proc_run_dt',
  refs: {
    excludedHeaders,
    filterFlowFk,
    itemsPerPage,
    sortBy,
  },
})

function getFilterStr(): string {
  let ret = ''
  if (filterFlowFk.value) { ret += '&flow_fk=' + filterFlowFk.value }
  return ret
}

function mapUrl(url: string, options: { page: number, itemsPerPage: number, sortBy: SortItem[] }): string {

  // allow passing of flow_fk via URL param
  if (route.query['flow_fk']) {

    filterFlowFk.value = Number(route.query['flow_fk'])

    const nextQuery = { ...route.query }
    delete nextQuery.flow_fk
    router.replace({ path: route.path, query: nextQuery })
  }

  return url
}

</script>

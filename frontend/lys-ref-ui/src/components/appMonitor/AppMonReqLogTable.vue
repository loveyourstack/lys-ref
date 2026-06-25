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
      <l-dt-top :ax="ax" :title="props.title ?? 'Request log'" :headers="headers" :excelDlUrl="excelDlUrl" v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
        <v-btn icon flat v-tooltip="'Refresh'" @click="refreshItems()">
          <v-icon icon="mdi-refresh"></v-icon>
        </v-btn>
      </l-dt-top>

      <v-row density="compact">
        <v-col>
          <app-mon-req-log-table-filters @update="refreshItems()" @updateDebounced="refreshItemsDebounced()"
            v-model:filterCreatedAtDate="filterCreatedAtDate"
            v-model:filterDurationMs="filterDurationMs"
            v-model:filterEndpoint="filterEndpoint"
            v-model:filterStatusCodeOk="filterStatusCodeOk"
            v-model:filterUserName="filterUserName"
          />
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.created_at`]="{ item }">
      <span>{{ useDateFormat(item.created_at, 'DD MMM YYYY HH:mm:ss') }}</span>
    </template>

    <template v-slot:[`item.duration_ms`]="{ item }">
      {{ formatter.format(item.duration_ms) }}
    </template>

    <template #bottom>
      <l-dt-bottom :itemsPerPage="itemsPerPage" :page="page" :totalItemsIsEstimate="totalItemsIsEstimate" :totalItemsEstimated="totalItemsEstimated"></l-dt-bottom>
    </template>
  </v-data-table-server>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useDateFormat } from '@vueuse/core'
import { type DateFilter, type NumericFilter, getDateFilterUrlParams, getNumericFilterUrlParams, getTextFilterUrlParam } from 'lys-vue'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import { type ServerRequest } from '@/types/tech'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Created at', key: 'created_at' },
  { title: 'Method', key: 'method' },
  { title: 'Endpoint', key: 'endpoint' },
  { title: 'User', key: 'user_name' },
  { title: 'IP', key: 'ip' },
  { title: 'Status code', key: 'status_code', align: 'end' },
  { title: 'Duration (ms)', key: 'duration_ms', align: 'end' },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/tech/server-requests'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems, refreshItemsDebounced
} = useTableState<ServerRequest>({ ax, baseUrl, getFilterStr })

const filterCreatedAtDate = ref<DateFilter>()
const filterDurationMs = ref<NumericFilter>()
const filterEndpoint = ref<string>()
const filterStatusCodeOk = ref<boolean>()
const filterUserName = ref<string>()

const formatter = new Intl.NumberFormat()

const { resetTable } = useJsonLs({
  lsKey: 'app_mon_req_log_dt',
  refs: {
    excludedHeaders,
    filterCreatedAtDate,
    filterDurationMs,
    filterEndpoint,
    filterStatusCodeOk,
    filterUserName,
    itemsPerPage,
    sortBy,
  },
})

function getFilterStr(): string {
  let ret = ''

  if (filterStatusCodeOk.value != undefined) {
    if (filterStatusCodeOk.value) {
      ret += '&status_code=200'
    } else {
      ret += '&status_code=!200'
    }
  }

  ret += getDateFilterUrlParams('created_at_date', filterCreatedAtDate.value)
  ret += getNumericFilterUrlParams('duration_ms', filterDurationMs.value)
  ret += getTextFilterUrlParam('endpoint', filterEndpoint.value)
  ret += getTextFilterUrlParam('user_name', filterUserName.value)

  return ret
}

</script>
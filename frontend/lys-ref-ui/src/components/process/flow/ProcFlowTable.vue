<template>
  <l-dialog-card v-model="showEdit">
    <proc-flow-form :id="editID"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @delete="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></proc-flow-form>
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
      <l-dt-top :ax="ax" :title="props.title ?? $t('parallel_processing.flows.title')" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()">
        <v-btn color="secondary" @click="editID = 0; showEdit = true">{{ $t('actions.add') }}</v-btn>
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('parallel_processing.flows.p1') }}</div>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col>
          <proc-flow-table-filters @update="refreshItems()" @updateDebounced="refreshItemsDebounced()"
            v-model:filterName="filterName"
          />
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.params`]="{ item }">
      <span v-if="item.params.length > 0">{{ item.params.join(' ') }}</span>
    </template>

     <template v-slot:[`item.run_count`]="{ item }">
      <span>
        {{ item.run_count }}
        <v-btn v-if="item.run_count > 0" icon flat size="small" v-tooltip:bottom="`${$t('parallel_processing.flows.view_runs')}`" :to="{ name: 'Runs', query: { flow_fk: item.id }}">
          <v-icon color="secondary" icon="mdi-repeat"></v-icon>
        </v-btn>
      </span>
    </template>

     <template v-slot:[`item.step_count`]="{ item }">
      <span>
        {{ item.step_count }}
        <v-btn icon flat size="small" v-tooltip:bottom="`${$t('parallel_processing.flows.view_steps')}`" :to="{ name: 'Steps', params: { id: item.id }}">
          <v-icon color="primary" icon="mdi-focus-field"></v-icon>
        </v-btn>
      </span>
    </template>

    <template v-slot:[`item.actions`]="{ item }">
      <v-btn icon flat size="small" v-tooltip:bottom="`${$t('actions.edit')}`" @click="editID = item.id; showEdit = true">
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
import { getTextFilterUrlParam } from 'lys-vue'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import { type Flow } from '@/types/process'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Name', key: 'name' },
  { title: 'Params', key: 'params' },
  { title: '# steps', key: 'step_count', align: 'end' },
  { title: '# runs', key: 'run_count', align: 'end' },
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/process/flows'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl, getFilterStr)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems, refreshItemsDebounced
} = useTableState<Flow>({ ax, baseUrl, getFilterStr })

const filterName = ref<string>()

const editID = ref(0)
const showEdit = ref(false)

const { resetTable } = useJsonLs({
  lsKey: 'proc_flow_dt',
  refs: {
    excludedHeaders,
    filterName,
    itemsPerPage,
    sortBy,
  },
})

function getFilterStr(): string {
  let ret = ''
  ret += getTextFilterUrlParam('name', filterName.value)
  return ret
}

</script>

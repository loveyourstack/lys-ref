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
      <l-dt-top :ax="ax" :title="props.title ?? 'Login attempts'" :headers="headers" :excelDlUrl="excelDlUrl" v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()">
        <v-btn icon flat v-tooltip="'Refresh'" @click="refreshItems()">
          <v-icon icon="mdi-refresh"></v-icon>
        </v-btn>
      </l-dt-top>
    </template>

    <template v-slot:[`item.actions`]="{ item }">
      <v-btn v-if="item.is_blocked" icon flat size="small" v-tooltip="'Unblock IP'" @click="unblockIp(item.ip)">
        <v-icon color="primary" icon="mdi-lock-open-remove-outline"></v-icon>
      </v-btn>
    </template>

    <template v-slot:[`item.created_at`]="{ item }">
      <span>{{ useDateFormat(item.created_at, 'DD MMM YYYY HH:mm:ss') }}</span>
    </template>

    <template v-slot:[`item.is_blocked`]="{ item }">
      <v-icon v-if="item.is_blocked" size="small" icon="mdi-check"></v-icon>
    </template>

    <template #bottom>
      <l-dt-bottom :itemsPerPage="itemsPerPage" :page="page" :totalItemsIsEstimate="totalItemsIsEstimate" :totalItemsEstimated="totalItemsEstimated"></l-dt-bottom>
    </template>
  </v-data-table-server>
</template>

<script lang="ts" setup>
import { useDateFormat } from '@vueuse/core'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import { type LoginAttempt } from '@/types/tech'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'IP', key: 'ip', sortable: false },
  { title: 'Created at', key: 'created_at', sortable: false },
  { title: 'Is blocked', key: 'is_blocked', sortable: false },
  { title: '# attempts', key: 'num_attempts', sortable: false },
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/tech/login-attempts'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems, refreshItemsDebounced
} = useTableState<LoginAttempt>({ ax, baseUrl })

const { resetTable } = useJsonLs({
  lsKey: 'app_mon_login_attempt_dt',
  refs: {
    excludedHeaders,
    itemsPerPage,
    sortBy,
  },
})

function unblockIp(ip: string) {
  if (!confirm(`Unblock IP ${ip}?`)) return

  const myUrl = baseUrl + '/unblock-ip/' + ip
  ax.post(myUrl).then(() => {
    refreshItemsDebounced()
  })
}

</script>
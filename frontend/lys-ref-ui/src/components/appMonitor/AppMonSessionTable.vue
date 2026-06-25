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
      <l-dt-top :ax="ax" :title="props.title ?? 'Sessions'" :headers="headers" :excelDlUrl="excelDlUrl" v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
        <v-btn icon flat v-tooltip="'Refresh'" @click="refreshItems()">
          <v-icon icon="mdi-refresh"></v-icon>
        </v-btn>
      </l-dt-top>
    </template>

    <template v-slot:[`item.actions`]="{ item }">
      <v-btn icon flat size="small" v-tooltip="'Block IP'" @click="blockSessionIp(item.ip)">
        <v-icon color="error" icon="mdi-block-helper"></v-icon>
      </v-btn>
    </template>

    <template v-slot:[`item.created_at`]="{ item }">
      <span>{{ useDateFormat(item.created_at, 'DD MMM YYYY HH:mm:ss') }}</span>
    </template>

    <template v-slot:[`item.expires_at`]="{ item }">
      <span>{{ useDateFormat(item.expires_at, 'DD MMM YYYY HH:mm:ss') }}</span>
    </template>

    <template v-slot:[`item.last_access_at`]="{ item }">
      <span>{{ useDateFormat(item.last_access_at, 'DD MMM YYYY HH:mm:ss') }}</span>
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
import { type Session } from '@/types/tech'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Last access at', key: 'last_access_at', sortable: false },
  { title: 'User name', key: 'user_name', sortable: false },
  /*{ title: 'Email', key: 'email', sortable: false },*/
  { title: 'IP', key: 'ip', sortable: false },
  { title: 'GeoIP Location', key: 'geo_ip_location', sortable: false },
  { title: 'User agent', key: 'user_agent', sortable: false },
  { title: 'Created at', key: 'created_at', sortable: false },
  { title: 'Expires at', key: 'expires_at', sortable: false },
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/tech/sessions'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems, refreshItemsDebounced
} = useTableState<Session>({ ax, baseUrl })

const { resetTable } = useJsonLs({
  lsKey: 'app_mon_session_dt',
  refs: {
    excludedHeaders,
    itemsPerPage,
    sortBy,
  },
})

function blockSessionIp(ip: string) {
  if (!confirm(`Block IP ${ip}?`)) return

  const myUrl = baseUrl + '/block-ip/' + ip
  ax.post(myUrl).then(() => {
    refreshItemsDebounced()
  })
}

// autorefresh every 10 seconds
setInterval(() => {
  refreshItemsDebounced()
}, 10000)

</script>
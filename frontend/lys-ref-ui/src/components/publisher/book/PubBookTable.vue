<template>
  <l-dialog-card v-model="showEdit">
    <pub-book-form :id="editID"
      @archive="showEdit = false; refreshItems()"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></pub-book-form>
  </l-dialog-card>

  <l-dialog-card v-model="showAudit">
    <l-audit-updates :ax="ax" :baseUrl="auditsBaseUrl" schemaName="publisher" tableName="book" :id="auditID" :title="auditTitle"
      :fkReplacements="[
        { fk_column: 'author_fk', new_column_name: 'author', selectionItems: pubStore.authors },
      ]"
      @close="showAudit = false"
    ></l-audit-updates>
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
      <l-dt-top :ax="ax" :headers="headers" :excelDlUrl="excelDlUrl" v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
        <template #title>
          <v-icon v-if="showArchive" color="secondary" size="small" class="mb-1" icon="mdi-archive-arrow-down-outline"></v-icon>
          {{ showArchive ? $t('user_data_retention.books.archived_title') : $t('user_data_retention.books.title') }}
        </template>

        <v-btn v-if="!showArchive" color="secondary" @click="editID = 0; showEdit = true">{{ $t('actions.add') }}</v-btn>

        <template #menuItems>
          <v-list-item prepend-icon="mdi-archive-arrow-down-outline">
            <v-switch :label="`${$t('actions.show_archive')}`" v-model="showArchive" class="ml-2" color="secondary" hide-details density="comfortable"
              @update:model-value="refreshItems()"
            ></v-switch>
          </v-list-item>
        </template>
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('user_data_retention.books.p1') }}</div>
          <div class="dt-subtitle">{{ $t('user_data_retention.books.p2') }}</div>
        </v-col>
      </v-row>
    </template>

     <template v-slot:[`item.archived_at`]="{ item }">
      <span v-if="item.archived_at">{{ useDateFormat(item.archived_at, 'DD MMM YYYY HH:mm') }}</span>
    </template>

    <template v-slot:[`item.archived_by_cascade`]="{ item }">
      <v-icon v-if="item.archived_by_cascade" size="small" icon="mdi-check"></v-icon>
    </template>

    <template v-slot:[`item.actions`]="{ item }">
      <v-btn v-if="!showArchive" icon flat size="small" v-tooltip:bottom="`${$t('actions.edit')}`" @click="editID = item.id; showEdit = true">
        <v-icon color="primary" icon="mdi-square-edit-outline"></v-icon>
      </v-btn>
      <v-btn v-if="!showArchive" icon flat size="small" v-tooltip:bottom="`${$t('actions.show_update_history')}`" @click="auditID = item.id; showAudit = true">
        <v-icon color="secondary" icon="mdi-history"></v-icon>
      </v-btn>
      <v-btn v-if="showArchive && !item.archived_by_cascade && !item.author_is_archived" :disabled="!auth.isWriter()" icon flat 
        size="small" v-tooltip="`${$t('actions.restore')}`" @click="restoreItem(item.id)">
        <v-icon color="secondary" icon="mdi-restore"></v-icon>
      </v-btn>
    </template>

    <template #bottom>
      <l-dt-bottom :itemsPerPage="itemsPerPage" :page="page" :totalItemsIsEstimate="totalItemsIsEstimate" :totalItemsEstimated="totalItemsEstimated"></l-dt-bottom>
    </template>

  </v-data-table-server>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue'
import { useDateFormat } from '@vueuse/core'
import { type SortItem } from 'vuetify/lib/components/VDataTable/composables/sort.mjs'
import { callPost } from 'lys-vue'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { usePublisherStore } from '@/stores/publisher'
import { type Book } from '@/types/publisher'

const props = defineProps<{
  title?: string
}>()

const pubStore = usePublisherStore()

const allHeaders = [
  { title: 'Author', key: 'author' },
  { title: 'Name', key: 'name' },
  { title: 'Archived at', key: 'archived_at' },
  { title: 'Archived due to author', key: 'archived_by_cascade' },
  { title: 'Actions', key: 'actions', sortable: false },
] as const

const headers = computed(() => {
  return showArchive.value
    ? allHeaders
    : allHeaders.filter((h) => h.key !== 'archived_at' && h.key !== 'archived_by_cascade')
})
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const showArchive = ref(false)

const auditsBaseUrl = '/a/system/audit-updates'

const baseUrl = '/a/publisher/books'
const selectUrl = computed(() => {
  return showArchive.value ? baseUrl + '-archived' : baseUrl
})
const { excelDlUrl } = useTableExcelDlUrl(selectUrl) // passing computed selectUrl

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems,
} = useTableState<Book>({ ax, baseUrl: selectUrl, mapOptions }) // passing computed selectUrl

const editID = ref(0)
const showEdit = ref(false)

const auditID = ref(0)
const auditTitle = computed(() => {
  const item = items.value.find(i => i.id === auditID.value)
  return item ? `${item.name}` : undefined
})
const showAudit = ref(false)

const { resetTable } = useJsonLs({
  lsKey: 'books_dt',
  refs: {
    excludedHeaders,
    itemsPerPage,
    showArchive,
    sortBy,
  },
})

function mapOptions(options: { page: number, itemsPerPage: number, sortBy: SortItem[] }): { page: number, itemsPerPage: number, sortBy: SortItem[] } {

  // if not showing archive, remove sorting by keys that are archive-specific
  if (!showArchive.value) {
    sortBy.value = options.sortBy.filter((s: SortItem) => s.key !== 'archived_at' && s.key !== 'archived_by_cascade')
    options.sortBy = sortBy.value
  }

  return options
}

function restoreItem(id: number) {
  callPost({ ax, myUrl: baseUrl + '/' + id + '/restore', onSuccess: () => { refreshItems() } })
}

</script>

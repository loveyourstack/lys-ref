<template>
  <l-dialog-card v-model="showEdit">
    <core-default-value-form :id="editID"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @delete="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></core-default-value-form>
  </l-dialog-card>

  <l-dialog-card v-model="showImport">
    <l-text-array-entry :title="$t('default_values.import_enter')" :saving="importing" :max-items="maxImportItems" :enterDisabled="!auth.isWriter()"
      :subtitle="`The expected columns are: ${defaultValueImportColumns.join(', ')}`"
      sampleSheetLink="https://docs.google.com/spreadsheets/d/10klDBUMBk5ByLsXmJ5T2jMqp1aeoHJOV-hT6_7-IFrM/edit?pli=1&gid=80181138#gid=80181138"
      :enterLabel="$t('actions.enter')"
      @cancel="showImport = false"
      @enter="(valA: string[]) => { const success = importItems(valA); if (success) { showImport = false } }"
    ></l-text-array-entry>
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
      <l-dt-top :ax="ax" :title="props.title ?? $t('default_values.title')" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
        <v-btn color="secondary" @click="editID = 0; showEdit = true">{{ $t('actions.add') }}</v-btn>
        <v-btn color="secondary" :loading="importing" @click="showImport = true">{{ $t('actions.import') }}</v-btn>
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('default_values.p1') }}</div>
        </v-col>
      </v-row>
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
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { getDefaultValueImportItems } from '@/components/core/defaultValue/default_value_import'
import { type DefaultValue, defaultValueImportColumns } from '@/types/core'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Default text', key: 'c_default_text' },
  { title: 'Suggested text', key: 'c_suggested_text' },
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/core/default-values'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems
} = useTableState<DefaultValue>({ ax, baseUrl })

const editID = ref(0)
const showEdit = ref(false)

const showImport = ref(false)
const importing = ref(false)
const maxImportItems = 10

const { resetTable } = useJsonLs({
  lsKey: 'default_values_dt',
  refs: {
    excludedHeaders,
    itemsPerPage,
    sortBy,
  },
})

function importItems(entryA: string[]): boolean {

  const itemA = getDefaultValueImportItems(entryA, maxImportItems)
  if (!itemA.ok) { alert(itemA.error); return false }

  importing.value = true
  ax.post(baseUrl+'/import', JSON.stringify(itemA.value))
    .then(response => {
      refreshItems()
    })
    .catch() // handled by interceptor
    .finally(() => { importing.value = false })
  
  return true
}

</script>

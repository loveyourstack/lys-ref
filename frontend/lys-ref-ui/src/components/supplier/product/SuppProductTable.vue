<template>
  <l-dialog-card v-model="showEdit">
    <supp-product-form :id="editID" :internal="props.internal"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @delete="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></supp-product-form>
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
      <!-- note passing of computed axInstance and reqHeaders for file download -->
      <l-dt-top :ax="axToUse" :title="props.title ?? 'Supplier products'" :headers="headers" :excelDlUrl="excelDlUrl" v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()"
        :reqHeaders="reqHeaders">
        <v-btn color="secondary" @click="editID = 0; showEdit = true">{{ $t('actions.add') }}</v-btn>
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle" v-if="props.internal">Internal view: all supplier products are visible but not editable.</div>
          <div class="dt-subtitle" v-else>Tenant view: only products of the tenant are visible, and are editable by the tenant.</div>
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
import { ref, computed } from 'vue'
import { type SortItem } from 'vuetify/lib/components/VDataTable/composables/sort.mjs'
import type { AxiosInstance } from 'axios'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import axDefault, { axSupplier } from '@/api'
import { useSupplierStore } from '@/stores/supplier'
import { type Product } from '@/types/supplier'

const props = defineProps<{
  internal: boolean
  title?: string
}>()

const suppStore = useSupplierStore()

const allHeaders = [
  { title: 'Company', key: 'company' },
  { title: 'Name', key: 'name' },
  { title: 'Category', key: 'category' },
  { title: 'Units on order', key: 'units_on_order', align: 'end' },
  { title: 'Created by', key: 'created_by' },
  { title: 'Last updated by', key: 'last_user_update_by' },
  { title: 'Actions', key: 'actions', sortable: false },
] as const

// only show actions when tenant view
const headers = computed(() => {
  return props.internal
    ? allHeaders.filter((h) => h.key !== 'actions')
    : allHeaders
})
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/supplier/products'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl)

const axToUse: AxiosInstance = computed(() => props.internal ? axDefault : axSupplier).value
const reqHeaders = computed(() => props.internal ? undefined : { 'Employee-Email': suppStore.selectedEmpEmail })

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems,
} = useTableState<Product>({ ax: axToUse, baseUrl, reqHeaders: reqHeaders.value })

const editID = ref(0)
const showEdit = ref(false)

const { resetTable } = useJsonLs({
  lsKey: 'products_dt',
  refs: {
    excludedHeaders,
    itemsPerPage,
    sortBy,
  },
})

</script>

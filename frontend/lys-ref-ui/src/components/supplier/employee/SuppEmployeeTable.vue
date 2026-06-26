<template>
  <l-dialog-card v-model="showEdit">
    <supp-employee-form :id="editID" :internal="props.internal"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @delete="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></supp-employee-form>
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
      <l-dt-top :ax="axToUse" :title="props.title ?? $t('supplier_employees.title')" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()" :reqHeaders="reqHeaders"
        :resetTableLabel="$t('actions.reset_table')" :adjustColumnsLabel="$t('actions.adjust_columns')" :downloadToExcelLabel="$t('actions.download_to_excel')">
        <v-btn v-if="props.internal" color="secondary" @click="editID = 0; showEdit = true">{{ $t('actions.add') }}</v-btn>
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle" v-if="props.internal">{{ $t('internal_view.supplier_employees.p1') }}</div>
          <div class="dt-subtitle" v-else>{{ $t('tenant_view.supplier_employees.p1') }}</div>
        </v-col>
      </v-row>
    </template>

    <template v-slot:[`item.email`]="{ item }">
      <span :class="suppStore.selectedEmpEmail === item.email ? 'font-weight-bold' : ''">{{ item.email }}</span>
    </template>

    <template v-slot:[`item.actions`]="{ item }">
      <v-btn icon flat size="small" v-tooltip:bottom="`${$t('actions.edit')}`" @click="editID = item.id; showEdit = true">
        <v-icon color="primary" icon="mdi-square-edit-outline"></v-icon>
      </v-btn>
      <v-btn v-if="!suppStore.selectedEmpEmail" icon flat size="small" v-tooltip:bottom="`${$t('actions.login')}`" @click="suppStore.selectedEmpEmail = item.email!; suppStore.selectedCompId = item.company_fk!">
        <v-icon color="primary" icon="mdi-login"></v-icon>
      </v-btn>
      <v-btn v-if="suppStore.selectedEmpEmail && suppStore.selectedEmpEmail === item.email" 
        icon flat size="small" v-tooltip:bottom="`${$t('actions.logout')}`" @click="suppStore.selectedEmpEmail = ''; suppStore.selectedCompId = 0">
        <v-icon color="primary" icon="mdi-logout"></v-icon>
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
import { type Employee } from '@/types/supplier'

const props = defineProps<{
  internal: boolean
  title?: string
}>()

const suppStore = useSupplierStore()

const allHeaders = [
  { title: 'Company', key: 'company' },
  { title: 'Name', key: 'name' },
  { title: 'Email', key: 'email' },
  { title: 'Actions', key: 'actions', sortable: false },
] as const

// only show actions when internal view
const headers = computed(() => {
  return props.internal
    ? allHeaders
    : allHeaders.filter((h) => h.key !== 'actions')
})
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/supplier/employees'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl)

// internal view: use default ax instance. Tenant view: use ax instance instantiated with supplier API url
const axToUse: AxiosInstance = computed(() => props.internal ? axDefault : axSupplier).value

// if tenant view, include Employee-Email header for fake tenant authentication
const reqHeaders = computed(() => props.internal ? undefined : { 'Employee-Email': suppStore.selectedEmpEmail })

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems,
} = useTableState<Employee>({ ax: axToUse, baseUrl, reqHeaders: reqHeaders.value, mapOptions })

const editID = ref(0)
const showEdit = ref(false)

const { resetTable } = useJsonLs({
  lsKey: 'employees_dt',
  refs: {
    excludedHeaders,
    itemsPerPage,
    sortBy,
  },
})

function mapOptions(options: { page: number, itemsPerPage: number, sortBy: SortItem[] }): { page: number, itemsPerPage: number, sortBy: SortItem[] } {

  // if in tenant view, remove any sorting by company (as that column is not visible), otherwise the backend will return an error
  if (!props.internal) {
    sortBy.value = options.sortBy.filter((s: SortItem) => s.key !== 'company')
    options.sortBy = sortBy.value
  }

  return options
}

</script>

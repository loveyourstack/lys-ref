<template>
  <l-dialog-card v-model="showEdit">
    <core-variant-type-form :id="editID"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @delete="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></core-variant-type-form>
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
      <l-dt-top :ax="ax" :title="props.title ?? 'Variants'" :headers="headers" :excelDlUrl="excelDlUrl" v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()">
        <v-btn color="secondary" @click="editID = 0; showEdit = true">Add</v-btn>
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">Common variants of basic data types.</div>
        </v-col>
      </v-row>
    </template>

     <template v-slot:[`item.c_money_amount`]="{ item }">
      <span :class="item.c_money_amount! < 0 ? 'text-error' : ''">
        {{ formatterDec2.format(item.c_money_amount!) + ' €' }}
      </span>
    </template>

    <template v-slot:[`item.c_percent`]="{ item }">
      {{ formatterPctDec2.format(item.c_percent!) }}
    </template>

    <template v-slot:[`item.actions`]="{ item }">
      <v-btn icon flat size="small" v-tooltip="'Edit'" @click="editID = item.id; showEdit = true">
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
import { type SortItem } from 'vuetify/lib/components/VDataTable/composables/sort.mjs'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import { type VariantType } from '@/types/core'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Constrained text', key: 'c_constrained_text' },
  { title: 'IP', key: 'c_ip' },
  { title: 'Long text', key: 'c_long_text_short' },
  { title: 'Money amount', key: 'c_money_amount', align: 'end' },
  { title: 'Percent', key: 'c_percent', align: 'end' },
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/core/variant-types'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems
} = useTableState<VariantType>({ ax, baseUrl, mapUrl })

const editID = ref(0)
const showEdit = ref(false)

// not using currency option in Intl.NumberFormat for c_money_amount: it adds the symbol as a prefix with no space
const formatterDec2 = new Intl.NumberFormat(undefined , { maximumFractionDigits: 2, minimumFractionDigits: 2})

const formatterPctDec2 = new Intl.NumberFormat(undefined , { style: 'percent', maximumFractionDigits: 2, minimumFractionDigits: 2})

const { resetTable } = useJsonLs({
  lsKey: 'variant_types_dt',
  refs: {
    excludedHeaders,
    itemsPerPage,
    sortBy,
  },
})

function mapUrl(url: string, options: { page: number, itemsPerPage: number, sortBy: SortItem[] }): string {

  // add fields param to prevent selection of c_long_text
  url += '&xfields=-c_long_text'

  return url
}

</script>

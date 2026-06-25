<template>
  <l-dialog-card v-model="showEdit">
    <core-array-type-form :id="editID"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @delete="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></core-array-type-form>
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
      <l-dt-top :ax="ax" :title="props.title ?? $t('type_handling.arrays.title')" :headers="headers" :excelDlUrl="excelDlUrl" 
        v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()">
        <v-btn color="secondary" @click="editID = 0; showEdit = true">{{ $t('actions.add') }}</v-btn>
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('type_handling.arrays.p1') }}</div>
        </v-col>
      </v-row>
    </template>

    <!-- hide zero length arrays, join values with a separator -->
    <template v-slot:[`item.c_bool`]="{ item }">
      <template v-if="item.c_bool.length > 0">
        <span v-for="(i, idx) in item.c_bool" :key="idx">
          <v-icon size="small" :icon="i ? 'mdi-check' : 'mdi-close'"></v-icon>
          <span v-if="idx < item.c_bool.length - 1"> | </span>
        </span>
      </template>
    </template>

    <template v-slot:[`item.c_date`]="{ item }">
      <span v-if="item.c_date.length > 0">{{ item.c_date.map(v => useDateFormat(v, 'DD MMM YYYY').value).join(' | ') }}</span>
    </template>

    <template v-slot:[`item.c_enum`]="{ item }">
      <span v-if="item.c_enum.length > 0">{{ item.c_enum.join(' | ') }}</span>
    </template>

    <template v-slot:[`item.c_int`]="{ item }">
      <template v-if="item.c_int.length > 0">
        <span v-for="(i, idx) in item.c_int" :key="idx">
          <span :class="i < 0 ? 'text-error' : ''">{{ formatter.format(i) }}</span>
          <span v-if="idx < item.c_int.length - 1"> | </span>
        </span>
      </template>
    </template>

    <template v-slot:[`item.c_numeric`]="{ item }">
      <template v-if="item.c_numeric.length > 0">
        <span v-for="(i, idx) in item.c_numeric" :key="idx">
          <span :class="i < 0 ? 'text-error' : ''">{{ formatterDec2.format(i) }}</span>
          <span v-if="idx < item.c_numeric.length - 1"> | </span>
        </span>
      </template>
    </template>

    <template v-slot:[`item.c_text`]="{ item }">
      <span v-if="item.c_text.length > 0">{{ item.c_text.join(' | ') }}</span>
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
import { useDateFormat } from '@vueuse/core'
import { useJsonLs, useTableExcelDlUrl, useTableHeaders, useTableState } from 'lys-vue'
import ax from '@/api'
import { type ArrayType } from '@/types/core'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Bool', key: 'c_bool' },
  { title: 'Date', key: 'c_date' },
  { title: 'Enum', key: 'c_enum' },
  { title: 'Int', key: 'c_int', align: 'end' },
  { title: 'Numeric', key: 'c_numeric', align: 'end' },
  { title: 'Text', key: 'c_text' },
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/core/array-types'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems
} = useTableState<ArrayType>({ ax, baseUrl })

const editID = ref(0)
const showEdit = ref(false)

const formatter = new Intl.NumberFormat()
const formatterDec2 = new Intl.NumberFormat(undefined , { maximumFractionDigits: 2, minimumFractionDigits: 2})

const { resetTable } = useJsonLs({
  lsKey: 'array_types_dt',
  refs: {
    excludedHeaders,
    itemsPerPage,
    sortBy,
  },
})

</script>

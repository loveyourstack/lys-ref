<template>
  <l-dialog-card v-model="showEdit">
    <dm-vertical-form :id="editID"
      @cancel="showEdit = false"
      @create="showEdit = false; refreshItems()"
      @delete="showEdit = false; refreshItems()"
      @update="showEdit = false; refreshItems()"
    ></dm-vertical-form>
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
      <l-dt-top :ax="ax" :title="props.title ?? $t('entity_relationships.verticals.title')" :headers="headers" :excelDlUrl="excelDlUrl" v-model:excludedHeaders="excludedHeaders" @resetTable="resetTable()">
        <v-btn color="secondary" @click="editID = 0; showEdit = true">{{ $t('actions.add') }}</v-btn>
      </l-dt-top>

      <v-row density="comfortable">
        <v-col class="mb-2">
          <div class="dt-subtitle">{{ $t('entity_relationships.verticals.p1') }}</div>

          <i18n-t scope="global" keypath="entity_relationships.verticals.p2" tag="div" class="dt-subtitle">
            <template #viewCampsIcon>
              <v-icon color="secondary" icon="mdi-bullhorn-outline"></v-icon>
            </template>
          </i18n-t>
          
          <div class="dt-subtitle">{{ $t('entity_relationships.verticals.p3') }}</div>
        </v-col>
      </v-row>
    </template>

     <template v-slot:[`item.campaign_count`]="{ item }">
      <span>
        {{ item.campaign_count }}
        <v-btn v-if="item.campaign_count > 0" icon flat size="small" v-tooltip="`${$t('entity_relationships.verticals.view_campaigns')}`" 
          :to="{ name: 'Campaigns', query: { vertical_fk: item.id }}">
          <v-icon color="secondary" icon="mdi-bullhorn-outline"></v-icon>
        </v-btn>
      </span>
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
import { type Vertical } from '@/types/digmark'

const props = defineProps<{
  title?: string
}>()

const headers = [
  { title: 'Name', key: 'name' },
  { title: '# campaigns', key: 'campaign_count', align: 'end' },
  { title: 'Actions', key: 'actions', sortable: false },
] as const
const { excludedHeaders, selectedHeaders } = useTableHeaders(headers)

const baseUrl = '/a/digmark/verticals'
const { excelDlUrl } = useTableExcelDlUrl(baseUrl)

const { items, itemsPerPage, page, sortBy, search, totalItems, totalItemsIsEstimate, totalItemsEstimated,
  loadItems, refreshItems,
} = useTableState<Vertical>({ ax, baseUrl })

const editID = ref(0)
const showEdit = ref(false)

const { resetTable } = useJsonLs({
  lsKey: 'verticals_dt',
  refs: {
    excludedHeaders,
    itemsPerPage,
    sortBy,
  },
})

</script>

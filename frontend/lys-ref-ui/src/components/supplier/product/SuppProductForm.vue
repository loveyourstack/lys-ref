<template>
  <v-card-title class="pl-1 mb-1">
    {{ formTitle }}
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">
    <v-row>
      <v-col class="form-col">

        <v-autocomplete label="Company" v-model="item.company_fk" disabled
          :items="supStore.companies" item-value="id" item-title="name"
          :rules="[(v: number) => !!v || 'Company is required']"
        ></v-autocomplete>

        <v-text-field label="Name" v-model="item.name"
          :rules="[(v: string) => !!v || 'Name is required']"
        ></v-text-field>

        <v-autocomplete label="Category" v-model="item.category_fk"
          :items="supStore.productCategories" item-value="id" item-title="name"
          :rules="[(v: number) => !!v || 'Category is required']"
        ></v-autocomplete>

        <v-text-field label="Units on order" v-model.number="item.units_on_order" type="number"
          :rules="[(v: number) => v >= 0 || 'Units on order must be 0 or more']"
        ></v-text-field>

      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="saveBtnLabel" :saveDisabled="!auth.isWriter()"
      @cancel="emit('cancel')" @save="saveItem">
      <template #extra>
        <v-spacer />
        <v-btn v-if="props.id !== 0" :disabled="!auth.isWriter()" color="error" @click="deleteItem">Delete</v-btn>
      </template>
    </l-cancel-and-save-actions>

  </v-form>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import type { AxiosInstance } from 'axios'
import { useFormCrud } from 'lys-vue'
import axDefault, { axSupplier } from '@/api'
import auth from '@/auth'
import { type Product, NewProduct, GetProductInputFromItem } from '@/types/supplier'
import { useSupplierStore } from '@/stores/supplier'

const props = defineProps<{
  id: number
  internal: boolean
}>()

const emit = defineEmits<{
  (e: 'cancel'): void
  (e: 'create', newID: number): void
  (e: 'delete'): void
  (e: 'load', id: number): void
  (e: 'update'): void
}>()

const supStore = useSupplierStore()

const axToUse: AxiosInstance = computed(() => props.internal ? axDefault : axSupplier).value

const { item, itemForm, saving, saveBtnLabel, showSaved, deleteItem, saveItem } =
  useFormCrud<Product>({
    ax: axToUse,
    id: props.id,
    baseUrl: '/a/supplier/products',
    reqHeaders: props.internal ? undefined : { 'Employee-Email': supStore.selectedEmpEmail },
    newItem: () => NewProduct(supStore.selectedCompId),
    getInput: GetProductInputFromItem,

    onCreate: (id) => { emit('create', id) },
    onDelete: () => { emit('delete') },
    onLoad: (id) => emit('load', id),
    onUpdate: () => { emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? (item.value?.name ?? '') : 'New product'
})

</script>

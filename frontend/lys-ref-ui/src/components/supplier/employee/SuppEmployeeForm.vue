<template>
  <v-card-title class="pl-1 mb-1">
    {{ formTitle }}
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">
    <v-row>
      <v-col class="form-col">

        <v-autocomplete label="Company" v-model="item.company_fk"
          :items="supStore.companies" item-value="id" item-title="name"
          :rules="[(v: number) => !!v || 'Company is required']"
        ></v-autocomplete>

        <v-text-field label="Given name" v-model="item.given_name"
          :rules="[(v: string) => !!v || 'Given name is required']"
        ></v-text-field>

        <v-text-field label="Family name" v-model="item.family_name"
          :rules="[(v: string) => !!v || 'Family name is required']"
        ></v-text-field>

        <v-text-field label="Email" :disabled="props.id !== 0" v-model="item.email"
          :rules="[
            (v: string) => !!v || 'Email is required',
            (v: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v) || 'Email must be valid'
          ]"
        ></v-text-field>

      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="saveBtnLabel" :saveDisabled="!auth.isWriter()"
      @cancel="emit('cancel')" @save="saveItem">
      <template #extra>
        <v-spacer />
        <v-btn v-if="props.id !== 0" disabled color="error" @click="deleteItem">{{ $t('actions.delete') }}</v-btn>
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
import { type Employee, NewEmployee, GetEmployeeInputFromItem } from '@/types/supplier'
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
  useFormCrud<Employee>({
    ax: axToUse,
    id: props.id,
    baseUrl: '/a/supplier/employees',
    reqHeaders: props.internal ? undefined : { 'Employee-Email': supStore.selectedEmpEmail },
    newItem: NewEmployee,
    getInput: GetEmployeeInputFromItem,

    onCreate: (id) => { emit('create', id) },
    onDelete: () => { emit('delete') },
    onLoad: (id) => emit('load', id),
    onUpdate: () => { emit('update') },
  })

const formTitle = computed(() => {
  if (props.id === 0) return 'New employee'
  if (!item.value) return ''
  return `${item.value.given_name ?? ''} ${item.value.family_name ?? ''}`.trim()
})

</script>

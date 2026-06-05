<template>
  <v-card-title class="pl-1 mb-1">
    {{ formTitle }}
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">
    <v-row>
      <v-col class="form-col">

        <v-switch label="Active" v-model="item.is_active" color="primary" hide-details></v-switch>
        
        <v-text-field label="Symbol" v-model="item.symbol" clearable
          :rules="[(v: string) => !v || v.length <= 5 || 'Symbol must be 5 characters or less']"
        ></v-text-field>

      </v-col>
    </v-row>

    <l-cancel-and-save-actions v-model:saving="saving" v-model:showSaved="showSaved" saveBtnLabel="Save" :saveDisabled="!auth.isWriter()"
      @cancel="emit('cancel')" @save="saveItem">
    </l-cancel-and-save-actions>

  </v-form>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useFormCrud, useFormPatch } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { useEcbStore } from '@/stores/ecb'
import { type Currency } from '@/types/ecb'

const props = defineProps<{
  id: number
}>()

const emit = defineEmits<{
  (e: 'cancel'): void
  (e: 'load', id: number): void
  (e: 'update'): void
}>()

const ecbStore = useEcbStore()

const { item, loadItem } =
  useFormCrud<Currency>({
    ax,
    id: props.id,
    baseUrl: '/a/ecb/currencies',

    onLoad: (id) => emit('load', id),
  })

const { itemForm, saving, showSaved, saveItem } =
  useFormPatch({
    ax,
    patch_id: () => item.value?.metadata_id, // getter, not eager value
    patchUrl: '/a/ecb/currency-metadata',
    getPatchInput: () => ({
      code: item.value?.code,
      is_active: item.value?.is_active,
      symbol: item.value?.symbol ? item.value?.symbol : '',
    }),
    loadItem,
    onUpdate: () => { ecbStore.loadActiveCurrencies(); emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? item.value?.name : ''
})

</script>

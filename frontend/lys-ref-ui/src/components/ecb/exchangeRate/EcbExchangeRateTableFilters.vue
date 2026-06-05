<template>
  <v-chip-group column>

    <l-filter-chip-date name="Day" v-model="filterDay"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterDay = undefined; emit('update')">
    </l-filter-chip-date>

    <l-filter-chip-select name="To currency" v-model="filterToCurrFk" :items="ecbStore.activeCurrenciesExEur"
      @update:modelValue="emit('update')" @close="filterToCurrFk = undefined; emit('update')">
    </l-filter-chip-select>

  </v-chip-group>
</template>

<script lang="ts" setup>
import {type DateFilter } from 'lys-vue'
import { useEcbStore } from '@/stores/ecb'

const filterDay = defineModel<DateFilter>('filterDay')
const filterToCurrFk = defineModel<number>('filterToCurrFk')

const ecbStore = useEcbStore()

const emit = defineEmits<{
  (e: 'update'): void
  (e: 'updateDebounced'): void
}>()

</script>

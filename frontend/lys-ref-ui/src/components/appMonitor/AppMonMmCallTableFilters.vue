<template>
  <v-chip-group column>

    <l-filter-chip-date name="Created at" v-model="filterCreatedAtDate"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterCreatedAtDate = undefined; emit('update')">
    </l-filter-chip-date>

    <l-filter-chip-text name="Endpoint search" v-model="filterEndpoint"
      @update:modelValue="emit('updateDebounced')" @close="filterEndpoint = undefined; emit('update')">
    </l-filter-chip-text>

    <l-filter-chip-bool name="Status OK" v-model="filterStatusCodeOk"
      @update:modelValue="emit('update')" @close="filterStatusCodeOk = undefined; emit('update')">
    </l-filter-chip-bool>

    <l-filter-chip-numeric name="Duration (ms)" v-model="filterDurationMs"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterDurationMs = undefined; emit('update')">
    </l-filter-chip-numeric>

  </v-chip-group>
</template>

<script lang="ts" setup>
import { type DateFilter, type NumericFilter } from 'lys-vue'

const filterCreatedAtDate = defineModel<DateFilter>('filterCreatedAtDate')
const filterDurationMs = defineModel<NumericFilter>('filterDurationMs')
const filterEndpoint = defineModel<string>('filterEndpoint')
const filterStatusCodeOk = defineModel<boolean>('filterStatusCodeOk')

const emit = defineEmits<{
  (e: 'update'): void
  (e: 'updateDebounced'): void
}>()

</script>

<template>
  <v-chip-group column>

    <l-filter-chip-date name="Created at" v-model="filterCreatedAtDate"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterCreatedAtDate = undefined; emit('update')">
    </l-filter-chip-date>

    <l-filter-chip-text name="Endpoint search" v-model="filterEndpoint"
      @update:modelValue="emit('updateDebounced')" @close="filterEndpoint = undefined; emit('update')">
    </l-filter-chip-text>

    <l-filter-chip-text name="User search" v-model="filterUserName"
      @update:modelValue="emit('updateDebounced')" @close="filterUserName = undefined; emit('update')">
    </l-filter-chip-text>

    <l-filter-chip-text name="Status search" v-model="filterStatus"
      @update:modelValue="emit('updateDebounced')" @close="filterStatus = undefined; emit('update')">
    </l-filter-chip-text>

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
const filterStatus = defineModel<string>('filterStatus')
const filterUserName = defineModel<string>('filterUserName')

const emit = defineEmits<{
  (e: 'update'): void
  (e: 'updateDebounced'): void
}>()

</script>

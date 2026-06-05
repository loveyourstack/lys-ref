<template>
  <v-chip-group column>

    <l-filter-chip-select name="Flow" v-model="filterFlowFk" :items="processStore.flows"
      @update:modelValue="emit('update')" @close="filterFlowFk = undefined; emit('update')">
    </l-filter-chip-select>

  </v-chip-group>
</template>

<script lang="ts" setup>
import { useProcessStore } from '@/stores/process'

const processStore = useProcessStore()

const filterFlowFk = defineModel<number>('filterFlowFk')

const emit = defineEmits<{
  (e: 'update'): void
  (e: 'updateDebounced'): void
}>()

</script>

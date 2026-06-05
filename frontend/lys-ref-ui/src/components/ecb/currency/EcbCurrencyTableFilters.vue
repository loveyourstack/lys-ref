<template>
  <v-chip-group column>

    <l-filter-chip-text name="Code search" v-model="filterCode"
      @update:modelValue="emit('updateDebounced')" @close="filterCode = undefined; emit('update')">
    </l-filter-chip-text>

    <l-filter-chip-bool name="Active" v-model="filterIsActive"
      @update:modelValue="emit('update')" @close="filterIsActive = undefined; emit('update')">
    </l-filter-chip-bool>

  </v-chip-group>
</template>

<script lang="ts" setup>

const filterCode = defineModel<string>('filterCode')
const filterIsActive = defineModel<boolean>('filterIsActive')

const emit = defineEmits<{
  (e: 'update'): void
  (e: 'updateDebounced'): void
}>()

</script>

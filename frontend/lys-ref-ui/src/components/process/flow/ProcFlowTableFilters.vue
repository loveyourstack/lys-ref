<template>
  <v-chip-group column>

    <l-filter-chip-text name="Name search" v-model="filterName"
      @update:modelValue="emit('updateDebounced')" @close="filterName = undefined; emit('update')">
    </l-filter-chip-text>

  </v-chip-group>
</template>

<script lang="ts" setup>

const filterName = defineModel<string>('filterName')

const emit = defineEmits<{
  (e: 'update'): void
  (e: 'updateDebounced'): void
}>()

</script>

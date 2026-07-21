<template>
  <v-chip-group column>

    <l-filter-chip-select name="Country" v-model="filterCountryFk" :items="geoStore.euCountries"
      @update:modelValue="emit('update')" @close="filterCountryFk = undefined; emit('update')">
    </l-filter-chip-select>

  </v-chip-group>
</template>

<script lang="ts" setup>
import { useGeoStore } from '@/stores/geo'

const filterCountryFk = defineModel<number>('filterCountryFk')

const emit = defineEmits<{
  (e: 'update'): void
}>()

const geoStore = useGeoStore()

</script>

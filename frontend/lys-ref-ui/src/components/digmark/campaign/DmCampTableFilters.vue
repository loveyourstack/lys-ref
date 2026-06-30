<template>
  <v-chip-group column>

    <l-filter-chip-text name="Name search" v-model="filterName"
      @update:modelValue="emit('updateDebounced')" @close="filterName = undefined; emit('update')">
    </l-filter-chip-text>

    <l-filter-chip-enum name="Managers" v-model="filterManagers" :items="digmarkStore.managersSelectable" multiple
      @update:modelValue="emit('update')" @close="filterManagers = undefined; emit('update')">
    </l-filter-chip-enum>

    <l-filter-chip-select name="Countries" v-model="filterCountryFKs" :items="geoStore.mandatoryCountries" multiple
      @update:modelValue="emit('update')" @close="filterCountryFKs = undefined; emit('update')">
    </l-filter-chip-select>

    <l-filter-chip-select name="Verticals" v-model="filterVerticalFks" :items="digmarkStore.mandatoryVerticals" multiple
      @update:modelValue="emit('update')" @close="filterVerticalFks = undefined; emit('update')">
    </l-filter-chip-select>

    <l-filter-chip-bool name="Active" v-model="filterIsActive"
      @update:modelValue="emit('update')" @close="filterIsActive = undefined; emit('update')">
    </l-filter-chip-bool>

    <l-filter-chip-numeric name="Daily budget" v-model="filterDailyBudget"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterDailyBudget = undefined; emit('update')">
    </l-filter-chip-numeric>

  </v-chip-group>
</template>

<script lang="ts" setup>
import { type NumericFilter } from 'lys-vue'
import { useDigmarkStore } from '@/stores/digmark'
import { useGeoStore } from '@/stores/geo'

const filterCountryFKs = defineModel<number[]>('filterCountryFKs')
const filterDailyBudget = defineModel<NumericFilter>('filterDailyBudget')
const filterIsActive = defineModel<boolean>('filterIsActive')
const filterManagers = defineModel<string[]>('filterManagers')
const filterName = defineModel<string>('filterName')
const filterVerticalFks = defineModel<number[]>('filterVerticalFks')

const emit = defineEmits<{
  (e: 'update'): void
  (e: 'updateDebounced'): void
}>()

const digmarkStore = useDigmarkStore()
const geoStore = useGeoStore()

</script>

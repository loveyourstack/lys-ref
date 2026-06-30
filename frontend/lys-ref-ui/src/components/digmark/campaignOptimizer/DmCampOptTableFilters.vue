<template>
  <v-chip-group column>

    <l-filter-chip-text name="Name search" v-model="filterName"
      @update:modelValue="emit('updateDebounced')" @close="filterName = undefined; emit('update')">
    </l-filter-chip-text>

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

    <l-filter-chip-numeric name="Impressions" v-model="filterImpressions"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterImpressions = undefined; emit('update')">
    </l-filter-chip-numeric>

    <l-filter-chip-numeric name="Clicks" v-model="filterClicks"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterClicks = undefined; emit('update')">
    </l-filter-chip-numeric>

    <l-filter-chip-numeric name="Conversions" v-model="filterConversions"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterConversions = undefined; emit('update')">
    </l-filter-chip-numeric>

    <l-filter-chip-numeric name="Revenue" v-model="filterRevenue"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterRevenue = undefined; emit('update')">
    </l-filter-chip-numeric>

    <l-filter-chip-numeric name="Spend" v-model="filterSpend"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterSpend = undefined; emit('update')">
    </l-filter-chip-numeric>

    <l-filter-chip-numeric name="Profit" v-model="filterProfit"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterProfit = undefined; emit('update')">
    </l-filter-chip-numeric>

    <l-filter-chip-numeric name="ROI" v-model="filterRoi" isPercent
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterRoi = undefined; emit('update')">
    </l-filter-chip-numeric>

  </v-chip-group>
</template>

<script lang="ts" setup>
import { type NumericFilter } from 'lys-vue'
import { useDigmarkStore } from '@/stores/digmark'
import { useGeoStore } from '@/stores/geo'

const filterClicks = defineModel<NumericFilter>('filterClicks')
const filterConversions = defineModel<NumericFilter>('filterConversions')
const filterCountryFKs = defineModel<number[]>('filterCountryFKs')
const filterDailyBudget = defineModel<NumericFilter>('filterDailyBudget')
const filterImpressions = defineModel<NumericFilter>('filterImpressions')
const filterIsActive = defineModel<boolean>('filterIsActive')
const filterName = defineModel<string>('filterName')
const filterProfit = defineModel<NumericFilter>('filterProfit')
const filterRoi = defineModel<NumericFilter>('filterRoi')
const filterRevenue = defineModel<NumericFilter>('filterRevenue')
const filterSpend = defineModel<NumericFilter>('filterSpend')
const filterVerticalFks = defineModel<number[]>('filterVerticalFks')

const emit = defineEmits<{
  (e: 'update'): void
  (e: 'updateDebounced'): void
}>()

const digmarkStore = useDigmarkStore()
const geoStore = useGeoStore()

</script>

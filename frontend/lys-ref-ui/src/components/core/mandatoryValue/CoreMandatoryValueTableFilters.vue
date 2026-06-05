<template>
  <v-chip-group column>

    <l-filter-chip-bool name="Bool" v-model="filterBool" 
      @update:modelValue="emit('update')" @close="filterBool = undefined; emit('update')">
    </l-filter-chip-bool>

    <l-filter-chip-date name="Date" v-model="filterDateCet"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterDateCet = undefined; emit('update')">
    </l-filter-chip-date>

    <l-filter-chip-enum name="Enum" v-model="filterEnum" :items="coreStore.mandatoryEnums"
      @update:modelValue="emit('update')" @close="filterEnum = undefined; emit('update')">
    </l-filter-chip-enum>

    <l-filter-chip-enum name="Enum (multi)" v-model="filterEnums" :items="coreStore.mandatoryEnums" multiple
      @update:modelValue="emit('update')" @close="filterEnums = undefined; emit('update')">
    </l-filter-chip-enum>

    <l-filter-chip-numeric name="Int" v-model="filterInt"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterInt = undefined; emit('update')">
    </l-filter-chip-numeric>

    <l-filter-chip-numeric name="Numeric" v-model="filterNumeric"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterNumeric = undefined; emit('update')">
    </l-filter-chip-numeric>

    <l-filter-chip-select name="Table" v-model="filterTableFk" :items="geoStore.oceans"
      @update:modelValue="emit('update')" @close="filterTableFk = undefined; emit('update')">
    </l-filter-chip-select>

    <l-filter-chip-select name="Table (multi)" v-model="filterTableFks" :items="geoStore.oceans" multiple
      @update:modelValue="emit('update')" @close="filterTableFks = undefined; emit('update')">
    </l-filter-chip-select>

    <l-filter-chip-text name="Text search" v-model="filterText"
      @update:modelValue="emit('updateDebounced')" @close="filterText = undefined; emit('update')">
    </l-filter-chip-text>

    <l-filter-chip-date name="Timestamp" v-model="filterTimestamp"
      @change="emit('update')" @changeDebounced="emit('updateDebounced')" @close="filterTimestamp = undefined; emit('update')">
    </l-filter-chip-date>

  </v-chip-group>
</template>

<script lang="ts" setup>
import { type DateFilter, type NumericFilter } from 'lys-vue'
import { useCoreStore } from '@/stores/core'
import { useGeoStore } from '@/stores/geo'

const filterBool = defineModel<boolean>('filterBool')
const filterDateCet = defineModel<DateFilter>('filterDateCet')
const filterEnum = defineModel<string>('filterEnum')
const filterEnums = defineModel<string[]>('filterEnums')
const filterInt = defineModel<NumericFilter>('filterInt')
const filterNumeric = defineModel<NumericFilter>('filterNumeric')
const filterTableFk = defineModel<number>('filterTableFk')
const filterTableFks = defineModel<number[]>('filterTableFks')
const filterText = defineModel<string>('filterText')
const filterTimestamp = defineModel<DateFilter>('filterTimestamp')

/* emit changes rather than watching the parent component filter variables in order to avoid triggering multiple data loads
 when parent filter variables are updated from local storage */
const emit = defineEmits<{
  (e: 'update'): void
  (e: 'updateDebounced'): void
}>()

const coreStore = useCoreStore()
const geoStore = useGeoStore()

</script>

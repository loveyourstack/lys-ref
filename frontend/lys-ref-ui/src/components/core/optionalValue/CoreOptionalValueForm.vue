<template>
  <v-card-title class="pl-1 mb-1">
    {{ formTitle }}
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">
    <v-row>
      <v-col class="form-col">
        
        <v-switch label="Bool" v-model="item.c_bool" color="primary" hide-details></v-switch>

        <!-- optional fields: add clearable handling where needed (note that by default it sets value to null) -->

        <v-text-field label="Date" type="date" v-model="item.c_date_cet" clearable
        ></v-text-field>

        <v-autocomplete label="Enum" v-model="item.c_enum" clearable
          :items="coreStore.optionalEnums"
        ></v-autocomplete>

        <!-- define custom clear action if setting value to null is not desired -->
        <v-text-field label="Int" type="number" v-model.number="item.c_int" clearable @click:clear="item.c_int = 0"
          :rules="[(v: number) => Number.isInteger(v) || 'Int must be a whole number']"
        ></v-text-field>

        <v-text-field label="Numeric" type="number" v-model.number="item.c_numeric" clearable
        ></v-text-field>

        <v-autocomplete label="Table" v-model="item.c_table_fk" clearable @click:clear="item.c_table_fk = -1"
          :items="geoStore.countries" item-value="id" item-title="name"
        ></v-autocomplete>

        <v-text-field label="Text" v-model="item.c_text" clearable
        ></v-text-field>

        <l-text-field-time label="Time" v-model="item.c_time"
          clearable @cleared="item.c_time = '00:00'"
        ></l-text-field-time>

      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="saveBtnLabel" :saveDisabled="!auth.isWriter()"
      @cancel="emit('cancel')" @save="saveItem">
      <template #extra>
        <v-spacer />
        <v-btn v-if="props.id !== 0" :disabled="!auth.isWriter()" color="error" @click="deleteItem">Delete</v-btn>
      </template>
    </l-cancel-and-save-actions>

  </v-form>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useFormCrud } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { useCoreStore } from '@/stores/core'
import { useGeoStore } from '@/stores/geo'
import { type OptionalValue, NewOptionalValue, GetOptionalValueInputFromItem } from '@/types/core'

const props = defineProps<{
  id: number
}>()

const emit = defineEmits<{
  (e: 'cancel'): void
  (e: 'create', newID: number): void
  (e: 'delete'): void
  (e: 'load', id: number): void
  (e: 'update'): void
}>()

const coreStore = useCoreStore()
const geoStore = useGeoStore()

const { item, itemForm, saving, saveBtnLabel, showSaved, deleteItem, saveItem } =
  useFormCrud<OptionalValue>({
    ax,
    id: props.id,
    baseUrl: '/a/core/optional-values',
    newItem: NewOptionalValue,
    getInput: GetOptionalValueInputFromItem,

    onCreate: (id) => { emit('create', id) },
    onDelete: () => { emit('delete') },
    onLoad: (id) => { 
      emit('load', id)
      item.value!.c_int = item.value?.c_int ?? 0
      item.value!.c_numeric = item.value?.c_numeric ?? 0
    },
    onUpdate: () => { emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? 'ID ' + props.id : 'New optional value'
})

</script>

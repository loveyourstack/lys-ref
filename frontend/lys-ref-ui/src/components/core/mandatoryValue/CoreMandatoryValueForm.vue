<template>
  <v-card-title class="pl-1 mb-1">
    {{ formTitle }}
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">
    <v-row>
      <v-col class="form-col">
        
        <v-switch label="Bool" v-model="item.c_bool" color="primary" hide-details></v-switch>

        <v-text-field label="Date" type="date" v-model="item.c_date_cet"
          :rules="[(v: string) => !!v || 'Date is required', (v: string) => new Date(v) >= new Date('1900-01-01') || 'Date must be >= 1 Jan 1900']"
        ></v-text-field>

        <v-autocomplete label="Enum" v-model="item.c_enum"
          :items="coreStore.mandatoryEnums"
          :rules="[(v: string) => !!v || 'Enum is required']"
        ></v-autocomplete>

        <v-text-field label="Int" type="number" v-model.number="item.c_int"
          :rules="[(v: number) => (v != undefined) || 'Int is required', (v: number) => Number.isInteger(v) || 'Int must be a whole number']"
        ></v-text-field>

        <v-text-field label="Numeric" type="number" v-model.number="item.c_numeric"
          :rules="[(v: number) => v != undefined || 'Numeric is required']"
        ></v-text-field>

        <v-autocomplete label="Table" v-model="item.c_table_fk"
          :items="geoStore.oceans" item-value="id" item-title="name"
          :rules="[(v: number) => !!v || 'Table is required']"
        ></v-autocomplete>

        <v-text-field label="Text" v-model="item.c_text"
          :rules="[(v: string) => !!v || 'Text is required']"
        ></v-text-field>

        <l-text-field-time label="Time" v-model="item.c_time"
          :rules="[(v: string) => !!v || 'Time is required']"
        ></l-text-field-time>

      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="$t(`actions.${saveBtnLabel.toLowerCase()}`)" :saveDisabled="!auth.isWriter()"
      @cancel="emit('cancel')" @save="saveItem">
      <template #extra>
        <v-spacer />
        <v-btn v-if="props.id !== 0" :disabled="!auth.isWriter()" color="error" @click="deleteItem">{{ $t('actions.delete') }}</v-btn>
      </template>
    </l-cancel-and-save-actions>

  </v-form>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useFormCrud } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { useCoreStore } from '@/stores/core'
import { useGeoStore } from '@/stores/geo'
import { type MandatoryValue, NewMandatoryValue, GetMandatoryValueInputFromItem } from '@/types/core'

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

const { t } = useI18n()
const coreStore = useCoreStore()
const geoStore = useGeoStore()

const { item, itemForm, saving, saveBtnLabel, showSaved, deleteItem, saveItem } =
  useFormCrud<MandatoryValue>({
    ax,
    id: props.id,
    baseUrl: '/a/core/mandatory-values',
    newItem: NewMandatoryValue,
    getInput: GetMandatoryValueInputFromItem,

    onCreate: (id: number) => { emit('create', id) },
    onDelete: () => { emit('delete') },

    // onLoad: set default values for int and numeric since they are required but might be missing in JSON response due to omitempty in backend
    onLoad: (id: number) => { 
      emit('load', id)
      item.value!.c_int = item.value?.c_int ?? 0
      item.value!.c_numeric = item.value?.c_numeric ?? 0
    },
    onUpdate: () => { emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? 'ID ' + props.id : t('type_handling.mandatory_values.new_item')
})

</script>

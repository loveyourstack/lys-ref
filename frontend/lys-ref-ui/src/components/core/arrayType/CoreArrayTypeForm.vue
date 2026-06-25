<template>
  <v-card-title class="pl-1 mb-1">
    {{ formTitle }}
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">
    <v-row>
      <v-col class="form-col">

        <l-array-bool label="Bool" v-model="item.c_bool" :rows="3" />

        <l-array-date label="Date" v-model="item.c_date" :rows="3" />

        <v-autocomplete label="Enum" v-model="item.c_enum" multiple chips clearable
          :items="coreStore.mandatoryEnums"
          :rules="[(v: string) => !!v || 'Enum is required']"
        ></v-autocomplete>

      </v-col>

      <v-col class="form-col">

        <l-array-int label="Int" v-model="item.c_int" :rows="3" />

        <l-array-numeric label="Numeric" v-model="item.c_numeric" :rows="3" />

        <l-array-text label="Text" v-model="item.c_text" :rows="3" />

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
import { type ArrayType, NewArrayType, GetArrayTypeInputFromItem } from '@/types/core'

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

const { item, itemForm, saving, saveBtnLabel, showSaved, deleteItem, saveItem } =
  useFormCrud<ArrayType>({
    ax,
    id: props.id,
    baseUrl: '/a/core/array-types',
    newItem: NewArrayType,
    getInput: GetArrayTypeInputFromItem,

    onCreate: (id) => { emit('create', id) },
    onDelete: () => { emit('delete') },
    onLoad: (id) => emit('load', id),
    onUpdate: () => { emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? 'ID ' + props.id : t('type_handling.arrays.new_item')
})

</script>

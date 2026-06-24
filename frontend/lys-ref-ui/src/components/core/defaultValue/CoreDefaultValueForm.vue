<template>
  <v-card-title class="pl-1 mb-1">
    {{ formTitle }}
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">
    <v-row>
      <v-col class="form-col">
        
        <!-- disable control and rules on insert, but enable on update. API will apply the default value on insert -->
        <v-text-field :disabled="props.id == 0" label="Default Text" v-model="item.c_default_text"
          :rules="props.id != 0 ? [(v: string) => !!v || 'Default text is required'] : []"
        ></v-text-field>

        <v-text-field label="Suggested Text" v-model="item.c_suggested_text"
          :rules="[(v: string) => !!v || 'Suggested text is required']"
        ></v-text-field>

      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="saveBtnLabel" :saveDisabled="!auth.isWriter()"
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
import { useFormCrud } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { type DefaultValue, NewDefaultValue, GetDefaultValueInputFromItem } from '@/types/core'

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

const { item, itemForm, saving, saveBtnLabel, showSaved, deleteItem, saveItem } =
  useFormCrud<DefaultValue>({
    ax,
    id: props.id,
    baseUrl: '/a/core/default-values',
    newItem: NewDefaultValue,
    getInput: GetDefaultValueInputFromItem,

    onCreate: (id) => { emit('create', id) },
    onDelete: () => { emit('delete') },
    onLoad: (id) => emit('load', id),
    onUpdate: () => { emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? 'ID ' + props.id : 'New default value'
})

</script>

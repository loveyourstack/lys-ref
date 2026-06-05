<template>
  <v-card-title class="pl-1 mb-1">
    {{ formTitle }}
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">

    <v-row>
      <v-col class="form-col">

        <v-text-field label="Name" v-model="item.name"
          :rules="[(v: string) => !!v || 'Name is required']"
        ></v-text-field>

        <l-array-text label="Params" v-model="item.params" :rows="3" placeholder="key=value"
          hint="Each param is a key=value pair. Enter each on a new line." />
        
      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="saveBtnLabel" :saveDisabled="!auth.isWriter()"
      @cancel="emit('cancel')" @save="saveItem">
      <template #extra>
        <v-spacer />
        <v-btn v-if="props.id !== 0" :disabled="!auth.isWriter() || item.run_count > 0" color="error" @click="deleteItem">Delete</v-btn>
      </template>
    </l-cancel-and-save-actions>

  </v-form>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useFormCrud } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { useProcessStore } from '@/stores/process'
import { type Flow, NewFlow, GetFlowInputFromItem } from '@/types/process'

const props = defineProps<{
  id: number
}>()

const emit = defineEmits<{
  (e: 'cancel'): void
  (e: 'create', newId: number): void
  (e: 'delete'): void
  (e: 'load', id: number): void
  (e: 'update'): void
}>()

const processStore = useProcessStore()

const { item, itemForm, saving, saveBtnLabel, showSaved, deleteItem, saveItem } =
  useFormCrud<Flow>({
    ax,
    id: props.id,
    baseUrl: '/a/process/flows',
    newItem: NewFlow,
    getInput: GetFlowInputFromItem,

    onCreate: (id) => { processStore.loadFlows(); emit('create', id) },
    onDelete: () => { processStore.loadFlows(); emit('delete') },
    onLoad: (id) => emit('load', id),
    onUpdate: () => { processStore.loadFlows(); emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? (item.value?.name ?? '') : 'New flow'
})

</script>

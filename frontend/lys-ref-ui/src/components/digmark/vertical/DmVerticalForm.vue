<template>
  <v-card-title class="pl-1 mb-1">
    {{ formTitle }}
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">
    <v-row>
      <v-col class="form-col">

        <v-text-field label="Name" v-model="item!.name"
          :rules="[(v: string) => !!v || 'Name is required']"
        ></v-text-field>

      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="saveBtnLabel" :saveDisabled="!auth.isWriter()"
      @cancel="emit('cancel')" @save="saveItem">
      <template #extra>
        <v-spacer />
        <v-btn v-if="props.id !== 0" :disabled="!auth.isWriter() || item!.campaign_count > 0" color="error" @click="deleteItem">Delete</v-btn>
      </template>
    </l-cancel-and-save-actions>

  </v-form>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useFormCrud } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { useDigmarkStore } from '@/stores/digmark'
import { type Vertical, NewVertical, GetVerticalInputFromItem } from '@/types/digmark'

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

const digmarkStore = useDigmarkStore()

const { item, itemForm, saving, saveBtnLabel, showSaved, deleteItem, saveItem } =
  useFormCrud<Vertical>({
    ax,
    id: props.id,
    baseUrl: '/a/digmark/verticals',
    newItem: NewVertical,
    getInput: GetVerticalInputFromItem,

    onCreate: (id) => { digmarkStore.loadVerticals(); emit('create', id) },
    onDelete: () => { digmarkStore.loadVerticals(); emit('delete') },
    onLoad: (id) => emit('load', id),
    onUpdate: () => { digmarkStore.loadVerticals(); emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? (item.value?.name ?? '') : 'New vertical'
})

</script>

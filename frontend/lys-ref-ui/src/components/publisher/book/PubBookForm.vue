<template>
  <v-card-title class="pl-1 mb-1">
    {{ formTitle }}
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">
    <v-row>
      <v-col class="form-col">

        <v-autocomplete label="Author" v-model="item.author_fk"
          :items="pubStore.authors" item-value="id" item-title="name"
          :rules="[(v: number) => !!v || 'Author is required']"
        ></v-autocomplete>

        <v-text-field label="Name" v-model="item.name"
          :rules="[(v: string) => !!v || 'Name is required']"
        ></v-text-field>

      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="saveBtnLabel" :saveDisabled="!auth.isWriter()"
      @cancel="emit('cancel')" @save="saveItem">
      <template #extra>
        <v-spacer />
        <v-btn v-if="props.id !== 0" :disabled="!auth.isWriter()" color="error" @click="archiveItem">{{ $t('actions.archive') }}</v-btn>
      </template>
    </l-cancel-and-save-actions>

  </v-form>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { callDelete, useFormCrud } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { type Book, NewBook, GetBookInputFromItem } from '@/types/publisher'
import { usePublisherStore } from '@/stores/publisher'

const props = defineProps<{
  id: number
}>()

const emit = defineEmits<{
  (e: 'archive'): void
  (e: 'cancel'): void
  (e: 'create', newID: number): void
  (e: 'load', id: number): void
  (e: 'update'): void
}>()

const pubStore = usePublisherStore()

const { item, itemUrl, itemForm, saving, saveBtnLabel, showSaved, saveItem } =
  useFormCrud<Book>({
    ax,
    id: props.id,
    baseUrl: '/a/publisher/books',
    newItem: NewBook,
    getInput: GetBookInputFromItem,

    onCreate: (id) => { emit('create', id) },
    onLoad: (id) => emit('load', id),
    onUpdate: () => { emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? (item.value?.name ?? '') : 'New book'
})

function archiveItem() {
  callDelete({ ax, myUrl: itemUrl + '/archive', onSuccess: () => { emit('archive') } })
}

</script>

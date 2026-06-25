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

      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="$t(`actions.${saveBtnLabel.toLowerCase()}`)" :saveDisabled="!auth.isWriter()"
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
import { useI18n } from 'vue-i18n'
import { callDelete, useFormCrud } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { type Author, NewAuthor, GetAuthorInputFromItem } from '@/types/publisher'
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

const { t } = useI18n()
const pubStore = usePublisherStore()

const { item, itemUrl, itemForm, saving, saveBtnLabel, showSaved, saveItem } =
  useFormCrud<Author>({
    ax,
    id: props.id,
    baseUrl: '/a/publisher/authors',
    newItem: NewAuthor,
    getInput: GetAuthorInputFromItem,

    onCreate: (id) => { pubStore.loadAuthors(); emit('create', id) },
    onLoad: (id) => emit('load', id),
    onUpdate: () => { pubStore.loadAuthors(); emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? (item.value?.name ?? '') : t('user_data_retention.authors.new_item')
})

function archiveItem() {
  callDelete({ ax, myUrl: itemUrl + '/archive', onSuccess: () => { pubStore.loadAuthors(); emit('archive') } })
}

</script>

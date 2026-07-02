<template>
  <fieldset class="pl-4 pr-4 pb-4 fs-std mb-2">
    <legend class="pl-2 pr-2">{{ launchIds.length }} {{ $t('launchers.selected_launchers') }}</legend>

      <div class="d-flex align-center ga-6 pt-2">
        <v-btn color="primary" disabled>{{ $t('actions.queue') }}</v-btn>
        <v-btn color="secondary" disabled>{{ $t('actions.cancel') }}</v-btn>

        <v-spacer></v-spacer>

        <v-btn :disabled="!auth.isWriter" :loading="deleting" color="error" @click="deleteByIds">{{ $t('actions.delete') }}</v-btn>
      </div>

  </fieldset>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { notify } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  launchIds: number[]
  baseUrl: string
  partner: string
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()

const appStore = useAppStore()

const deleting = ref(false)

function deleteByIds() {
  deleting.value = true

  ax.post(props.baseUrl + '/delete-many', props.launchIds)
    .then((resp) => {
      const numDeleted = resp.data.data
      notify(appStore.company, numDeleted + ' ' + props.partner + ' launcher(s) deleted', appStore.logoUrl)
    })
    .catch() // handled by interceptor
    .finally(() => {
      deleting.value = false
      emit('update')
    })
}

</script>

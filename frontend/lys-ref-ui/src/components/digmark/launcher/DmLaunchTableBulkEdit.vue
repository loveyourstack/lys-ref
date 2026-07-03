<template>
  <fieldset class="pl-4 pr-4 pb-4 fs-std mb-2">
    <legend class="pl-2 pr-2">{{ launchIds.length }} {{ $t('launchers.selected_launchers') }}</legend>

      <div class="d-flex align-center ga-6 pt-2">
        <v-btn :disabled="!auth.isWriter" :loading="queueing" color="primary" @click="queueByIds">{{ $t('actions.queue') }}</v-btn>
        <v-btn :disabled="!auth.isWriter" :loading="cancelling" color="grey" @click="cancelByIds">{{ $t('actions.cancel') }}</v-btn>

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

const cancelling = ref(false)
const deleting = ref(false)
const queueing = ref(false)

function cancelByIds() {
  cancelling.value = true

  ax.post(props.baseUrl + '/cancel-many', props.launchIds)
    .then((resp) => {
      const numCancelled = resp.data.data
      notify(appStore.company, numCancelled + ' ' + props.partner + ' launcher(s) cancelled', appStore.logoUrl)
    })
    .catch() // handled by interceptor
    .finally(() => {
      cancelling.value = false
      emit('update')
    })
}

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

function queueByIds() {
  queueing.value = true

  ax.post(props.baseUrl + '/queue-many', props.launchIds)
    .then((resp) => {
      const numQueued = resp.data.data
      notify(appStore.company, numQueued + ' ' + props.partner + ' launcher(s) queued', appStore.logoUrl)
    })
    .catch() // handled by interceptor
    .finally(() => {
      queueing.value = false
      emit('update')
    })
}

</script>

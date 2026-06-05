<template>
  <v-dialog v-model="showDialog" width="auto">
    <v-card>
      <v-card-title>Error message</v-card-title>
      <v-card-text>
        <div>{{ errMsg }}</div>
      </v-card-text>
      <v-card-actions>
        <v-btn color="primary" @click="showDialog = false">Close</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <v-row density="compact">
    <v-col class="d-flex align-center">
      <v-breadcrumbs :items="[
          { title: 'Runs', disabled: false },
          { title: props.flow, disabled: false },
          { title: props.step_name, disabled: false }
        ]" 
        density="compact"
      ></v-breadcrumbs>

      <v-spacer />

      <v-btn icon flat v-tooltip="'Refresh'" @click="loadItems()">
        <v-icon icon="mdi-refresh"></v-icon>
      </v-btn>
    </v-col>
  </v-row>

  <v-row density="comfortable">
    <v-col class="mb-6">
      <div class="dt-subtitle">Each card below shows the result of the execution of a process step.</div>
      <div class="dt-subtitle">The cards are updated automatically every 2 seconds as long as there is at least one process still running.</div>
      <div class="dt-subtitle">Each fake running process has a 5% chance per second of failing.</div>
    </v-col>
  </v-row>

  <v-row class="ma-2 ga-16">
    <v-col v-for="mCol in m" class="d-flex flex-column justify-space-evenly ga-6">
      <v-card density="compact" v-for="item in mCol" :variant="current.dark ? 'outlined' : undefined">

        <v-card-title class="step-card-title">
          {{ item.step_name }}
          <span class="float-right" :class="getStatusClass(item.status)">{{ item.status }}</span>
        </v-card-title>
        
        <v-card-subtitle class="pb-1">{{ item.cmd }}</v-card-subtitle>

        <v-card-text v-if="new Date(item.started_at) > new Date(2000,1,1)" class="pt-2 pb-2">
          <span>{{ useDateFormat(item.started_at, 'HH:mm:ss').value }}</span>
          <span v-if="new Date(item.finished_at) > new Date(2000,1,1)">
            <span> - {{ useDateFormat(item.finished_at, 'HH:mm:ss').value }}</span>
            <span class="ml-4">
              {{ (new Date(item.finished_at).getTime() - new Date(item.started_at).getTime()) / 1000 + 's' }}
            </span>
          </span>

          <span v-if="item.status === 'Error'" class="float-right">
            <v-icon color="primary" icon="mdi-information-outline" @click="errMsg = item.err_msg; showDialog = true"></v-icon>
          </span>
        </v-card-text>

      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useDateFormat } from '@vueuse/core'
import { useTheme } from 'vuetify'
import { fetchOnce } from 'lys-vue'
import ax from '@/api'
import { type Point } from '@/types/process'

const props = defineProps<{
  flow: string
  run_id: number
  step_name: string
}>()

const { current } = useTheme()

const items = ref<Point[]>([])
const m = ref<Point[][]>([])

const showDialog = ref(false)
const errMsg = ref('')

const baseUrl = computed(() => {
  return '/a/process/points?run_fk=' + props.run_id + '&xsort=display_order'
})

const isRunning = computed(() => items.value.some(item => item.status === 'Running'))
let pollInterval: ReturnType<typeof setInterval> | null = null

function getStatusClass(status: string) {
  switch (status) {
    case 'Cancelled': return 'text-grey'
    case 'Completed': return 'text-success'
    case 'Error': return 'text-error'
    case 'Interrupted': return 'text-grey'
    case 'Running': return 'text-amber'
    case 'Waiting': return 'text-grey'
    default: return ''
  }
}

function loadItems() {
  m.value = []
  fetchOnce({ ax, myUrl: baseUrl.value, result: items, onSuccess: () => {

    let assignedIds: number[] = []
    let iter: number = 0

    const includesAll = (arr: number[], values: number[]) => values.every(v => arr.includes(v))

    // display a card per part
    do {
      iter++
      //console.log('iter', iter)

      if (iter > 1000) {
        console.log('circular dependency: too many iterations')
        return
      }

      let a: number[] = [] // assigned ids this iteration
      let mCol: Point[] = []

      items.value.forEach(item => {

        // skip items already assigned
        if (assignedIds.includes(item.id)) {
          return
        }

        // assign items with no deps
        if (!item.depends_on) {
          mCol.push(item)
          a.push(item.id)
          //console.log('assigned no dep', item.id)
          return
        }

        // item has deps: assign if all deps are included
        if (includesAll(assignedIds, item.depends_on)) {
          mCol.push(item)
          a.push(item.id)
          //console.log('assigned with dep', item.id)
        }
      })

      assignedIds = assignedIds.concat(a) // merges a into assignedIds
      m.value.push(mCol)
    }
    while (assignedIds.length < items.value.length)
  }})
}

function startPolling() {
  if (pollInterval) return

  // as long as there is at least 1 running process, set a timer to refresh data every 2 seconds
  pollInterval = setInterval(() => {
    // guard in case state changed between ticks
    if (!isRunning.value) {
      stopPolling()
      return
    }
    loadItems()
  }, 2000)
}

function stopPolling() {
  if (!pollInterval) return
  clearInterval(pollInterval)
  pollInterval = null
}

watch(isRunning, (isRunning) => {
  if (isRunning) startPolling()
  else stopPolling()
})

onMounted(() => {
  loadItems()
})

onUnmounted(() => {
  stopPolling()
})

</script>
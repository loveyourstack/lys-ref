<template>
  <l-dialog-card v-model="showEdit">
    <proc-step-form :id="editID" :flow_id="props.flow_id"
      @cancel="showEdit = false"
      @create="showEdit = false; loadItems()"
      @delete="showEdit = false; loadItems()"
      @update="showEdit = false; loadItems()"
    ></proc-step-form>
  </l-dialog-card>

  <l-dialog-card v-model="showLinkEdit">
    <proc-step-link-add-form :step_id="stepID"
      @cancel="showLinkEdit = false"
      @create="showLinkEdit = false; loadItems()"
    ></proc-step-link-add-form>
  </l-dialog-card>

  <v-row density="compact">
    <v-col class="d-flex align-center">
      <v-breadcrumbs density="compact" :items="[
          { title: $t('parallel_processing.flows.title'), disabled: false },
          { title: props.flow_name, disabled: false },
          { title: $t('parallel_processing.flows.steps.title'), disabled: false }
        ]" 
      ></v-breadcrumbs>

      <v-spacer />

      <v-btn color="secondary" @click="editID = 0; showEdit = true">{{ $t('parallel_processing.flows.steps.add_item') }}</v-btn>
    </v-col>
  </v-row>

  <v-row density="comfortable">
    <v-col class="mb-6">
      <div class="dt-subtitle">{{ $t('parallel_processing.flows.steps.p1') }}</div>
      <div class="dt-subtitle">{{ $t('parallel_processing.flows.steps.p2') }}</div>
      <div class="dt-subtitle">{{ $t('parallel_processing.flows.steps.p3') }}
        <ul class="mt-1 mb-1">
          <li>{{ $t('parallel_processing.flows.steps.p3_list.item_1') }}</li>
          <li>{{ $t('parallel_processing.flows.steps.p3_list.item_2') }}</li>
        </ul>
      </div>
    </v-col>
  </v-row>

  <v-row density="compact">
    <v-col class="d-flex align-center">
      <v-text-field label="Params" v-model="paramString"
        prepend-inner-icon="mdi-restore" @click:prepend-inner="paramString = props.default_params"
        :hint="'Default: ' + props.default_params" persistent-hint
      ></v-text-field>

      <v-spacer />

      <v-switch :label="$t('parallel_processing.flows.steps.cancel_on_error')" v-model="stopOnErr" color="primary"></v-switch>
    </v-col>
  </v-row>

  <v-row class="mt-6 mr-4 ma-2 ga-16">
    <v-col v-for="mCol in m" class="d-flex flex-column justify-space-evenly ga-6">

      <v-card density="compact" v-for="(item, idx) in mCol" :variant="current.dark ? 'outlined' : undefined">
        <v-card-title class="step-card-title">
          {{ item.name }}
        </v-card-title>
        <v-card-subtitle class="pb-2">{{ item.cmd }}</v-card-subtitle>

        <v-card-text v-if="item.depends_on_names" class="pt-1 pb-0 opacity-60">
          <v-list lines="one" density="compact" class="pt-0 pb-0">
            <v-list-item v-for="name in item.depends_on_names" prepend-icon="mdi-arrow-left-bottom-bold" class="pl-0">
              {{ name.split(' | ')[0] }} <!-- name is in form: <name> | <link id> -->
              <v-btn :disabled="!auth.isWriter()" icon flat v-tooltip="'Delete dependency'" size="x-small" class="bg-transparent" @click="deleteLink(Number(name.split(' | ')[1]))">
                <v-icon color="error" icon="mdi-delete"></v-icon>
              </v-btn>
            </v-list-item>
          </v-list>
        </v-card-text>

        <v-card-actions class="pt-0">
          <v-btn icon flat v-tooltip:bottom="`${$t('actions.edit')}`" size="small" @click="editID = item.id; showEdit = true">
            <v-icon color="secondary" icon="mdi-square-edit-outline"></v-icon>
          </v-btn>
          <v-btn icon flat v-tooltip:bottom="`${$t('parallel_processing.flows.steps.links.add_item')}`" size="small" @click="stepID = item.id; showLinkEdit = true">
            <v-icon color="secondary" icon="mdi-arrow-left-bottom-bold"></v-icon>
          </v-btn>
          <v-btn :disabled="!auth.isWriter() || !mCol[idx-1]" icon flat v-tooltip:bottom="`${$t('actions.move_up')}`" size="small" 
            @click="swapDisplayOrder(item.id, mCol[idx-1]!.id)">
            <v-icon color="secondary" icon="mdi-arrow-up"></v-icon>
          </v-btn>
          <v-btn :disabled="!auth.isWriter() || !mCol[idx+1]" icon flat v-tooltip:bottom="`${$t('actions.move_down')}`" size="small" 
            @click="swapDisplayOrder(item.id, mCol[idx+1]!.id)">
            <v-icon color="secondary" icon="mdi-arrow-down"></v-icon>
          </v-btn>

          <v-spacer></v-spacer>

          <v-btn :disabled="!auth.isWriter()" icon flat v-tooltip:bottom="`${$t('actions.run')}`" size="small" @click="runStep(item.id, false)">
            <v-icon color="primary" icon="mdi-play"></v-icon>
          </v-btn>
          <v-btn v-if="item.depends_on" :disabled="!auth.isWriter()" icon flat size="small" v-tooltip:bottom="`${$t('actions.run_with_dependencies')}`" @click="runStep(item.id, true)">
            <v-icon color="primary" icon="mdi-sitemap"></v-icon>
          </v-btn>
        </v-card-actions>
      </v-card>

    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useTheme } from 'vuetify'
import { callDelete, fetchOnce, useJsonLs } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { type Step } from '@/types/process'

const props = defineProps<{
  default_params: string
  flow_id: number
  flow_name: string
}>()

const { current } = useTheme()
const router = useRouter()

const baseUrl = '/a/process/steps'
const loadUrl = baseUrl + '?flow_fk=' + props.flow_id + '&xsort=display_order'
const linkUrl = '/a/process/step-links'
const items = ref<Step[]>([])
const m = ref<Step[][]>([])

const editID = ref(0)
const showEdit = ref(false)

const stepID = ref(0)
const showLinkEdit = ref(false)

const paramString = ref<string>('')
const stopOnErr = ref<boolean>(false)

useJsonLs({
  lsKey: 'proc_step_diag',
  refs: {
    stopOnErr,
  },
})

function deleteLink(linkId: number) {
  callDelete({ ax, myUrl: linkUrl + '/' + linkId, onSuccess: () => loadItems() })
}

function loadItems() {
  m.value = []
  fetchOnce({ ax, myUrl: loadUrl, result: items, onSuccess: () => {

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
      let mCol: Step[] = []

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

function runStep(stepId: number, withDeps: boolean) {
  const myUrl = baseUrl + '/' + stepId + '/run'

  const input = {
    'param_string': paramString.value,
    'stop_on_error': stopOnErr.value,
    'with_dependencies': withDeps,
  }

  ax.post(myUrl, JSON.stringify(input))
    .then((resp) => {
      // navigate to the new run's points page
      router.push({name: 'Points', params: { id: resp.data.data }})
    })
    .catch() // handled by interceptor
}

function swapDisplayOrder(id1: number, id2: number) {
  const myUrl = baseUrl + '/swap-display-order'
  const saveItem = {
    'step_id1': id1,
    'step_id2': id2,
  }

  ax.put(myUrl, saveItem)
    .then(() => {
      loadItems()
    })
    .catch() // handled by interceptor
}

onMounted(() => {
  paramString.value = props.default_params
  loadItems()
})

</script>
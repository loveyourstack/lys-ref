<template>
  <v-card-title class="pl-1 mb-1">
    New dependency
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">

    <v-row>
      <v-col class="form-col">

        <v-autocomplete label="Depends on" v-model="item.depends_on_fk" :items="stepItems" item-title="name" item-value="id"
          :rules="[(v: number) => !!v || 'Depends on is required']"
        ></v-autocomplete>

      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="saveBtnLabel" :saveDisabled="!auth.isWriter()"
      @cancel="emit('cancel')" @save="saveItem">
    </l-cancel-and-save-actions>

  </v-form>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { useFormCrud } from 'lys-vue'
import { fetchOnce } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { type Step, type StepLink, NewStepLink, GetStepLinkInputFromItem } from '@/types/process'

const props = defineProps<{
  step_id: number
}>()

const emit = defineEmits<{
  (e: 'cancel'): void
  (e: 'create', newId: number): void
}>()

const { item, itemForm, saving, saveBtnLabel, showSaved, saveItem } =
  useFormCrud<StepLink>({
    ax,
    id: 0,
    baseUrl: '/a/process/step-links',
    newItem: () => NewStepLink(props.step_id), // step_id is mandatory and not changeable by user, so set it here
    getInput: GetStepLinkInputFromItem,

    onCreate: (id) => { emit('create', id) },
  })

const stepItems = ref<Step[]>([])
const stepUrl = '/a/process/steps/' + props.step_id + '/available-dependencies'

function loadStepParts() {
  fetchOnce({ ax, myUrl: stepUrl, result: stepItems })
}

onMounted(() => {
  loadStepParts()
})
</script>

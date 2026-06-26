<template>
  <fieldset class="pl-4 pr-4 pb-4 fs-std mb-2">
    <legend class="pl-2 pr-2">{{ campaignIds.length }} selected campaign(s)</legend>

      <v-tabs v-model="selectedTab">
        <v-tab value="active">Active</v-tab>
        <v-tab value="budgetPercent">Budget - Δ %</v-tab>
      </v-tabs>

      <v-window v-model="selectedTab" class="mt-5 mb-2">

        <v-window-item value="active">
          <v-form ref="activeForm">
            <v-row>
              <v-col class="d-flex">
                <v-autocomplete label="Active" v-model="newActive" :items="BooleanOptions" density="comfortable" max-width="300" hide-details
                  :rules="[(v: string) => v != undefined || 'Active is required']"
                ></v-autocomplete>
                <v-btn color="secondary" class="ml-5 mt-3" :loading="patchingActive" :disabled="!auth.isWriter()" @click="patchActiveByIds">{{ $t('actions.save') }}</v-btn>
              </v-col>
            </v-row>
          </v-form>
        </v-window-item>

        <v-window-item value="budgetPercent">
          <v-form ref="budgetPercentForm">
            <v-row>
              <v-col class="d-flex">
                <v-text-field :label="$t('optimizer.budget_percent_change')" v-model.number="budgetChangeByPercent" type="number" 
                  density="comfortable" max-width="300" hide-details
                  :class="getTextClass(budgetChangeByPercent)"
                  :rules="[(v: string) => !!v || `${$t('optimizer.budget_percent_change')} is required`]">
                </v-text-field>
                <v-btn color="secondary" class="ml-5 mt-3" :loading="patchingBudgetPercent" :disabled="!auth.isWriter()" @click="patchBudgetPercentByIds">{{ $t('actions.save') }}</v-btn>
              </v-col>
            </v-row>
          </v-form>
        </v-window-item>

      </v-window>
  </fieldset>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { BooleanOptions, notify, useJsonLs } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  campaignIds: number[]
  patchUrl: string
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()

const appStore = useAppStore()

const selectedTab = ref('active')

const newActive = ref<string>()
const budgetChangeByPercent = ref<number>()

const activeForm = ref()
const budgetPercentForm = ref()

const patchingActive = ref(false)
const patchingBudgetPercent = ref(false)

useJsonLs({
  lsKey: 'campaign_opt_bulk_edit',
  refs: {
    selectedTab,
  },
})

async function patchActiveByIds() {
  const {valid} = await activeForm.value?.validate()
  if (!valid) {
    return
  }

  const patchInput = {
    ids: props.campaignIds,
    new_active: newActive.value,
  }

  patchingActive.value = true

  ax.patch(props.patchUrl + '/active-by-ids', patchInput)
    .then((resp) => {
      notify(appStore.company, 'Set active', appStore.logoUrl, resp.data.data)
    })
    .catch() // handled by interceptor
    .finally(() => {
      patchingActive.value = false
      emit('update')
    })
}

async function patchBudgetPercentByIds() {
  const {valid} = await budgetPercentForm.value?.validate()
  if (!valid) {
    return
  }

  const patchInput = {
    ids: props.campaignIds,
    budget_percent_change: budgetChangeByPercent.value,
  }

  patchingBudgetPercent.value = true

  ax.patch(props.patchUrl + '/budget-percent-by-ids', patchInput)
    .then((resp) => {
      notify(appStore.company, 'Change budget percent', appStore.logoUrl, resp.data.data)
    })
    .catch() // handled by interceptor
    .finally(() => {
      patchingBudgetPercent.value = false
      emit('update')
    })
}

function getTextClass(val: number | undefined): string {
  if (!val) { return '' }
  if (val > 0) { return 'text-success' }
  return 'text-error'
}

</script>

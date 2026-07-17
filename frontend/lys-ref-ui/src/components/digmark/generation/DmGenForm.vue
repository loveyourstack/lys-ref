<template>
  <v-form ref="genForm" class="ml-4 mt-2" @submit.prevent="generate">
    <div class="d-flex ga-4 align-center">
      <v-text-field label="Product or service" :disabled="showEditForm" v-model="product" clearable
        placeholder="Trip to London, Car insurance" max-width="400"
        :rules="[(v: string) => !!v || 'Product or service is required',
          (v: string) => v.length <= 100 || 'Product or service must be 100 characters or less']"
      ></v-text-field>
      <v-text-field label="Text model" v-model="textModel" max-width="250" disabled
        :rules="[(v: string) => !!v || 'Text model is required']"
      ></v-text-field>
      <v-btn type="submit" class="mb-2" color="primary" :disabled="!auth.isWriter()" :loading="generating">{{ $t('actions.generate') }}</v-btn>
    </div>
  </v-form>

  <v-card variant="flat" max-width="650" class="mt-2">
    <v-card-title>{{ $t('generation.generated_campaign') }}</v-card-title>
    <v-card-text v-if="showEditForm">
      <v-form ref="editForm" class="ml-4 mt-2">
        <v-text-field label="Headline" v-model="headline"
          :rules="[(v: string) => !!v || 'Headline is required']"
        ></v-text-field>
        <v-textarea label="Body" v-model="body"
          :rules="[(v: string) => !!v || 'Body is required']"
        </v-textarea>
        <v-text-field label="Call to action" v-model="callToAction"
          :rules="[(v: string) => !!v || 'Call to action is required']"
        ></v-text-field>
      </v-form>
      <div class="ml-4 mt-2 d-flex ga-4 align-center">
        <v-btn class="mb-2" color="primary" :disabled="!auth.isWriter()" :loading="saving" @click="save">Save</v-btn>
        <v-btn class="mb-2" @click="clear">Clear</v-btn>
      </div>
    </v-card-text>
    <v-card-text v-else>
      <div class="mt-2 dt-subtitle">{{ $t('generation.generated_campaign_placeholder') }}</div>
    </v-card-text>
  </v-card>

</template>

<script lang="ts" setup>
import { ref, computed } from 'vue'
import { type VForm } from 'vuetify/components'
import ax from '@/api'
import auth from '@/auth'
import { type GeneratedCampaignInput } from '@/types/digmark'

const emit = defineEmits<{
  (e: 'generated'): void
}>()

const genForm = ref<InstanceType<typeof VForm>>()
const product = ref('')
const textModel = ref('gemini-3.1-flash-lite')
const genUrl = '/a/gemini/generate-marketing-campaign'
const generating = ref(false)

const editForm = ref<InstanceType<typeof VForm>>()
const headline = ref('')
const body = ref('')
const callToAction = ref('')
const showEditForm = computed(() => !!headline.value || !!body.value || !!callToAction.value)
const saveUrl = '/a/digmark/generated-campaigns'
const saving = ref(false)

function clear() {
  headline.value = ''
  body.value = ''
  callToAction.value = ''
}

async function generate() {
  const result = await genForm.value?.validate()
  if (!result?.valid) { return }

  generating.value = true
  ax.post(genUrl, {'product': product.value, 'model': textModel.value})
    .then(resp => {
      headline.value = resp.data.data.headline ?? ''
      body.value = resp.data.data.body ?? ''
      callToAction.value = resp.data.data.call_to_action ?? ''
    })
    .catch() // handled by interceptor
    .finally(() => generating.value = false)
}

async function save() {
  const result = await editForm.value?.validate()
  if (!result?.valid) { return }

  const input: GeneratedCampaignInput = {
    body: body.value,
    call_to_action: callToAction.value,
    headline: headline.value,
    image_filename: 'placeholder.png',
    model: textModel.value,
    product: product.value,
  }

  saving.value = true
  ax.post(saveUrl, input)
    .then(resp => {
      emit('generated')
      clear()
    })
    .catch() // handled by interceptor
    .finally(() => saving.value = false)
}

</script>

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

        <v-text-field label="Cmd" v-model="item.cmd"
          placeholder="refcli fake ..."
          :rules="[
            (v: string) => !!v || 'Cmd is required',
            (v: string) => !v.includes('&') || 'Cmd must not contain &quot;&amp;&quot;',
            (v: string) => !v.includes('|') || 'Cmd must not contain &quot;|&quot;',
            (v: string) => !v.includes(';') || 'Cmd must not contain &quot;;&quot;',
          ]"
        ></v-text-field>

        <v-text-field label="Display order" v-model.number="item.display_order"
          :rules="[(v: number) => !!v || 'Display order is required']"
        ></v-text-field>

      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="$t(`actions.${saveBtnLabel.toLowerCase()}`)" :saveDisabled="!auth.isWriter()"
      @cancel="emit('cancel')" @save="saveItem">
      <template #extra>
        <v-spacer />
        <v-btn v-if="props.id !== 0" :disabled="!auth.isWriter()" color="error" @click="deleteItem">{{ $t('actions.delete') }}</v-btn>
      </template>
    </l-cancel-and-save-actions>

  </v-form>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useFormCrud } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { type Step, NewStep, GetStepInputFromItem } from '@/types/process'

const props = defineProps<{
  id: number
  flow_id: number
}>()

const emit = defineEmits<{
  (e: 'cancel'): void
  (e: 'create', newId: number): void
  (e: 'delete'): void
  (e: 'load', id: number): void
  (e: 'update'): void
}>()

const { t } = useI18n()

const { item, itemForm, saving, saveBtnLabel, showSaved, deleteItem, saveItem } =
  useFormCrud<Step>({
    ax,
    id: props.id,
    baseUrl: '/a/process/steps',
    newItem: () => NewStep(props.flow_id), // flow_id is mandatory and not changeable by user, so set it here
    getInput: GetStepInputFromItem,

    onCreate: (id) => { emit('create', id) },
    onDelete: () => { emit('delete') },
    onLoad: (id) => emit('load', id),
    onUpdate: () => { emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? (item.value?.name ?? '') : t('flows.steps.new_item')
})

</script>

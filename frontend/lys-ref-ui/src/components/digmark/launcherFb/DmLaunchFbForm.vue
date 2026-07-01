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

        <v-autocomplete label="Manager" v-model="item.manager"
          :items="digmarkStore.managersSelectable"
          :rules="[(v: number) => !!v || 'Manager is required']"
        ></v-autocomplete>

        <v-text-field label="Fan page" v-model="item.fan_page"
          :rules="[(v: string) => !!v || 'Fan page is required']"
        ></v-text-field>

        <v-text-field label="Daily budget (EUR)" type="number" v-model.number="item.daily_budget_eur"
          :rules="[
            (v: number) => v != undefined || 'Daily budget is required',
            (v: number) => v >= 0 && v <= 2000 || 'Daily budget must be between 0 and 2,000'
          ]"
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
import { useDigmarkStore } from '@/stores/digmark'
import { type LauncherFb, NewLauncherFb, GetLauncherInputFbFromItem } from '@/types/digmark'

const props = defineProps<{
  id: number
}>()

const emit = defineEmits<{
  (e: 'cancel'): void
  (e: 'create', newID: number): void
  (e: 'delete'): void
  (e: 'load', id: number): void
  (e: 'update'): void
}>()

const { t } = useI18n()
const digmarkStore = useDigmarkStore()

const { item, itemForm, saving, saveBtnLabel, showSaved, deleteItem, saveItem } =
  useFormCrud<LauncherFb>({
    ax,
    id: props.id,
    baseUrl: '/a/digmark/launchers-fb',
    newItem: NewLauncherFb,
    getInput: GetLauncherInputFbFromItem,

    onCreate: (id) => { emit('create', id) },
    onDelete: () => { emit('delete') },
    onLoad: (id) => emit('load', id),
    onUpdate: () => { emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? (item.value?.name ?? '') : t('launchers.new_item')
})

</script>

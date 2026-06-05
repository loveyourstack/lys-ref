<template>
  <v-card-title class="pl-1 mb-1">
    {{ formTitle }}
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">
    <v-row>
      <v-col class="form-col">

        <v-switch label="Active" v-model="item.is_active" color="primary" hide-details></v-switch>

        <v-text-field label="Name" v-model="item.name"
          :rules="[(v: string) => !!v || 'Name is required']"
        ></v-text-field>

        <v-autocomplete label="Manager" v-model="item.manager"
          :items="digmarkStore.managersSelectable"
          :rules="[(v: number) => !!v || 'Manager is required']"
        ></v-autocomplete>

        <v-autocomplete label="Country" v-model="item.country_fk"
          :items="geoStore.mandatoryCountries" item-value="id" item-title="name"
          :rules="[(v: number) => !!v || 'Country is required']"
        ></v-autocomplete>

        <v-autocomplete label="Vertical" v-model="item.vertical_fk"
          :items="digmarkStore.verticals" item-value="id" item-title="name"
          :rules="[(v: number) => !!v || 'Vertical is required']"
        ></v-autocomplete>

        <v-text-field label="Daily budget (EUR)" type="number" v-model.number="item.daily_budget_eur"
          :rules="[
            (v: number) => v != undefined || 'Daily budget is required',
            (v: number) => v >= 0 && v <= 2000 || 'Daily budget must be between 0 and 2,000'
          ]"
        ></v-text-field>

      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="saveBtnLabel" :saveDisabled="!auth.isWriter()"
      @cancel="emit('cancel')" @save="saveItem">
      <template #extra>
        <v-spacer />
        <v-btn v-if="props.id !== 0" :disabled="!auth.isWriter()" color="error" @click="deleteItem">Delete</v-btn>
      </template>
    </l-cancel-and-save-actions>

  </v-form>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useFormCrud } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { useDigmarkStore } from '@/stores/digmark'
import { useGeoStore } from '@/stores/geo'
import { type Campaign, NewCampaign, GetCampaignInputFromItem } from '@/types/digmark'

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

const digmarkStore = useDigmarkStore()
const geoStore = useGeoStore()

const { item, itemForm, saving, saveBtnLabel, showSaved, deleteItem, saveItem } =
  useFormCrud<Campaign>({
    ax,
    id: props.id,
    baseUrl: '/a/digmark/campaigns',
    newItem: NewCampaign,
    getInput: GetCampaignInputFromItem,

    onCreate: (id) => { emit('create', id) },
    onDelete: () => { emit('delete') },
    onLoad: (id) => emit('load', id),
    onUpdate: () => { emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? (item.value?.name ?? '') : 'New campaign'
})

</script>

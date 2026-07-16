<template>
  <v-row class="ml-4 pb-6">
    <v-col v-for="camp in items" :key="camp.id" cols="auto">
      <v-card elevation="2" max-width="650">
        <v-card-title>{{ camp.headline }}</v-card-title>

        <v-card-text>
          <v-row class="align-center">
            <v-rating :model-value="fakeStatsById[camp.id]?.rating" color="#D0D0D0" :active-color="'rgb(var(--v-theme-primary))'" 
              density="compact" size="small" half-increments readonly></v-rating>
            <div class="text-grey ms-4">{{ fakeStatsById[camp.id]?.rating }} ({{ fakeStatsById[camp.id]?.votes }})</div>
          </v-row>

          <div class="mt-4">{{ camp.body }}</div>
        </v-card-text>

        <v-card-actions>
          <v-btn :color="'rgb(var(--v-theme-secondary))'" border prepend-icon="mdi-star-four-points">{{ camp.call_to_action }}</v-btn>
        </v-card-actions>

        <v-divider class="mt-4"></v-divider>

        <v-list-item prepend-icon="mdi-account-arrow-right" :subtitle="camp.product + ' (' + camp.model + ')'"></v-list-item>
      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue'
import { useTableState } from 'lys-vue'
import ax from '@/api'
import { type GeneratedCampaign } from '@/types/digmark'

const props = defineProps<{
  refresh: string
}>()

const baseUrl = '/a/digmark/generated-campaigns'

// add fake rating and votes to each campaign for display purposes
type FakeCardStats = {
  rating: number
  votes: number
}
const fakeStatsById = ref<Record<number, FakeCardStats>>({})

function randomInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

const ratingChoices = [4, 4.5, 5]
function randomRating(): number {
  return ratingChoices[randomInt(0, ratingChoices.length - 1)] ?? 4.5
}

const { items, loadItems } = useTableState<GeneratedCampaign>({ ax, baseUrl, onFetchSuccess: (resp) => {
  for (const camp of resp.data.data) {
    if (!fakeStatsById.value[camp.id]) {
      fakeStatsById.value[camp.id] = {
        rating: randomRating(),
        votes: randomInt(500, 2500)
      }
    }
  }
}})

watch(() => props.refresh, () => {
  loadItems({ page: 1, itemsPerPage: 8, sortBy: [] })
})

</script>

<template>
  <v-container fluid>
    <v-responsive>

      <v-row v-if="item" density="compact">
        <v-col>
          <proc-point-diag :flow="item.flow" :run_id="props.run_id" :step_name="item.step_name" />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-btn icon class="mr-4 mb-1 ml-1" @click="router.back">
            <v-icon icon="mdi-arrow-left"></v-icon>
          </v-btn>
        </v-col>
      </v-row>

    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { fetchOnce } from 'lys-vue'
import ax from '@/api'
import { type Run } from '@/types/process'

const props = defineProps<{
  run_id: number
}>()

const router = useRouter()

const baseUrl = '/a/process/runs/' + props.run_id
const item = ref<Run>()

onMounted(() => {
  fetchOnce({ ax, myUrl: baseUrl, result: item })
})

</script>

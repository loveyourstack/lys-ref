<template>
  <v-container fluid>
    <v-responsive>

      <v-row v-if="item" density="compact">
        <v-col>
          <proc-step-diag :default_params="item.params_replaced" :flow_id="props.flow_id" :flow_name="item.name" />
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
import { type Flow } from '@/types/process'

const props = defineProps<{
  flow_id: number
}>()

const router = useRouter()

const baseUrl = '/a/process/flows/' + props.flow_id
const item = ref<Flow>()

onMounted(() => {
  fetchOnce({ ax, myUrl: baseUrl, result: item })
})

</script>

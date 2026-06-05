<template>
  <v-dialog v-model="showDialog" width="auto" min-width="300px" max-width="800px">
    <v-card>
      <v-card-title style="background-color: red;">
        <v-icon size="small" color="white" icon="mdi-alert" class="mr-2"></v-icon>
        <span style="color: white;">Server error</span>
      </v-card-title>

      <v-card-text v-if="appStore.apiErr">
        <div class="text-body-1 mb-2">{{ appStore.apiErr.errMsg }}</div>
        <div class="text-caption">{{ appStore.apiErr.method }} {{ appStore.apiErr.url }}</div>
      </v-card-text>

      <v-card-actions>
        <v-btn color="primary" block @click="appStore.apiErr = undefined">Close</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const showDialog = computed(() => {
  return appStore.apiErr !== undefined
})
</script>
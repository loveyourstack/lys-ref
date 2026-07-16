<template>
  <v-container fluid>
    <v-responsive>
      <v-row density="compact" class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text class="pb-0">
              <span class="dt-title">{{ $t('generation.title') }}</span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <div class="dt-subtitle">{{ $t('generation.p1') }}</div>

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <dm-gen-form @generated="refreshSavedCamps = String(Date.now())" />

      <v-row class="ml-4 mt-4">
        <v-col cols="auto">
          <div class="text-headline-small">{{ $t('generation.saved_campaigns') }}</div>
        </v-col>
      </v-row>

      <dm-gen-saved-camps :refresh="refreshSavedCamps" /> 

    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'

const refreshSavedCamps = ref('')

onMounted(() => {
  refreshSavedCamps.value = String(Date.now())
})

</script>

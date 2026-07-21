<template>
  <v-container fluid>
    <v-responsive>

      <v-row density="compact" class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text class="pb-0">
              <span class="dt-title">Application monitor</span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>
              <v-tabs v-model="selectedTab" density="comfortable" class="mb-4">
                <v-tab value="sessions">Sessions</v-tab>
                <v-tab value="reqLog">Request log</v-tab>
                <v-tab value="loginAttempts">Login attempts</v-tab>
                <v-tab value="hubStatus">Hub status</v-tab>
                <v-tab value="awsCalls">AWS API</v-tab>
                <v-tab value="ecbCalls">ECB API</v-tab>
                <v-tab value="gemCalls">Gemini API</v-tab>
                <v-tab value="mmCalls">Maxmind API</v-tab>
                <v-tab value="tedbCalls">TEDB API</v-tab>
              </v-tabs>

              <v-window v-model="selectedTab">

                <v-window-item value="sessions">
                  <app-mon-session-table />
                </v-window-item>

                <v-window-item value="reqLog">
                  <app-mon-req-log-table />
                </v-window-item>

                <v-window-item value="loginAttempts">
                  <app-mon-login-attempt-table />
                </v-window-item>

                <v-window-item value="hubStatus">
                  <app-mon-hub-status />
                </v-window-item>

                <v-window-item value="awsCalls">
                  <app-mon-aws-call-table />
                </v-window-item>

                <v-window-item value="ecbCalls">
                  <app-mon-ecb-call-table />
                </v-window-item>

                <v-window-item value="gemCalls">
                  <app-mon-gem-call-table />
                </v-window-item>

                <v-window-item value="mmCalls">
                  <app-mon-mm-call-table />
                </v-window-item>

                <v-window-item value="tedbCalls">
                  <app-mon-tedb-call-table />
                </v-window-item>

              </v-window>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useJsonLs } from 'lys-vue'

const selectedTab = ref('reqLog')

useJsonLs({
  lsKey: 'app_monitor',
  refs: {
    selectedTab,
  },
})

</script>

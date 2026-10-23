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

              <v-tabs v-model="selectedGroup" density="comfortable" class="mb-2"
                @update:model-value="onGroupChange">
                <v-tab value="httpServer">HTTP server</v-tab>
                <v-tab value="rpcServer">RPC server</v-tab>
                <v-tab value="externalApis">External APIs</v-tab>
              </v-tabs>

              <!-- using key to force re-render of tabs when group changes -->
              <v-tabs :key="selectedGroup" v-model="selectedTab" density="comfortable" class="mb-4">
                <v-tab
                  v-for="tab in tabsByGroup[selectedGroup]"
                  :key="tab.value"
                  :value="tab.value"
                >
                  {{ tab.label }}
                </v-tab>
              </v-tabs>

              <v-window v-model="selectedTab">

                <!-- HTTP server -->

                <v-window-item value="httpSessions">
                  <app-mon-http-session-table />
                </v-window-item>

                <v-window-item value="httpReqLog">
                  <app-mon-http-req-log-table />
                </v-window-item>

                <v-window-item value="httpLoginAttempts">
                  <app-mon-http-login-attempt-table />
                </v-window-item>

                <v-window-item value="httpHubStatus">
                  <app-mon-http-hub-status />
                </v-window-item>

                <!-- RPC server -->

                <v-window-item value="rpcReqLog">
                  <app-mon-rpc-req-log-table />
                </v-window-item>

                <!-- External APIs -->

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
import { ref, nextTick } from 'vue'
import { useJsonLs } from 'lys-vue'

const tabsByGroup = {
  httpServer: [
    { value: 'httpSessions', label: 'Sessions' },
    { value: 'httpReqLog', label: 'Request log' },
    { value: 'httpLoginAttempts', label: 'Login attempts' },
    { value: 'httpHubStatus', label: 'Hub status' },
  ],
  rpcServer: [
    { value: 'rpcReqLog', label: 'Request log' },
  ],
  externalApis: [
    { value: 'awsCalls', label: 'AWS' },
    { value: 'ecbCalls', label: 'ECB' },
    { value: 'gemCalls', label: 'Gemini' },
    { value: 'mmCalls', label: 'Maxmind' },
    { value: 'tedbCalls', label: 'TEDB' },
  ],
} as const

type TabGroup = keyof typeof tabsByGroup

const selectedGroup = ref<TabGroup>('httpServer')
const selectedTab = ref('httpReqLog')

async function onGroupChange(group: TabGroup) {
  // wait for tabs to re-render before setting the selected tab
  await nextTick()
  selectedTab.value = tabsByGroup[group][0].value
}

useJsonLs({
  lsKey: 'app_monitor',
  refs: {
    selectedGroup,
    selectedTab,
  },
})

</script>

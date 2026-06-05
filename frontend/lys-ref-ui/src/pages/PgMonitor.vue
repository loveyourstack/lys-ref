<template>
  <v-container fluid>
    <v-responsive>

      <v-row density="compact" class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text class="pb-0">
              <span class="dt-title">Database monitor</span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row density="compact" class="mt-4">
        <v-col class="d-flex align-center text-medium-emphasis">
          <v-card variant="flat">
            <v-card-text class="pb-0">
              <div>{{ version }}</div>
              <div>{{ dbSize }}</div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>
              <v-tabs v-model="selectedTab" density="comfortable" class="mb-4">
                <v-tab value="queries">Queries</v-tab>
                <v-tab value="table-size">Table size</v-tab>
                <v-tab value="bloat">Bloat</v-tab>
                <v-tab value="unused-indexes">Unused indexes</v-tab>
                <v-tab value="settings">Settings</v-tab>
              </v-tabs>

              <v-window v-model="selectedTab">

                <v-window-item value="queries">
                  <l-pg-mon-query-table :ax="ax" baseUrl="/a/pg-monitor/queries" />
                </v-window-item>

                <v-window-item value="table-size">
                  <l-pg-mon-table-size-table :ax="ax" baseUrl="/a/pg-monitor/table-size" />
                </v-window-item>

                <v-window-item value="bloat">
                  <l-pg-mon-bloat-table :ax="ax" baseUrl="/a/pg-monitor/bloat" />
                </v-window-item>

                <v-window-item value="unused-indexes">
                  <l-pg-mon-unused-idx-table :ax="ax" baseUrl="/a/pg-monitor/unused-indexes" />
                </v-window-item>

                <v-window-item value="settings">
                  <l-pg-mon-setting-table :ax="ax" baseUrl="/a/pg-monitor/settings" />
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
import { ref, onMounted } from 'vue'
import { useJsonLs } from 'lys-vue'
import ax from '@/api'

const selectedTab = ref('queries')

const dbSizeUrl = '/a/pg-monitor/database-size'
const versionUrl = '/a/pg-monitor/version'

const dbSize = ref('')
const version = ref('')

function fetchDbInfo() {
  Promise.all([
    ax.get(dbSizeUrl),
    ax.get(versionUrl),
  ]).then(([
    dbSizeResp,
    versionResp,
  ]) => {
    dbSize.value = dbSizeResp.data.data
    version.value = versionResp.data.data
  })
  .catch() // handled by interceptor
}

useJsonLs({
  lsKey: 'pg_monitor',
  refs: {
    selectedTab,
  },
})

onMounted(() => {
  fetchDbInfo()
})
</script>

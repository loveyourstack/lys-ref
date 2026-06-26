<template>
  <v-container fluid>
    <v-responsive>
      <v-row density="compact" class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text class="pb-0">
              <span class="dt-title">{{ $t('mcp_server.title') }}</span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <i18n-t scope="global" keypath="mcp_server.p1" tag="div" class="dt-subtitle">
                <template #MCPServer>
                  <a href="https://modelcontextprotocol.io/docs/learn/server-concepts" target="_blank" rel="noopener noreferrer">{{ $t('mcp_server.title') }}</a>
                </template>
              </i18n-t>

              <div class="dt-subtitle">{{ $t('mcp_server.p2') }}</div>

              <v-table class="mt-4">
                <thead>
                  <tr>
                    <th>Natural language</th>
                    <th>Gets translated to: MCP tool -> params</th>
                    <th>Action</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="query in queries" :key="query.id">
                    <td>{{ query.naturalLanguage }}</td>
                    <td><code>{{ query.mcpTool }} -> {{ formatParams(query.params) }}</code></td>
                    <td>
                      <v-btn color="primary" :loading="isLoading" @click="activeQuery = query; run(query.naturalLanguage)">{{ $t('actions.run') }}</v-btn>
                    </td>
                  </tr>
                </tbody>
              </v-table>

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="auto">
          <v-card variant="flat" class="mt-2">
            <v-card-title>Response</v-card-title>
            <v-card-text>
              <span v-if="!mcpResp" class="dt-subtitle">The response from the MCP server will be shown here.</span>
              <span v-else-if="tableRows.length === 0" class="dt-subtitle">No data available.</span>

              <v-table v-else class="response-table">
                <thead>
                  <tr>
                    <th v-for="col in tableColumns" :key="String(col.key)" :class="col.align === 'end' ? 'text-end' : ''">
                      {{ col.label }}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in tableRows" :key="idx">
                    <td v-for="col in tableColumns" :key="String(col.key)" :class="col.align === 'end' ? 'text-end' : ''">
                      {{ col.format ? col.format(row[col.key]) : row[col.key] }}
                    </td>
                  </tr>
                </tbody>
              </v-table>

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue'
import ax from '@/api'
import { queries } from '@/components/digmark/mcpServer/mcpServer.queries'
import { type McpQueryDef } from '@/types/digmark'

const baseUrl = '/a/digmark/mcp-query'
const isLoading = ref(false)

const activeQuery = ref<McpQueryDef<any> | null>(null)
const mcpResp = ref<unknown>(null)

const formatParams = (p: Record<string, unknown>) => Object.entries(p).map(([k, v]) => `${k} = ${String(v)}`).join(', ')
const tableColumns = computed(() => activeQuery.value?.columns ?? [])
const tableRows = computed(() =>
  activeQuery.value ? activeQuery.value.normalize(mcpResp.value) : []
)

function run(natLangQuery: string) {
  isLoading.value = true
  ax.post(baseUrl, {'query': natLangQuery })
    .then(response => {
      mcpResp.value = response.data.data.mcp_result.structuredContent ?? null
    })
    .catch() // handled by interceptor
    .finally(() => isLoading.value = false)
}
</script>

<style scoped>
.response-table { min-width: 500px; max-width: 100%; }
@media (max-width: 600px) { 
  .response-table { min-width: 0; width: 100%; }
}
</style>
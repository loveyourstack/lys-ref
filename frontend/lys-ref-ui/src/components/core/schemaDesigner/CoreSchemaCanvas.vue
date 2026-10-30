<template>
  <v-snackbar v-model="showError" color="error" timeout="3000">{{ $t('schema_designer.invalid_file') }}</v-snackbar>

  <v-dialog v-model="showSql" max-width="800">
    <v-card :title="$t('schema_designer.sql_title')">
      <v-card-text>
        <pre class="schema-sql">{{ sql }}</pre>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn prepend-icon="mdi-content-copy" @click="copySql">{{ $t('actions.copy') }}</v-btn>
        <v-btn @click="showSql = false">{{ $t('actions.close') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <input ref="fileInput" type="file" accept="application/json,.json" hidden @change="onFileChosen" />

  <!-- marker defs referenced by CoreSchemaRelationEdge -->
  <svg width="0" height="0" style="position: absolute;">
    <defs>
      <marker id="schema-one" viewBox="0 0 24 16" refX="24" refY="8" markerWidth="24" markerHeight="16" markerUnits="userSpaceOnUse" orient="auto">
        <path d="M18 2 V14" stroke="#888" stroke-width="1.5" fill="none" />
      </marker>
      <marker id="schema-many" viewBox="0 0 24 16" refX="24" refY="8" markerWidth="24" markerHeight="16" markerUnits="userSpaceOnUse" orient="auto-start-reverse">
        <path d="M24 2 L12 8 L24 14 M24 8 H12" stroke="#888" stroke-width="1.5" fill="none" />
        <circle cx="8" cy="8" r="3" stroke="#888" stroke-width="1.5" fill="none" />
      </marker>
    </defs>
  </svg>

  <v-toolbar density="compact" flat class="mb-2">
    <v-btn prepend-icon="mdi-plus" @click="addTable">{{ $t('schema_designer.add_table') }}</v-btn>
    <v-btn prepend-icon="mdi-delete-outline" :disabled="!selectedId" @click="deleteSelected">{{ $t('schema_designer.delete_table') }}</v-btn>
    <v-btn prepend-icon="mdi-fit-to-screen-outline" @click="fitView({ padding: 0.2 })">{{ $t('schema_designer.fit') }}</v-btn>
    <v-spacer />
    <v-btn prepend-icon="mdi-download" @click="exportJson">{{ $t('schema_designer.export_json') }}</v-btn>
    <v-btn :disabled="!auth.isWriter()" prepend-icon="mdi-upload" @click="fileInput?.click()">{{ $t('schema_designer.import_json') }}</v-btn>
    <v-btn prepend-icon="mdi-database-export-outline" @click="openSql">{{ $t('schema_designer.export_sql') }}</v-btn>
    <v-btn prepend-icon="mdi-restore" @click="resetSchema">{{ $t('schema_designer.reset') }}</v-btn>
  </v-toolbar>

  <div class="d-flex" style="height: calc(100vh - 270px); min-height: 500px; min-width: 900px;">
    <div ref="canvasEl" class="flex-grow-1 schema-canvas">
      <VueFlow :nodes="nodes" :edges="edges" :min-zoom="0.2" :max-zoom="2" :nodes-connectable="true"
        :connection-mode="ConnectionMode.Loose" fit-view-on-init
        @node-drag-stop="onNodeDragStop" @node-click="onNodeClick" @pane-click="selectedId = null"
        @nodes-initialized="centerPendingNode" @connect="onConnect">
        <template #node-table="props">
          <core-schema-table-node v-bind="props" />
        </template>
        <template #edge-relation="props">
          <core-schema-relation-edge v-bind="props" />
        </template>
        <Background />
        <Controls :show-fit-view="false" />
        <MiniMap pannable zoomable />
      </VueFlow>
    </div>
    <core-schema-edit-panel :table-id="selectedId" class="ms-2" />
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue'
import { ConnectionMode, VueFlow, useVueFlow, type Connection, type Edge, type Node, type NodeDragEvent, type NodeMouseEvent } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import auth from '@/auth'
import { generateSchemaSql } from '@/composables/schemaSql'
import { useCoreStore } from '@/stores/core'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

const MIN_NODE_WIDTH = 240

const core = useCoreStore()
const { fitView, findNode, screenToFlowCoordinate } = useVueFlow()

const selectedId = ref<string | null>(null)
const pendingCenterNodeId = ref<string | null>(null)
const canvasEl = ref<HTMLElement>()
const fileInput = ref<HTMLInputElement>()
const showError = ref(false)
const showSql = ref(false)
const sql = ref('')

const nodes = computed<Node[]>(() => core.schemaTables.map(t => ({
  id: t.id,
  type: 'table',
  position: { x: t.x, y: t.y },
  data: { table: t },
  selected: t.id === selectedId.value,
})))

// the referencing column's handle faces the referenced table, so lines leave and enter on the facing sides
const edges = computed<Edge[]>(() => {
  const result: Edge[] = []
  for (const t of core.schemaTables) {
    for (const c of t.columns) {
      if (!c.ref) continue
      const parent = core.getSchemaTable(c.ref.table_id)
      if (!parent) continue
      const childWidth = findNode(t.id)?.dimensions.width || MIN_NODE_WIDTH
      const parentWidth = findNode(parent.id)?.dimensions.width || MIN_NODE_WIDTH
      const childLeft = t.x >= parent.x + parentWidth
      const childRight = t.x + childWidth <= parent.x
      const sourceSide = childRight ? 'r' : 'l'
      const targetSide = childRight ? 'l' : childLeft ? 'r' : 'l'
      result.push({
        id: c.id,
        type: 'relation',
        source: t.id,
        sourceHandle: `${c.id}-${sourceSide}`,
        target: parent.id,
        targetHandle: `${c.ref.column_id}-${targetSide}`,
      })
    }
  }
  return result
})

function addTable() {
  const rect = canvasEl.value?.getBoundingClientRect()
  const pos = rect
    ? screenToFlowCoordinate({ x: rect.left + rect.width / 2 - MIN_NODE_WIDTH / 2, y: rect.top + rect.height / 3 })
    : { x: 0, y: 0 }
  const table = core.addSchemaTable(pos.x, pos.y)
  selectedId.value = table.id
  pendingCenterNodeId.value = rect ? table.id : null
}

function centerPendingNode() {
  const id = pendingCenterNodeId.value
  const node = id ? findNode(id) : undefined
  const table = id ? core.getSchemaTable(id) : undefined
  const rect = canvasEl.value?.getBoundingClientRect()
  if (!id || !node || !table || !rect) return

  const center = screenToFlowCoordinate({ x: rect.left + rect.width / 2, y: rect.top + rect.height / 3 })
  core.setSchemaTablePosition(id, center.x - node.dimensions.width / 2, table.y)
  pendingCenterNodeId.value = null
}

function deleteSelected() {
  if (!selectedId.value) return
  core.deleteSchemaTable(selectedId.value)
  selectedId.value = null
}

function onNodeDragStop(e: NodeDragEvent) {
  for (const n of e.nodes) core.setSchemaTablePosition(n.id, n.position.x, n.position.y)
}

function onNodeClick(e: NodeMouseEvent) {
  selectedId.value = e.node.id
}

// handle ids are "<columnId>-<side>"; the dragged-from column becomes the foreign key
function onConnect(c: Connection) {
  if (!c.sourceHandle || !c.targetHandle || c.source === c.target) return
  const childColId = c.sourceHandle.slice(0, -2)
  const parentColId = c.targetHandle.slice(0, -2)
  core.setSchemaColumnRef(c.source, childColId, { table_id: c.target, column_id: parentColId })
}

function resetSchema() {
  core.resetSchemaTables()
  selectedId.value = null
  setTimeout(() => fitView({ padding: 0.2 }))
}

function exportJson() {
  const blob = new Blob([JSON.stringify(core.schemaTables, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'schema.json'
  a.click()
  URL.revokeObjectURL(url)
}

async function onFileChosen(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  try {
    core.importSchemaTables(await file.text())
    selectedId.value = null
    setTimeout(() => fitView({ padding: 0.2 }))
  } catch {
    showError.value = true
  }
}

function openSql() {
  sql.value = generateSchemaSql(core.schemaTables)
  showSql.value = true
}

function copySql() {
  navigator.clipboard.writeText(sql.value)
}
</script>

<style scoped>
.schema-canvas {
  border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
  border-radius: 4px;
  background: rgb(var(--v-theme-background));
}
.schema-sql {
  font-family: monospace;
  font-size: 0.8rem;
  white-space: pre;
  overflow: auto;
  max-height: 60vh;
}
</style>

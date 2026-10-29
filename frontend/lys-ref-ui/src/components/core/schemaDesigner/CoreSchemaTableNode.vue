<template>
  <v-card class="schema-node" :class="{ 'schema-node-selected': selected }" elevation="2">
    <div class="schema-node-header">
      <v-icon size="small" class="me-2">mdi-table</v-icon>
      <span class="text-truncate">{{ schemaTableLabel(table) }}</span>
    </div>
    <div v-for="c in table.columns" :key="c.id" class="schema-node-row">
      <Handle :id="`${c.id}-l`" type="source" :position="Position.Left" />
      <span class="schema-node-col-name" :class="{ 'font-weight-bold': c.pk }">{{ c.name }}</span>
      <v-icon v-if="c.pk" size="x-small" class="ms-1">mdi-key</v-icon>
      <v-icon v-else-if="c.ref" size="x-small" class="ms-1">mdi-link</v-icon>
      <v-spacer />
      <span class="schema-node-col-type">{{ c.type }}</span>
      <span v-if="c.not_null" class="schema-node-nn ms-1">NN</span>
      <Handle :id="`${c.id}-r`" type="source" :position="Position.Right" />
    </div>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, nextTick, watch } from 'vue'
import { Handle, Position, useVueFlow } from '@vue-flow/core'
import { schemaTableLabel, type SchemaTable } from '@/types/core'

// Vue Flow passes its own `position` ({x, y}) to node components, which would fall through to v-card's `position` prop
defineOptions({ inheritAttrs: false })

const props = defineProps<{
  id: string
  data: { table: SchemaTable }
  selected?: boolean
}>()

const table = computed(() => props.data.table)

// handle positions are cached by Vue Flow, so re-measure when columns are reordered
const { updateNodeInternals } = useVueFlow()
watch(() => table.value.columns.map(c => c.id).join(), async () => {
  await nextTick()
  updateNodeInternals([props.id])
})
</script>

<style scoped>
.schema-node {
  width: max-content;
  min-width: 240px;
  overflow: visible;
  border: 1px solid transparent;
}
.schema-node-selected {
  border-color: rgb(var(--v-theme-primary));
}
.schema-node-header {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  font-weight: 600;
  color: #fff;
  background: #2f6694;
  border-radius: 4px 4px 0 0;
}
.schema-node-row {
  position: relative;
  display: flex;
  align-items: center;
  padding: 6px 12px;
  font-size: 0.875rem;
  gap: 0;
}
.schema-node-col-name {
  white-space: nowrap;
  margin-right: 16px;
}
.schema-node-col-type {
  white-space: nowrap;
  margin-left: 12px;
  font-family: monospace;
  font-size: 0.75rem;
  opacity: 0.7;
}
.schema-node-nn {
  font-size: 0.65rem;
  padding: 0 4px;
  border-radius: 3px;
  background: rgba(var(--v-theme-on-surface), 0.1);
  opacity: 0.8;
}
</style>

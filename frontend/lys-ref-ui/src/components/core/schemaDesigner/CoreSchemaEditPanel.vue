<template>
  <v-card v-if="table" variant="flat" class="pa-3" style="width: 380px; overflow-y: auto;">
    <v-text-field v-model="table.schema" :label="$t('schema_designer.schema_name')" density="compact" hide-details class="mb-2" />
    <v-text-field v-model="table.name" :label="$t('schema_designer.table_name')" density="compact" hide-details class="mb-4" />

    <div class="d-flex align-center mb-2">
      <span class="text-title-small">{{ $t('schema_designer.columns') }}</span>
      <v-spacer />
      <v-btn size="small" variant="text" prepend-icon="mdi-plus" @click="core.addSchemaColumn(table.id)">
        {{ $t('schema_designer.add_column') }}
      </v-btn>
    </div>

    <v-card v-for="(c, i) in table.columns" :key="c.id" variant="flat" class="schema-column pa-3 mb-4">
      <div class="d-flex align-center ga-1 mb-2">
        <v-text-field v-model="c.name" :label="$t('schema_designer.name')" density="compact" hide-details />
        <v-select v-model="c.type" :items="SchemaColumnTypes" :label="$t('schema_designer.type')" density="compact" hide-details
          style="max-width: 140px;" />
      </div>
      <div class="d-flex align-center">
        <v-checkbox v-model="c.pk" :label="$t('schema_designer.pk')" density="compact" hide-details />
        <v-checkbox v-model="c.not_null" :label="$t('schema_designer.not_null')" density="compact" hide-details />
        <v-spacer />
        <v-btn icon="mdi-arrow-up" size="x-small" variant="text" :disabled="i === 0" @click="core.moveSchemaColumn(table.id, i, -1)" />
        <v-btn icon="mdi-arrow-down" size="x-small" variant="text" :disabled="i === table.columns.length - 1" @click="core.moveSchemaColumn(table.id, i, 1)" />
        <v-btn icon="mdi-delete-outline" size="x-small" variant="text" @click="core.deleteSchemaColumn(table.id, c.id)" />
      </div>
      <v-select :model-value="refValue(c.ref)" :items="refItems" :label="$t('schema_designer.references')" density="compact"
        hide-details clearable @update:model-value="v => core.setSchemaColumnRef(table!.id, c.id, parseRef(v))" />
    </v-card>
  </v-card>
  <v-card v-else variant="flat" class="pa-4" style="width: 380px;">
    <span class="text-medium-emphasis">{{ $t('schema_designer.select_table') }}</span>
  </v-card>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useCoreStore } from '@/stores/core'
import { SchemaColumnTypes, schemaTableLabel, type SchemaColumnRef } from '@/types/core'

const props = defineProps<{ tableId: string | null }>()

const core = useCoreStore()

const table = computed(() => props.tableId ? core.getSchemaTable(props.tableId) : undefined)

const refItems = computed(() =>
  core.schemaTables.flatMap(t => t.columns.map(c => ({ title: `${schemaTableLabel(t)}.${c.name}`, value: `${t.id}|${c.id}` }))))

function refValue(ref?: SchemaColumnRef) {
  return ref ? `${ref.table_id}|${ref.column_id}` : null
}

function parseRef(v: string | null): SchemaColumnRef | undefined {
  if (!v) return undefined
  const [table_id, column_id] = v.split('|')
  return { table_id: table_id!, column_id: column_id! }
}
</script>

<style scoped>
.schema-column {
  background: rgb(var(--v-theme-surface));
  border: 1px solid rgba(var(--v-border-color), 0.3);
  border-left: 4px solid rgb(var(--v-theme-primary));
  border-radius: 8px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.18);
}
</style>

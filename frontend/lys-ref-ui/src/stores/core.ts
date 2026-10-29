import { ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { fetchOnce } from 'lys-vue'
import ax from '@/api'
import { NewSchemaColumn, NewSchemaTable, type SchemaColumn, type SchemaColumnRef, type SchemaTable } from '@/types/core'

export const useCoreStore = defineStore('core', () => {

  const mandatoryEnums = ref<string[]>([])
  const optionalEnums = ref<string[]>([])
  const periods = ref<string[]>([])
  const xrPeriods = ref<string[]>([])
  
  function loadMandatoryEnums() {
    const myUrl = '/a/core/mandatory-enums'
    fetchOnce({ ax, myUrl, result: mandatoryEnums })
  }

  function loadOptionalEnums() {
    const myUrl = '/a/core/optional-enums'
    fetchOnce({ ax, myUrl, result: optionalEnums })
  }

  function loadPeriods() {
    const myUrl = '/a/core/performance-periods'
    fetchOnce({ ax, myUrl, result: periods })
  }

  // use watch rather than onSuccess callback since store values are set directly on initialization in Main.vue, not just via load funcs

  watch([periods], () => {
    xrPeriods.value = periods.value.filter((p: string) => !['Today', 'Yesterday', 'Last 3 days'].includes(p))
  })

  // schema designer

  const SCHEMA_STORAGE_KEY = 'lys-ref-schema-designer'

  function seedSchemaTables(): SchemaTable[] {
    const col = (name: string, type: string, o: Partial<SchemaColumn> = {}): SchemaColumn =>
      ({ ...NewSchemaColumn(), name, type, ...o })
    const users: SchemaTable = {
      id: crypto.randomUUID(), schema: 'public', name: 'users', x: 40, y: 80,
      columns: [
        col('id', 'bigint', { pk: true, not_null: true }),
        col('username', 'varchar'),
        col('role', 'varchar'),
        col('created_at', 'timestamptz'),
      ],
    }
    const posts: SchemaTable = {
      id: crypto.randomUUID(), schema: 'public', name: 'posts', x: 480, y: 40,
      columns: [
        col('id', 'bigint', { pk: true, not_null: true }),
        col('title', 'varchar'),
        col('body', 'text'),
        col('user_id', 'bigint', { not_null: true, ref: { table_id: users.id, column_id: users.columns[0]!.id } }),
        col('status', 'varchar'),
        col('created_at', 'timestamptz'),
      ],
    }
    const follows: SchemaTable = {
      id: crypto.randomUUID(), schema: 'public', name: 'follows', x: 40, y: 420,
      columns: [
        col('id', 'bigint', { pk: true, not_null: true }),
        col('following_user_id', 'bigint', { not_null: true, ref: { table_id: users.id, column_id: users.columns[0]!.id } }),
        col('followed_user_id', 'bigint', { not_null: true, ref: { table_id: users.id, column_id: users.columns[0]!.id } }),
        col('created_at', 'timestamptz'),
      ],
    }
    return [users, posts, follows]
  }

  function withDefaultSchema(t: SchemaTable): SchemaTable {
    return { ...t, schema: t.schema || 'public' }
  }

  function loadSchemaTables(): SchemaTable[] {
    try {
      const raw = localStorage.getItem(SCHEMA_STORAGE_KEY)
      if (raw) {
        const parsed = JSON.parse(raw)
        if (Array.isArray(parsed)) return parsed.map(withDefaultSchema)
      }
    } catch {
      // fall through to seed data
    }
    return seedSchemaTables()
  }

  const schemaTables = ref<SchemaTable[]>(loadSchemaTables())

  watch(schemaTables, () => {
    localStorage.setItem(SCHEMA_STORAGE_KEY, JSON.stringify(schemaTables.value))
  }, { deep: true })

  function getSchemaTable(id: string) {
    return schemaTables.value.find(t => t.id === id)
  }

  function addSchemaTable(x: number, y: number) {
    const t = NewSchemaTable(`table_${schemaTables.value.length + 1}`, x, y)
    schemaTables.value.push(t)
    return t
  }

  function deleteSchemaTable(id: string) {
    schemaTables.value = schemaTables.value.filter(t => t.id !== id)
    for (const t of schemaTables.value) {
      for (const c of t.columns) {
        if (c.ref?.table_id === id) c.ref = undefined
      }
    }
  }

  function addSchemaColumn(tableId: string) {
    const c = NewSchemaColumn()
    getSchemaTable(tableId)?.columns.push(c)
    return c
  }

  function deleteSchemaColumn(tableId: string, columnId: string) {
    const t = getSchemaTable(tableId)
    if (!t) return
    t.columns = t.columns.filter(c => c.id !== columnId)
    for (const o of schemaTables.value) {
      for (const c of o.columns) {
        if (c.ref?.column_id === columnId) c.ref = undefined
      }
    }
  }

  function moveSchemaColumn(tableId: string, index: number, delta: number) {
    const cols = getSchemaTable(tableId)?.columns
    const to = index + delta
    if (!cols || to < 0 || to >= cols.length) return
    cols.splice(to, 0, ...cols.splice(index, 1))
  }

  function setSchemaColumnRef(tableId: string, columnId: string, ref?: SchemaColumnRef) {
    const c = getSchemaTable(tableId)?.columns.find(c => c.id === columnId)
    if (c) c.ref = ref
  }

  function setSchemaTablePosition(id: string, x: number, y: number) {
    const t = getSchemaTable(id)
    if (t) {
      t.x = x
      t.y = y
    }
  }

  function resetSchemaTables() {
    schemaTables.value = seedSchemaTables()
  }

  // throws if the JSON is not a schema table array
  function importSchemaTables(json: string) {
    const parsed = JSON.parse(json)
    const valid = Array.isArray(parsed) && parsed.every((t: SchemaTable) =>
      typeof t.id === 'string' && typeof t.name === 'string' && (t.schema === undefined || typeof t.schema === 'string') && Array.isArray(t.columns) &&
      t.columns.every((c: SchemaColumn) => typeof c.id === 'string' && typeof c.name === 'string'))
    if (!valid) throw new Error('invalid schema')
    schemaTables.value = parsed.map(withDefaultSchema)
  }

  return { 
    mandatoryEnums, optionalEnums, periods, xrPeriods,
    loadMandatoryEnums, loadOptionalEnums, loadPeriods,
    schemaTables, getSchemaTable, addSchemaTable, deleteSchemaTable, addSchemaColumn, deleteSchemaColumn,
    moveSchemaColumn, setSchemaColumnRef, setSchemaTablePosition, resetSchemaTables, importSchemaTables,
  }
})


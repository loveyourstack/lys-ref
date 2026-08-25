<template>
  <v-treeview v-model:selected="selected" return-object :items="items"
    @update:selected="emit('selected', selected[0] ?? undefined)"
    item-value="id" item-title="full_name"
    open-all indent-lines="default"
    selectable select-strategy="single-independent" selected-color="success">
    <template v-slot:prepend="{ item }">
      <v-avatar v-if="item.profile_pic" :image="`${s3Bucket}profiles/${item.profile_pic}`" 
        size="32" class="mr-3"></v-avatar>
    </template>
    <template v-slot:append="{ item }">
      <span v-if="item.children && item.children.length > 0 && (item.job_title.includes('Chief') || item.job_title.includes('VP'))" 
        class="text-body-small ml-4">{{ item.department }}</span>

      <!-- hack to ensure that department is right-aligned to all items, not just the ones shown by the v-if -->
      <span v-else class="text-body-small ml-4 opacity-0">{{ item.department }}</span>
     </template>
    <template v-slot:title="{ item }">
      <span class="font-weight-medium">{{ item.full_name }}</span>
      <span class="ml-4 text-body-small">{{ item.job_title }}</span>
    </template>
  </v-treeview>
</template>

<script setup lang="ts">
/*
Note: I wanted to clear selected on expanding/collapsing nodes, but currently when I bind v-model:opened to ref(<Employee[]>([])), the treeview nodes become unclickable.
*/

import { ref, onMounted } from 'vue'
import ax from '@/api'
import { type Employee } from '@/types/lysinc'

const emit = defineEmits<{
  (e: 'selected', item: Employee | undefined): void
}>()

const baseUrl = '/a/lysinc/employees/tree'
const s3Bucket = import.meta.env.VITE_S3_BUCKET

const items = ref(<Employee[]>([]))
const selected = ref<Employee[]>([])

function loadItems() {
  ax.get(baseUrl).then((res) => {
    items.value = res.data.data
  })
}

onMounted(() => {
  loadItems()
})

</script>

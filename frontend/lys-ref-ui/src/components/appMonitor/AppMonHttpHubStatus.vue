<template>
  <v-row>
    <v-col class="d-flex align-center">
      <div class="dt-title">Hub status</div>
      <div class="v-spacer"></div>
      <v-btn icon flat v-tooltip="$t('actions.refresh')" @click="loadItems()">
        <v-icon icon="mdi-refresh"></v-icon>
      </v-btn>
    </v-col>
  </v-row>

  <v-table>
    <thead>
      <tr>
        <th>User</th>
        <th># connections</th>
      </tr>
    </thead>

    <tbody>
      <tr v-for="(count, user) in items" :key="user">
        <td>{{ user }}</td>
        <td>{{ count }}</td>
      </tr>
    </tbody>
  </v-table>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import ax from '@/api'

interface statusRecord {
   [userName: string]: number
}

const myUrl = '/a/tech/hub/status'
const items = ref<statusRecord>()

function loadItems() {
  ax.get(myUrl).then(res => {
    items.value = res.data.data as statusRecord
  })
}

onMounted(() => {
  loadItems()
})

</script>
<template>
  <v-list density="compact" class="nav-list nav-list-left">
    <v-list-item link title="Home" to="/home" prepend-icon="mdi-home"></v-list-item>

    <v-list-subheader title="Type handling" class="mt-2 clickable" @click="showTypeItems = !showTypeItems"></v-list-subheader>

    <div v-if="showTypeItems">
      <v-list-item link title="Mandatory values" to="/core/mandatory-values" prepend-icon="mdi-focus-field"></v-list-item>
      <v-list-item link title="Optional values" to="/core/optional-values" prepend-icon="mdi-focus-field"></v-list-item>
      <v-list-item link title="Default values" to="/core/default-values" prepend-icon="mdi-focus-field"></v-list-item>
      <v-list-item link title="Variants" to="/core/variants" prepend-icon="mdi-focus-field"></v-list-item>
      <v-list-item link title="Arrays" to="/core/arrays" prepend-icon="mdi-focus-field"></v-list-item>
    </div>

    <v-list-subheader title="Entity relationships" class="mt-2 clickable" @click="showEntityRelItems = !showEntityRelItems"></v-list-subheader>

    <div v-if="showEntityRelItems">
      <v-list-item link title="Verticals" to="/digital-marketing/verticals" prepend-icon="mdi-domain"></v-list-item>
      <v-list-item link title="Campaigns" to="/digital-marketing/campaigns" prepend-icon="mdi-bullhorn-outline"></v-list-item>
      <v-list-item link title="Campaign Perf" to="/digital-marketing/campaign-performance" prepend-icon="mdi-poll"></v-list-item>
    </div>

    <v-list-subheader title="Advanced tables" class="mt-2 clickable" @click="showAdvTableItems = !showAdvTableItems"></v-list-subheader>

    <div v-if="showAdvTableItems">
      <v-list-item link title="Optimizer" to="/digital-marketing/campaign-optimizer" prepend-icon="mdi-counter"></v-list-item>
    </div>

    <v-list-subheader title="User data retention" class="mt-2 clickable" @click="showUserDataRetItems = !showUserDataRetItems"></v-list-subheader> 

    <div v-if="showUserDataRetItems">
      <v-list-item link title="Authors" to="/publisher/authors" prepend-icon="mdi-book-open-outline"></v-list-item>
      <v-list-item link title="Books" to="/publisher/books" prepend-icon="mdi-book-open-outline"></v-list-item>
    </div>

  </v-list>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useJsonLs } from 'lys-vue'

const showTypeItems = ref(true)
const showEntityRelItems = ref(true)
const showAdvTableItems = ref(true)
const showUserDataRetItems = ref(true)

useJsonLs({
  lsKey: 'left_nav_list',
  refs: {
    showTypeItems,
    showEntityRelItems,
    showAdvTableItems,
    showUserDataRetItems,
  },
})

</script>
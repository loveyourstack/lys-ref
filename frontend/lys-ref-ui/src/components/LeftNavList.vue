<template>
  <v-list density="compact" class="nav-list nav-list-left">
    <v-list-item link :title="$t('home.nav')" to="/home" prepend-icon="mdi-home"></v-list-item>

    <v-list-subheader :title="$t('nav_header.type_handling')" class="mt-2 clickable" @click="showTypeItems = !showTypeItems"></v-list-subheader>

    <div v-if="showTypeItems">
      <v-list-item link :title="$t('mandatory_values.nav')" to="/core/mandatory-values" prepend-icon="mdi-focus-field"></v-list-item>
      <v-list-item link :title="$t('optional_values.nav')" to="/core/optional-values" prepend-icon="mdi-focus-field"></v-list-item>
      <v-list-item link :title="$t('default_values.nav')" to="/core/default-values" prepend-icon="mdi-focus-field"></v-list-item>
      <v-list-item link :title="$t('variants.nav')" to="/core/variants" prepend-icon="mdi-focus-field"></v-list-item>
      <v-list-item link :title="$t('arrays.nav')" to="/core/arrays" prepend-icon="mdi-focus-field"></v-list-item>
    </div>

    <v-list-subheader :title="$t('nav_header.entity_relationships')" class="mt-2 clickable" @click="showEntityRelItems = !showEntityRelItems"></v-list-subheader>

    <div v-if="showEntityRelItems">
      <v-list-item link :title="$t('verticals.nav')" to="/digital-marketing/verticals" prepend-icon="mdi-domain"></v-list-item>
      <v-list-item link :title="$t('campaigns.nav')" to="/digital-marketing/campaigns" prepend-icon="mdi-bullhorn-outline"></v-list-item>
      <v-list-item link :title="$t('campaign_perf.nav')" to="/digital-marketing/campaign-performance" prepend-icon="mdi-poll"></v-list-item>
    </div>

    <v-list-subheader :title="$t('nav_header.advanced_tables')" class="mt-2 clickable" @click="showAdvTableItems = !showAdvTableItems"></v-list-subheader>

    <div v-if="showAdvTableItems">
      <v-list-item link :title="$t('optimizer.nav')" to="/digital-marketing/campaign-optimizer" prepend-icon="mdi-counter"></v-list-item>
    </div>

    <v-list-subheader :title="$t('nav_header.ai')" class="mt-2 clickable" @click="showAiItems = !showAiItems"></v-list-subheader>

    <div v-if="showAiItems">
      <v-list-item link :title="$t('mcp_server.nav')" to="/digital-marketing/mcp-server" prepend-icon="mdi-robot"></v-list-item>
    </div>

    <v-list-subheader :title="$t('nav_header.user_data_retention')" class="mt-2 clickable" @click="showUserDataRetItems = !showUserDataRetItems"></v-list-subheader> 

    <div v-if="showUserDataRetItems">
      <v-list-item link :title="$t('authors.nav')" to="/publisher/authors" prepend-icon="mdi-book-open-outline"></v-list-item>
      <v-list-item link :title="$t('books.nav')" to="/publisher/books" prepend-icon="mdi-book-open-outline"></v-list-item>
    </div>

  </v-list>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useJsonLs } from 'lys-vue'

const showTypeItems = ref(true)
const showEntityRelItems = ref(true)
const showAdvTableItems = ref(true)
const showAiItems = ref(true)
const showUserDataRetItems = ref(true)

useJsonLs({
  lsKey: 'left_nav_list',
  refs: {
    showTypeItems,
    showEntityRelItems,
    showAdvTableItems,
    showAiItems,
    showUserDataRetItems,
  },
})

</script>
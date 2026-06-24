<template>
  <v-list density="compact" class="nav-list nav-list-left">
    <v-list-item link :title="$t('left_nav.home')" to="/home" prepend-icon="mdi-home"></v-list-item>

    <v-list-subheader :title="$t('left_nav.type_handling.title')" class="mt-2 clickable" @click="showTypeItems = !showTypeItems"></v-list-subheader>

    <div v-if="showTypeItems">
      <v-list-item link :title="$t('left_nav.type_handling.mandatory_values')" to="/core/mandatory-values" prepend-icon="mdi-focus-field"></v-list-item>
      <v-list-item link :title="$t('left_nav.type_handling.optional_values')" to="/core/optional-values" prepend-icon="mdi-focus-field"></v-list-item>
      <v-list-item link :title="$t('left_nav.type_handling.default_values')" to="/core/default-values" prepend-icon="mdi-focus-field"></v-list-item>
      <v-list-item link :title="$t('left_nav.type_handling.variants')" to="/core/variants" prepend-icon="mdi-focus-field"></v-list-item>
      <v-list-item link :title="$t('left_nav.type_handling.arrays')" to="/core/arrays" prepend-icon="mdi-focus-field"></v-list-item>
    </div>

    <v-list-subheader :title="$t('left_nav.entity_relationships.title')" class="mt-2 clickable" @click="showEntityRelItems = !showEntityRelItems"></v-list-subheader>

    <div v-if="showEntityRelItems">
      <v-list-item link :title="$t('left_nav.entity_relationships.verticals')" to="/digital-marketing/verticals" prepend-icon="mdi-domain"></v-list-item>
      <v-list-item link :title="$t('left_nav.entity_relationships.campaigns')" to="/digital-marketing/campaigns" prepend-icon="mdi-bullhorn-outline"></v-list-item>
      <v-list-item link :title="$t('left_nav.entity_relationships.campaign_perf')" to="/digital-marketing/campaign-performance" prepend-icon="mdi-poll"></v-list-item>
    </div>

    <v-list-subheader :title="$t('left_nav.advanced_tables.title')" class="mt-2 clickable" @click="showAdvTableItems = !showAdvTableItems"></v-list-subheader>

    <div v-if="showAdvTableItems">
      <v-list-item link :title="$t('left_nav.advanced_tables.optimizer')" to="/digital-marketing/campaign-optimizer" prepend-icon="mdi-counter"></v-list-item>
    </div>

    <v-list-subheader :title="$t('left_nav.ai.title')" class="mt-2 clickable" @click="showAiItems = !showAiItems"></v-list-subheader>

    <div v-if="showAiItems">
      <v-list-item link :title="$t('left_nav.ai.mcp_server')" to="/digital-marketing/mcp-server" prepend-icon="mdi-robot"></v-list-item>
    </div>

    <v-list-subheader :title="$t('left_nav.user_data_retention.title')" class="mt-2 clickable" @click="showUserDataRetItems = !showUserDataRetItems"></v-list-subheader> 

    <div v-if="showUserDataRetItems">
      <v-list-item link :title="$t('left_nav.user_data_retention.authors')" to="/publisher/authors" prepend-icon="mdi-book-open-outline"></v-list-item>
      <v-list-item link :title="$t('left_nav.user_data_retention.books')" to="/publisher/books" prepend-icon="mdi-book-open-outline"></v-list-item>
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
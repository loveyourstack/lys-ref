<template>
  <v-card class="py-4" color="surface-variant" rounded="lg" variant="tonal">
    <v-card-item>
      <template #prepend>
        <v-avatar class="ml-2 mr-4" icon="mdi-rocket-launch-outline" size="60" variant="tonal" />
      </template>

      <v-card-title class="text-headline-medium font-weight-bold">{{ $t("home.project_stack.title") }}</v-card-title>
      <v-card-subtitle class="text-body-large mb-3">{{ $t("home.project_stack.subtitle") }}</v-card-subtitle>

      <i18n-t scope="global" keypath="home.project_stack.p1" tag="div" class="text-body-large mb-3">
        <template #PostgreSQL>
          <a href="https://www.postgresql.org/" target="_blank" rel="noopener noreferrer">PostgreSQL</a>
        </template>
        <template #Go>
          <a href="https://go.dev/" target="_blank" rel="noopener noreferrer">Go</a>
        </template>
        <template #VueJs>
          <a href="https://vuejs.org/" target="_blank" rel="noopener noreferrer">Vue.js</a>
        </template>
        <template #Vuetify>
          <a href="https://vuetifyjs.com/" target="_blank" rel="noopener noreferrer">Vuetify</a>
        </template>
      </i18n-t>

      <div class="text-body-large">
        {{ $t("home.project_stack.p2") }}
        <ul>
          <i18n-t scope="global" keypath="home.project_stack.p2_list.item_1" tag="li">
            <template #lysRef>
              <a href="https://github.com/loveyourstack/lys-ref" target="_blank" rel="noopener noreferrer">lys-ref</a>
            </template>
          </i18n-t>

          <i18n-t scope="global" keypath="home.project_stack.p2_list.item_2" tag="li">
            <template #lys>
              <a href="https://github.com/loveyourstack/lys" target="_blank" rel="noopener noreferrer">lys</a>
            </template>
          </i18n-t>

          <i18n-t scope="global" keypath="home.project_stack.p2_list.item_3" tag="li">
            <template #connectors>
              <a href="https://github.com/loveyourstack/connectors" target="_blank" rel="noopener noreferrer">connectors</a>
            </template>
          </i18n-t>

          <i18n-t scope="global" keypath="home.project_stack.p2_list.item_4" tag="li">
            <template #lysVue>
              <a href="https://github.com/loveyourstack/lys-vue" target="_blank" rel="noopener noreferrer">lys-vue</a>
            </template>
            <template #npm>
              <a href="https://www.npmjs.com/package/lys-vue" target="_blank" rel="noopener noreferrer">npm</a>
            </template>
          </i18n-t>
        </ul>
      </div>

      <div class="text-body-large">
        {{ $t("home.project_stack.p3") }}
      </div>
    </v-card-item>
  </v-card>
</template>

<script lang="ts" setup>
</script>

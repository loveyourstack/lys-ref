<template>
  <v-card class="py-4" color="surface-variant" rounded="lg" variant="tonal">
    <v-card-item>
      <template #prepend>
        <v-avatar class="ml-2 mr-4" icon="mdi-rocket-launch-outline" size="60" variant="tonal" />
      </template>

      <v-card-title class="text-headline-medium font-weight-bold">Project stack and repositories</v-card-title>
      <v-card-subtitle class="text-body-large mb-3">Modern technologies with custom open source libraries on top.</v-card-subtitle>

      <div class="text-body-large mb-3">
        This project uses
        <a href="https://www.postgresql.org/" target="_blank" rel="noopener noreferrer">PostgreSQL</a> for the database, 
        <a href="https://go.dev/" target="_blank" rel="noopener noreferrer">Go</a> on the server, 
        and <a href="https://vuejs.org/" target="_blank" rel="noopener noreferrer">Vue.js</a> 
        with <a href="https://vuetifyjs.com/" target="_blank" rel="noopener noreferrer">Vuetify</a> components in the UI.
      </div>

      <div class="text-body-large">
        The LoveYourStack open source repositories relevant to this project are:
        <ul>
          <li>
            <a href="https://github.com/loveyourstack/lys-ref" target="_blank" rel="noopener noreferrer">lys-ref</a> for this project's code.
          </li>
          <li>
            <a href="https://github.com/loveyourstack/lys" target="_blank" rel="noopener noreferrer">lys</a> for the main backend library.
          </li>
          <li>
            <a href="https://github.com/loveyourstack/connectors" target="_blank" rel="noopener noreferrer">connectors</a> for the ECB and MaxMind connectors.
          </li>
          <li>
            <a href="https://github.com/loveyourstack/lys-vue" target="_blank" rel="noopener noreferrer">lys-vue</a> for the UI library, 
            which is available via <a href="https://www.npmjs.com/package/lys-vue" target="_blank" rel="noopener noreferrer">npm</a>.
          </li>
        </ul>
      </div>

      <div class="text-body-large">
        The cards below give an overview of the subheadings in the left and right navigation panels.
      </div>
    </v-card-item>
  </v-card>
</template>

<script lang="ts" setup>
</script>

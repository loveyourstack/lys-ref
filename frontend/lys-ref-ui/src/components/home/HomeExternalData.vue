<template>
  <v-card class="py-4" color="surface-variant" rounded="lg" variant="tonal">
    <v-card-item>
      <template #prepend>
        <v-avatar class="ml-2 mr-4" icon="mdi-currency-eur" size="60" variant="tonal" />
      </template>

      <v-card-title class="text-headline-medium font-weight-bold">External data</v-card-title>
      <v-card-subtitle class="text-body-large mb-3">Importing and managing external data sources.</v-card-subtitle>

      <div class="text-body-large mb-3">
        This section shows data from the European Central Bank's API and the MaxMind API.
      </div>

      <div class="text-body-large mb-3">
        In the backend, the it shows how to merge data efficiently using hash tables
        and how to stream from file directly into the database.
      </div>
    </v-card-item>
  </v-card>
</template>

<script lang="ts" setup>
</script>

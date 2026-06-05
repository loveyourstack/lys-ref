<template>
  <v-card class="py-4" color="surface-variant" rounded="lg" variant="tonal">
    <v-card-item>
      <template #prepend>
        <v-avatar class="ml-2 mr-4" icon="mdi-sitemap-outline" size="60" variant="tonal" />
      </template>

      <v-card-title class="text-headline-medium font-weight-bold">Parallel processing</v-card-title>
      <v-card-subtitle class="text-body-large mb-3">Improved data throughput.</v-card-subtitle>

      <div class="text-body-large mb-3">
        This section shows an implementation of job workflows, where each step is a fake backend command to manipulate data.
      </div>

      <div class="text-body-large">
        When a job is started, each step that can be executed in parallel runs concurrently, and structures are in place to manage failure gracefully.
      </div>
    </v-card-item>
  </v-card>
</template>

<script lang="ts" setup>
</script>

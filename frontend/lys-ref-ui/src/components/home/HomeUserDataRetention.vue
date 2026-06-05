<template>
  <v-card class="py-4" color="surface-variant" rounded="lg" variant="tonal">
    <v-card-item>
      <template #prepend>
        <v-avatar class="ml-2 mr-4" icon="mdi-book-open-outline" size="60" variant="tonal" />
      </template>

      <v-card-title class="text-headline-medium font-weight-bold">User data retention</v-card-title>
      <v-card-subtitle class="text-body-large mb-3">Techniques for preserving user data.</v-card-subtitle>

      <div class="text-body-large">
        This section demonstrates how to properly archive (soft delete) data including cascading effects, and how to track user changes using a database audit log.
      </div>
    </v-card-item>
  </v-card>
</template>

<script lang="ts" setup>
</script>

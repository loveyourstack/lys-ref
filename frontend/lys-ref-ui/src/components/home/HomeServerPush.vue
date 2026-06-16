<template>
  <v-card class="py-4" color="surface-variant" rounded="lg" variant="tonal">
    <v-card-item>
      <template #prepend>
        <v-avatar class="ml-2 mr-4" icon="mdi-web" size="60" variant="tonal" />
      </template>

      <v-card-title class="text-headline-medium font-weight-bold">Server push</v-card-title>
      <v-card-subtitle class="text-body-large mb-3">Real-time notifications and updates.</v-card-subtitle>

      <div class="text-body-large mb-3">
        This section demonstrates how data is pushed from database to app (via listen/notify) and from app to frontend (via websockets).
      </div>
    </v-card-item>
  </v-card>
</template>

<script lang="ts" setup>
</script>

<template>
  <v-card class="py-4" color="surface-variant" rounded="lg" variant="tonal">
    <v-card-item>
      <template #prepend>
        <v-avatar class="ml-2 mr-4" icon="mdi-robot" size="60" variant="tonal" />
      </template>

      <v-card-title class="text-headline-medium font-weight-bold">{{ $t("home.ai.title") }}</v-card-title>
      <v-card-subtitle class="text-body-large mb-3">{{ $t("home.ai.subtitle") }}</v-card-subtitle>

      <div class="text-body-large mb-3">
        {{ $t("home.ai.p1") }}
      </div>

    </v-card-item>
  </v-card>
</template>

<script lang="ts" setup>
</script>

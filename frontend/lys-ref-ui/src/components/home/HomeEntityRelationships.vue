<template>
  <v-card class="py-4" color="surface-variant" rounded="lg" variant="tonal">
    <v-card-item>
      <template #prepend>
        <v-avatar class="ml-2 mr-4" icon="mdi-domain" size="60" variant="tonal" />
      </template>

      <v-card-title class="text-headline-medium font-weight-bold">{{ $t("home.entity_relationships.title") }}</v-card-title>
      <v-card-subtitle class="text-body-large mb-3">{{ $t("home.entity_relationships.subtitle") }}</v-card-subtitle>

      <div class="text-body-large">
        {{ $t("home.entity_relationships.p1") }}
      </div>
    </v-card-item>
  </v-card>
</template>

<script lang="ts" setup>
</script>

<template>
  <v-card class="py-4" color="surface-variant" rounded="lg" variant="tonal">
    <v-card-item>
      <template #prepend>
        <v-avatar class="ml-2 mr-4" icon="mdi-currency-eur" size="60" variant="tonal" />
      </template>

      <v-card-title class="text-headline-medium font-weight-bold">{{ $t("home.external_data.title") }}</v-card-title>
      <v-card-subtitle class="text-body-large mb-3">{{ $t("home.external_data.subtitle") }}</v-card-subtitle>

      <div class="text-body-large mb-3">
        {{ $t("home.external_data.p1") }}
      </div>

      <div class="text-body-large mb-3">
        {{ $t("home.external_data.p2") }}
      </div>
    </v-card-item>
  </v-card>
</template>

<script lang="ts" setup>
</script>

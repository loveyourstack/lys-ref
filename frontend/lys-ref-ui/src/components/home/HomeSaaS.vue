<template>
  <v-card class="py-4" color="surface-variant" rounded="lg" variant="tonal">
    <v-card-item>
      <template #prepend>
        <v-avatar class="ml-2 mr-4" icon="mdi-web" size="60" variant="tonal" />
      </template>

      <v-card-title class="text-headline-medium font-weight-bold">SaaS / multi-tenancy</v-card-title>
      <v-card-subtitle class="text-body-large mb-3">Solving SaaS on the data level.</v-card-subtitle>

      <div class="text-body-large mb-3">
        This section shows how to use Postgres RLS (Row-Level Security) to implement multi-tenancy in a SaaS application.
      </div>

      <div class="text-body-large">
        It shows how to access the data both from an internal database connection, and also from a tenant application using a special RLS-enabled database connection.
      </div>
    </v-card-item>
  </v-card>
</template>

<script lang="ts" setup>
</script>

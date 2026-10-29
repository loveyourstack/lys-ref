<template>
  <v-container fluid>
    <v-responsive>
      <v-row>
        <v-col cols="12">
          <v-card variant="flat">
            <v-card-text>
              <span class="dt-title">{{ $t('schema_designer.title') }}</span>
              <div class="dt-subtitle mt-6 mb-4">{{ $t('schema_designer.p1') }}</div>

              <core-schema-canvas />

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
</script>

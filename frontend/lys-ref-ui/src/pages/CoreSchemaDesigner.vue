<template>
  <v-container fluid>
    <v-responsive>
      <v-row>
        <v-col cols="12">
          <v-card variant="flat">
            <v-card-text>
              <span class="dt-title">{{ $t('schema_designer.title') }}</span>

              <i18n-t scope="global" keypath="schema_designer.p1" tag="div" class="dt-subtitle mt-6 mb-4">
                <template #VueFlow>
                  <a href="https://vueflow.dev/" target="_blank" rel="noopener noreferrer">Vue-Flow</a>
                </template>
              </i18n-t>

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

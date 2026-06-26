<template>
  <v-container fluid>
    <v-responsive>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <v-row density="comfortable">
                <v-col class="mb-6">
                  <div class="dt-subtitle">{{ $t('internal_view.p1') }}</div>
                  <div class="dt-subtitle">{{ $t('internal_view.p2') }}</div>
                  <div class="dt-subtitle">{{ $t('internal_view.p3') }}</div>
                  <div class="dt-subtitle">{{ $t('internal_view.p4') }}</div>
                </v-col>
              </v-row>

              <supp-employee-table :internal="true" />

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <supp-product-table :internal="true" />

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
</script>

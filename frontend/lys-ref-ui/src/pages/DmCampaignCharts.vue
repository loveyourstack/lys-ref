<template>
  <v-container fluid>
    <v-responsive>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <v-row density="comfortable">
                <v-col class="mb-2">
                  <div class="dt-subtitle">Using the campaign data from the 'Entity relationships' section, demonstrates further chart types.</div>
                  <div class="dt-subtitle">The title and label text colors in the chart library have been overriden to respect the application-level light or dark mode.</div>
                </v-col>
              </v-row>

              <dm-camp-budget-charts />

              <dm-camp-perf-summ-chart />

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
</script>

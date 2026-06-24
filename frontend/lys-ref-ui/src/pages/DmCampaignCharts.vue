<template>
  <v-container fluid>
    <v-responsive>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <v-row density="comfortable">
                <v-col class="mb-2">
                  <div class="dt-subtitle">{{ $t('charts.campaign_charts.p1') }}</div>
                  <div class="dt-subtitle">{{ $t('charts.campaign_charts.p2') }}</div>
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

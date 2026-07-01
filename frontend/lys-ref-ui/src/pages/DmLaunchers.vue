<template>
  <v-container fluid>
    <v-responsive>

      <v-row density="compact" class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text class="pb-0">
              <span class="dt-title">{{ $t('launchers.title') }}</span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row density="comfortable">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text class="pb-0">
              <span class="dt-subtitle">{{ $t('launchers.p1') }}</span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>
              <v-tabs v-model="selectedTab" density="comfortable" class="mb-4">
                <v-tab value="fb">Facebook</v-tab>
                <v-tab value="gads">Google Ads</v-tab>
              </v-tabs>

              <v-window v-model="selectedTab">

                <v-window-item value="fb">
                  <dm-launch-fb-table />
                </v-window-item>

                <v-window-item value="gads">
                  <dm-launch-gads-table />
                </v-window-item>

              </v-window>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useJsonLs } from 'lys-vue'

const selectedTab = ref('fb')

useJsonLs({
  lsKey: 'launchers',
  refs: {
    selectedTab,
  },
})

</script>

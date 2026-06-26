<template>
  <v-container fluid>
    <v-responsive>
      <v-row density="compact" class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text class="pb-0">
              <span class="dt-title">{{ $t('notifications.title') }}</span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <div class="dt-subtitle">{{ $t('notifications.p1') }}</div>
              <div class="dt-subtitle">{{ $t('notifications.p2') }}</div>
              <div class="dt-subtitle">{{ $t('notifications.p3') }}</div>

              <v-list>
                <v-list-item>
                  <v-btn color="info" @click="addNotification('Info')">{{ $t('notifications.add_info') }}</v-btn>
                </v-list-item>
                <v-list-item>
                  <v-btn color="warning" @click="addNotification('Warning')">{{ $t('notifications.add_warning') }}</v-btn>
                </v-list-item>
              </v-list>

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import ax from '@/api'

const baseUrl = '/a/system/notifications/add-fake'

function addNotification(type: string) {
  ax.post(`${baseUrl}?type=${type}`)
}
</script>

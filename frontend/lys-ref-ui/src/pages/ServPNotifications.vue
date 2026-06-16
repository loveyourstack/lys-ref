<template>
  <v-container fluid>
    <v-responsive>
      <v-row density="compact" class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text class="pb-0">
              <span class="dt-title">Server push: notifications</span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <div class="dt-subtitle">Notifications are added in the database so they are persistent and can be used by any component. 
                They are pushed to the app using listen/notify, and from the app to the frontend via websockets.</div>

              <div class="dt-subtitle">If you allow browser notifications for this site, you will see a pop-up notification when you click one of the buttons below.</div>

              <v-list>
                <v-list-item>
                  <v-btn color="info" @click="addNotification('Info')">Add Info notification</v-btn>
                </v-list-item>
                <v-list-item>
                  <v-btn color="warning" @click="addNotification('Warning')">Add Warning notification</v-btn>
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

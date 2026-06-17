<template>
  <v-menu v-model="menuOpen" :close-on-content-click="false" location="start">
    <template v-slot:activator="{ props }">
      <v-btn class="text-none mr-2 opacity-70" stacked v-bind="props">
        <v-badge v-if="notsStore.unreadCount > 0" location="top right" :content="notsStore.unreadCount" color="primary">
          <v-icon icon="mdi-bell-outline"></v-icon>
        </v-badge>
        <v-icon v-else icon="mdi-bell-outline"></v-icon>
      </v-btn>
    </template>

    <v-card class="mr-2" max-width="400">
      <v-list density="compact" lines="two">
        <div class="d-flex">
          <v-list-subheader>NOTIFICATIONS</v-list-subheader>

          <v-spacer></v-spacer>

          <v-menu :close-on-content-click="false">
            <template v-slot:activator="{ props }">
              <v-btn icon="mdi-dots-vertical" flat v-bind="props"></v-btn>
            </template>

            <v-card>
              <v-list density="compact">
                <v-list-item class="clickable" @click="notsStore.setAllRead()">
                  <v-btn prepend-icon="mdi-email-open-outline" style="letter-spacing: normal !important;" 
                    :loading="notsStore.markingAllAsRead">Mark all as read</v-btn>
                </v-list-item>
              </v-list>
            </v-card>
           </v-menu>
        </div>
        
        <v-list-item v-for="item in notsStore.items" :key="item.id" :class="item.is_read ? '' : 'bg-primary-lighten-5'">

          <template v-slot:prepend>
            <v-icon :icon="getIconDetails(item.not_type).icon" :color="getIconDetails(item.not_type).color"></v-icon>
          </template>

          <v-list-item-title v-text="item.not_type"></v-list-item-title>
          <v-list-item-subtitle v-text="item.message"></v-list-item-subtitle>
        </v-list-item>

      </v-list>
    </v-card>
  </v-menu>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useNotsStore } from '@/stores/notifications'

const notsStore = useNotsStore()

const menuOpen = ref(false)

function getIconDetails(type: string): { icon: string, color: string } {
  switch (type) {
    case 'Info':
      return { icon: 'mdi-information-outline', color: 'blue' }
    case 'Warning':
      return { icon: 'mdi-alert-outline', color: 'orange' }
    default:
      return { icon: 'mdi-bell-outline', color: 'grey' }
  }
}

watch(menuOpen, isOpen => {
  if (isOpen) {
    notsStore.loadItems()
  }
})

onMounted(() => {
  notsStore.loadUnreadCount()
})
</script>

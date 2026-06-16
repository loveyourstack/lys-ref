import { ref } from 'vue'
import { defineStore } from 'pinia'
import { useWebSocket } from '@vueuse/core'
import { notify } from 'lys-vue'
import { useAppStore } from '@/stores/app'

type IncomingNotification = {
  type: string
  body: string
}

function toNotificationsWsUrl(apiBaseUrl: string, token: string): string {
  const url = new URL('/a/ws/notifications/register', apiBaseUrl)
  url.searchParams.set('token', token)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  return url.toString()
}

export const useNotsStore = defineStore('notifications', () => {
  const appStore = useAppStore()

  const wsUrl = ref('')
  const isStarted = ref(false)

  const { status, close } = useWebSocket(wsUrl, {
    immediate: false,
    autoReconnect: { 
      retries: 3,
      delay: 2000,
      onFailed() {
        console.log('autoReconnect websocket failed after 3 retries')
      },
    },
    onDisconnected(_, event) {
      if (event.code === 4429) {
        // server rejected due to max connections — stop reconnecting
        close()
      }
    },
    onMessage(_, event) {
      try {
        const msg = JSON.parse(String(event.data)) as IncomingNotification

        if (!msg?.type || !msg?.body) return

        notify(appStore.company, msg.type, appStore.logoUrl, msg.body)
      } catch {
        // ignore malformed message
      }
    },
  })

  function start() {
    const token = sessionStorage.getItem('token')
    if (!token || isStarted.value) return

    wsUrl.value = toNotificationsWsUrl(import.meta.env.VITE_API_URL, token)
    isStarted.value = true
    // no open(): wsUrl change triggers useWebSocket to connect
  }

  function stop() {
    isStarted.value = false
    close()
  }

  return { status, start, stop }
})
import { ref } from 'vue'
import { defineStore } from 'pinia'
import { useWebSocket } from '@vueuse/core'
import { fetchOnce, notify, type GetMetadata } from 'lys-vue'
import { useAppStore } from '@/stores/app'
import { type Notification } from '@/types/system'
import ax from '@/api'

type IncomingWsNotification = {
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
  const baseUrl = '/a/system/notifications'  

  const isStarted = ref(false)
  const items = ref<Notification[]>([])
  const listHasMore = ref(false)
  const listPage = ref(1)
  const markingAllAsRead = ref(false)
  const wsUrl = ref('')
  const unreadCount = ref(0)

  const { status: wsStatus, close } = useWebSocket(wsUrl, {
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
        const msg = JSON.parse(String(event.data)) as IncomingWsNotification

        if (!msg?.type || !msg?.body) return

        notify(appStore.company, msg.type, appStore.logoUrl, msg.body)
        loadUnreadCount()
      } catch {
        // ignore malformed message
      }
    },
  })

  function loadItems() {
    const metaData = ref<GetMetadata>({count: 0, total_count: 0, total_count_is_estimated: false})
    const myUrl = `${baseUrl}?xpage=${listPage.value}&xper_page=10`
    fetchOnce({ ax, myUrl, result: items, metaData, onSuccess: () => {

      // determine whether there are more pages to load
      if (metaData.value && metaData.value.total_count && metaData.value.total_count > listPage.value * 10) {
        listHasMore.value = true
      } else {
        listHasMore.value = false
      }

      // mark any unread loaded notifications as read
      const unreadIds = items.value.filter(i => !i.is_read).map(i => i.id)
      if (unreadIds.length > 0) {
        setIdsRead(unreadIds)
      }
    }})
  }

  function loadUnreadCount() {
    const myUrl = `${baseUrl}/unread-count`
    fetchOnce({ ax, myUrl, result: unreadCount })
  }

  function setAllRead() {
    markingAllAsRead.value = true
    ax.patch(`${baseUrl}/set-all-read`).then(() => {
      loadUnreadCount()
    }).finally(() => {
      markingAllAsRead.value = false
    })
  }

  function setIdsRead(ids: number[]) {
    ax.patch(`${baseUrl}/set-read`, { ids }).then(() => {
      loadUnreadCount()
    })
  }

  function wsStart() {
    const token = sessionStorage.getItem('token')
    if (!token || isStarted.value) return

    wsUrl.value = toNotificationsWsUrl(import.meta.env.VITE_API_URL, token)
    isStarted.value = true
    // no open(): wsUrl change triggers useWebSocket to connect
  }

  function wsStop() {
    isStarted.value = false
    close()
  }

  return { items, listHasMore, listPage, markingAllAsRead, unreadCount, wsStatus, wsStart, wsStop,
    loadItems, loadUnreadCount, setAllRead, setIdsRead,
   }
})
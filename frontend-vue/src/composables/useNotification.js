import { ref } from 'vue'

const API_BASE_URL = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api`

// Shared state across components
const notifications = ref([])
const unreadCount = ref(0)
const isOpen = ref(false)

let eventSource = null
let sseReconnectAttempt = 0
const SSE_MAX_RECONNECT = 10
let sseReconnectTimer = null

export function useNotification() {
  const loadNotifications = async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/notifications?limit=20`)
      if (!res.ok) return
      const data = await res.json()
      notifications.value = data.data || []
      unreadCount.value = data.unread_count || 0
    } catch (err) {
      console.error('[Notification] Load failed:', err)
    }
  }

  const markAsRead = async (id) => {
    try {
      await fetch(`${API_BASE_URL}/notifications/${id}/read`, { method: 'PUT' })
      const n = notifications.value.find(n => n.id === id)
      if (n && !n.is_read) {
        n.is_read = true
        unreadCount.value = Math.max(0, unreadCount.value - 1)
      }
    } catch (err) {
      console.error('[Notification] Mark read failed:', err)
    }
  }

  const markAllAsRead = async () => {
    try {
      await fetch(`${API_BASE_URL}/notifications/read-all`, { method: 'PUT' })
      notifications.value.forEach(n => n.is_read = true)
      unreadCount.value = 0
    } catch (err) {
      console.error('[Notification] Mark all read failed:', err)
    }
  }

  const connectSSE = () => {
    if (eventSource) return
    try {
      eventSource = new EventSource(`${API_BASE_URL}/notifications/stream`)

      eventSource.onopen = () => {
        sseReconnectAttempt = 0 // reset on successful connection
      }

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          if (data.type === 'notification') {
            notifications.value.unshift({
              id: Date.now(),
              content_id: data.data.content_id,
              title: data.data.title,
              summary: data.data.summary,
              innovation_score: data.data.innovation_score,
              depth_score: data.data.depth_score,
              decision: data.data.decision,
              is_read: false,
              created_at: new Date().toISOString(),
            })
            unreadCount.value++
            sendSystemNotification(data.data.title, data.data.summary)
          }
        } catch (e) {
          console.warn('[Notification] Failed to parse SSE message:', e)
        }
      }

      eventSource.onerror = () => {
        disconnectSSE()
        if (sseReconnectAttempt < SSE_MAX_RECONNECT) {
          const delay = Math.min(1000 * Math.pow(2, sseReconnectAttempt), 60000) // max 60s
          sseReconnectAttempt++
          console.warn(`[Notification] SSE reconnecting in ${delay}ms (attempt ${sseReconnectAttempt}/${SSE_MAX_RECONNECT})`)
          sseReconnectTimer = setTimeout(connectSSE, delay)
        } else {
          console.error('[Notification] SSE max reconnect attempts reached, giving up')
        }
      }
    } catch (e) {
      console.error('[Notification] Failed to create EventSource:', e)
    }
  }

  const disconnectSSE = () => {
    if (sseReconnectTimer) {
      clearTimeout(sseReconnectTimer)
      sseReconnectTimer = null
    }
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
  }

  const sendSystemNotification = async (title, body) => {
    // Try Tauri notification first (only available in Tauri desktop app)
    if (window.__TAURI_INTERNALS__) {
      try {
        const { sendNotification, isPermissionGranted, requestPermission } = await import('@tauri-apps/plugin-notification')
        let permission = await isPermissionGranted()
        if (!permission) {
          permission = (await requestPermission()) === 'granted'
        }
        if (permission) {
          sendNotification({ title: `Junk Filter: ${title}`, body: body || '' })
          return
        }
      } catch { /* Tauri plugin not available */ }
    }

    // Fallback to browser notification
    if ('Notification' in window) {
      if (Notification.permission === 'granted') {
        new Notification(`Junk Filter: ${title}`, { body: body || '' })
      } else if (Notification.permission !== 'denied') {
        const perm = await Notification.requestPermission()
        if (perm === 'granted') {
          new Notification(`Junk Filter: ${title}`, { body: body || '' })
        }
      }
    }
  }

  const toggle = () => {
    isOpen.value = !isOpen.value
  }

  return {
    notifications,
    unreadCount,
    isOpen,
    loadNotifications,
    markAsRead,
    markAllAsRead,
    connectSSE,
    disconnectSSE,
    toggle,
  }
}

import { ref, readonly } from 'vue'

const toasts = ref([])

export function useToast() {
  const show = (message, type = 'success', duration = 3000) => {
    const id = Date.now()
    const toast = { id, message, type }

    toasts.value.push(toast)

    if (duration > 0) {
      setTimeout(() => {
        dismiss(id)
      }, duration)
    }

    return id
  }

  const dismiss = (id) => {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  return {
    toasts: readonly(toasts),
    show,
    dismiss,
  }
}

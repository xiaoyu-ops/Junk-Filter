import { ref } from 'vue'

export function useDetailDrawer() {
  const isOpen = ref(false)
  const selectedCard = ref(null)

  const openDrawer = (card) => {
    selectedCard.value = card
    isOpen.value = true
  }

  const closeDrawer = () => {
    isOpen.value = false
    selectedCard.value = null
  }

  return {
    isOpen,
    selectedCard,
    openDrawer,
    closeDrawer,
  }
}

import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { cnyToUsdt, fetchCnyPerUsdt, getDefaultCnyPerUsdt } from '@/utils/exchangeRate'

const REFRESH_INTERVAL_MS = 60_000

export function useCnyUsdtRate(active = () => true) {
  const cnyPerUsdt = ref(getDefaultCnyPerUsdt())
  const loading = ref(false)
  const lastUpdated = ref<Date | null>(null)
  let refreshTimer: number | undefined

  const rateDisplay = computed(() => cnyPerUsdt.value.toFixed(4))

  async function refreshRate() {
    if (!active()) return

    loading.value = true
    try {
      cnyPerUsdt.value = await fetchCnyPerUsdt()
      lastUpdated.value = new Date()
    } catch {
      if (!lastUpdated.value) {
        cnyPerUsdt.value = getDefaultCnyPerUsdt()
      }
    } finally {
      loading.value = false
    }
  }

  function convertCnyToUsdt(cnyAmount: number | null | undefined): number {
    if (!cnyAmount || cnyAmount <= 0) return 0
    return cnyToUsdt(cnyAmount, cnyPerUsdt.value)
  }

  function formatUsdtAmount(cnyAmount: number | null | undefined, digits = 2): string {
    return convertCnyToUsdt(cnyAmount).toFixed(digits)
  }

  function startAutoRefresh() {
    stopAutoRefresh()
    refreshTimer = window.setInterval(() => {
      void refreshRate()
    }, REFRESH_INTERVAL_MS)
  }

  function stopAutoRefresh() {
    if (refreshTimer !== undefined) {
      window.clearInterval(refreshTimer)
      refreshTimer = undefined
    }
  }

  onMounted(() => {
    void refreshRate()
    startAutoRefresh()
  })

  onUnmounted(() => {
    stopAutoRefresh()
  })

  watch(active, (isActive) => {
    if (isActive) {
      void refreshRate()
      startAutoRefresh()
      return
    }
    stopAutoRefresh()
  })

  return {
    cnyPerUsdt,
    loading,
    lastUpdated,
    rateDisplay,
    refreshRate,
    convertCnyToUsdt,
    formatUsdtAmount,
  }
}

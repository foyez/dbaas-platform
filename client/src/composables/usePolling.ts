import { onUnmounted, ref } from 'vue'

/**
 * Repeatedly calls fn every intervalMs, using recursive setTimeout rather
 * than setInterval so a slow request can never overlap with the next tick.
 * Stops automatically when the component unmounts.
 */
export function usePolling(fn: () => void | Promise<void>, intervalMs: number) {
  const isPolling = ref(false)
  let timer: ReturnType<typeof setTimeout> | null = null

  async function tick() {
    try {
      await fn()
    } finally {
      if (isPolling.value) {
        timer = setTimeout(tick, intervalMs)
      }
    }
  }

  function start() {
    if (isPolling.value) return

    isPolling.value = true
    timer = setTimeout(tick, intervalMs)
  }

  function stop() {
    isPolling.value = false

    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  onUnmounted(stop)

  return { start, stop, isPolling }
}
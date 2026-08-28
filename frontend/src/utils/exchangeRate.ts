const DEFAULT_CNY_PER_USDT = 7.2
const FETCH_TIMEOUT_MS = 8000

async function fetchWithTimeout(url: string, timeoutMs = FETCH_TIMEOUT_MS): Promise<Response> {
  const controller = new AbortController()
  const timer = window.setTimeout(() => controller.abort(), timeoutMs)
  try {
    return await fetch(url, { signal: controller.signal })
  } finally {
    window.clearTimeout(timer)
  }
}

async function fetchUsdtCnyFromCoinGecko(): Promise<number> {
  const response = await fetchWithTimeout(
    'https://api.coingecko.com/api/v3/simple/price?ids=tether&vs_currencies=cny',
  )
  if (!response.ok) {
    throw new Error(`CoinGecko ${response.status}`)
  }
  const data = await response.json() as { tether?: { cny?: number } }
  const rate = data.tether?.cny
  if (!rate || rate <= 0) {
    throw new Error('CoinGecko rate unavailable')
  }
  return rate
}

async function fetchUsdCnyRate(): Promise<number> {
  const response = await fetchWithTimeout(
    'https://api.exchangerate-api.com/v4/latest/USD',
  )
  if (!response.ok) {
    throw new Error(`ExchangeRate ${response.status}`)
  }
  const data = await response.json() as { rates?: { CNY?: number } }
  const rate = data.rates?.CNY
  if (!rate || rate <= 0) {
    throw new Error('USD/CNY rate unavailable')
  }
  return rate
}

export async function fetchCnyPerUsdt(): Promise<number> {
  try {
    return await fetchUsdtCnyFromCoinGecko()
  } catch {
    return fetchUsdCnyRate()
  }
}

export function cnyToUsdt(cnyAmount: number, cnyPerUsdt: number): number {
  if (!cnyAmount || cnyAmount <= 0 || !cnyPerUsdt || cnyPerUsdt <= 0) {
    return 0
  }
  return cnyAmount / cnyPerUsdt
}

export function getDefaultCnyPerUsdt(): number {
  return DEFAULT_CNY_PER_USDT
}

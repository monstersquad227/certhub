import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  connectSolanaWallet,
  disconnectSolanaWallet,
  formatWalletAddress,
  getSolanaProvider,
  isSolanaWalletAvailable,
  transferUsdt,
} from '@/utils/solana'

export function useSolanaWallet() {
  const walletAvailable = ref(false)
  const connected = ref(false)
  const walletAddress = ref<string | null>(null)
  const connecting = ref(false)
  const transferring = ref(false)

  const displayAddress = computed(() => {
    if (!walletAddress.value) return ''
    return formatWalletAddress(walletAddress.value)
  })

  let provider = getSolanaProvider()

  function syncProviderState() {
    provider = getSolanaProvider()
    walletAvailable.value = isSolanaWalletAvailable()
    if (provider?.isConnected && provider.publicKey) {
      connected.value = true
      walletAddress.value = provider.publicKey.toBase58()
      return
    }
    connected.value = false
    walletAddress.value = null
  }

  async function connect(onlyIfTrusted = false) {
    provider = getSolanaProvider()
    if (!provider) {
      throw new Error('未检测到 Solana 钱包，请安装 Phantom 或 Solflare')
    }

    connecting.value = true
    try {
      walletAddress.value = await connectSolanaWallet(provider, onlyIfTrusted)
      connected.value = true
    } finally {
      connecting.value = false
    }
  }

  async function disconnect() {
    provider = getSolanaProvider()
    if (provider) {
      await disconnectSolanaWallet(provider)
    }
    connected.value = false
    walletAddress.value = null
  }

  async function sendUsdt(recipientAddress: string, amountUsdt: number) {
    provider = getSolanaProvider()
    if (!provider?.publicKey) {
      throw new Error('请先连接钱包')
    }

    transferring.value = true
    try {
      return await transferUsdt(provider, recipientAddress, amountUsdt)
    } finally {
      transferring.value = false
    }
  }

  function handleAccountChanged(publicKey: unknown) {
    if (publicKey && typeof publicKey === 'object' && 'toBase58' in publicKey) {
      walletAddress.value = (publicKey as { toBase58: () => string }).toBase58()
      connected.value = true
      return
    }
    connected.value = false
    walletAddress.value = null
  }

  function handleDisconnect() {
    connected.value = false
    walletAddress.value = null
  }

  onMounted(() => {
    syncProviderState()
    provider = getSolanaProvider()
    provider?.on?.('connect', handleAccountChanged)
    provider?.on?.('accountChanged', handleAccountChanged)
    provider?.on?.('disconnect', handleDisconnect)
  })

  onUnmounted(() => {
    provider = getSolanaProvider()
    provider?.on?.('disconnect', handleDisconnect)
  })

  return {
    walletAvailable,
    connected,
    walletAddress,
    displayAddress,
    connecting,
    transferring,
    connect,
    disconnect,
    sendUsdt,
    syncProviderState,
  }
}

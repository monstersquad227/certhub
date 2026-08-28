import type { PublicKey, Transaction } from '@solana/web3.js'

export interface SolanaWalletProvider {
  isPhantom?: boolean
  isSolflare?: boolean
  isConnected: boolean
  publicKey: PublicKey | null
  connect: (options?: { onlyIfTrusted?: boolean }) => Promise<{ publicKey: PublicKey }>
  disconnect: () => Promise<void>
  signAndSendTransaction: (
    transaction: Transaction,
    options?: { skipPreflight?: boolean; maxRetries?: number },
  ) => Promise<{ signature: string }>
  on?: (event: string, handler: (...args: unknown[]) => void) => void
}

declare global {
  interface Window {
    phantom?: { solana?: SolanaWalletProvider }
    solana?: SolanaWalletProvider
    solflare?: SolanaWalletProvider
  }
}

export {}

import {
  Connection,
  PublicKey,
  Transaction,
} from '@solana/web3.js'
import {
  createAssociatedTokenAccountIdempotentInstruction,
  createTransferInstruction,
  getAccount,
  getAssociatedTokenAddressSync,
} from '@solana/spl-token'
import type { SolanaWalletProvider } from '@/types/solana'

export const USDT_MINT = new PublicKey('Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB')
export const USDT_DECIMALS = 6

const DEFAULT_RPC_ENDPOINTS = [
  import.meta.env.VITE_SOLANA_RPC_URL,
  'https://solana-rpc.publicnode.com',
  'https://api.mainnet-beta.solana.com',
].filter((url): url is string => Boolean(url))

export const SOLANA_RPC_URL = DEFAULT_RPC_ENDPOINTS[0]

async function createSolanaConnection(): Promise<Connection> {
  let lastError: unknown

  for (const endpoint of [...new Set(DEFAULT_RPC_ENDPOINTS)]) {
    try {
      const connection = new Connection(endpoint, 'confirmed')
      await connection.getLatestBlockhash('confirmed')
      return connection
    } catch (error) {
      lastError = error
    }
  }

  throw lastError instanceof Error
    ? lastError
    : new Error('无法连接 Solana RPC 节点，请稍后重试')
}

export function getSolanaProvider(): SolanaWalletProvider | null {
  if (typeof window === 'undefined') return null
  if (window.phantom?.solana?.isPhantom) return window.phantom.solana
  if (window.solana?.isPhantom) return window.solana
  if (window.solflare?.isSolflare) return window.solflare
  return null
}

export function isSolanaWalletAvailable(): boolean {
  return getSolanaProvider() !== null
}

export function formatWalletAddress(address: string, chars = 4): string {
  if (address.length <= chars * 2) return address
  return `${address.slice(0, chars)}...${address.slice(-chars)}`
}

export function toUsdtBaseUnits(amountUsdt: number): bigint {
  return BigInt(Math.round(amountUsdt * 10 ** USDT_DECIMALS))
}

export async function connectSolanaWallet(
  provider: SolanaWalletProvider,
  onlyIfTrusted = false,
): Promise<string> {
  const response = await provider.connect({ onlyIfTrusted })
  return response.publicKey.toBase58()
}

export async function disconnectSolanaWallet(provider: SolanaWalletProvider): Promise<void> {
  await provider.disconnect()
}

export async function transferUsdt(
  provider: SolanaWalletProvider,
  recipientAddress: string,
  amountUsdt: number,
): Promise<string> {
  if (!provider.publicKey) {
    throw new Error('钱包未连接')
  }

  const connection = await createSolanaConnection()
  const fromPubkey = provider.publicKey
  const toPubkey = new PublicKey(recipientAddress)
  const fromAta = getAssociatedTokenAddressSync(USDT_MINT, fromPubkey)
  const toAta = getAssociatedTokenAddressSync(USDT_MINT, toPubkey)

  const transaction = new Transaction()

  try {
    await getAccount(connection, toAta)
  } catch {
    transaction.add(
      createAssociatedTokenAccountIdempotentInstruction(
        fromPubkey,
        toAta,
        toPubkey,
        USDT_MINT,
      ),
    )
  }

  transaction.add(
    createTransferInstruction(
      fromAta,
      toAta,
      fromPubkey,
      toUsdtBaseUnits(amountUsdt),
    ),
  )

  const { blockhash, lastValidBlockHeight } = await connection.getLatestBlockhash()
  transaction.recentBlockhash = blockhash
  transaction.feePayer = fromPubkey

  const { signature } = await provider.signAndSendTransaction(transaction)

  const confirmation = await connection.confirmTransaction(
    { signature, blockhash, lastValidBlockHeight },
    'confirmed',
  )
  if (confirmation.value.err) {
    throw new Error('链上交易执行失败')
  }

  return signature
}

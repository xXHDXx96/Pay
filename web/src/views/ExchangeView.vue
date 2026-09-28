<script setup lang="ts">
import { ref } from 'vue'
import {
  CurrencyDollarIcon,
  QrCodeIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  ArrowLeftIcon,
} from '@heroicons/vue/24/outline'
import { paymentApi, ExchangeQuoteRequest, ExchangeRedeemRequest, CryptoRedemption } from '@/api/client'

const activeTab = ref<'quote' | 'redeem' | 'history'>('quote')
const loading = ref(false)
const error = ref<string | null>(null)
const quoteResult = ref<CryptoRedemption | null>(null)
const redeemResult = ref<CryptoRedemption | null>(null)

const quoteForm = ref<ExchangeQuoteRequest>({
  asset: 'USDT',
  fiat_cny: 0,
  order_id: '',
})

const redeemForm = ref<ExchangeRedeemRequest>({
  order_id: '',
  wallet_addr: '',
  asset: 'USDT',
  network: 'TRON',
})

const networks = [
  { value: 'TRON', label: 'TRON (USDT-TRON)' },
  { value: 'ERC20', label: 'Ethereum (USDT-ERC20)' },
  { value: 'BEP20', label: 'BSC (USDT-BEP20)' },
]

const assets = [
  { value: 'USDT', label: 'USDT (Tether)' },
  { value: 'USDC', label: 'USDC (USD Coin)' },
  { value: 'BTC', label: 'BTC (Bitcoin)' },
  { value: 'ETH', label: 'ETH (Ethereum)' },
]

async function getQuote() {
  loading.value = true
  error.value = null
  quoteResult.value = null
  try {
    const res = await paymentApi.cryptoQuote(quoteForm.value)
    quoteResult.value = res.data
    redeemForm.value.order_id = String(res.data.order_id)
    redeemForm.value.asset = res.data.crypto_asset
  } catch (err: any) {
    error.value = err.response?.data || err.message || 'Failed to get quote'
  } finally {
    loading.value = false
  }
}

async function doRedeem() {
  loading.value = true
  error.value = null
  redeemResult.value = null
  try {
    const res = await paymentApi.cryptoRedeem(redeemForm.value)
    redeemResult.value = res.data
  } catch (err: any) {
    error.value = err.response?.data || err.message || 'Failed to redeem'
  } finally {
    loading.value = false
  }
}

function getStatusColor(status: string) {
  if (status === 'CONFIRMED') return 'bg-green-100 text-green-800'
  if (status === 'PENDING') return 'bg-yellow-100 text-yellow-800'
  if (status === 'BROADCASTED') return 'bg-blue-100 text-blue-800'
  return 'bg-red-100 text-red-800'
}

function formatDate(dateStr: string | undefined): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex gap-4 mb-6">
      <button
        @click="activeTab = 'quote'"
        :class="activeTab === 'quote' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-700'"
        class="px-4 py-2 rounded-lg font-medium transition-colors"
      >
        Get Quote
      </button>
      <button
        @click="activeTab = 'redeem'"
        :class="activeTab === 'redeem' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-700'"
        class="px-4 py-2 rounded-lg font-medium transition-colors"
      >
        Redeem
      </button>
    </div>

    <div v-if="error" class="bg-red-50 text-red-700 p-3 rounded-lg">
      {{ error }}
    </div>

    <div v-if="activeTab === 'quote'" class="space-y-6">
      <div class="bg-white rounded-xl border border-gray-200 p-6">
        <h3 class="text-lg font-semibold mb-4">Get Crypto Quote</h3>
        <p class="text-sm text-gray-500 mb-4">
          Get a quote for exchanging fiat (CNY) to cryptocurrency.
        </p>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Asset *</label>
            <select
              v-model="quoteForm.asset"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            >
              <option v-for="asset in assets" :key="asset.value" :value="asset.value">
                {{ asset.label }}
              </option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Amount (CNY) *</label>
            <input
              v-model.number="quoteForm.fiat_cny"
              type="number"
              step="0.01"
              min="0.01"
              placeholder="0.00"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Order ID</label>
            <input
              v-model="quoteForm.order_id"
              type="text"
              placeholder="Optional (link to payment order)"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>
        </div>

        <div class="mt-6">
          <button
            @click="getQuote"
            :disabled="loading || quoteForm.fiat_cny <= 0"
            class="btn-primary"
          >
            <span v-if="loading" class="flex items-center gap-2">
              <svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              Quoting...
            </span>
            <span v-else>Get Quote</span>
          </button>
        </div>
      </div>

      <div v-if="quoteResult" class="bg-white rounded-xl border border-gray-200 p-6">
        <h3 class="text-lg font-semibold mb-4">Quote Result</h3>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="space-y-3">
            <div>
              <p class="text-xs text-gray-500">Asset</p>
              <p class="font-medium text-gray-900">{{ quoteResult.crypto_asset }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Crypto Amount</p>
              <p class="font-bold text-xl">{{ quoteResult.crypto_amount.toFixed(6) }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Exchange Rate</p>
              <p class="font-medium">1 CNY = {{ quoteResult.exchange_rate.toFixed(6) }} {{ quoteResult.crypto_asset }}</p>
            </div>
          </div>
          <div class="space-y-3">
            <div>
              <p class="text-xs text-gray-500">Fiat Amount</p>
              <p class="font-medium">¥{{ quoteResult.fiat_cny.toFixed(2) }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Status</p>
              <span :class="getStatusColor(quoteResult.status)" class="status-badge">
                {{ quoteResult.status }}
              </span>
            </div>
            <div>
              <p class="text-xs text-gray-500">Created</p>
              <p class="font-medium">{{ formatDate(quoteResult.created_at) }}</p>
            </div>
          </div>
        </div>
        <div class="mt-4 flex justify-end">
          <button
            @click="activeTab = 'redeem'"
            class="btn-primary"
          >
            Proceed to Redeem
          </button>
        </div>
      </div>
    </div>

    <div v-else-if="activeTab === 'redeem'" class="space-y-6">
      <div class="bg-white rounded-xl border border-gray-200 p-6">
        <h3 class="text-lg font-semibold mb-4">Redeem Crypto</h3>
        <p class="text-sm text-gray-500 mb-4">
          Broadcast a crypto transfer to the customer's wallet.
        </p>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Order ID *</label>
            <input
              v-model="redeemForm.order_id"
              type="text"
              placeholder="Redemption order ID"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Wallet Address *</label>
            <input
              v-model="redeemForm.wallet_addr"
              type="text"
              placeholder="Customer wallet address"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Asset</label>
            <select
              v-model="redeemForm.asset"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            >
              <option v-for="asset in assets" :key="asset.value" :value="asset.value">
                {{ asset.label }}
              </option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Network</label>
            <select
              v-model="redeemForm.network"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            >
              <option v-for="net in networks" :key="net.value" :value="net.value">
                {{ net.label }}
              </option>
            </select>
          </div>
        </div>

        <div class="mt-6">
          <button
            @click="doRedeem"
            :disabled="loading || !redeemForm.order_id || !redeemForm.wallet_addr"
            class="btn-primary"
          >
            <span v-if="loading" class="flex items-center gap-2">
              <svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              Redeeming...
            </span>
            <span v-else>Redeem Now</span>
          </button>
        </div>
      </div>

      <div v-if="redeemResult" class="bg-white rounded-xl border border-gray-200 p-6">
        <h3 class="text-lg font-semibold mb-4">Redemption Result</h3>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="space-y-3">
            <div>
              <p class="text-xs text-gray-500">Status</p>
              <span :class="getStatusColor(redeemResult.status)" class="status-badge">
                {{ redeemResult.status }}
              </span>
            </div>
            <div>
              <p class="text-xs text-gray-500">Crypto Amount</p>
              <p class="font-bold text-xl">{{ redeemResult.crypto_amount.toFixed(6) }} {{ redeemResult.crypto_asset }}</p>
            </div>
            <div v-if="redeemResult.tx_hash" class="break-all">
              <p class="text-xs text-gray-500">Transaction Hash</p>
              <p class="font-medium text-gray-900">{{ redeemResult.tx_hash }}</p>
            </div>
          </div>
          <div class="space-y-3">
            <div>
              <p class="text-xs text-gray-500">Wallet Address</p>
              <p class="font-medium">{{ redeemResult.wallet_address }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Network</p>
              <p class="font-medium">{{ redeemResult.network }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Fiat Amount</p>
              <p class="font-medium">¥{{ redeemResult.fiat_cny.toFixed(2) }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

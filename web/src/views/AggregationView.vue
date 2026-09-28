<script setup lang="ts">
import { ref } from 'vue'
import {
  QrCodeIcon,
  PlusIcon,
  CheckIcon,
  XMarkIcon,
} from '@heroicons/vue/24/outline'
import { paymentApi, AggregationPaymentRequest, AggregationPaymentResponse } from '@/api/client'

const loading = ref(false)
const error = ref<string | null>(null)
const result = ref<AggregationPaymentResponse | null>(null)
const selectedChannel = ref<string>('')

const channels = [
  { value: 'ALIPAY', label: 'Alipay', color: 'bg-alipay-500' },
  { value: 'WECHAT', label: 'WeChat Pay', color: 'bg-wechat-500' },
  { value: 'UNIONPAY', label: 'UnionPay', color: 'bg-unionpay-500' },
  { value: 'CRYPTO', label: 'Crypto', color: 'bg-crypto-500' },
  { value: 'BANK_CARD', label: 'Bank Card', color: 'bg-gray-500' },
]

const form = ref<AggregationPaymentRequest>({
  merchant_id: '',
  out_order_no: '',
  subject: '',
  amount_cny: 0,
  expire_minutes: 15,
})

function generateOrderId(): string {
  return 'ORD' + Date.now().toString().slice(-8) + Math.floor(Math.random() * 1000).toString().padStart(3, '0')
}

async function createAggregationPayment() {
  if (!form.value.out_order_no) {
    form.value.out_order_no = generateOrderId()
  }
  loading.value = true
  error.value = null
  result.value = null
  try {
    const res = await paymentApi.createAggregationPayment(form.value)
    result.value = res.data
  } catch (err: any) {
    error.value = err.response?.data || err.message || 'Failed to create aggregation payment'
  } finally {
    loading.value = false
  }
}

async function resolveChannel(channel: string) {
  loading.value = true
  error.value = null
  try {
    const res = await paymentApi.resolveAggregation(result.value!.out_order_no, channel)
    result.value!.qr_code_url = res.data.qr_code_url
  } catch (err: any) {
    error.value = err.response?.data || err.message || 'Failed to resolve channel'
  } finally {
    loading.value = false
  }
}

function reset() {
  result.value = null
  error.value = null
  selectedChannel.value = ''
  form.value = {
    merchant_id: '',
    out_order_no: '',
    subject: '',
    amount_cny: 0,
    expire_minutes: 15,
  }
}

function getChannelClass(channel: string): string {
  const map: Record<string, string> = {
    ALIPAY: 'channel-alipay',
    WECHAT: 'channel-wechat',
    UNIONPAY: 'channel-unionpay',
    CRYPTO: 'channel-crypto',
    BANK_CARD: 'bg-gray-100 text-gray-800',
  }
  return map[channel] || 'bg-gray-100 text-gray-800'
}
</script>

<template>
  <div class="space-y-6">
    <div v-if="result" class="space-y-6">
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold">Aggregation Payment Created</h3>
        <button @click="reset" class="btn-secondary text-sm">New Payment</button>
      </div>

      <div class="bg-white rounded-xl border border-gray-200 p-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <p class="text-xs text-gray-500 mb-2">QR Code</p>
            <img
              :src="result.qr_code_url"
              alt="QR Code"
              class="w-48 h-48 object-contain border rounded-lg"
            />
          </div>

          <div class="space-y-4">
            <div>
              <p class="text-xs text-gray-500">Order Number</p>
              <p class="font-medium text-gray-900">{{ result.out_order_no }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Amount</p>
              <p class="font-bold text-xl">¥{{ result.amount_cny.toFixed(2) }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Status</p>
              <span class="status-badge bg-yellow-100 text-yellow-800">{{ result.status }}</span>
            </div>
            <div>
              <p class="text-xs text-gray-500">Expires</p>
              <p class="font-medium text-gray-900">{{ new Date(result.expires_at).toLocaleString() }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-xl border border-gray-200 p-6">
        <h4 class="text-md font-semibold mb-4">Available Channels</h4>
        <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
          <button
            v-for="channel in result.channels"
            :key="channel.channel"
            @click="selectedChannel = channel.channel; resolveChannel(channel.channel)"
            :disabled="loading"
            class="flex items-center gap-3 p-3 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors text-left"
          >
            <span :class="getChannelClass(channel.channel)" class="status-badge">
              {{ channel.channel }}
            </span>
            <span class="text-sm font-medium">Select</span>
          </button>
        </div>
        <p v-if="result.channels.length === 0" class="text-gray-500">No channels available</p>
      </div>
    </div>

    <div v-else class="bg-white rounded-xl border border-gray-200 p-6">
      <h3 class="text-lg font-semibold mb-4">Create Aggregation Payment</h3>
      <p class="text-sm text-gray-500 mb-4">
        Create a payment that allows customers to choose from multiple payment channels.
      </p>

      <div v-if="error" class="bg-red-50 text-red-700 p-3 rounded-lg mb-4">
        {{ error }}
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Merchant ID</label>
          <input
            v-model="form.merchant_id"
            type="text"
            placeholder="Optional"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Subject *</label>
          <input
            v-model="form.subject"
            type="text"
            placeholder="Product/service description"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Amount (CNY) *</label>
          <input
            v-model.number="form.amount_cny"
            type="number"
            step="0.01"
            min="0.01"
            placeholder="0.00"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Order No.</label>
          <input
            v-model="form.out_order_no"
            type="text"
            placeholder="Auto-generated if left empty"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Expire (minutes)</label>
          <input
            v-model.number="form.expire_minutes"
            type="number"
            min="1"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Callback URL</label>
          <input
            v-model="form.callback_url"
            type="url"
            placeholder="https://your-domain.com/callback"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
      </div>

      <div class="mt-6">
        <button
          @click="createAggregationPayment"
          :disabled="loading || !form.subject || form.amount_cny <= 0"
          class="btn-primary"
        >
          <span v-if="loading" class="flex items-center gap-2">
            <svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            Creating...
          </span>
          <span v-else>Create Aggregation Payment</span>
        </button>
      </div>
    </div>
  </div>
</template>

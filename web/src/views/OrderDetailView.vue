<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  QrCodeIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  ArrowLeftIcon,
} from '@heroicons/vue/24/outline'
import { paymentApi, Order } from '@/api/client'

const route = useRoute()
const router = useRouter()
const orderId = parseInt(route.params.id as string)
const loading = ref(true)
const error = ref<string | null>(null)
const order = ref<Order | null>(null)

async function loadOrder() {
  loading.value = true
  error.value = null
  try {
    const res = await paymentApi.getOrder(orderId)
    order.value = res.data
  } catch (err: any) {
    error.value = err.response?.data || err.message || 'Failed to load order'
  } finally {
    loading.value = false
  }
}

function getStatusIcon(status: string) {
  switch (status) {
    case 'PAID':
    case 'SETTLED':
      return CheckCircleIcon
    case 'PENDING':
      return ClockIcon
    case 'FAILED':
    case 'CLOSED':
    case 'REFUNDED':
      return XCircleIcon
    default:
      return ClockIcon
  }
}

function getStatusColor(status: string) {
  if (status === 'PAID' || status === 'SETTLED') return 'text-green-600'
  if (status === 'PENDING') return 'text-yellow-600'
  return 'text-red-600'
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

function formatDate(dateStr: string | undefined): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(loadOrder)
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center gap-3">
      <button
        @click="router.back()"
        class="p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-colors"
      >
        <ArrowLeftIcon class="w-5 h-5" />
      </button>
      <h2 class="text-xl font-semibold text-gray-900">Order Detail</h2>
    </div>

    <div v-if="loading" class="bg-white rounded-xl border border-gray-200 p-8 text-center">
      <p class="text-gray-500">Loading order...</p>
    </div>

    <div v-else-if="error" class="bg-red-50 text-red-700 p-4 rounded-lg">
      {{ error }}
    </div>

    <div v-else-if="order" class="space-y-6">
      <div class="bg-white rounded-xl border border-gray-200 p-6">
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-3">
            <component :is="getStatusIcon(order.status)" class="w-6 h-6" :class="getStatusColor(order.status)" />
            <h3 class="text-lg font-semibold">{{ order.subject }}</h3>
          </div>
          <span :class="getChannelClass(order.channel)" class="status-badge">
            {{ order.channel }}
          </span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="space-y-3">
            <div>
              <p class="text-xs text-gray-500">Order Number</p>
              <p class="font-medium text-gray-900">{{ order.out_order_no }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Internal ID</p>
              <p class="font-medium text-gray-900">{{ order.id }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Merchant ID</p>
              <p class="font-medium text-gray-900">{{ order.merchant_id || '-' }}</p>
            </div>
          </div>
          <div class="space-y-3">
            <div>
              <p class="text-xs text-gray-500">Amount</p>
              <p class="font-bold text-2xl text-gray-900">¥{{ order.amount_cny.toFixed(2) }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Status</p>
              <span :class="getStatusColor(order.status)" class="status-badge">
                {{ order.status }}
              </span>
            </div>
            <div>
              <p class="text-xs text-gray-500">Created</p>
              <p class="font-medium text-gray-900">{{ formatDate(order.created_at) }}</p>
            </div>
          </div>
        </div>

        <div v-if="order.trade_no" class="mt-4 pt-4 border-t">
          <p class="text-xs text-gray-500">Trade No.</p>
          <p class="font-medium text-gray-900">{{ order.trade_no }}</p>
        </div>

        <div v-if="order.transaction_id" class="mt-2">
          <p class="text-xs text-gray-500">Transaction ID</p>
          <p class="font-medium text-gray-900">{{ order.transaction_id }}</p>
        </div>

        <div v-if="order.qr_code_url" class="mt-4 pt-4 border-t">
          <p class="text-xs text-gray-500 mb-2">QR Code</p>
          <img :src="order.qr_code_url" alt="QR Code" class="w-32 h-32 object-contain" />
        </div>
      </div>
    </div>
  </div>
</template>

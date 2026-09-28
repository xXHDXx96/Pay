<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { FunnelIcon, MagnifyingGlassIcon, PlusIcon } from '@heroicons/vue/24/outline'
import { paymentApi, Order } from '@/api/client'

const router = useRouter()
const loading = ref(true)
const error = ref<string | null>(null)
const orders = ref<Order[]>([])
const searchTerm = ref('')
const filterChannel = ref('')
const filterStatus = ref('')

const channelOptions = ['', 'ALIPAY', 'WECHAT', 'UNIONPAY', 'CRYPTO', 'BANK_CARD']
const statusOptions = ['', 'PENDING', 'PAID', 'SETTLED', 'FAILED', 'CLOSED', 'REFUNDED']

async function loadOrders() {
  loading.value = true
  error.value = null
  try {
    const res = await paymentApi.listOrders()
    orders.value = res.data.orders
  } catch (err: any) {
    error.value = err.message || 'Failed to load orders'
  } finally {
    loading.value = false
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

function getStatusClass(status: string): string {
  if (status === 'PAID' || status === 'SETTLED') return 'bg-green-100 text-green-800'
  if (status === 'PENDING') return 'bg-yellow-100 text-yellow-800'
  return 'bg-red-100 text-red-800'
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const filteredOrders = computed(() => {
  let result = orders.value
  if (searchTerm.value) {
    const term = searchTerm.value.toLowerCase()
    result = result.filter(o =>
      o.out_order_no.toLowerCase().includes(term) ||
      o.subject.toLowerCase().includes(term)
    )
  }
  if (filterChannel.value) {
    result = result.filter(o => o.channel === filterChannel.value)
  }
  if (filterStatus.value) {
    result = result.filter(o => o.status === filterStatus.value)
  }
  return result
})

function viewOrder(id: number) {
  router.push(`/orders/${id}`)
}

onMounted(loadOrders)
</script>

<template>
  <div class="space-y-4">
    <div class="flex gap-3 items-center">
      <div class="relative flex-1">
        <MagnifyingGlassIcon class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
        <input
          v-model="searchTerm"
          type="text"
          placeholder="Search by order number or subject..."
          class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
        />
      </div>
      <select
        v-model="filterChannel"
        class="border border-gray-300 rounded-lg px-3 py-2 focus:ring-primary-500 focus:border-primary-500"
      >
        <option value="">All Channels</option>
        <option v-for="ch in channelOptions.slice(1)" :key="ch" :value="ch">
          {{ ch }}
        </option>
      </select>
      <select
        v-model="filterStatus"
        class="border border-gray-300 rounded-lg px-3 py-2 focus:ring-primary-500 focus:border-primary-500"
      >
        <option value="">All Statuses</option>
        <option v-for="st in statusOptions.slice(1)" :key="st" :value="st">
          {{ st }}
        </option>
      </select>
      <button
        @click="router.push('/qrcode')"
        class="btn-primary flex items-center gap-2"
      >
        <PlusIcon class="w-5 h-5" />
        New Payment
      </button>
    </div>

    <div v-if="error" class="bg-red-50 text-red-700 p-4 rounded-lg">
      {{ error }}
    </div>

    <div v-if="loading" class="bg-white rounded-xl border border-gray-200">
      <div class="p-8 text-center">
        <div class="text-gray-400">Loading orders...</div>
      </div>
    </div>

    <div v-else class="bg-white rounded-xl border border-gray-200">
      <div class="overflow-x-auto">
        <table class="w-full text-left">
          <thead>
            <tr class="border-b bg-gray-50">
              <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Order</th>
              <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Subject</th>
              <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Amount</th>
              <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Channel</th>
              <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Status</th>
              <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Created</th>
              <th class="px-4 py-2"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="order in filteredOrders"
              :key="order.id"
              class="border-b hover:bg-gray-50"
            >
              <td class="px-4 py-3">
                <p class="text-sm font-medium text-gray-900">{{ order.out_order_no }}</p>
                <p class="text-xs text-gray-500">ID: {{ order.id }}</p>
              </td>
              <td class="px-4 py-3 text-sm text-gray-700">{{ order.subject }}</td>
              <td class="px-4 py-3 text-sm font-medium">¥{{ order.amount_cny.toFixed(2) }}</td>
              <td class="px-4 py-3">
                <span :class="getChannelClass(order.channel)" class="status-badge">
                  {{ order.channel }}
                </span>
              </td>
              <td class="px-4 py-3">
                <span :class="getStatusClass(order.status)" class="status-badge">
                  {{ order.status }}
                </span>
              </td>
              <td class="px-4 py-3 text-sm text-gray-500">{{ formatDate(order.created_at) }}</td>
              <td class="px-4 py-3 text-right">
                <button
                  @click="viewOrder(order.id)"
                  class="text-primary-600 hover:text-primary-700 font-medium text-sm"
                >
                  View
                </button>
              </td>
            </tr>
            <tr v-if="filteredOrders.length === 0">
              <td colspan="7" class="px-4 py-8 text-center text-gray-500">
                No orders found
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

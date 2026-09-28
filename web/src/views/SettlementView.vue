<script setup lang="ts">
import { ref } from 'vue'
import {
  BanknotesIcon,
  BuildingOfficeIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
} from '@heroicons/vue/24/outline'

const activeTab = ref<'batches' | 'orders'>('batches')
const loading = ref(false)

const settlementBatches = ref([
  {
    id: 'BATCH001',
    merchant: 'Merchant ABC',
    total_amount: 12500.00,
    order_count: 45,
    fee: 375.00,
    net_amount: 12125.00,
    status: 'PAID',
    created_at: '2026-09-25 10:30:00',
  },
  {
    id: 'BATCH002',
    merchant: 'Merchant XYZ',
    total_amount: 8750.50,
    order_count: 23,
    fee: 262.52,
    net_amount: 8487.98,
    status: 'PROCESSING',
    created_at: '2026-09-26 14:15:00',
  },
  {
    id: 'BATCH003',
    merchant: 'Merchant ABC',
    total_amount: 3200.00,
    order_count: 12,
    fee: 96.00,
    net_amount: 3104.00,
    status: 'PENDING',
    created_at: '2026-09-27 09:00:00',
  },
])

const settlementOrders = ref([
  {
    id: 1001,
    order_no: 'ORD12345678',
    batch_id: 'BATCH001',
    amount: 128.00,
    fee: 3.84,
    net: 124.16,
    status: 'SETTLED',
  },
  {
    id: 1002,
    order_no: 'ORD12345679',
    batch_id: 'BATCH001',
    amount: 56.50,
    fee: 1.70,
    net: 54.80,
    status: 'SETTLED',
  },
  {
    id: 1003,
    order_no: 'ORD12345680',
    batch_id: 'BATCH002',
    amount: 89.00,
    fee: 2.67,
    net: 86.33,
    status: 'PENDING',
  },
])

function getStatusClass(status: string): string {
  switch (status) {
    case 'PAID':
    case 'SETTLED':
      return 'bg-green-100 text-green-800'
    case 'PROCESSING':
    case 'PENDING':
      return 'bg-yellow-100 text-yellow-800'
    case 'FAILED':
      return 'bg-red-100 text-red-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString()
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex gap-4 mb-6">
      <button
        @click="activeTab = 'batches'"
        :class="activeTab === 'batches' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-700'"
        class="px-4 py-2 rounded-lg font-medium transition-colors"
      >
        Settlement Batches
      </button>
      <button
        @click="activeTab = 'orders'"
        :class="activeTab === 'orders' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-700'"
        class="px-4 py-2 rounded-lg font-medium transition-colors"
      >
        Settlement Orders
      </button>
    </div>

    <div v-if="activeTab === 'batches'">
      <div class="bg-white rounded-xl border border-gray-200">
        <div class="p-4 border-b">
          <h3 class="text-lg font-semibold text-gray-900">Settlement Batches</h3>
          <p class="text-sm text-gray-500 mt-1">View payment settlement batches</p>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left">
            <thead>
              <tr class="border-b bg-gray-50">
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Batch ID</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Merchant</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Orders</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Total (CNY)</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Fee (CNY)</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Net (CNY)</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Status</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Created</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="batch in settlementBatches" :key="batch.id" class="border-b">
                <td class="px-4 py-3 text-sm font-medium text-gray-900">{{ batch.id }}</td>
                <td class="px-4 py-3 text-sm text-gray-700">{{ batch.merchant }}</td>
                <td class="px-4 py-3 text-sm text-gray-700">{{ batch.order_count }}</td>
                <td class="px-4 py-3 text-sm font-medium">¥{{ batch.total_amount.toFixed(2) }}</td>
                <td class="px-4 py-3 text-sm text-gray-500">¥{{ batch.fee.toFixed(2) }}</td>
                <td class="px-4 py-3 text-sm font-medium">¥{{ batch.net_amount.toFixed(2) }}</td>
                <td class="px-4 py-3">
                  <span :class="getStatusClass(batch.status)" class="status-badge">
                    {{ batch.status }}
                  </span>
                </td>
                <td class="px-4 py-3 text-sm text-gray-500">{{ batch.created_at }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div v-else>
      <div class="bg-white rounded-xl border border-gray-200">
        <div class="p-4 border-b">
          <h3 class="text-lg font-semibold text-gray-900">Settlement Order Details</h3>
          <p class="text-sm text-gray-500 mt-1">View individual order settlements within batches</p>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left">
            <thead>
              <tr class="border-b bg-gray-50">
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Order ID</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Order No</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Batch ID</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Amount (CNY)</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Fee (CNY)</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Net (CNY)</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="order in settlementOrders" :key="order.id" class="border-b">
                <td class="px-4 py-3 text-sm font-medium text-gray-900">{{ order.id }}</td>
                <td class="px-4 py-3 text-sm text-gray-700">{{ order.order_no }}</td>
                <td class="px-4 py-3 text-sm text-gray-700">{{ order.batch_id }}</td>
                <td class="px-4 py-3 text-sm font-medium">¥{{ order.amount.toFixed(2) }}</td>
                <td class="px-4 py-3 text-sm text-gray-500">¥{{ order.fee.toFixed(2) }}</td>
                <td class="px-4 py-3 text-sm font-medium">¥{{ order.net.toFixed(2) }}</td>
                <td class="px-4 py-3">
                  <span :class="getStatusClass(order.status)" class="status-badge">
                    {{ order.status }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

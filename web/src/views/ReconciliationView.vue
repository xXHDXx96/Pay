<script setup lang="ts">
import { ref } from 'vue'
import {
  ClipboardDocumentListIcon,
  CheckCircleIcon,
  XCircleIcon,
  ExclamationTriangleIcon,
} from '@heroicons/vue/24/outline'

const activeTab = ref<'list' | 'detail'>('list')
const currentRecon = ref<any>(null)

const reconciliations = ref([
  {
    id: 'RECON001',
    channel: 'ALIPAY',
    provider: 'Alipay',
    trade_date: '2026-09-26',
    total_remote: 150,
    total_count: 148,
    total_amount: 85230.00,
    matched_count: 145,
    mismatch_count: 2,
    missing_local: 1,
    missing_remote: 0,
    status: 'MATCHED',
    file_name: 'alipay_20260926.csv',
  },
  {
    id: 'RECON002',
    channel: 'WECHAT',
    provider: 'WeChat Pay',
    trade_date: '2026-09-26',
    total_remote: 92,
    total_count: 90,
    total_amount: 51200.50,
    matched_count: 88,
    mismatch_count: 1,
    missing_local: 1,
    missing_remote: 2,
    status: 'MISMATCH',
    file_name: 'wechat_20260926.csv',
  },
  {
    id: 'RECON003',
    channel: 'UNIONPAY',
    provider: 'UnionPay',
    trade_date: '2026-09-25',
    total_remote: 68,
    total_count: 68,
    total_amount: 23400.75,
    matched_count: 68,
    mismatch_count: 0,
    missing_local: 0,
    missing_remote: 0,
    status: 'MATCHED',
    file_name: 'unionpay_20260925.csv',
  },
])

const reconciliationDetails = ref([
  {
    id: 1,
    recon_id: 'RECON001',
    local_order: 'ORD20260926001',
    remote_order: 'ALI20260926001',
    local_amount: 120.00,
    remote_amount: 120.00,
    status: 'MATCHED',
  },
  {
    id: 2,
    recon_id: 'RECON001',
    local_order: 'ORD20260926002',
    remote_order: 'ALI20260926002',
    local_amount: 88.50,
    remote_amount: 85.00,
    status: 'MISMATCH',
    discrepancy: 'Amount mismatch: 3.50 CNY',
  },
  {
    id: 3,
    recon_id: 'RECON001',
    local_order: '',
    remote_order: 'ALI20260926003',
    local_amount: 0,
    remote_amount: 56.00,
    status: 'MISSING_LOCAL',
    discrepancy: 'Order exists on provider but not locally',
  },
])

function getStatusClass(status: string): string {
  switch (status) {
    case 'MATCHED':
      return 'bg-green-100 text-green-800'
    case 'MISMATCH':
      return 'bg-red-100 text-red-800'
    case 'MISSING_LOCAL':
    case 'MISSING_REMOTE':
      return 'bg-yellow-100 text-yellow-800'
    case 'COMPLETED':
      return 'bg-blue-100 text-blue-800'
    case 'PENDING':
      return 'bg-gray-100 text-gray-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

function viewDetail(recon: any) {
  currentRecon.value = recon
  activeTab.value = 'detail'
}

function backToList() {
  activeTab.value = 'list'
  currentRecon.value = null
}

function getChannelClass(channel: string): string {
  const map: Record<string, string> = {
    ALIPAY: 'channel-alipay',
    WECHAT: 'channel-wechat',
    UNIONPAY: 'channel-unionpay',
    CRYPTO: 'channel-crypto',
  }
  return map[channel] || 'bg-gray-100 text-gray-800'
}
</script>

<template>
  <div class="space-y-6">
    <div v-if="activeTab === 'list'" class="space-y-6">
      <div class="bg-white rounded-xl border border-gray-200">
        <div class="p-4 border-b flex justify-between items-center">
          <div>
            <h3 class="text-lg font-semibold text-gray-900">Reconciliation History</h3>
            <p class="text-sm text-gray-500 mt-1">View reconciliation results by channel</p>
          </div>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left">
            <thead>
              <tr class="border-b bg-gray-50">
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">ID</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Channel</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Provider</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Trade Date</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Remote Txns</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Matched</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Mismatch</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Missing</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Total Amount</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Status</th>
                <th class="px-4 py-2"></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="recon in reconciliations" :key="recon.id" class="border-b hover:bg-gray-50">
                <td class="px-4 py-3 text-sm font-medium text-gray-900">{{ recon.id }}</td>
                <td class="px-4 py-3">
                  <span :class="getChannelClass(recon.channel)" class="status-badge">
                    {{ recon.channel }}
                  </span>
                </td>
                <td class="px-4 py-3 text-sm text-gray-700">{{ recon.provider }}</td>
                <td class="px-4 py-3 text-sm text-gray-700">{{ recon.trade_date }}</td>
                <td class="px-4 py-3 text-sm text-gray-700">{{ recon.total_count }}/{{ recon.total_remote }}</td>
                <td class="px-4 py-3 text-sm text-green-700">{{ recon.matched_count }}</td>
                <td class="px-4 py-3 text-sm text-red-700">{{ recon.mismatch_count }}</td>
                <td class="px-4 py-3 text-sm text-yellow-700">{{ recon.missing_local }} / {{ recon.missing_remote }}</td>
                <td class="px-4 py-3 text-sm font-medium">¥{{ recon.total_amount.toFixed(2) }}</td>
                <td class="px-4 py-3">
                  <span :class="getStatusClass(recon.status)" class="status-badge">
                    {{ recon.status }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right">
                  <button
                    @click="viewDetail(recon)"
                    class="text-primary-600 hover:text-primary-700 font-medium text-sm"
                  >
                    Details
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div v-if="currentRecon" class="bg-white rounded-xl border border-gray-200">
        <div class="p-4 border-b">
          <h3 class="text-lg font-semibold text-gray-900">Details for {{ currentRecon.id }}</h3>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left">
            <thead>
              <tr class="border-b bg-gray-50">
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Local Order</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Remote Order</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Local Amount</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Remote Amount</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Status</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Discrepancy</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="detail in reconciliationDetails.filter(d => d.recon_id === currentRecon.id)" :key="detail.id" class="border-b">
                <td class="px-4 py-3 text-sm text-gray-700">{{ detail.local_order || '-' }}</td>
                <td class="px-4 py-3 text-sm text-gray-700">{{ detail.remote_order || '-' }}</td>
                <td class="px-4 py-3 text-sm font-medium">¥{{ detail.local_amount.toFixed(2) }}</td>
                <td class="px-4 py-3 text-sm font-medium">¥{{ detail.remote_amount.toFixed(2) }}</td>
                <td class="px-4 py-3">
                  <span :class="getStatusClass(detail.status)" class="status-badge">
                    {{ detail.status }}
                  </span>
                </td>
                <td class="px-4 py-3 text-sm text-gray-600">{{ detail.discrepancy || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

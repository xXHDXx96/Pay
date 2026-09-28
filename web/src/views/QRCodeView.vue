<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { QrCodeIcon, ShoppingCartIcon } from '@heroicons/vue/24/outline'
import QRCode from 'qrcode'
import { paymentApi, QRPaymentRequest, QRPaymentResponse, PaymentStatusResponse } from '@/api/client'

const router = useRouter()
const activeTab = ref<'single' | 'multi'>('single')
const loading = ref(false)
const error = ref<string | null>(null)
const paymentResult = ref<QRPaymentResponse | null>(null)
const statusResult = ref<PaymentStatusResponse | null>(null)

const channels = [
  { value: 'ALIPAY', label: 'Alipay', color: 'bg-alipay-500' },
  { value: 'WECHAT', label: 'WeChat Pay', color: 'bg-wechat-500' },
  { value: 'UNIONPAY', label: 'UnionPay', color: 'bg-unionpay-500' },
  { value: 'CRYPTO', label: 'Crypto', color: 'bg-crypto-500' },
  { value: 'BANK_CARD', label: 'Bank Card', color: 'bg-gray-500' },
]

const singleForm = ref<QRPaymentRequest>({
  merchant_id: '',
  out_order_no: '',
  subject: '',
  amount_cny: 0,
  channel: 'ALIPAY',
  expire_minutes: 15,
})

const statusForm = ref({
  out_order_no: '',
})

const statusQR = ref<string>('')

function generateLocalQR(text: string) {
  try {
    statusQR.value = QRCode.toDataURL(text)
  } catch {
    statusQR.value = text
  }
}

function generateOrderId(): string {
  return 'ORD' + Date.now().toString().slice(-8) + Math.floor(Math.random() * 1000).toString().padStart(3, '0')
}

async function createPayment() {
  if (!singleForm.value.out_order_no) {
    singleForm.value.out_order_no = generateOrderId()
  }
  loading.value = true
  error.value = null
  paymentResult.value = null
  try {
    const res = await paymentApi.createQRCodePayment(singleForm.value)
    paymentResult.value = res.data
  } catch (err: any) {
    error.value = err.response?.data || err.message || 'Failed to create payment'
  } finally {
    loading.value = false
  }
}

async function queryStatus() {
  if (!statusForm.value.out_order_no) {
    error.value = 'Please enter an order number'
    return
  }
  loading.value = true
  error.value = null
  statusResult.value = null
  try {
    const res = await paymentApi.queryPayment(statusForm.value.out_order_no)
    statusResult.value = res.data
  } catch (err: any) {
    error.value = err.response?.data || err.message || 'Failed to query payment'
  } finally {
    loading.value = false
  }
}

function reset() {
  paymentResult.value = null
  statusResult.value = null
  error.value = null
  singleForm.value = {
    merchant_id: '',
    out_order_no: '',
    subject: '',
    amount_cny: 0,
    channel: 'ALIPAY',
    expire_minutes: 15,
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex gap-4 mb-6">
      <button
        @click="activeTab = 'single'"
        :class="activeTab === 'single' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-700'"
        class="px-4 py-2 rounded-lg font-medium transition-colors"
      >
        Single Channel Payment
      </button>
      <button
        @click="router.push('/aggregation')"
        class="px-4 py-2 rounded-lg font-medium bg-gray-100 text-gray-700 hover:bg-gray-200 transition-colors"
      >
        Multi Channel (Aggregation)
      </button>
    </div>

    <div v-if="paymentResult" class="bg-white rounded-xl border border-gray-200 p-6">
      <h3 class="text-lg font-semibold mb-4">Payment Created</h3>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <div class="mb-4">
            <p class="text-xs text-gray-500">QR Code</p>
            <img
              v-if="paymentResult.qr_code_base64"
              :src="`data:image/png;base64,${paymentResult.qr_code_base64}`"
              alt="QR Code"
              class="w-48 h-48 object-contain border rounded-lg"
            />
            <img
              v-else-if="paymentResult.qr_code_url"
              :src="paymentResult.qr_code_url"
              alt="QR Code"
              class="w-48 h-48 object-contain border rounded-lg"
            />
          </div>
          <p class="text-xs text-gray-500">Or scan: {{ paymentResult.qr_code_url }}</p>
        </div>

        <div class="space-y-3">
          <div>
            <p class="text-xs text-gray-500">Order Number</p>
            <p class="font-medium text-gray-900">{{ paymentResult.out_order_no }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500">Amount</p>
            <p class="font-bold text-xl">¥{{ paymentResult.amount_cny.toFixed(2) }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500">Channel</p>
            <span class="status-badge channel-alipay">{{ paymentResult.channel }}</span>
          </div>
          <div>
            <p class="text-xs text-gray-500">Status</p>
            <span class="status-badge bg-yellow-100 text-yellow-800">{{ paymentResult.status }}</span>
          </div>
          <div>
            <p class="text-xs text-gray-500">Expires</p>
            <p class="font-medium text-gray-900">{{ new Date(paymentResult.expires_at).toLocaleString() }}</p>
          </div>
        </div>
      </div>

      <div class="mt-6 pt-4 border-t flex gap-3">
        <input
          v-model="statusForm.out_order_no"
          type="hidden"
        />
        <button @click="reset" class="btn-secondary">
          New Payment
        </button>
        <button @click="statusForm.out_order_no = paymentResult.out_order_no; queryStatus()" class="btn-primary">
          Check Status
        </button>
      </div>
    </div>

    <div v-else class="space-y-6">
      <div class="bg-white rounded-xl border border-gray-200 p-6">
        <h3 class="text-lg font-semibold mb-4">Create Payment</h3>

        <div v-if="error" class="bg-red-50 text-red-700 p-3 rounded-lg mb-4">
          {{ error }}
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Merchant ID</label>
            <input
              v-model="singleForm.merchant_id"
              type="text"
              placeholder="Optional"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Subject *</label>
            <input
              v-model="singleForm.subject"
              type="text"
              placeholder="Product/service description"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Amount (CNY) *</label>
            <input
              v-model.number="singleForm.amount_cny"
              type="number"
              step="0.01"
              min="0.01"
              placeholder="0.00"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Channel *</label>
            <select
              v-model="singleForm.channel"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            >
              <option v-for="ch in channels" :key="ch.value" :value="ch.value">
                {{ ch.label }}
              </option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Order No.</label>
            <input
              v-model="singleForm.out_order_no"
              type="text"
              placeholder="Auto-generated if left empty"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Expire (minutes)</label>
            <input
              v-model.number="singleForm.expire_minutes"
              type="number"
              min="1"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>
        </div>

        <div class="mt-6 flex gap-3">
          <button
            @click="createPayment"
            :disabled="loading || !singleForm.subject || singleForm.amount_cny <= 0"
            class="btn-primary"
          >
            <span v-if="loading" class="flex items-center gap-2">
              <svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              Creating...
            </span>
            <span v-else>Create Payment</span>
          </button>
        </div>
      </div>

      <div class="bg-white rounded-xl border border-gray-200 p-6">
        <h3 class="text-lg font-semibold mb-4">Check Order Status</h3>
        <div class="flex gap-3">
          <input
            v-model="statusForm.out_order_no"
            type="text"
            placeholder="Enter order number"
            class="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          />
          <button
            @click="queryStatus"
            :disabled="loading || !statusForm.out_order_no"
            class="btn-primary"
          >
            Check
          </button>
        </div>

        <div v-if="statusResult" class="mt-4 p-4 bg-gray-50 rounded-lg">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <p class="text-xs text-gray-500">Status</p>
              <p class="font-medium">{{ statusResult.status }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Amount</p>
              <p class="font-medium">¥{{ statusResult.amount_cny.toFixed(2) }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Channel</p>
              <p class="font-medium">{{ statusResult.channel }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500">Paid At</p>
              <p class="font-medium">{{ statusResult.paid_at ? new Date(statusResult.paid_at).toLocaleString() : 'Not paid' }}</p>
            </div>
          </div>

          <div v-if="statusResult.qr_code_url" class="mt-3">
            <p class="text-xs text-gray-500 mb-1">QR Code</p>
            <img :src="statusResult.qr_code_url" alt="QR Code" class="w-32 h-32" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import {
  WifiIcon,
  PlayIcon,
  KeyIcon,
  ClockIcon,
} from '@heroicons/vue/24/outline'
import { paymentApi, WebhookEvent } from '@/api/client'

const loading = ref(false)
const error = ref<string | null>(null)
const success = ref<string | null>(null)
const result = ref<any>(null)

const form = ref<WebhookEvent>({
  out_order_no: '',
  status: 'PAID',
  amount_cny: 0,
  trade_no: '',
  nonce: '',
  timestamp: Math.floor(Date.now() / 1000),
})

const testSignature = ref('')

const statusOptions = [
  { value: 'PAID', label: 'Paid' },
  { value: 'SUCCESS', label: 'Success' },
  { value: 'FAILED', label: 'Failed' },
]

function generateNonce(): string {
  return Math.random().toString(36).substring(2, 12) + Date.now().toString(36)
}

function updateTimestamp() {
  form.value.timestamp = Math.floor(Date.now() / 1000)
}

function generateSignature() {
  const data = JSON.stringify(form.value)
  testSignature.value = btoa(data).substring(0, 32)
}

async function sendWebhook() {
  if (!form.value.out_order_no) {
    error.value = 'Order number is required'
    return
  }
  loading.value = true
  error.value = null
  success.value = null
  result.value = null
  try {
    const headers = testSignature.value ? { 'X-Signature': testSignature.value } : undefined
    const res = await paymentApi.sendWebhook(form.value, testSignature.value)
    result.value = res.data
    success.value = 'Webhook sent successfully'
  } catch (err: any) {
    error.value = err.response?.data || err.message || 'Failed to send webhook'
  } finally {
    loading.value = false
  }
}

function presetNotification() {
  form.value = {
    out_order_no: '',
    status: 'PAID',
    amount_cny: 0,
    trade_no: '',
    nonce: generateNonce(),
    timestamp: Math.floor(Date.now() / 1000),
  }
}

function presetFailure() {
  form.value = {
    out_order_no: '',
    status: 'FAILED',
    amount_cny: 0,
    trade_no: '',
    nonce: generateNonce(),
    timestamp: Math.floor(Date.now() / 1000),
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="bg-white rounded-xl border border-gray-200 p-6">
      <h3 class="text-lg font-semibold mb-4">Webhook Testing</h3>
      <p class="text-sm text-gray-500 mb-4">
        Simulate webhook notifications to test payment status updates.
      </p>

      <div v-if="success" class="bg-green-50 text-green-700 p-3 rounded-lg mb-4">
        {{ success }}
      </div>

      <div v-if="error" class="bg-red-50 text-red-700 p-3 rounded-lg mb-4">
        {{ error }}
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Out Order No. *</label>
          <input
            v-model="form.out_order_no"
            type="text"
            placeholder="Order number to notify"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Status *</label>
          <select
            v-model="form.status"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          >
            <option v-for="s in statusOptions" :key="s.value" :value="s.value">
              {{ s.label }}
            </option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Amount (CNY)</label>
          <input
            v-model.number="form.amount_cny"
            type="number"
            step="0.01"
            min="0"
            placeholder="0.00"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Trade No.</label>
          <input
            v-model="form.trade_no"
            type="text"
            placeholder="Provider transaction number"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Nonce</label>
          <input
            v-model="form.nonce"
            type="text"
            placeholder="Unique identifier"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Timestamp</label>
          <div class="flex gap-2">
            <input
              v-model.number="form.timestamp"
              type="number"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
            <button @click="updateTimestamp" class="btn-secondary">
              Now
            </button>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Signature (optional)</label>
          <div class="flex gap-2">
            <input
              v-model="testSignature"
              type="text"
              placeholder="HMAC signature"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
            <button @click="generateSignature" class="btn-secondary">
              <KeyIcon class="w-5 h-5" />
            </button>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Presets</label>
          <div class="flex gap-2">
            <button @click="presetNotification" class="btn-secondary text-sm">
              Paid
            </button>
            <button @click="presetFailure" class="btn-secondary text-sm">
              Failed
            </button>
          </div>
        </div>
      </div>

      <div class="mt-6 flex gap-3">
        <button
          @click="sendWebhook"
          :disabled="loading || !form.out_order_no"
          class="btn-primary"
        >
          <span v-if="loading" class="flex items-center gap-2">
            <svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            Sending...
          </span>
          <span v-else>Send Webhook</span>
        </button>
      </div>
    </div>

    <div v-if="result" class="bg-white rounded-xl border border-gray-200 p-6">
      <h3 class="text-lg font-semibold mb-4">Response</h3>
      <pre class="bg-gray-50 p-4 rounded-lg text-sm overflow-x-auto">
        {{ JSON.stringify(result, null, 2) }}
      </pre>
    </div>
  </div>
</template>

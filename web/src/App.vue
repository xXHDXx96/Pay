<script setup lang="ts">
import { ref } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import {
  QrCodeIcon,
  ShoppingCartIcon,
  CurrencyDollarIcon,
  ChartBarIcon,
  SquaresPlusIcon,
  ClipboardDocumentListIcon,
  BuildingOfficeIcon,
  WifiIcon,
  CogIcon,
  Bars3Icon,
  XMarkIcon,
} from '@heroicons/vue/24/outline'

const route = useRoute()
const sidebarOpen = ref(false)

interface NavItem {
  name: string
  path: string
  icon: any
  label: string
}

const navItems: NavItem[] = [
  { name: 'dashboard', path: '/', icon: ChartBarIcon, label: 'Dashboard' },
  { name: 'orders', path: '/orders', icon: ShoppingCartIcon, label: 'Orders' },
  { name: 'qrcode', path: '/qrcode', icon: QrCodeIcon, label: 'QR Code Payment' },
  { name: 'aggregation', path: '/aggregation', icon: SquaresPlusIcon, label: 'Aggregation Payment' },
  { name: 'exchange', path: '/exchange', icon: CurrencyDollarIcon, label: 'Crypto Exchange' },
  { name: 'webhook', path: '/webhook', icon: WifiIcon, label: 'Webhook Testing' },
  { name: 'settlement', path: '/settlement', icon: BuildingOfficeIcon, label: 'Settlement' },
  { name: 'reconciliation', path: '/reconciliation', icon: ClipboardDocumentListIcon, label: 'Reconciliation' },
]

function isActive(path: string): boolean {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <div v-if="sidebarOpen" class="fixed inset-0 z-50 lg:hidden">
      <div class="fixed inset-0 bg-black/30" @click="sidebarOpen = false"></div>
      <div class="fixed inset-y-0 left-0 w-64 bg-white shadow-xl overflow-y-auto">
        <div class="flex items-center justify-between p-4 border-b">
          <h2 class="text-xl font-semibold text-gray-900">Payment Gateway</h2>
          <button @click="sidebarOpen = false" class="p-2 rounded-lg hover:bg-gray-100">
            <XMarkIcon class="w-5 h-5" />
          </button>
        </div>
        <nav class="mt-4 space-y-1">
          <router-link
            v-for="item in navItems"
            :key="item.name"
            :to="item.path"
            @click="sidebarOpen = false"
            :class="[
              'flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
              isActive(item.path)
                ? 'bg-primary-50 text-primary-700'
                : 'text-gray-700 hover:bg-gray-100',
            ]"
          >
            <component :is="item.icon" class="w-5 h-5" />
            {{ item.label }}
          </router-link>
        </nav>
      </div>
    </div>

    <div class="flex h-screen overflow-hidden">
      <aside class="hidden lg:flex lg:flex-col lg:w-64 lg:border-r lg:bg-white lg:overflow-y-auto">
        <div class="flex items-center gap-3 px-6 py-5 border-b">
          <div class="w-8 h-8 rounded-lg bg-primary-600 flex items-center justify-center">
            <span class="text-white font-bold text-sm">AP</span>
          </div>
          <h1 class="text-xl font-bold text-gray-900">Payment Gateway</h1>
        </div>
        <nav class="flex-1 mt-4 space-y-1 px-4">
          <router-link
            v-for="item in navItems"
            :key="item.name"
            :to="item.path"
            :class="[
              'flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
              isActive(item.path)
                ? 'bg-primary-50 text-primary-700'
                : 'text-gray-700 hover:bg-gray-100',
            ]"
          >
            <component :is="item.icon" class="w-5 h-5" />
            {{ item.label }}
          </router-link>
        </nav>
        <div class="border-t p-4">
          <button class="flex items-center gap-3 w-full px-3 py-2 text-sm font-medium text-gray-700 rounded-lg hover:bg-gray-100 transition-colors">
            <CogIcon class="w-5 h-5" />
            Settings
          </button>
        </div>
      </aside>

      <main class="flex-1 overflow-y-auto">
        <header class="bg-white border-b px-4 py-3 lg:px-6 flex items-center justify-between">
          <button
            @click="sidebarOpen = true"
            class="lg:hidden p-2 rounded-lg text-gray-600 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <Bars3Icon class="w-6 h-6" />
          </button>
          <h2 class="text-xl font-semibold text-gray-900">{{ route.meta.title || 'Dashboard' }}</h2>
          <div class="w-10"></div>
        </header>
        <div class="p-4 lg:p-6">
          <RouterView />
        </div>
      </main>
    </div>
  </div>
</template>

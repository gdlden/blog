<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const tabs = [
  { name: 'ledger', path: '/ledger', label: '交易' },
  { name: 'ledgerAccounts', path: '/ledger/accounts', label: '账户' },
  { name: 'ledgerCategories', path: '/ledger/categories', label: '分类' },
  { name: 'ledgerRecurring', path: '/ledger/recurring', label: '周期' },
  { name: 'ledgerReport', path: '/ledger/report', label: '报表' },
]

const currentRouteName = computed(() => route.name as string)
</script>

<template>
  <div class="inline-flex items-center gap-1 bg-white rounded-xl p-1 border border-[#f0f0f0]">
    <router-link
      v-for="tab in tabs"
      :key="tab.name"
      :to="tab.path"
      class="px-4 py-1.5 text-[13px] font-medium rounded-lg transition-all duration-200"
      :class="
        currentRouteName === tab.name
          ? 'text-[#0071e3] bg-[#0071e3]/10'
          : 'text-[#86868b] hover:text-[#1d1d1f]'
      "
    >
      {{ tab.label }}
    </router-link>
  </div>
</template>

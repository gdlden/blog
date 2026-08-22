<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useLedgerStore } from '@/stores/ledgerStore'
import {
  ACCOUNT_SUBTYPE_LABELS,
  ACCOUNT_TYPE_LABELS,
  ASSET_SUBTYPES,
  LIABILITY_SUBTYPES,
  formatAmount,
  ledgerErrorMessage,
} from '@/api/ledger'
import type { LedgerAccount, LedgerAccountType } from '@/api/ledger'
import { reminderBadgeText } from '@/utils/creditReminder'
import LedgerNav from '@/components/LedgerNav.vue'

const ledgerStore = useLedgerStore()
const { accounts, accountsLoading, creditReminders } = storeToRefs(ledgerStore)

onMounted(async () => {
  try {
    await ledgerStore.fetchAccounts()
  } catch {
    return // 账户加载失败已 toast，跳过提醒计算
  }
  ledgerStore.fetchCreditReminders()
})

// 角标视图：accountId → 文案 + 等级
const reminderView = computed<Record<string, { text: string; level: 'warning' | 'info' }>>(() => {
  const view: Record<string, { text: string; level: 'warning' | 'info' }> = {}
  creditReminders.value.forEach((reminder) => {
    view[reminder.accountId] = { text: reminderBadgeText(reminder), level: reminder.level }
  })
  return view
})

const showModal = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const formData = ref({
  id: '',
  name: '',
  type: 'asset' as LedgerAccountType,
  subtype: 'cash',
  creditLimit: '',
  billingDay: '' as string | number,
  paymentDueDay: '' as string | number,
  remark: '',
  openingBalance: '',
  archived: false,
})

// subtype 二级联动：type 切换时重置为该类型第一个 subtype
const subtypeOptions = computed<readonly string[]>(() =>
  formData.value.type === 'asset' ? ASSET_SUBTYPES : LIABILITY_SUBTYPES,
)
watch(
  () => formData.value.type,
  () => {
    formData.value.subtype = subtypeOptions.value[0] || 'other'
  },
)

// 信用卡三字段仅 subtype=credit_card 时显示
const isCreditCard = computed(() => formData.value.subtype === 'credit_card')

const canSubmit = computed(() => !!formData.value.name.trim())

function subtypeLabel(subtype: string): string {
  return ACCOUNT_SUBTYPE_LABELS[subtype] || subtype || '其他'
}

function openCreateModal() {
  isEditing.value = false
  formData.value = {
    id: '',
    name: '',
    type: 'asset',
    subtype: 'cash',
    creditLimit: '',
    billingDay: '',
    paymentDueDay: '',
    remark: '',
    openingBalance: '',
    archived: false,
  }
  showModal.value = true
}

function openEditModal(account: LedgerAccount) {
  isEditing.value = true
  formData.value = {
    id: account.id,
    name: account.name,
    type: account.type,
    subtype: account.subtype,
    creditLimit: account.creditLimit || '',
    billingDay: account.billingDay ?? '',
    paymentDueDay: account.paymentDueDay ?? '',
    remark: account.remark || '',
    // 初始余额只在新增时填写，编辑不改
    openingBalance: '',
    archived: !!account.archived,
  }
  showModal.value = true
}

async function handleSubmit() {
  if (!canSubmit.value) return
  isSubmitting.value = true
  try {
    await ledgerStore.saveAccount({ ...formData.value }, isEditing.value)
    showModal.value = false
  } catch (err) {
    alert(ledgerErrorMessage(err, '操作失败'))
  } finally {
    isSubmitting.value = false
  }
}

async function handleDelete(account: LedgerAccount) {
  if (!confirm(`确定要删除账户「${account.name}」吗？`)) return
  try {
    await ledgerStore.removeAccount(account.id)
  } catch (err) {
    // store 已 toast 后端错误信息，这里兜底
    alert(ledgerErrorMessage(err, '删除失败'))
  }
}

async function toggleArchive(account: LedgerAccount) {
  try {
    await ledgerStore.saveAccount(
      {
        id: account.id,
        name: account.name,
        type: account.type,
        subtype: account.subtype,
        creditLimit: account.creditLimit || '',
        billingDay: account.billingDay ?? '',
        paymentDueDay: account.paymentDueDay ?? '',
        remark: account.remark || '',
        archived: !account.archived,
      },
      true,
    )
  } catch (err) {
    alert(ledgerErrorMessage(err, '操作失败'))
  }
}

function closeModal() {
  showModal.value = false
}
</script>

<template>
  <div class="max-w-[1100px] mx-auto px-5 py-10">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between mb-6">
      <div>
        <h1 class="text-[32px] font-semibold tracking-tight text-[#1d1d1f]">账户管理</h1>
        <p class="mt-1 text-sm text-[#86868b]">维护资产与负债账户</p>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <LedgerNav />
        <button
          @click="openCreateModal"
          class="inline-flex items-center justify-center gap-2 px-5 py-2.5 text-white text-[15px] font-medium rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] shadow-sm"
          style="background: linear-gradient(135deg, #0071e3, #0063c7)"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 4v16m8-8H4"
            />
          </svg>
          新增账户
        </button>
      </div>
    </div>

    <div v-if="accountsLoading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div v-for="n in 6" :key="n" class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="h-5 bg-[#f5f5f7] rounded-lg w-3/4 mb-3 animate-pulse" />
        <div class="h-3 bg-[#f5f5f7] rounded w-full mb-2 animate-pulse" />
        <div class="h-3 bg-[#f5f5f7] rounded w-2/3 animate-pulse" />
      </div>
    </div>

    <div
      v-else-if="accounts.length === 0"
      class="text-center py-24 bg-white rounded-2xl border border-[#f0f0f0]"
    >
      <h3 class="text-xl font-semibold text-[#1d1d1f] mb-2">暂无账户</h3>
      <p class="text-sm text-[#86868b] mb-6">先添加账户，再开始记账</p>
      <button
        @click="openCreateModal"
        class="inline-flex items-center gap-2 px-5 py-2.5 text-white text-[15px] font-medium rounded-xl transition-all hover:scale-[1.02]"
        style="background: linear-gradient(135deg, #0071e3, #0063c7)"
      >
        新增账户
      </button>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div
        v-for="account in accounts"
        :key="account.id"
        class="ledger-account-card group bg-white rounded-2xl p-5 flex flex-col border border-[#f0f0f0] transition-all duration-300 hover:shadow-lg hover:border-[#e8e8ed]"
        :class="account.archived ? 'opacity-60' : ''"
      >
        <div class="flex justify-between items-start mb-2">
          <h3
            class="text-[17px] font-semibold text-[#1d1d1f] leading-snug line-clamp-1 flex-1"
            :title="account.name"
          >
            {{ account.name }}
          </h3>
          <div class="flex items-center gap-1.5 ml-2 shrink-0">
            <span
              v-if="reminderView[account.id]"
              class="credit-reminder-badge text-xs px-2 py-0.5 rounded-full border font-medium"
              :class="
                reminderView[account.id]?.level === 'warning'
                  ? 'bg-[#ff3b30]/10 text-[#ff3b30] border-[#ff3b30]/30'
                  : 'bg-[#0071e3]/5 text-[#0071e3] border-[#0071e3]/20'
              "
            >
              {{ reminderView[account.id]?.text }}
            </span>
            <span
              class="text-xs px-2 py-0.5 rounded-full border font-medium"
              :class="
                account.type === 'asset'
                  ? 'bg-blue-50 text-blue-700 border-blue-200'
                  : 'bg-orange-50 text-orange-700 border-orange-200'
              "
            >
              {{ ACCOUNT_TYPE_LABELS[account.type] || account.type }}
            </span>
          </div>
        </div>
        <p class="text-sm text-[#86868b] mb-3">
          {{ subtypeLabel(account.subtype) }}
          <span v-if="account.archived" class="ml-1 text-xs">（已归档）</span>
        </p>
        <p class="text-[22px] font-semibold text-[#1d1d1f] mb-1">
          ¥{{ formatAmount(account.balance) }}
        </p>
        <p v-if="account.subtype === 'credit_card' && account.creditLimit" class="text-xs text-[#86868b] mb-2">
          额度 ¥{{ formatAmount(account.creditLimit) }} · 账单日 {{ account.billingDay || '-' }} 号 ·
          还款日 {{ account.paymentDueDay || '-' }} 号
        </p>
        <p class="text-sm text-[#86868b] line-clamp-2 flex-1 mb-4">{{ account.remark || '无备注' }}</p>
        <div
          class="flex justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity pt-3 border-t border-[#f5f5f7]"
        >
          <button
            @click="toggleArchive(account)"
            class="px-2 py-1.5 rounded-lg text-[#86868b] hover:bg-[#f5f5f7] text-xs font-medium transition-colors"
            :title="account.archived ? '取消归档' : '归档'"
          >
            {{ account.archived ? '取消归档' : '归档' }}
          </button>
          <button
            @click="openEditModal(account)"
            class="p-1.5 rounded-lg text-[#0071e3] hover:bg-[#0071e3]/10 transition-colors"
            title="编辑"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
              />
            </svg>
          </button>
          <button
            @click="handleDelete(account)"
            class="ledger-account-delete p-1.5 rounded-lg text-[#ff3b30] hover:bg-[#ff3b30]/10 transition-colors"
            title="删除"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>

  <Transition
    enter-active-class="transition duration-200 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition duration-150 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="showModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      style="background-color: rgba(0, 0, 0, 0.35)"
      @click.self="closeModal"
    >
      <Transition
        enter-active-class="transition duration-300 ease-out"
        enter-from-class="opacity-0 scale-95 translate-y-2"
        enter-to-class="opacity-100 scale-100 translate-y-0"
        leave-active-class="transition duration-200 ease-in"
        leave-from-class="opacity-100 scale-100 translate-y-0"
        leave-to-class="opacity-0 scale-95 translate-y-2"
      >
        <div
          v-if="showModal"
          class="ledger-account-modal bg-white w-full max-w-2xl rounded-2xl shadow-2xl overflow-hidden max-h-[90vh] overflow-y-auto"
        >
          <div class="flex justify-between items-center px-6 py-4 border-b border-[#f0f0f0]">
            <h3 class="text-lg font-semibold text-[#1d1d1f]">
              {{ isEditing ? '编辑账户' : '新增账户' }}
            </h3>
            <button
              @click="closeModal"
              class="p-1 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] transition-colors"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
          <div class="px-6 py-5 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">账户名称</label>
              <input
                v-model="formData.name"
                type="text"
                placeholder="例如：招商储蓄卡"
                class="account-name w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">账户类型</label>
              <select
                v-model="formData.type"
                class="account-type w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              >
                <option value="asset">资产</option>
                <option value="liability">负债</option>
              </select>
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">子类型</label>
              <select
                v-model="formData.subtype"
                class="account-subtype w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              >
                <option v-for="sub in subtypeOptions" :key="sub" :value="sub">
                  {{ subtypeLabel(sub) }}
                </option>
              </select>
            </div>
            <div v-if="!isEditing">
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">初始余额</label>
              <input
                v-model="formData.openingBalance"
                type="number"
                step="0.01"
                placeholder="请输入初始余额"
                class="account-opening-balance w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <template v-if="isCreditCard">
              <div>
                <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">信用额度</label>
                <input
                  v-model="formData.creditLimit"
                  type="number"
                  step="0.01"
                  placeholder="请输入信用额度"
                  class="account-credit-limit w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                />
              </div>
              <div>
                <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">账单日</label>
                <input
                  v-model="formData.billingDay"
                  type="number"
                  min="1"
                  max="31"
                  placeholder="1-31"
                  class="account-billing-day w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                />
              </div>
              <div>
                <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">还款日</label>
                <input
                  v-model="formData.paymentDueDay"
                  type="number"
                  min="1"
                  max="31"
                  placeholder="1-31"
                  class="account-payment-due-day w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                />
              </div>
            </template>
            <div class="md:col-span-2">
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">备注</label>
              <textarea
                v-model="formData.remark"
                rows="3"
                placeholder="请输入备注"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none resize-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <label
              v-if="isEditing"
              class="md:col-span-2 flex items-center gap-2 text-sm font-medium text-[#1d1d1f]"
            >
              <input
                v-model="formData.archived"
                type="checkbox"
                class="account-archived w-4 h-4 rounded border-[#d2d2d7]"
              />
              归档该账户（归档后记账时不再可选）
            </label>
          </div>
          <div class="flex justify-end gap-2.5 px-6 py-4 border-t border-[#f0f0f0] bg-[#fafafc]/50">
            <button
              @click="closeModal"
              class="px-5 py-2 text-sm font-medium text-[#1d1d1f] bg-white border border-[#e8e8ed] rounded-xl hover:bg-[#f5f5f7] transition-colors"
            >
              取消
            </button>
            <button
              @click="handleSubmit"
              :disabled="!canSubmit || isSubmitting"
              class="px-5 py-2 text-sm font-medium text-white rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] disabled:opacity-40 disabled:scale-100 disabled:cursor-not-allowed"
              style="background: linear-gradient(135deg, #0071e3, #0063c7)"
            >
              {{ isSubmitting ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

<style scoped>
.line-clamp-1 {
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>

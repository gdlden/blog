<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useDebtStore } from '@/stores/debtStore'
import { useDebtDetailStore } from '@/stores/debtDetailStore'

const route = useRoute()
const router = useRouter()
const debtId = route.params.id as string

const debtStore = useDebtStore()
const detailStore = useDebtDetailStore()
const { currentDebt, loading: debtLoading } = storeToRefs(debtStore)
const { details, loading: detailLoading, totalPrincipal, totalInterest, totalRepaid } = storeToRefs(detailStore)

const showModal = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const formData = ref({
  id: '', debtId: '', postingDate: '', principal: '', interest: '', period: ''
})

onMounted(() => {
  debtStore.fetchDebtById(debtId)
  detailStore.fetchDetails(debtId)
})

function goBack() {
  router.push({ name: 'debt' })
}

function openCreateModal() {
  isEditing.value = false
  formData.value = { id: '', debtId, postingDate: '', principal: '', interest: '', period: '' }
  showModal.value = true
}

function openEditModal(detail: any) {
  isEditing.value = true
  formData.value = { ...detail }
  showModal.value = true
}

async function handleSubmit() {
  if (!formData.value.postingDate.trim()) return
  isSubmitting.value = true
  try {
    const payload = {
      ...formData.value,
      postingDate: formData.value.postingDate + ' 00:00:00'
    }
    if (isEditing.value) await detailStore.updateDetail(payload)
    else { const { id, ...rest } = payload; await detailStore.createDetail(rest) }
    showModal.value = false
  } catch (err: any) { alert(err.message || '操作失败') }
  finally { isSubmitting.value = false }
}

function statusText(s: string) {
  if (s === '1' || s === '已结清') return '已结清'
  return '进行中'
}

async function handleDelete(id: string) {
  if (confirm('确定要删除这条明细吗？')) {
    try { await detailStore.deleteDetail(id, debtId) } catch (err: any) { alert(err.message || '删除失败') }
  }
}

function closeModal() { showModal.value = false }
function formatAmount(n: number) { return n.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }
function statusClass(s: string) {
  if (s === '1' || s === '已结清') return 'bg-green-50 text-green-700 border-green-200'
  return 'bg-blue-50 text-blue-700 border-blue-200'
}
</script>

<template>
  <div class="max-w-[1100px] mx-auto px-5 py-10">
    <!-- Back + Title -->
    <div class="flex items-center gap-4 mb-8">
      <button @click="goBack" class="p-2 rounded-xl text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] transition-all">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
      </button>
      <div>
        <h1 class="text-[28px] font-semibold tracking-tight text-[#1d1d1f]">{{ currentDebt?.name || '债务详情' }}</h1>
        <p class="mt-0.5 text-sm text-[#86868b]">{{ currentDebt?.bankName || '' }}</p>
      </div>
    </div>

    <!-- Debt Info Card -->
    <div v-if="currentDebt" class="bg-white rounded-2xl p-6 border border-[#f0f0f0] mb-6">
      <div class="flex flex-wrap gap-y-4">
        <div class="w-1/2 md:w-1/4">
          <p class="text-xs text-[#86868b] mb-1">金额</p>
          <p class="text-[18px] font-semibold text-[#1d1d1f]">¥{{ currentDebt.amount }}</p>
        </div>
        <div class="w-1/2 md:w-1/4">
          <p class="text-xs text-[#86868b] mb-1">年利率</p>
          <p class="text-[18px] font-semibold text-[#1d1d1f]">{{ currentDebt.apr }}%</p>
        </div>
        <div class="w-1/2 md:w-1/4">
          <p class="text-xs text-[#86868b] mb-1">期限</p>
          <p class="text-[18px] font-semibold text-[#1d1d1f]">{{ currentDebt.tenor }} 个月</p>
        </div>
        <div class="w-1/2 md:w-1/4">
          <p class="text-xs text-[#86868b] mb-1">状态</p>
          <span class="text-xs px-2 py-0.5 rounded-full border font-medium" :class="statusClass(currentDebt.status)">{{ statusText(currentDebt.status) }}</span>
        </div>
        <div class="w-1/2 md:w-1/4" v-if="currentDebt.applyTime">
          <p class="text-xs text-[#86868b] mb-1">申请日期</p>
          <p class="text-[15px] font-medium text-[#1d1d1f]">{{ currentDebt.applyTime?.split(' ')?.[0] || currentDebt.applyTime }}</p>
        </div>
        <div class="w-1/2 md:w-1/4" v-if="currentDebt.endTime">
          <p class="text-xs text-[#86868b] mb-1">结束日期</p>
          <p class="text-[15px] font-medium text-[#1d1d1f]">{{ currentDebt.endTime?.split(' ')?.[0] || currentDebt.endTime }}</p>
        </div>
        <div class="w-full md:w-1/2" v-if="currentDebt.remark">
          <p class="text-xs text-[#86868b] mb-1">备注</p>
          <p class="text-[15px] text-[#1d1d1f]">{{ currentDebt.remark }}</p>
        </div>
      </div>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 rounded-lg bg-blue-50 flex items-center justify-center">
            <svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08.402-2.599 1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
          </div>
          <span class="text-sm text-[#86868b]">已还本金</span>
        </div>
        <p class="text-[22px] font-semibold text-[#1d1d1f]">¥{{ formatAmount(totalPrincipal) }}</p>
      </div>
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 rounded-lg bg-purple-50 flex items-center justify-center">
            <svg class="w-4 h-4 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"/></svg>
          </div>
          <span class="text-sm text-[#86868b]">已还利息</span>
        </div>
        <p class="text-[22px] font-semibold text-[#1d1d1f]">¥{{ formatAmount(totalInterest) }}</p>
      </div>
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 rounded-lg bg-green-50 flex items-center justify-center">
            <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
          </div>
          <span class="text-sm text-[#86868b]">合计还款</span>
        </div>
        <p class="text-[22px] font-semibold text-[#1d1d1f]">¥{{ formatAmount(totalRepaid) }}</p>
      </div>
    </div>

    <!-- Detail List Header -->
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-lg font-semibold text-[#1d1d1f]">还款明细</h2>
      <button @click="openCreateModal" class="inline-flex items-center gap-2 px-4 py-2 text-white text-sm font-medium rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] shadow-sm" style="background: linear-gradient(135deg, #0071e3, #0063c7);">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
        新增明细
      </button>
    </div>

    <!-- Loading -->
    <div v-if="detailLoading" class="space-y-3">
      <div v-for="n in 4" :key="n" class="bg-white rounded-xl p-4 border border-[#f0f0f0] flex justify-between items-center">
        <div class="h-4 bg-[#f5f5f7] rounded w-24 animate-pulse"/>
        <div class="h-4 bg-[#f5f5f7] rounded w-20 animate-pulse"/>
        <div class="h-4 bg-[#f5f5f7] rounded w-20 animate-pulse"/>
        <div class="h-4 bg-[#f5f5f7] rounded w-16 animate-pulse"/>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="details.length === 0" class="text-center py-16 bg-white rounded-2xl border border-[#f0f0f0]">
      <div class="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-[#f5f5f7] mb-4">
        <svg class="w-7 h-7 text-[#86868b]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"/></svg>
      </div>
      <h3 class="text-lg font-semibold text-[#1d1d1f] mb-1">暂无还款明细</h3>
      <p class="text-sm text-[#86868b] mb-5">添加你的第一笔还款记录</p>
      <button @click="openCreateModal" class="inline-flex items-center gap-2 px-5 py-2.5 text-white text-sm font-medium rounded-xl transition-all hover:scale-[1.02]" style="background: linear-gradient(135deg, #0071e3, #0063c7);">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
        新增明细
      </button>
    </div>

    <!-- Detail Table -->
    <div v-else class="bg-white rounded-2xl border border-[#f0f0f0] overflow-hidden">
      <table class="w-full text-left">
        <thead>
          <tr class="border-b border-[#f0f0f0]">
            <th class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide">入账日期</th>
            <th class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide text-right">本金</th>
            <th class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide text-right">利息</th>
            <th class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide text-center">期数</th>
            <th class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide text-right">合计</th>
            <th class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide text-right">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="detail in details" :key="detail.id" class="border-b border-[#f5f5f7] last:border-b-0 hover:bg-[#fafafc] transition-colors">
            <td class="px-5 py-4 text-sm text-[#1d1d1f]">{{ detail.postingDate?.split(' ')?.[0] || detail.postingDate }}</td>
            <td class="px-5 py-4 text-sm text-[#1d1d1f] text-right">¥{{ detail.principal }}</td>
            <td class="px-5 py-4 text-sm text-[#1d1d1f] text-right">¥{{ detail.interest }}</td>
            <td class="px-5 py-4 text-sm text-[#1d1d1f] text-center">{{ detail.period }}</td>
            <td class="px-5 py-4 text-sm font-semibold text-[#1d1d1f] text-right">¥{{ (parseFloat(detail.principal) + parseFloat(detail.interest)).toFixed(2) }}</td>
            <td class="px-5 py-4 text-right">
              <div class="inline-flex gap-1">
                <button @click="openEditModal(detail)" class="p-1.5 rounded-lg text-[#0071e3] hover:bg-[#0071e3]/10 transition-colors" title="编辑">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
                </button>
                <button @click="handleDelete(detail.id)" class="p-1.5 rounded-lg text-[#ff3b30] hover:bg-[#ff3b30]/10 transition-colors" title="删除">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <!-- Modal -->
  <Transition enter-active-class="transition duration-200 ease-out" enter-from-class="opacity-0" enter-to-class="opacity-100" leave-active-class="transition duration-150 ease-in" leave-from-class="opacity-100" leave-to-class="opacity-0">
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4" style="background-color: rgba(0,0,0,0.35);" @click.self="closeModal">
      <Transition enter-active-class="transition duration-300 ease-out" enter-from-class="opacity-0 scale-95 translate-y-2" enter-to-class="opacity-100 scale-100 translate-y-0" leave-active-class="transition duration-200 ease-in" leave-from-class="opacity-100 scale-100 translate-y-0" leave-to-class="opacity-0 scale-95 translate-y-2">
        <div v-if="showModal" class="bg-white w-full max-w-lg rounded-2xl shadow-2xl overflow-hidden">
          <div class="flex justify-between items-center px-6 py-4 border-b border-[#f0f0f0]">
            <h3 class="text-lg font-semibold text-[#1d1d1f]">{{ isEditing ? '编辑明细' : '新增明细' }}</h3>
            <button @click="closeModal" class="p-1 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] transition-colors">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
            </button>
          </div>
          <div class="px-6 py-5 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">入账日期</label><input v-model="formData.postingDate" type="date" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">期数</label><input v-model="formData.period" type="number" placeholder="请输入期数" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">本金</label><input v-model="formData.principal" type="number" step="0.01" placeholder="请输入本金" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">利息</label><input v-model="formData.interest" type="number" step="0.01" placeholder="请输入利息" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
          </div>
          <div class="flex justify-end gap-2.5 px-6 py-4 border-t border-[#f0f0f0] bg-[#fafafc]/50">
            <button @click="closeModal" class="px-5 py-2 text-sm font-medium text-[#1d1d1f] bg-white border border-[#e8e8ed] rounded-xl hover:bg-[#f5f5f7] transition-colors">取消</button>
            <button @click="handleSubmit" :disabled="!formData.postingDate.trim() || isSubmitting" class="px-5 py-2 text-sm font-medium text-white rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] disabled:opacity-40 disabled:scale-100 disabled:cursor-not-allowed" style="background: linear-gradient(135deg, #0071e3, #0063c7);">{{ isSubmitting ? '保存中...' : '保存' }}</button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

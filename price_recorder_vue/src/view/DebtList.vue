<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useDebtStore } from '@/stores/debtStore'

const router = useRouter()
const debtStore = useDebtStore()
const { debts, loading, totalPages, currentPage, totalDebt, repaidAmount, outstandingAmount } = storeToRefs(debtStore)

const showModal = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const formData = ref({
  id: '', name: '', bankName: '', bankAccount: '', applyTime: '',
  endTime: '', amount: '', status: '', remark: '', apr: '', fee: '', tenor: ''
})

onMounted(() => debtStore.fetchDebts())

function openCreateModal() {
  isEditing.value = false
  formData.value = { id: '', name: '', bankName: '', bankAccount: '', applyTime: '', endTime: '', amount: '', status: '', remark: '', apr: '', fee: '', tenor: '' }
  showModal.value = true
}

function openEditModal(debt: any) {
  isEditing.value = true
  formData.value = { ...debt }
  showModal.value = true
}

async function handleSubmit() {
  if (!formData.value.name.trim()) return
  isSubmitting.value = true
  try {
    if (isEditing.value) await debtStore.updateDebt(formData.value)
    else { const { id, ...rest } = formData.value; await debtStore.createDebt(rest) }
    showModal.value = false
  } catch (err: any) { alert(err.message || '操作失败') }
  finally { isSubmitting.value = false }
}

async function handleDelete(id: string) {
  if (confirm('确定要删除这条债务记录吗？')) {
    try { await debtStore.deleteDebt(id) } catch (err: any) { alert(err.message || '删除失败') }
  }
}

function changePage(page: number) {
  if (page >= 1 && page <= totalPages.value) debtStore.fetchDebts(page)
}

function goToDetail(id: string) {
  router.push({ name: 'debtDetail', params: { id } })
}

function closeModal() { showModal.value = false }
function formatAmount(n: number) { return n.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }
function statusClass(s: string) {
  if (s === '已结清') return 'bg-green-50 text-green-700 border-green-200'
  return 'bg-blue-50 text-blue-700 border-blue-200'
}
</script>

<template>
  <div class="max-w-[1100px] mx-auto px-5 py-10">
    <!-- Header -->
    <div class="flex justify-between items-center mb-8">
      <div>
        <h1 class="text-[32px] font-semibold tracking-tight text-[#1d1d1f]">债务</h1>
        <p class="mt-1 text-sm text-[#86868b]">管理你的债务和还款计划</p>
      </div>
      <button @click="openCreateModal" class="inline-flex items-center gap-2 px-5 py-2.5 text-white text-[15px] font-medium rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] shadow-sm" style="background: linear-gradient(135deg, #0071e3, #0063c7);">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
        新建债务
      </button>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 rounded-lg bg-red-50 flex items-center justify-center">
            <svg class="w-4 h-4 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
          </div>
          <span class="text-sm text-[#86868b]">总债务</span>
        </div>
        <p class="text-[24px] font-semibold text-[#1d1d1f]">¥{{ formatAmount(totalDebt) }}</p>
      </div>
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 rounded-lg bg-green-50 flex items-center justify-center">
            <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
          </div>
          <span class="text-sm text-[#86868b]">已还款</span>
        </div>
        <p class="text-[24px] font-semibold text-[#1d1d1f]">¥{{ formatAmount(repaidAmount) }}</p>
      </div>
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 rounded-lg bg-orange-50 flex items-center justify-center">
            <svg class="w-4 h-4 text-orange-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>
          </div>
          <span class="text-sm text-[#86868b]">未还金额</span>
        </div>
        <p class="text-[24px] font-semibold text-[#1d1d1f]">¥{{ formatAmount(outstandingAmount) }}</p>
      </div>
    </div>

    <!-- Skeleton Loading -->
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div v-for="n in 6" :key="n" class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="h-5 bg-[#f5f5f7] rounded-lg w-3/4 mb-3 animate-pulse"/>
        <div class="h-3 bg-[#f5f5f7] rounded w-full mb-2 animate-pulse"/>
        <div class="h-3 bg-[#f5f5f7] rounded w-2/3 mb-4 animate-pulse"/>
        <div class="flex justify-between pt-3 border-t border-[#f5f5f7]">
          <div class="h-3 bg-[#f5f5f7] rounded w-20 animate-pulse"/>
          <div class="flex gap-3"><div class="h-3 bg-[#f5f5f7] rounded w-8 animate-pulse"/><div class="h-3 bg-[#f5f5f7] rounded w-8 animate-pulse"/></div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="debts.length === 0" class="text-center py-24">
      <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-[#f5f5f7] mb-5">
        <svg class="w-8 h-8 text-[#86868b]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z"/></svg>
      </div>
      <h3 class="text-xl font-semibold text-[#1d1d1f] mb-2">暂无债务记录</h3>
      <p class="text-sm text-[#86868b] mb-6">创建你的第一条债务记录</p>
      <button @click="openCreateModal" class="inline-flex items-center gap-2 px-5 py-2.5 text-white text-[15px] font-medium rounded-xl transition-all hover:scale-[1.02]" style="background: linear-gradient(135deg, #0071e3, #0063c7);">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
        新建债务
      </button>
    </div>

    <!-- Card Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div v-for="debt in debts" :key="debt.id" @click="goToDetail(debt.id)" class="group bg-white rounded-2xl p-5 flex flex-col border border-[#f0f0f0] transition-all duration-300 hover:shadow-lg hover:border-[#e8e8ed] hover:-translate-y-0.5 cursor-pointer">
        <div class="flex justify-between items-start mb-2">
          <h3 class="text-[17px] font-semibold text-[#1d1d1f] leading-snug line-clamp-1 flex-1" :title="debt.name">{{ debt.name }}</h3>
          <span v-if="debt.status" class="text-xs px-2 py-0.5 rounded-full border font-medium ml-2" :class="statusClass(debt.status)">{{ debt.status }}</span>
        </div>
        <p class="text-sm text-[#86868b] mb-3">{{ debt.bankName }}</p>
        <p class="text-[22px] font-semibold text-[#1d1d1f] mb-1">¥{{ debt.amount }}</p>
        <p class="text-sm text-[#86868b] line-clamp-2 flex-1 mb-4">{{ debt.remark || '无备注' }}</p>
        <div class="flex justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity pt-3 border-t border-[#f5f5f7]">
          <button @click.stop="openEditModal(debt)" class="p-1.5 rounded-lg text-[#0071e3] hover:bg-[#0071e3]/10 transition-colors" title="编辑">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
          </button>
          <button @click.stop="handleDelete(debt.id)" class="p-1.5 rounded-lg text-[#ff3b30] hover:bg-[#ff3b30]/10 transition-colors" title="删除">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1 && !loading" class="flex justify-center items-center gap-1.5 mt-10">
      <button @click="changePage(currentPage - 1)" :disabled="currentPage <= 1" class="p-2 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] disabled:opacity-30 disabled:cursor-not-allowed transition-all">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
      </button>
      <button v-for="page in totalPages" :key="page" @click="changePage(page)" class="min-w-[36px] h-9 px-2.5 rounded-lg text-sm font-medium transition-all" :class="page === currentPage ? 'bg-[#1d1d1f] text-white' : 'text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7]'">{{ page }}</button>
      <button @click="changePage(currentPage + 1)" :disabled="currentPage >= totalPages" class="p-2 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] disabled:opacity-30 disabled:cursor-not-allowed transition-all">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
      </button>
    </div>
  </div>

  <!-- Modal -->
  <Transition enter-active-class="transition duration-200 ease-out" enter-from-class="opacity-0" enter-to-class="opacity-100" leave-active-class="transition duration-150 ease-in" leave-from-class="opacity-100" leave-to-class="opacity-0">
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4" style="background-color: rgba(0,0,0,0.35);" @click.self="closeModal">
      <Transition enter-active-class="transition duration-300 ease-out" enter-from-class="opacity-0 scale-95 translate-y-2" enter-to-class="opacity-100 scale-100 translate-y-0" leave-active-class="transition duration-200 ease-in" leave-from-class="opacity-100 scale-100 translate-y-0" leave-to-class="opacity-0 scale-95 translate-y-2">
        <div v-if="showModal" class="bg-white w-full max-w-2xl rounded-2xl shadow-2xl overflow-hidden max-h-[90vh] overflow-y-auto">
          <div class="flex justify-between items-center px-6 py-4 border-b border-[#f0f0f0]">
            <h3 class="text-lg font-semibold text-[#1d1d1f]">{{ isEditing ? '编辑债务' : '新建债务' }}</h3>
            <button @click="closeModal" class="p-1 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] transition-colors">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
            </button>
          </div>
          <div class="px-6 py-5 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">债务名称</label><input v-model="formData.name" type="text" placeholder="请输入债务名称" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">银行名称</label><input v-model="formData.bankName" type="text" placeholder="请输入银行名称" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">银行账号</label><input v-model="formData.bankAccount" type="text" placeholder="请输入银行账号" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">金额</label><input v-model="formData.amount" type="number" step="0.01" placeholder="请输入金额" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">申请日期</label><input v-model="formData.applyTime" type="date" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">结束日期</label><input v-model="formData.endTime" type="date" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">状态</label><select v-model="formData.status" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"><option value="">请选择</option><option value="进行中">进行中</option><option value="已结清">已结清</option></select></div>
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">年利率 (%)</label><input v-model="formData.apr" type="number" step="0.01" placeholder="请输入年利率" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">手续费</label><input v-model="formData.fee" type="number" step="0.01" placeholder="请输入手续费" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
            <div><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">期限 (月)</label><input v-model="formData.tenor" type="number" placeholder="请输入期限" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
            <div class="md:col-span-2"><label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">备注</label><textarea v-model="formData.remark" rows="3" placeholder="请输入备注" class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none resize-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"/></div>
          </div>
          <div class="flex justify-end gap-2.5 px-6 py-4 border-t border-[#f0f0f0] bg-[#fafafc]/50">
            <button @click="closeModal" class="px-5 py-2 text-sm font-medium text-[#1d1d1f] bg-white border border-[#e8e8ed] rounded-xl hover:bg-[#f5f5f7] transition-colors">取消</button>
            <button @click="handleSubmit" :disabled="!formData.name.trim() || isSubmitting" class="px-5 py-2 text-sm font-medium text-white rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] disabled:opacity-40 disabled:scale-100 disabled:cursor-not-allowed" style="background: linear-gradient(135deg, #0071e3, #0063c7);">{{ isSubmitting ? '保存中...' : '保存' }}</button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

<style scoped>
.line-clamp-1 { display: -webkit-box; -webkit-line-clamp: 1; -webkit-box-orient: vertical; overflow: hidden; }
.line-clamp-2 { display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
</style>
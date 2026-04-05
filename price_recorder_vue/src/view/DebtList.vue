<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useDebtStore } from '@/stores/debtStore'

const debtStore = useDebtStore()
const { debts, loading, totalPages, currentPage, totalDebt, repaidAmount, outstandingAmount } = storeToRefs(debtStore)

const showModal = ref(false)
const isEditing = ref(false)
const formData = ref({
  id: '',
  name: '',
  bankName: '',
  bankAccount: '',
  applyTime: '',
  endTime: '',
  amount: '',
  status: '',
  remark: '',
  apr: '',
  fee: '',
  tenor: ''
})

onMounted(() => {
  debtStore.fetchDebts()
})

function openCreateModal() {
  isEditing.value = false
  formData.value = {
    id: '',
    name: '',
    bankName: '',
    bankAccount: '',
    applyTime: '',
    endTime: '',
    amount: '',
    status: '',
    remark: '',
    apr: '',
    fee: '',
    tenor: ''
  }
  showModal.value = true
}

function openEditModal(debt: {
  id: string
  name: string
  bankName: string
  bankAccount: string
  applyTime: string
  endTime: string
  amount: string
  status: string
  remark: string
  apr: string
  fee: string
  tenor: string
}) {
  isEditing.value = true
  formData.value = { ...debt }
  showModal.value = true
}

async function handleSubmit() {
  try {
    if (isEditing.value) {
      await debtStore.updateDebt(formData.value)
    } else {
      const { id, ...createData } = formData.value
      await debtStore.createDebt(createData)
    }
    showModal.value = false
    resetForm()
  } catch (err: any) {
    alert(err.message || '操作失败')
  }
}

async function handleDelete(id: string) {
  if (confirm('确定要删除这条债务记录吗？')) {
    try {
      await debtStore.deleteDebt(id)
    } catch (err: any) {
      alert(err.message || '删除失败')
    }
  }
}

function changePage(page: number) {
  if (page >= 1 && page <= totalPages.value) {
    debtStore.fetchDebts(page)
  }
}

function closeModal() {
  showModal.value = false
  resetForm()
}

function resetForm() {
  formData.value = {
    id: '',
    name: '',
    bankName: '',
    bankAccount: '',
    applyTime: '',
    endTime: '',
    amount: '',
    status: '',
    remark: '',
    apr: '',
    fee: '',
    tenor: ''
  }
}

function formatAmount(amount: number): string {
  return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 py-6">
    <!-- Header -->
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-semibold text-gray-800">债务</h2>
      <button
        @click="openCreateModal"
        class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors"
      >
        新建债务
      </button>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
      <div class="bg-white rounded-lg shadow p-4 border border-gray-200">
        <p class="text-sm text-gray-600 mb-1">总债务</p>
        <p class="text-2xl font-bold text-gray-800">¥{{ formatAmount(totalDebt) }}</p>
      </div>
      <div class="bg-white rounded-lg shadow p-4 border border-gray-200">
        <p class="text-sm text-gray-600 mb-1">已还款</p>
        <p class="text-2xl font-bold text-green-600">¥{{ formatAmount(repaidAmount) }}</p>
      </div>
      <div class="bg-white rounded-lg shadow p-4 border border-gray-200">
        <p class="text-sm text-gray-600 mb-1">未还金额</p>
        <p class="text-2xl font-bold text-red-600">¥{{ formatAmount(outstandingAmount) }}</p>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-12">
      <p class="text-gray-600 text-lg">加载中...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="debts.length === 0" class="bg-gray-50 rounded-lg p-12 text-center border border-gray-200">
      <p class="text-xl font-semibold text-gray-700 mb-2">暂无内容</p>
      <p class="text-gray-500 mb-4">点击"新建债务"创建第一条记录</p>
    </div>

    <!-- Card Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="debt in debts"
        :key="debt.id"
        class="bg-white rounded-lg shadow border border-gray-200 p-4 flex flex-col"
      >
        <h3 class="text-lg font-semibold text-gray-800 truncate mb-2" :title="debt.name">
          {{ debt.name }}
        </h3>
        <p class="text-gray-600 text-sm mb-1">{{ debt.bankName }}</p>
        <p class="text-gray-800 font-medium mb-1">金额: ¥{{ debt.amount }}</p>
        <p class="text-sm mb-2">
          <span class="text-gray-600">状态: </span>
          <span
            :class="{
              'text-green-600': debt.status === '已结清' || debt.status === 'repaid',
              'text-yellow-600': debt.status === '进行中' || debt.status === 'active',
              'text-gray-600': !debt.status || (debt.status !== '已结清' && debt.status !== 'repaid' && debt.status !== '进行中' && debt.status !== 'active')
            }"
          >
            {{ debt.status || '未知' }}
          </span>
        </p>
        <p class="text-gray-500 text-sm line-clamp-2 flex-1">
          {{ debt.remark || '无备注' }}
        </p>
        <div class="flex justify-end gap-2 mt-4 pt-4 border-t border-gray-100">
          <button
            @click="openEditModal(debt)"
            class="text-blue-600 hover:text-blue-800 text-sm font-medium px-3 py-1 rounded hover:bg-blue-50 transition-colors"
          >
            编辑
          </button>
          <button
            @click="handleDelete(debt.id)"
            class="text-red-600 hover:text-red-800 text-sm font-medium px-3 py-1 rounded hover:bg-red-50 transition-colors"
          >
            删除
          </button>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1 && !loading" class="flex justify-center gap-2 mt-8">
      <button
        @click="changePage(currentPage - 1)"
        :disabled="currentPage <= 1"
        class="px-3 py-1 rounded border border-gray-300 text-sm font-medium transition-colors"
        :class="currentPage <= 1 ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-white text-gray-700 hover:bg-gray-50'"
      >
        上一页
      </button>
      <button
        v-for="page in totalPages"
        :key="page"
        @click="changePage(page)"
        class="px-3 py-1 rounded border text-sm font-medium transition-colors"
        :class="page === currentPage ? 'bg-blue-600 text-white border-blue-600' : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'"
      >
        {{ page }}
      </button>
      <button
        @click="changePage(currentPage + 1)"
        :disabled="currentPage >= totalPages"
        class="px-3 py-1 rounded border border-gray-300 text-sm font-medium transition-colors"
        :class="currentPage >= totalPages ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-white text-gray-700 hover:bg-gray-50'"
      >
        下一页
      </button>
    </div>

    <!-- Modal -->
    <div
      v-if="showModal"
      class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4"
      @click.self="closeModal"
    >
      <div class="bg-white rounded-lg shadow-xl w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto">
        <!-- Modal Header -->
        <div class="flex justify-between items-center px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-semibold text-gray-800">
            {{ isEditing ? '编辑债务' : '新建债务' }}
          </h3>
          <button
            @click="closeModal"
            class="text-gray-400 hover:text-gray-600 text-xl leading-none"
          >
            ×
          </button>
        </div>

        <!-- Modal Body -->
        <div class="px-6 py-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">债务名称</label>
              <input
                v-model="formData.name"
                type="text"
                placeholder="请输入债务名称"
                class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">银行名称</label>
              <input
                v-model="formData.bankName"
                type="text"
                placeholder="请输入银行名称"
                class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">银行账号</label>
              <input
                v-model="formData.bankAccount"
                type="text"
                placeholder="请输入银行账号"
                class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">金额</label>
              <input
                v-model="formData.amount"
                type="number"
                step="0.01"
                placeholder="请输入金额"
                class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">申请日期</label>
              <input
                v-model="formData.applyTime"
                type="date"
                class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">结束日期</label>
              <input
                v-model="formData.endTime"
                type="date"
                class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">状态</label>
              <select
                v-model="formData.status"
                class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent"
              >
                <option value="">请选择状态</option>
                <option value="进行中">进行中</option>
                <option value="已结清">已结清</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">年利率 (%)</label>
              <input
                v-model="formData.apr"
                type="number"
                step="0.01"
                placeholder="请输入年利率"
                class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">手续费</label>
              <input
                v-model="formData.fee"
                type="number"
                step="0.01"
                placeholder="请输入手续费"
                class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">期限 (月)</label>
              <input
                v-model="formData.tenor"
                type="number"
                placeholder="请输入期限"
                class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent"
              />
            </div>
            <div class="md:col-span-2">
              <label class="block text-sm font-medium text-gray-700 mb-1">备注</label>
              <textarea
                v-model="formData.remark"
                rows="3"
                placeholder="请输入备注"
                class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent resize-none"
              ></textarea>
            </div>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="flex justify-end gap-3 px-6 py-4 border-t border-gray-200">
          <button
            @click="closeModal"
            class="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg font-medium transition-colors"
          >
            取消
          </button>
          <button
            @click="handleSubmit"
            :disabled="!formData.name.trim() || loading"
            class="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-300 text-white rounded-lg font-medium transition-colors"
          >
            {{ loading ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>

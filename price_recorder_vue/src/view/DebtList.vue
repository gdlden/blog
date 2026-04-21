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
  <div class="max-w-[980px] mx-auto px-5 py-12">
    <!-- Header -->
    <div class="flex justify-between items-center mb-10">
      <h2 style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 40px; font-weight: 600; line-height: 1.10; letter-spacing: normal; color: #1d1d1f;">
        债务
      </h2>
      <button
        @click="openCreateModal"
        class="text-white transition-colors"
        style="background-color: #0071e3; border-radius: 8px; padding: 8px 15px; font-size: 17px; font-weight: 400; line-height: 2.41;"
      >
        新建债务
      </button>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-5 mb-10">
      <div class="bg-white p-5" style="border-radius: 8px; box-shadow: rgba(0, 0, 0, 0.22) 3px 5px 30px 0px;">
        <p style="font-size: 14px; font-weight: 400; line-height: 1.29; letter-spacing: -0.224px; color: rgba(0, 0, 0, 0.48); margin-bottom: 4px;">
          总债务
        </p>
        <p style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 28px; font-weight: 400; line-height: 1.14; letter-spacing: 0.196px; color: #1d1d1f;">
          ¥{{ formatAmount(totalDebt) }}
        </p>
      </div>
      <div class="bg-white p-5" style="border-radius: 8px; box-shadow: rgba(0, 0, 0, 0.22) 3px 5px 30px 0px;">
        <p style="font-size: 14px; font-weight: 400; line-height: 1.29; letter-spacing: -0.224px; color: rgba(0, 0, 0, 0.48); margin-bottom: 4px;">
          已还款
        </p>
        <p style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 28px; font-weight: 400; line-height: 1.14; letter-spacing: 0.196px; color: #1d1d1f;">
          ¥{{ formatAmount(repaidAmount) }}
        </p>
      </div>
      <div class="bg-white p-5" style="border-radius: 8px; box-shadow: rgba(0, 0, 0, 0.22) 3px 5px 30px 0px;">
        <p style="font-size: 14px; font-weight: 400; line-height: 1.29; letter-spacing: -0.224px; color: rgba(0, 0, 0, 0.48); margin-bottom: 4px;">
          未还金额
        </p>
        <p style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 28px; font-weight: 400; line-height: 1.14; letter-spacing: 0.196px; color: #1d1d1f;">
          ¥{{ formatAmount(outstandingAmount) }}
        </p>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-20">
      <p style="font-size: 17px; letter-spacing: -0.374px; line-height: 1.47; color: rgba(0, 0, 0, 0.48);">
        加载中...
      </p>
    </div>

    <!-- Empty State -->
    <div v-else-if="debts.length === 0" class="text-center py-20">
      <p style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 28px; font-weight: 400; line-height: 1.14; letter-spacing: 0.196px; color: #1d1d1f; margin-bottom: 8px;">
        暂无内容
      </p>
      <p style="font-size: 14px; letter-spacing: -0.224px; line-height: 1.29; color: rgba(0, 0, 0, 0.48);">
        点击"新建债务"创建第一条记录
      </p>
    </div>

    <!-- Card Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="debt in debts"
        :key="debt.id"
        class="bg-white p-5 flex flex-col"
        style="border-radius: 8px; box-shadow: rgba(0, 0, 0, 0.22) 3px 5px 30px 0px;"
      >
        <h3
          class="truncate mb-2"
          style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 21px; font-weight: 700; line-height: 1.19; letter-spacing: 0.231px; color: #1d1d1f;"
          :title="debt.name"
        >
          {{ debt.name }}
        </h3>
        <p style="font-size: 14px; font-weight: 400; line-height: 1.29; letter-spacing: -0.224px; color: rgba(0, 0, 0, 0.8); margin-bottom: 4px;">
          {{ debt.bankName }}
        </p>
        <p style="font-size: 17px; font-weight: 600; line-height: 1.24; letter-spacing: -0.374px; color: #1d1d1f; margin-bottom: 4px;">
          金额: ¥{{ debt.amount }}
        </p>
        <p style="font-size: 14px; font-weight: 400; line-height: 1.29; letter-spacing: -0.224px; color: rgba(0, 0, 0, 0.48); margin-bottom: 4px;">
          <span>状态: </span>
          <span>{{ debt.status || '未知' }}</span>
        </p>
        <p
          class="line-clamp-2 flex-1"
          style="font-size: 14px; font-weight: 400; line-height: 1.29; letter-spacing: -0.224px; color: rgba(0, 0, 0, 0.48);"
        >
          {{ debt.remark || '无备注' }}
        </p>
        <div class="flex justify-end gap-4 mt-4 pt-4" style="border-top: 1px solid rgba(0, 0, 0, 0.04);">
          <button
            @click="openEditModal(debt)"
            class="transition-colors"
            style="font-size: 14px; font-weight: 400; line-height: 1.43; letter-spacing: -0.224px; color: #0066cc;"
          >
            编辑
          </button>
          <button
            @click="handleDelete(debt.id)"
            class="transition-colors"
            style="font-size: 14px; font-weight: 400; line-height: 1.43; letter-spacing: -0.224px; color: #0066cc;"
          >
            删除
          </button>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1 && !loading" class="flex justify-center gap-2 mt-10">
      <button
        @click="changePage(currentPage - 1)"
        :disabled="currentPage <= 1"
        class="px-3 py-1.5 text-sm font-medium transition-colors"
        :class="currentPage <= 1 ? 'cursor-not-allowed' : 'hover:bg-white'"
        :style="currentPage <= 1
          ? 'background-color: transparent; color: rgba(0, 0, 0, 0.48); border-radius: 8px;'
          : 'background-color: #ffffff; color: #1d1d1f; border-radius: 8px;'"
      >
        上一页
      </button>
      <button
        v-for="page in totalPages"
        :key="page"
        @click="changePage(page)"
        class="px-3 py-1.5 text-sm font-medium transition-colors"
        :style="page === currentPage
          ? 'background-color: #0071e3; color: #ffffff; border-radius: 8px;'
          : 'background-color: #ffffff; color: #1d1d1f; border-radius: 8px;'"
      >
        {{ page }}
      </button>
      <button
        @click="changePage(currentPage + 1)"
        :disabled="currentPage >= totalPages"
        class="px-3 py-1.5 text-sm font-medium transition-colors"
        :class="currentPage >= totalPages ? 'cursor-not-allowed' : 'hover:bg-white'"
        :style="currentPage >= totalPages
          ? 'background-color: transparent; color: rgba(0, 0, 0, 0.48); border-radius: 8px;'
          : 'background-color: #ffffff; color: #1d1d1f; border-radius: 8px;'"
      >
        下一页
      </button>
    </div>

    <!-- Modal -->
    <div
      v-if="showModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      style="background-color: rgba(0, 0, 0, 0.4);"
      @click.self="closeModal"
    >
      <div class="bg-white w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto" style="border-radius: 12px;">
        <!-- Modal Header -->
        <div class="flex justify-between items-center px-6 py-4" style="border-bottom: 1px solid rgba(0, 0, 0, 0.04);">
          <h3 style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 21px; font-weight: 700; line-height: 1.19; letter-spacing: 0.231px; color: #1d1d1f;">
            {{ isEditing ? '编辑债务' : '新建债务' }}
          </h3>
          <button
            @click="closeModal"
            class="text-2xl leading-none transition-colors"
            style="color: rgba(0, 0, 0, 0.48);"
          >
            &times;
          </button>
        </div>

        <!-- Modal Body -->
        <div class="px-6 py-5">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
                债务名称
              </label>
              <input
                v-model="formData.name"
                type="text"
                placeholder="请输入债务名称"
                class="w-full px-3 py-2.5 outline-none"
                style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
              />
            </div>
            <div>
              <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
                银行名称
              </label>
              <input
                v-model="formData.bankName"
                type="text"
                placeholder="请输入银行名称"
                class="w-full px-3 py-2.5 outline-none"
                style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
              />
            </div>
            <div>
              <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
                银行账号
              </label>
              <input
                v-model="formData.bankAccount"
                type="text"
                placeholder="请输入银行账号"
                class="w-full px-3 py-2.5 outline-none"
                style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
              />
            </div>
            <div>
              <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
                金额
              </label>
              <input
                v-model="formData.amount"
                type="number"
                step="0.01"
                placeholder="请输入金额"
                class="w-full px-3 py-2.5 outline-none"
                style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
              />
            </div>
            <div>
              <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
                申请日期
              </label>
              <input
                v-model="formData.applyTime"
                type="date"
                class="w-full px-3 py-2.5 outline-none"
                style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
              />
            </div>
            <div>
              <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
                结束日期
              </label>
              <input
                v-model="formData.endTime"
                type="date"
                class="w-full px-3 py-2.5 outline-none"
                style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
              />
            </div>
            <div>
              <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
                状态
              </label>
              <select
                v-model="formData.status"
                class="w-full px-3 py-2.5 outline-none"
                style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
              >
                <option value="">请选择状态</option>
                <option value="进行中">进行中</option>
                <option value="已结清">已结清</option>
              </select>
            </div>
            <div>
              <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
                年利率 (%)
              </label>
              <input
                v-model="formData.apr"
                type="number"
                step="0.01"
                placeholder="请输入年利率"
                class="w-full px-3 py-2.5 outline-none"
                style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
              />
            </div>
            <div>
              <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
                手续费
              </label>
              <input
                v-model="formData.fee"
                type="number"
                step="0.01"
                placeholder="请输入手续费"
                class="w-full px-3 py-2.5 outline-none"
                style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
              />
            </div>
            <div>
              <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
                期限 (月)
              </label>
              <input
                v-model="formData.tenor"
                type="number"
                placeholder="请输入期限"
                class="w-full px-3 py-2.5 outline-none"
                style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
              />
            </div>
            <div class="md:col-span-2">
              <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
                备注
              </label>
              <textarea
                v-model="formData.remark"
                rows="3"
                placeholder="请输入备注"
                class="w-full px-3 py-2.5 outline-none resize-none"
                style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
              ></textarea>
            </div>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="flex justify-end gap-3 px-6 py-4" style="border-top: 1px solid rgba(0, 0, 0, 0.04);">
          <button
            @click="closeModal"
            class="px-4 py-2 transition-colors"
            style="background-color: #1d1d1f; color: #ffffff; border-radius: 8px; font-size: 17px; font-weight: 400; line-height: 2.41;"
          >
            取消
          </button>
          <button
            @click="handleSubmit"
            :disabled="!formData.name.trim() || loading"
            class="px-4 py-2 text-white transition-colors disabled:opacity-50"
            style="background-color: #0071e3; border-radius: 8px; font-size: 17px; font-weight: 400; line-height: 2.41;"
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

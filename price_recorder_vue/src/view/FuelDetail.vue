<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useFuelStore } from '@/stores/fuelStore'
import type { RefuelRecord } from '@/api/fuel'

const route = useRoute()
const router = useRouter()
const vehicleId = route.params.vehicleId as string
const fuelStore = useFuelStore()
const { currentVehicle, records, stats, recordLoading } = storeToRefs(fuelStore)

const showModal = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const formData = ref<RefuelRecord>({
  id: '',
  vehicleId,
  refuelTime: '',
  odometer: '',
  volume: '',
  unitPrice: '',
  amount: '',
  station: '',
  isFull: true,
  remark: '',
  intervalConsumption: '',
})

const trendPolyline = computed(() => {
  const points = stats.value?.trend || []
  if (points.length === 0) return ''
  const values = points.map((point) => Number(point.consumption || 0))
  const min = Math.min(...values)
  const max = Math.max(...values)
  const spread = Math.max(max - min, 1)
  return values
    .map((value, index) => {
      const x = points.length === 1 ? 50 : (index / (points.length - 1)) * 100
      const y = 90 - ((value - min) / spread) * 70
      return `${x},${y}`
    })
    .join(' ')
})

onMounted(() => {
  fuelStore.fetchVehicleById(vehicleId)
  fuelStore.fetchVehicleDashboard(vehicleId)
})

function goBack() {
  router.push({ name: 'fuel' })
}

function openCreateModal() {
  isEditing.value = false
  formData.value = {
    id: '',
    vehicleId,
    refuelTime: '',
    odometer: '',
    volume: '',
    unitPrice: '',
    amount: '',
    station: '',
    isFull: true,
    remark: '',
    intervalConsumption: '',
  }
  showModal.value = true
}

function openEditModal(record: RefuelRecord) {
  isEditing.value = true
  formData.value = {
    ...record,
    refuelTime: record.refuelTime ? record.refuelTime.slice(0, 10) : '',
  }
  showModal.value = true
}

async function handleSubmit() {
  if (!formData.value.refuelTime.trim()) return
  isSubmitting.value = true
  const amount =
    formData.value.amount || computeAmount(formData.value.volume, formData.value.unitPrice)
  const payload = {
    ...formData.value,
    refuelTime: formData.value.refuelTime.includes(' ')
      ? formData.value.refuelTime
      : `${formData.value.refuelTime} 00:00:00`,
    amount,
  }
  try {
    if (isEditing.value) {
      await fuelStore.updateRefuelRecord(payload)
    } else {
      const { id, intervalConsumption, ...rest } = payload
      await fuelStore.createRefuelRecord(rest)
    }
    showModal.value = false
  } catch (err: any) {
    alert(err.message || '操作失败')
  } finally {
    isSubmitting.value = false
  }
}

async function handleDelete(id: string) {
  if (!confirm('确定要删除这条加油记录吗？')) return
  try {
    await fuelStore.deleteRefuelRecord(id, vehicleId)
  } catch (err: any) {
    alert(err.message || '删除失败')
  }
}

function closeModal() {
  showModal.value = false
}

function computeAmount(volume: string, unitPrice: string) {
  const value = Number(volume || 0) * Number(unitPrice || 0)
  return value > 0 ? value.toFixed(2) : ''
}

function formatNumber(value?: string, suffix = '') {
  const num = Number(value || 0)
  return `${num.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}${suffix}`
}
</script>

<template>
  <div class="max-w-[1100px] mx-auto px-5 py-10">
    <div class="flex items-center gap-4 mb-8">
      <button
        @click="goBack"
        class="p-2 rounded-xl text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] transition-all"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 19l-7-7 7-7"
          />
        </svg>
      </button>
      <div>
        <h1 class="text-[28px] font-semibold tracking-tight text-[#1d1d1f]">
          {{ currentVehicle?.name || '油耗详情' }}
        </h1>
        <p class="mt-0.5 text-sm text-[#86868b]">
          {{ currentVehicle?.plateNo || currentVehicle?.model || '' }}
        </p>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <p class="text-sm text-[#86868b] mb-2">平均油耗</p>
        <p class="text-[24px] font-semibold text-[#1d1d1f]">
          {{ formatNumber(stats?.averageConsumption, ' L/100km') }}
        </p>
      </div>
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <p class="text-sm text-[#86868b] mb-2">最近油耗</p>
        <p class="text-[24px] font-semibold text-[#1d1d1f]">
          {{ formatNumber(stats?.latestConsumption, ' L/100km') }}
        </p>
      </div>
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <p class="text-sm text-[#86868b] mb-2">每公里成本</p>
        <p class="text-[24px] font-semibold text-[#1d1d1f]">
          ¥{{ formatNumber(stats?.costPerKm) }}
        </p>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-5 mb-8">
      <div class="lg:col-span-2 bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold text-[#1d1d1f]">油耗趋势</h2>
          <span class="text-sm text-[#86868b]">{{ stats?.trend?.length || 0 }} 个有效区间</span>
        </div>
        <svg viewBox="0 0 100 100" preserveAspectRatio="none" class="w-full h-48">
          <line x1="0" y1="90" x2="100" y2="90" stroke="#f0f0f0" stroke-width="1" />
          <line x1="0" y1="20" x2="100" y2="20" stroke="#f5f5f7" stroke-width="1" />
          <polyline
            v-if="trendPolyline"
            :points="trendPolyline"
            fill="none"
            stroke="#0071e3"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        <p v-if="!trendPolyline" class="text-center text-sm text-[#86868b] -mt-24 mb-20">
          至少两次加满后生成趋势
        </p>
      </div>
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <p class="text-sm text-[#86868b] mb-2">总里程</p>
        <p class="text-[22px] font-semibold text-[#1d1d1f] mb-5">
          {{ formatNumber(stats?.totalDistance, ' km') }}
        </p>
        <p class="text-sm text-[#86868b] mb-2">总油量</p>
        <p class="text-[22px] font-semibold text-[#1d1d1f] mb-5">
          {{ formatNumber(stats?.totalVolume, ' L') }}
        </p>
        <p class="text-sm text-[#86868b] mb-2">总油费</p>
        <p class="text-[22px] font-semibold text-[#1d1d1f]">
          ¥{{ formatNumber(stats?.totalAmount) }}
        </p>
      </div>
    </div>

    <div class="flex justify-between items-center mb-4">
      <h2 class="text-lg font-semibold text-[#1d1d1f]">加油记录</h2>
      <button
        @click="openCreateModal"
        class="inline-flex items-center gap-2 px-4 py-2 text-white text-sm font-medium rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] shadow-sm"
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
        新增加油
      </button>
    </div>

    <div v-if="recordLoading" class="space-y-3">
      <div v-for="n in 4" :key="n" class="bg-white rounded-xl p-4 border border-[#f0f0f0]">
        <div class="h-4 bg-[#f5f5f7] rounded w-2/3 animate-pulse" />
      </div>
    </div>

    <div
      v-else-if="records.length === 0"
      class="text-center py-16 bg-white rounded-2xl border border-[#f0f0f0]"
    >
      <h3 class="text-lg font-semibold text-[#1d1d1f] mb-1">暂无加油记录</h3>
      <p class="text-sm text-[#86868b] mb-5">添加第一次加油记录</p>
      <button
        @click="openCreateModal"
        class="inline-flex items-center gap-2 px-5 py-2.5 text-white text-sm font-medium rounded-xl transition-all hover:scale-[1.02]"
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
        新增加油
      </button>
    </div>

    <div v-else class="bg-white rounded-2xl border border-[#f0f0f0] overflow-x-auto">
      <table class="w-full text-left min-w-[760px]">
        <thead>
          <tr class="border-b border-[#f0f0f0]">
            <th class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide">
              日期
            </th>
            <th
              class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide text-right"
            >
              里程
            </th>
            <th
              class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide text-right"
            >
              油量
            </th>
            <th
              class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide text-right"
            >
              金额
            </th>
            <th
              class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide text-center"
            >
              加满
            </th>
            <th
              class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide text-right"
            >
              区间油耗
            </th>
            <th
              class="px-5 py-3.5 text-xs font-medium text-[#86868b] uppercase tracking-wide text-right"
            >
              操作
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="record in records"
            :key="record.id"
            class="border-b border-[#f5f5f7] last:border-b-0 hover:bg-[#fafafc] transition-colors"
          >
            <td class="px-5 py-4 text-sm text-[#1d1d1f]">
              {{ record.refuelTime?.split(' ')?.[0] || record.refuelTime }}
            </td>
            <td class="px-5 py-4 text-sm text-[#1d1d1f] text-right">{{ record.odometer }} km</td>
            <td class="px-5 py-4 text-sm text-[#1d1d1f] text-right">{{ record.volume }} L</td>
            <td class="px-5 py-4 text-sm text-[#1d1d1f] text-right">¥{{ record.amount }}</td>
            <td class="px-5 py-4 text-sm text-center">
              <span
                class="text-xs px-2 py-0.5 rounded-full border font-medium"
                :class="
                  record.isFull
                    ? 'bg-green-50 text-green-700 border-green-200'
                    : 'bg-gray-50 text-gray-600 border-gray-200'
                "
                >{{ record.isFull ? '是' : '否' }}</span
              >
            </td>
            <td class="px-5 py-4 text-sm text-[#1d1d1f] text-right">
              {{ record.intervalConsumption || '-' }}
            </td>
            <td class="px-5 py-4 text-right">
              <div class="inline-flex gap-1">
                <button
                  @click="openEditModal(record)"
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
                  @click="handleDelete(record.id)"
                  class="p-1.5 rounded-lg text-[#ff3b30] hover:bg-[#ff3b30]/10 transition-colors"
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
            </td>
          </tr>
        </tbody>
      </table>
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
          class="bg-white w-full max-w-2xl rounded-2xl shadow-2xl overflow-hidden max-h-[90vh] overflow-y-auto"
        >
          <div class="flex justify-between items-center px-6 py-4 border-b border-[#f0f0f0]">
            <h3 class="text-lg font-semibold text-[#1d1d1f]">
              {{ isEditing ? '编辑加油记录' : '新增加油记录' }}
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
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">加油日期</label
              ><input
                v-model="formData.refuelTime"
                type="date"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">总里程 (km)</label
              ><input
                v-model="formData.odometer"
                type="number"
                step="0.01"
                placeholder="请输入总里程"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">油量 (L)</label
              ><input
                v-model="formData.volume"
                type="number"
                step="0.01"
                placeholder="请输入油量"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">单价</label
              ><input
                v-model="formData.unitPrice"
                type="number"
                step="0.01"
                placeholder="请输入单价"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">金额</label
              ><input
                v-model="formData.amount"
                type="number"
                step="0.01"
                placeholder="留空则按油量和单价计算"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">加油站</label
              ><input
                v-model="formData.station"
                type="text"
                placeholder="请输入加油站"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <label class="md:col-span-2 flex items-center gap-2 text-sm font-medium text-[#1d1d1f]">
              <input
                v-model="formData.isFull"
                type="checkbox"
                class="w-4 h-4 rounded border-[#d2d2d7]"
              />
              本次已加满
            </label>
            <div class="md:col-span-2">
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">备注</label
              ><textarea
                v-model="formData.remark"
                rows="3"
                placeholder="请输入备注"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none resize-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
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
              :disabled="!formData.refuelTime.trim() || isSubmitting"
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

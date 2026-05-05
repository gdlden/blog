import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useToast } from 'vue-toastification'
import * as fuelApi from '@/api/fuel'
import type { FuelStats, FuelVehicle, RefuelRecord } from '@/api/fuel'

export const useFuelStore = defineStore('fuel', () => {
  const toast = useToast()
  const vehicles = ref<FuelVehicle[]>([])
  const currentVehicle = ref<FuelVehicle | null>(null)
  const records = ref<RefuelRecord[]>([])
  const stats = ref<FuelStats | null>(null)
  const loading = ref(false)
  const recordLoading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)
  const recordTotal = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(12)
  const recordPage = ref(1)
  const recordPageSize = ref(10)

  const vehicleCount = computed(() => vehicles.value.length)
  const totalPages = computed(() => Math.ceil(total.value / pageSize.value))
  const recordTotalPages = computed(() => Math.ceil(recordTotal.value / recordPageSize.value))

  async function fetchVehicles(page?: number, size?: number): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const pageNum = page ?? currentPage.value
      const pageSizeNum = size ?? pageSize.value
      const response = await fuelApi.getVehicles(String(pageNum), String(pageSizeNum))
      vehicles.value = response.list || []
      total.value = parseInt(response.total || '0', 10)
      currentPage.value = pageNum
      pageSize.value = pageSizeNum
    } catch (err: any) {
      error.value = err.message || '获取车辆列表失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchVehicleById(id: string): Promise<FuelVehicle | null> {
    loading.value = true
    try {
      const vehicle = await fuelApi.getVehicleById(id)
      currentVehicle.value = vehicle
      return vehicle
    } catch (err: any) {
      toast.error(err.message || '获取车辆详情失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createVehicle(data: Omit<FuelVehicle, 'id'>): Promise<void> {
    loading.value = true
    try {
      await fuelApi.createVehicle(data)
      toast.success('车辆创建成功')
      await fetchVehicles()
    } catch (err: any) {
      toast.error(err.message || '创建失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateVehicle(data: FuelVehicle): Promise<void> {
    loading.value = true
    try {
      await fuelApi.updateVehicle(data)
      toast.success('车辆更新成功')
      await fetchVehicles()
      if (currentVehicle.value?.id === data.id) {
        currentVehicle.value = data
      }
    } catch (err: any) {
      toast.error(err.message || '更新失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteVehicle(id: string): Promise<void> {
    loading.value = true
    try {
      await fuelApi.deleteVehicle(id)
      toast.success('车辆删除成功')
      vehicles.value = vehicles.value.filter((vehicle) => vehicle.id !== id)
      total.value = Math.max(0, total.value - 1)
    } catch (err: any) {
      toast.error(err.message || '删除失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchRecords(vehicleId: string, page?: number, size?: number): Promise<void> {
    recordLoading.value = true
    error.value = null
    try {
      const pageNum = page ?? recordPage.value
      const pageSizeNum = size ?? recordPageSize.value
      const response = await fuelApi.getRefuelRecords(
        vehicleId,
        String(pageNum),
        String(pageSizeNum),
      )
      records.value = response.list || []
      recordTotal.value = parseInt(response.total || '0', 10)
      recordPage.value = pageNum
      recordPageSize.value = pageSizeNum
    } catch (err: any) {
      error.value = err.message || '获取加油记录失败'
      throw err
    } finally {
      recordLoading.value = false
    }
  }

  async function fetchStats(vehicleId: string): Promise<void> {
    stats.value = await fuelApi.getFuelStats(vehicleId)
  }

  async function fetchVehicleDashboard(vehicleId: string): Promise<void> {
    await Promise.all([fetchStats(vehicleId), fetchRecords(vehicleId, 1)])
  }

  async function createRefuelRecord(
    data: Omit<RefuelRecord, 'id' | 'intervalConsumption'>,
  ): Promise<void> {
    recordLoading.value = true
    try {
      await fuelApi.createRefuelRecord(data)
      toast.success('加油记录创建成功')
      await fetchVehicleDashboard(data.vehicleId)
    } catch (err: any) {
      toast.error(err.message || '创建失败')
      throw err
    } finally {
      recordLoading.value = false
    }
  }

  async function updateRefuelRecord(data: RefuelRecord): Promise<void> {
    recordLoading.value = true
    try {
      await fuelApi.updateRefuelRecord(data)
      toast.success('加油记录更新成功')
      await fetchVehicleDashboard(data.vehicleId)
    } catch (err: any) {
      toast.error(err.message || '更新失败')
      throw err
    } finally {
      recordLoading.value = false
    }
  }

  async function deleteRefuelRecord(id: string, vehicleId: string): Promise<void> {
    recordLoading.value = true
    try {
      await fuelApi.deleteRefuelRecord(id)
      toast.success('加油记录删除成功')
      await fetchVehicleDashboard(vehicleId)
    } catch (err: any) {
      toast.error(err.message || '删除失败')
      throw err
    } finally {
      recordLoading.value = false
    }
  }

  return {
    vehicles,
    currentVehicle,
    records,
    stats,
    loading,
    recordLoading,
    error,
    total,
    recordTotal,
    currentPage,
    pageSize,
    recordPage,
    recordPageSize,
    vehicleCount,
    totalPages,
    recordTotalPages,
    fetchVehicles,
    fetchVehicleById,
    createVehicle,
    updateVehicle,
    deleteVehicle,
    fetchRecords,
    fetchStats,
    fetchVehicleDashboard,
    createRefuelRecord,
    updateRefuelRecord,
    deleteRefuelRecord,
  }
})

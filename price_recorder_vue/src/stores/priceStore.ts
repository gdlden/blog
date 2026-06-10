import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useToast } from 'vue-toastification'
import * as priceApi from '@/api/price'
import type { PriceItem } from '@/api/price'

export const usePriceStore = defineStore('price', () => {
  const toast = useToast()
  const items = ref<PriceItem[]>([])
  const loading = ref(false)
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(12)

  const totalPages = computed(() => Math.ceil(total.value / pageSize.value))
  const itemCount = computed(() => items.value.length)

  async function fetchItems(page?: number, size?: number): Promise<void> {
    loading.value = true
    try {
      const pageNum = page ?? currentPage.value
      const pageSizeNum = size ?? pageSize.value
      const response = await priceApi.getPriceList(String(pageNum), String(pageSizeNum))
      items.value = response.data || []
      total.value = parseInt(response.total || '0', 10)
      currentPage.value = pageNum
      pageSize.value = pageSizeNum
    } catch (err: any) {
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createItem(data: Omit<PriceItem, 'id' | 'totalPrice'>): Promise<void> {
    loading.value = true
    try {
      await priceApi.createPrice(data)
      toast.success('价格记录创建成功')
      await fetchItems()
    } catch (err: any) {
      toast.error(err.message || '创建失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateItem(data: PriceItem): Promise<void> {
    loading.value = true
    try {
      await priceApi.updatePrice(data)
      toast.success('价格记录更新成功')
      await fetchItems()
    } catch (err: any) {
      toast.error(err.message || '更新失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteItem(id: string): Promise<void> {
    loading.value = true
    try {
      await priceApi.deletePrice(id)
      toast.success('价格记录删除成功')
      items.value = items.value.filter((item) => item.id !== id)
      total.value = Math.max(0, total.value - 1)
    } catch (err: any) {
      toast.error(err.message || '删除失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    items,
    loading,
    total,
    currentPage,
    pageSize,
    totalPages,
    itemCount,
    fetchItems,
    createItem,
    updateItem,
    deleteItem,
  }
})

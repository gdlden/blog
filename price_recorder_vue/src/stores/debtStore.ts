import { ref, computed } from "vue";
import { defineStore } from "pinia";
import { useToast } from "vue-toastification";
import * as debtApi from "@/api/debt";
import type { Debt } from "@/api/debt";

export const useDebtStore = defineStore("debt", () => {
  const toast = useToast();
  // State
  const debts = ref<Debt[]>([]);
  const currentDebt = ref<Debt | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const total = ref(0);
  const currentPage = ref(1);
  const pageSize = ref(10);

  // Getters
  const debtCount = computed(() => debts.value.length);
  const totalPages = computed(() => Math.ceil(total.value / pageSize.value));

  const totalDebt = computed(() => {
    return debts.value.reduce((sum, d) => sum + (parseFloat(d.amount) || 0), 0);
  });

  const repaidAmount = computed(() => {
    return debts.value
      .filter((d) => d.status === "已结清" || d.status === "repaid")
      .reduce((sum, d) => sum + (parseFloat(d.amount) || 0), 0);
  });

  const outstandingAmount = computed(() => {
    return debts.value
      .filter((d) => d.status !== "已结清" && d.status !== "repaid")
      .reduce((sum, d) => sum + (parseFloat(d.amount) || 0), 0);
  });

  // Actions
  async function fetchDebts(page?: number, size?: number): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      const pageNum = page ?? currentPage.value;
      const pageSizeNum = size ?? pageSize.value;
      const response = await debtApi.getDebts(String(pageNum), String(pageSizeNum));
      debts.value = response.list || [];
      total.value = parseInt(response.total || "0", 10);
      currentPage.value = pageNum;
      pageSize.value = pageSizeNum;
    } catch (err: any) {
      error.value = err.message || "获取债务列表失败";
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function createDebt(data: Omit<Debt, "id">): Promise<void> {
    loading.value = true;
    try {
      await debtApi.createDebt(data);
      toast.success('债务记录创建成功');
      await fetchDebts();
    } catch (err: any) {
      toast.error(err.message || '创建失败');
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function updateDebt(data: Debt): Promise<void> {
    loading.value = true;
    try {
      await debtApi.updateDebt(data);
      toast.success('债务记录更新成功');
      await fetchDebts();
    } catch (err: any) {
      toast.error(err.message || '更新失败');
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function deleteDebt(id: string): Promise<void> {
    loading.value = true;
    try {
      await debtApi.deleteDebt(id);
      toast.success('债务记录删除成功');
      debts.value = debts.value.filter((debt) => debt.id !== id);
      total.value = Math.max(0, total.value - 1);
    } catch (err: any) {
      toast.error(err.message || '删除失败');
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function fetchDebtById(id: string): Promise<Debt | null> {
    loading.value = true;
    try {
      const debt = await debtApi.getDebtById(id);
      currentDebt.value = debt;
      return debt;
    } catch (err: any) {
      toast.error(err.message || '获取债务详情失败');
      throw err;
    } finally {
      loading.value = false;
    }
  }

  function setCurrentDebt(debt: Debt | null): void {
    currentDebt.value = debt;
  }

  return {
    // State
    debts,
    currentDebt,
    loading,
    error,
    total,
    currentPage,
    pageSize,
    // Getters
    debtCount,
    totalPages,
    totalDebt,
    repaidAmount,
    outstandingAmount,
    // Actions
    fetchDebts,
    createDebt,
    updateDebt,
    deleteDebt,
    fetchDebtById,
    setCurrentDebt,
  };
});

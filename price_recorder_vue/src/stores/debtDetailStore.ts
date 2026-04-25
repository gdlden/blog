import { ref, computed } from "vue";
import { defineStore } from "pinia";
import { useToast } from "vue-toastification";
import * as debtDetailApi from "@/api/debtDetail";
import type { DebtDetail } from "@/api/debtDetail";

export const useDebtDetailStore = defineStore("debtDetail", () => {
  const toast = useToast();

  const details = ref<DebtDetail[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const totalPrincipal = computed(() => {
    return details.value.reduce((sum, d) => sum + (parseFloat(d.principal) || 0), 0);
  });

  const totalInterest = computed(() => {
    return details.value.reduce((sum, d) => sum + (parseFloat(d.interest) || 0), 0);
  });

  const totalRepaid = computed(() => totalPrincipal.value + totalInterest.value);

  async function fetchDetails(debtId: string): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      details.value = await debtDetailApi.getDebtDetails(debtId);
    } catch (err: any) {
      error.value = err.message || "获取债务明细失败";
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function createDetail(data: Omit<DebtDetail, "id">): Promise<void> {
    loading.value = true;
    try {
      await debtDetailApi.createDebtDetail(data);
      toast.success("明细创建成功");
      await fetchDetails(data.debtId);
    } catch (err: any) {
      toast.error(err.message || "创建失败");
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function updateDetail(data: DebtDetail): Promise<void> {
    loading.value = true;
    try {
      await debtDetailApi.updateDebtDetail(data);
      toast.success("明细更新成功");
      await fetchDetails(data.debtId);
    } catch (err: any) {
      toast.error(err.message || "更新失败");
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function deleteDetail(id: string, debtId: string): Promise<void> {
    loading.value = true;
    try {
      await debtDetailApi.deleteDebtDetail(id);
      toast.success("明细删除成功");
      details.value = details.value.filter((d) => d.id !== id);
    } catch (err: any) {
      toast.error(err.message || "删除失败");
      throw err;
    } finally {
      loading.value = false;
    }
  }

  return {
    details,
    loading,
    error,
    totalPrincipal,
    totalInterest,
    totalRepaid,
    fetchDetails,
    createDetail,
    updateDetail,
    deleteDetail,
  };
});

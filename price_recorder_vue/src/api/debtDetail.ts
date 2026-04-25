import instance from "@/utils/request.ts";

export interface DebtDetail {
  id: string;
  debtId: string;
  postingDate: string;
  principal: string;
  interest: string;
  period: string;
}

export async function getDebtDetails(debtId: string): Promise<DebtDetail[]> {
  const response = await instance.get("/debtDetail/list/v1", { params: { debtId } }) as { list: DebtDetail[] };
  return response.list || [];
}

export async function getDebtDetailById(id: string): Promise<DebtDetail> {
  return await instance.get("/debtDetail/get/v1", { params: { id } });
}

export async function createDebtDetail(data: Omit<DebtDetail, "id">): Promise<DebtDetail> {
  return await instance.post("/debtDetail/save/v1", data);
}

export async function updateDebtDetail(data: DebtDetail): Promise<DebtDetail> {
  return await instance.post("/debtDetail/edit/v1", data);
}

export async function deleteDebtDetail(id: string): Promise<boolean> {
  const response = await instance.post("/debtDetail/delete/v1", { id }) as { success: boolean };
  return response.success;
}

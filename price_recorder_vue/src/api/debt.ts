import instance from "@/utils/request.ts";

export interface Debt {
  id: string;
  name: string;
  bankName: string;
  bankAccount: string;
  applyTime: string;
  endTime: string;
  amount: string;
  status: string;
  remark: string;
  apr: string;
  fee: string;
  tenor: string;
}

export interface DebtPageResponse {
  page: string;
  total: string;
  list: Debt[];
}

export async function getDebts(
  page?: string,
  pageSize?: string
): Promise<DebtPageResponse> {
  const params: Record<string, string> = {};
  if (page) params.page = page;
  if (pageSize) params.pageSize = pageSize;
  return await instance.get("/debt/page/v1", { params }).then((res) => res.data);
}

export async function getDebtById(id: string): Promise<Debt> {
  return await instance.get("/debt/get/v1", { params: { id } }).then((res) => res.data);
}

export async function createDebt(data: Omit<Debt, "id">): Promise<Debt> {
  return await instance.post("/debt/save/v1", data).then((res) => res.data);
}

export async function updateDebt(data: Debt): Promise<Debt> {
  return await instance.post("/debt/update/v1", data).then((res) => res.data);
}

export async function deleteDebt(id: string): Promise<boolean> {
  const response = await instance.post("/debt/delete/v1", { id });
  return response.data.flag;
}

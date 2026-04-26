import instance from "@/utils/request.ts";

export interface DebtDetail {
  id: string;
  debtId: string;
  postingDate: string;
  principal: string;
  interest: string;
  period: string;
}

export interface DebtDetailOcrReply {
  rawText: string;
  items: Omit<DebtDetail, "id">[];
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

export async function recognizeDebtDetailOcr(
  file: File,
  debtId: string,
  year = new Date().getFullYear(),
): Promise<DebtDetailOcrReply> {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("debtId", debtId);
  formData.append("year", String(year));
  return await instance.post("/debtDetail/ocr/v1", formData, {
    headers: { "Content-Type": "multipart/form-data" },
  });
}

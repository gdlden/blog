import instance from "@/utils/request"

export interface AppVersionEntity {
  id: number
  version: string
  info: string[]
  iosUrl: string
  androidUrl: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface AppVersionPageResponse {
  page: number
  pageSize: number
  total: number
  list: AppVersionEntity[]
}

export async function getAppVersions(
  page?: number,
  pageSize?: number
): Promise<AppVersionPageResponse> {
  const params: Record<string, string | number> = {}
  if (page) params.page = page
  if (pageSize) params.pageSize = pageSize
  return await instance.get("/app/version/page/v1", { params })
}

export async function getAppVersionById(id: number): Promise<AppVersionEntity> {
  return await instance.get("/app/version/get/v1", { params: { id } })
}

export async function createAppVersion(
  data: Omit<AppVersionEntity, "id" | "createdAt" | "updatedAt">
): Promise<{ id: number }> {
  return await instance.post("/app/version/create/v1", data)
}

export async function updateAppVersion(
  data: Omit<AppVersionEntity, "createdAt" | "updatedAt">
): Promise<{ id: number }> {
  return await instance.post("/app/version/update/v1", data)
}

export async function deleteAppVersion(id: number): Promise<{ success: boolean }> {
  return await instance.post("/app/version/delete/v1", { id })
}

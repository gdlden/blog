import instance from '@/utils/request.ts'

export interface PriceItem {
  id: string
  productName: string
  weight: string
  unitPrice: string
  totalPrice: string
  priceDate: string
}

export interface PricePageResponse {
  current: string
  size: string
  total: string
  data: PriceItem[]
}

export interface SavePriceReply {
  id: string
}

type PriceCreateRequest = {
  productName: string
  weight: string | number
  unitPrice: string | number
  priceDate: string
}

type PriceUpdateRequest = {
  id: string
  productName: string
  weight: string | number
  unitPrice: string | number
  priceDate: string
}

export async function getPriceList(
  current?: string,
  size?: string,
): Promise<PricePageResponse> {
  const params: Record<string, string> = {}
  if (current) params.current = current
  if (size) params.size = size
  return await instance.get('/price/list/v1', { params })
}

export async function getPriceById(id: string): Promise<PriceItem> {
  return await instance.get(`/price/${id}/v1`)
}

export async function createPrice(data: PriceCreateRequest): Promise<SavePriceReply> {
  return await instance.post('/price/add/v1', data)
}

export async function updatePrice(data: PriceUpdateRequest): Promise<SavePriceReply> {
  return await instance.put(`/price/${data.id}/v1`, data)
}

export async function deletePrice(id: string): Promise<void> {
  return await instance.post(`/price/${id}/delete/v1`)
}

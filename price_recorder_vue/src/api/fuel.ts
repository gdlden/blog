import instance from '@/utils/request.ts'

export interface FuelVehicle {
  id: string
  name: string
  plateNo: string
  brand: string
  model: string
  tankCapacity: string
  remark: string
}

export interface RefuelRecord {
  id: string
  vehicleId: string
  refuelTime: string
  odometer: string
  volume: string
  unitPrice: string
  amount: string
  station: string
  isFull: boolean
  remark: string
  intervalConsumption: string
}

export interface FuelTrendPoint {
  refuelTime: string
  odometer: string
  consumption: string
  distance: string
  volume: string
  refuelRecordId: string
}

export interface FuelStats {
  vehicleId: string
  totalDistance: string
  totalVolume: string
  totalAmount: string
  averageConsumption: string
  latestConsumption: string
  costPerKm: string
  trend: FuelTrendPoint[]
}

export interface FuelPageResponse<T> {
  page: string
  total: string
  list: T[]
}

export interface SaveFuelReply {
  id: string
  message: string
}

type FuelVehicleRequest = Omit<FuelVehicle, 'id'> & {
  tankCapacity: string | number
}

type RefuelRecordRequest = Omit<RefuelRecord, 'id' | 'intervalConsumption'> & {
  odometer: string | number
  volume: string | number
  unitPrice: string | number
  amount: string | number
}

function decimalToString(value: string | number | undefined): string {
  if (value === undefined || value === null) return ''
  return String(value)
}

function serializeVehicle(data: FuelVehicle | FuelVehicleRequest) {
  return {
    ...data,
    tankCapacity: decimalToString(data.tankCapacity),
  }
}

function serializeRefuelRecord(data: RefuelRecord | RefuelRecordRequest) {
  return {
    ...data,
    odometer: decimalToString(data.odometer),
    volume: decimalToString(data.volume),
    unitPrice: decimalToString(data.unitPrice),
    amount: decimalToString(data.amount),
  }
}

export async function getVehicles(
  page?: string,
  pageSize?: string,
): Promise<FuelPageResponse<FuelVehicle>> {
  const params: Record<string, string> = {}
  if (page) params.page = page
  if (pageSize) params.pageSize = pageSize
  return await instance.get('/fuel/vehicle/page/v1', { params })
}

export async function getVehicleById(id: string): Promise<FuelVehicle> {
  return await instance.get('/fuel/vehicle/get/v1', { params: { id } })
}

export async function createVehicle(data: FuelVehicleRequest): Promise<SaveFuelReply> {
  return await instance.post('/fuel/vehicle/save/v1', serializeVehicle(data))
}

export async function updateVehicle(data: FuelVehicle): Promise<SaveFuelReply> {
  return await instance.post('/fuel/vehicle/update/v1', serializeVehicle(data))
}

export async function deleteVehicle(id: string): Promise<boolean> {
  const response = (await instance.post('/fuel/vehicle/delete/v1', { id })) as { flag: boolean }
  return response.flag
}

export async function getRefuelRecords(
  vehicleId: string,
  page?: string,
  pageSize?: string,
): Promise<FuelPageResponse<RefuelRecord>> {
  const params: Record<string, string> = { vehicleId }
  if (page) params.page = page
  if (pageSize) params.pageSize = pageSize
  return await instance.get('/fuel/refuel/page/v1', { params })
}

export async function getRefuelRecordById(id: string): Promise<RefuelRecord> {
  return await instance.get('/fuel/refuel/get/v1', { params: { id } })
}

export async function createRefuelRecord(data: RefuelRecordRequest): Promise<SaveFuelReply> {
  return await instance.post('/fuel/refuel/save/v1', serializeRefuelRecord(data))
}

export async function updateRefuelRecord(data: RefuelRecord): Promise<SaveFuelReply> {
  return await instance.post('/fuel/refuel/update/v1', serializeRefuelRecord(data))
}

export async function deleteRefuelRecord(id: string): Promise<boolean> {
  const response = (await instance.post('/fuel/refuel/delete/v1', { id })) as { flag: boolean }
  return response.flag
}

export async function getFuelStats(vehicleId: string): Promise<FuelStats> {
  return await instance.get('/fuel/stats/v1', { params: { vehicleId } })
}

import instance from '@/utils/request.ts'

/* ---------- 枚举与标签 ---------- */

export type LedgerAccountType = 'asset' | 'liability'
export type LedgerDirection = 'expense' | 'income'
export type LedgerTransactionType = 'expense' | 'income' | 'transfer'

export const ACCOUNT_TYPE_LABELS: Record<LedgerAccountType, string> = {
  asset: '资产',
  liability: '负债',
}

export const ASSET_SUBTYPES = [
  'cash',
  'debit_card',
  'e_wallet',
  'prepaid_card',
  'investment',
  'other',
] as const
export const LIABILITY_SUBTYPES = [
  'credit_card',
  'huabei_like',
  'loan_payable',
  'loan_receivable',
  'other',
] as const

export const ACCOUNT_SUBTYPE_LABELS: Record<string, string> = {
  cash: '现金',
  debit_card: '储蓄卡',
  e_wallet: '电子钱包',
  prepaid_card: '预付卡',
  investment: '投资账户',
  other: '其他',
  credit_card: '信用卡',
  huabei_like: '花呗/白条',
  loan_payable: '应付借款',
  loan_receivable: '应收借款',
}

export const TRANSACTION_TYPE_LABELS: Record<LedgerTransactionType, string> = {
  expense: '支出',
  income: '收入',
  transfer: '转账',
}

export const DIRECTION_LABELS: Record<LedgerDirection, string> = {
  expense: '支出',
  income: '收入',
}

/* ---------- 账户 ---------- */

// 契约：id 为 JSON 数字（响应边界已被 request.ts 统一转成字符串）；金额字段为 string
export interface LedgerAccount {
  id: string
  name: string
  type: LedgerAccountType
  subtype: string
  creditLimit?: string
  billingDay?: number
  paymentDueDay?: number
  remark?: string
  sort?: number
  archived?: boolean
  balance?: string
  isSystem?: boolean
}

export interface LedgerAccountListResponse {
  list: LedgerAccount[]
}

export interface SaveLedgerReply {
  id: string
  message: string
}

// openingBalance 仅 save 使用（初始余额）；update 带 id 整对象更新
export interface LedgerAccountRequest {
  id?: string
  name: string
  type: string
  subtype: string
  creditLimit?: string | number
  billingDay?: string | number
  paymentDueDay?: string | number
  remark?: string
  openingBalance?: string | number
  archived?: boolean
  sort?: number
}

/* ---------- 分类 ---------- */

export interface LedgerCategory {
  id: string
  parentId: string
  name: string
  direction: LedgerDirection
  sort?: number
  isSystem?: boolean
}

export interface LedgerCategoryListResponse {
  list: LedgerCategory[]
}

export interface LedgerCategoryRequest {
  id?: string
  parentId?: string | number
  name: string
  direction: LedgerDirection
  sort?: number
}

/* ---------- 交易 ---------- */

export interface LedgerPosting {
  id?: string
  accountId: string
  amount: string
  categoryId?: string
  sort: number
}

export interface LedgerTransaction {
  id: string
  type: LedgerTransactionType
  bookedAt: string
  remark: string
  postings: LedgerPosting[]
}

export interface PostingDraft {
  accountId: string | number
  amount: string | number
  categoryId?: string | number
  sort: number
}

export interface TransactionLeg {
  amount: string | number
  categoryId: string | number
}

export interface LedgerTransactionRequest {
  id?: string
  type: LedgerTransactionType
  bookedAt: string
  remark: string
  postings: PostingDraft[]
}

export interface LedgerTransactionPageResponse {
  page: string
  total: string
  list: LedgerTransaction[]
}

/* ---------- 月度统计 ---------- */

export interface LedgerCategoryAmount {
  categoryId: string
  categoryName: string
  amount: string
}

export interface LedgerMonthlyStats {
  month: string
  totalExpense: string
  totalIncome: string
  expenseByCategory: LedgerCategoryAmount[]
  incomeByCategory: LedgerCategoryAmount[]
}

/* ---------- 分类预算 ---------- */

// 契约：id/categoryId 为 JSON 数字（响应边界统一转成字符串）；amount/used 为 string
// used = 当月该分类含子分类支出合计
export interface LedgerBudget {
  id: string
  categoryId: string
  categoryName: string
  amount: string
  used: string
}

export interface LedgerBudgetListResponse {
  list: LedgerBudget[]
}

// upsert：同分类同月覆盖；仅 expense 方向分类；month 格式 "YYYY-MM"
export interface LedgerBudgetRequest {
  categoryId: string | number
  month: string
  amount: string | number
}

/* ---------- 余额走势 ---------- */

export interface LedgerBalancePoint {
  date: string
  balance: string
}

export interface LedgerBalanceTrendResponse {
  points: LedgerBalancePoint[]
}

// accountId 空 = 净资产（全部账户合计），传参时省略；startTime/endTime 格式 "YYYY-MM-DD"
export interface LedgerBalanceTrendParams {
  accountId?: string
  startTime?: string
  endTime?: string
}

/* ---------- 周期账单 ---------- */

// 契约：id/accountId/categoryId 为 JSON 数字（响应边界统一转成字符串）；amount 为 string
// nextDate 为下次生成日期 "YYYY-MM-DD"；lastGeneratedMonth/startMonth 格式 "YYYY-MM"
export interface LedgerRecurring {
  id: string
  accountId: string
  accountName: string
  categoryId: string
  categoryName: string
  type: LedgerDirection
  amount: string
  remark: string
  dayOfMonth: number
  startMonth: string
  lastGeneratedMonth: string
  enabled: boolean
  nextDate: string
}

export interface LedgerRecurringListResponse {
  list: LedgerRecurring[]
}

export interface LedgerRecurringApplyResponse {
  created: number
}

// save 幂等：无 id 创建、带 id 整对象更新
export interface LedgerRecurringRequest {
  id?: string
  accountId: string | number
  categoryId: string | number
  type: LedgerDirection
  amount: string | number
  remark: string
  dayOfMonth: string | number
  startMonth: string
  enabled: boolean
}

/* ---------- 序列化 ---------- */

function decimalToString(value: string | number | undefined): string {
  if (value === undefined || value === null) return ''
  return String(value)
}

function serializeAccount(data: LedgerAccountRequest) {
  const body: Record<string, unknown> = {
    name: data.name,
    type: data.type,
    subtype: data.subtype,
    remark: data.remark ?? '',
  }
  if (data.id) body.id = Number(data.id)
  if (data.creditLimit !== undefined && data.creditLimit !== '')
    body.creditLimit = decimalToString(data.creditLimit)
  if (data.billingDay !== undefined && data.billingDay !== '')
    body.billingDay = Number(data.billingDay)
  if (data.paymentDueDay !== undefined && data.paymentDueDay !== '')
    body.paymentDueDay = Number(data.paymentDueDay)
  if (data.openingBalance !== undefined && data.openingBalance !== '')
    body.openingBalance = decimalToString(data.openingBalance)
  if (data.archived !== undefined) body.archived = data.archived
  if (data.sort !== undefined) body.sort = data.sort
  return body
}

function serializeCategory(data: LedgerCategoryRequest) {
  const body: Record<string, unknown> = {
    name: data.name,
    direction: data.direction,
    // 契约：parentId=0/空 表示一级分类
    parentId: Number(data.parentId || 0),
  }
  if (data.id) body.id = Number(data.id)
  if (data.sort !== undefined) body.sort = data.sort
  return body
}

function serializeTransaction(data: LedgerTransactionRequest) {
  const body: Record<string, unknown> = {
    type: data.type,
    bookedAt: data.bookedAt,
    remark: data.remark ?? '',
    postings: data.postings.map((posting) => {
      const item: Record<string, unknown> = {
        accountId: Number(posting.accountId),
        amount: decimalToString(posting.amount),
        sort: posting.sort,
      }
      if (posting.categoryId !== undefined && posting.categoryId !== '') {
        item.categoryId = Number(posting.categoryId)
      }
      return item
    }),
  }
  if (data.id) body.id = Number(data.id)
  return body
}

// 预算：categoryId 传 JSON 数字，amount 传 string，month 原样 "YYYY-MM"
function serializeBudget(data: LedgerBudgetRequest) {
  return {
    categoryId: Number(data.categoryId),
    month: data.month,
    amount: decimalToString(data.amount),
  }
}

// 周期账单：id/accountId/categoryId/dayOfMonth 传 JSON 数字，amount 传 string，startMonth 原样
function serializeRecurring(data: LedgerRecurringRequest) {
  const body: Record<string, unknown> = {
    accountId: Number(data.accountId),
    categoryId: Number(data.categoryId),
    type: data.type,
    amount: decimalToString(data.amount),
    remark: data.remark ?? '',
    dayOfMonth: Number(data.dayOfMonth),
    startMonth: data.startMonth,
    enabled: data.enabled,
  }
  if (data.id) body.id = Number(data.id)
  return body
}

/* ---------- postings 构造（真复式，用户无感） ----------
 * 约定：前端只传"用户侧腿 + 分类腿"，后端把 accountId=0 的腿落到
 * 系统费用账户（expense）/系统收入账户（income）；前端保证自传腿 Σ=0。
 * 支出：付款账户 -总额 + 费用腿（accountId=0, +腿金额, categoryId）
 * 收入：收款账户 +总额 + 收入腿（accountId=0, -腿金额, categoryId）
 * 转账：转出 -X、转入 +X，两腿都是用户账户。
 */

function signedAmount(value: string | number, negative: boolean): string {
  const abs = String(value).trim().replace(/^-/, '')
  return negative ? `-${abs}` : abs
}

export function buildExpensePostings(
  accountId: string | number,
  total: string | number,
  legs: TransactionLeg[],
): PostingDraft[] {
  return [
    { accountId, amount: signedAmount(total, true), sort: 0 },
    ...legs.map((leg, index) => ({
      accountId: 0,
      amount: signedAmount(leg.amount, false),
      categoryId: leg.categoryId,
      sort: index + 1,
    })),
  ]
}

export function buildIncomePostings(
  accountId: string | number,
  total: string | number,
  legs: TransactionLeg[],
): PostingDraft[] {
  return [
    { accountId, amount: signedAmount(total, false), sort: 0 },
    ...legs.map((leg, index) => ({
      accountId: 0,
      amount: signedAmount(leg.amount, true),
      categoryId: leg.categoryId,
      sort: index + 1,
    })),
  ]
}

export function buildTransferPostings(
  fromAccountId: string | number,
  toAccountId: string | number,
  amount: string | number,
): PostingDraft[] {
  return [
    { accountId: fromAccountId, amount: signedAmount(amount, true), sort: 0 },
    { accountId: toAccountId, amount: signedAmount(amount, false), sort: 1 },
  ]
}

/* ---------- 月份/格式化工具 ---------- */

export function currentMonth(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

function parseMonth(month: string): { year: number; monthIndex: number } {
  const parts = month.split('-')
  return { year: Number(parts[0]), monthIndex: Number(parts[1]) - 1 }
}

export function shiftMonth(month: string, delta: number): string {
  const { year, monthIndex } = parseMonth(month)
  const date = new Date(year, monthIndex + delta, 1)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
}

export function monthRange(month: string): { startTime: string; endTime: string } {
  const { year, monthIndex } = parseMonth(month)
  const lastDay = new Date(year, monthIndex + 1, 0).getDate()
  return {
    startTime: `${month}-01 00:00:00`,
    endTime: `${month}-${String(lastDay).padStart(2, '0')} 23:59:59`,
  }
}

// 走势范围：结束今天，往前推 months 个月；契约日期格式 "YYYY-MM-DD"（缺省近 6 个月）
export function trendDateRange(months: number): { startTime: string; endTime: string } {
  const pad = (n: number) => String(n).padStart(2, '0')
  const formatDate = (date: Date) =>
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth() - months, now.getDate())
  return { startTime: formatDate(start), endTime: formatDate(now) }
}

export function formatAmount(value?: string | number): string {
  const num = Number(value || 0)
  return num.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

export function ledgerErrorMessage(err: unknown, fallback: string): string {
  return err instanceof Error && err.message ? err.message : fallback
}

/* ---------- API ---------- */

export async function getAccountList(): Promise<LedgerAccountListResponse> {
  return await instance.get('/ledger/account/list/v1')
}

export async function saveAccount(data: LedgerAccountRequest): Promise<SaveLedgerReply> {
  return await instance.post('/ledger/account/save/v1', serializeAccount(data))
}

export async function updateAccount(data: LedgerAccountRequest): Promise<SaveLedgerReply> {
  return await instance.post('/ledger/account/update/v1', serializeAccount(data))
}

export async function deleteAccount(id: string): Promise<boolean> {
  const response = (await instance.post('/ledger/account/delete/v1', { id: Number(id) })) as {
    flag: boolean
  }
  return response.flag
}

export async function getCategoryList(): Promise<LedgerCategoryListResponse> {
  return await instance.get('/ledger/category/list/v1')
}

export async function saveCategory(data: LedgerCategoryRequest): Promise<SaveLedgerReply> {
  return await instance.post('/ledger/category/save/v1', serializeCategory(data))
}

export async function updateCategory(data: LedgerCategoryRequest): Promise<SaveLedgerReply> {
  return await instance.post('/ledger/category/update/v1', serializeCategory(data))
}

export async function deleteCategory(id: string): Promise<boolean> {
  const response = (await instance.post('/ledger/category/delete/v1', { id: Number(id) })) as {
    flag: boolean
  }
  return response.flag
}

export interface LedgerTransactionPageParams {
  page?: string
  pageSize?: string
  accountId?: string
  categoryId?: string
  type?: string
  startTime?: string
  endTime?: string
}

export async function getTransactionPage(
  params: LedgerTransactionPageParams,
): Promise<LedgerTransactionPageResponse> {
  const query: Record<string, string> = {}
  if (params.page) query.page = params.page
  if (params.pageSize) query.pageSize = params.pageSize
  if (params.accountId) query.accountId = params.accountId
  if (params.categoryId) query.categoryId = params.categoryId
  if (params.type) query.type = params.type
  if (params.startTime) query.startTime = params.startTime
  if (params.endTime) query.endTime = params.endTime
  return await instance.get('/ledger/transaction/page/v1', { params: query })
}

export async function getTransactionById(id: string): Promise<LedgerTransaction> {
  return await instance.get('/ledger/transaction/get/v1', { params: { id } })
}

export async function createTransaction(data: LedgerTransactionRequest): Promise<SaveLedgerReply> {
  return await instance.post('/ledger/transaction/save/v1', serializeTransaction(data))
}

export async function updateTransaction(data: LedgerTransactionRequest): Promise<SaveLedgerReply> {
  return await instance.post('/ledger/transaction/update/v1', serializeTransaction(data))
}

export async function deleteTransaction(id: string): Promise<boolean> {
  const response = (await instance.post('/ledger/transaction/delete/v1', { id: Number(id) })) as {
    flag: boolean
  }
  return response.flag
}

export async function getMonthlyStats(month: string): Promise<LedgerMonthlyStats> {
  return await instance.get('/ledger/stats/monthly/v1', { params: { month } })
}

export async function saveBudget(data: LedgerBudgetRequest): Promise<SaveLedgerReply> {
  return await instance.post('/ledger/budget/save/v1', serializeBudget(data))
}

export async function deleteBudget(id: string): Promise<boolean> {
  const response = (await instance.post('/ledger/budget/delete/v1', { id: Number(id) })) as {
    flag: boolean
  }
  return response.flag
}

export async function getBudgetList(month: string): Promise<LedgerBudgetListResponse> {
  return await instance.get('/ledger/budget/list/v1', { params: { month } })
}

export async function getBalanceTrend(
  params: LedgerBalanceTrendParams,
): Promise<LedgerBalanceTrendResponse> {
  const query: Record<string, string> = {}
  if (params.accountId) query.accountId = params.accountId
  if (params.startTime) query.startTime = params.startTime
  if (params.endTime) query.endTime = params.endTime
  return await instance.get('/ledger/stats/balance-trend/v1', { params: query })
}

export async function saveRecurring(data: LedgerRecurringRequest): Promise<SaveLedgerReply> {
  return await instance.post('/ledger/recurring/save/v1', serializeRecurring(data))
}

export async function deleteRecurring(id: string): Promise<boolean> {
  const response = (await instance.post('/ledger/recurring/delete/v1', { id: Number(id) })) as {
    flag: boolean
  }
  return response.flag
}

export async function getRecurringList(): Promise<LedgerRecurringListResponse> {
  return await instance.get('/ledger/recurring/list/v1')
}

export async function applyRecurring(): Promise<LedgerRecurringApplyResponse> {
  return await instance.post('/ledger/recurring/apply/v1', {})
}

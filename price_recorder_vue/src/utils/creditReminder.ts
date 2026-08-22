import type { LedgerAccount, LedgerTransaction } from '@/api/ledger'

export interface CreditBillingPeriod {
  periodStart: string
  periodEnd: string
  dueDate: string
}

export interface CreditReminder {
  accountId: string
  accountName: string
  amountDue: number
  dueDate: string
  daysUntilDue: number
  level: 'warning' | 'info'
  periodStart: string
  periodEnd: string
}

const DAY_MS = 24 * 60 * 60 * 1000

function pad2(value: number): string {
  return String(value).padStart(2, '0')
}

function formatDate(date: Date): string {
  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())}`
}

function parseDate(value: string): Date {
  const [year, month, day] = value.split('-').map(Number)
  return new Date(year ?? 1970, (month ?? 1) - 1, day ?? 1)
}

// 账单/还款日取 1-31，短月钳到月末（如 31 号在 2 月落到 28 号）
function clampDay(year: number, monthIndex: number, day: number): number {
  const daysInMonth = new Date(year, monthIndex + 1, 0).getDate()
  return Math.min(day, daysInMonth)
}

function validDay(day: number | undefined): day is number {
  return typeof day === 'number' && Number.isInteger(day) && day >= 1 && day <= 31
}

// 账单日当天视为已过：本期区间 = (上月账单日, 本月账单日]；还款日大于账单日还当月，否则还次月
export function creditBillingPeriod(
  account: Pick<LedgerAccount, 'billingDay' | 'paymentDueDay'>,
  today: Date = new Date(),
): CreditBillingPeriod | null {
  const { billingDay, paymentDueDay } = account
  if (!validDay(billingDay) || !validDay(paymentDueDay)) return null

  const now = new Date(today.getFullYear(), today.getMonth(), today.getDate())
  const thisBilling = new Date(
    now.getFullYear(),
    now.getMonth(),
    clampDay(now.getFullYear(), now.getMonth(), billingDay),
  )

  const periodEnd =
    thisBilling.getTime() <= now.getTime()
      ? thisBilling
      : new Date(
          now.getFullYear(),
          now.getMonth() - 1,
          clampDay(now.getFullYear(), now.getMonth() - 1, billingDay),
        )
  const periodStart = new Date(
    periodEnd.getFullYear(),
    periodEnd.getMonth() - 1,
    clampDay(periodEnd.getFullYear(), periodEnd.getMonth() - 1, billingDay),
  )

  const dueMonth = new Date(
    periodEnd.getFullYear(),
    periodEnd.getMonth() + (paymentDueDay > billingDay ? 0 : 1),
    1,
  )
  const dueDate = new Date(
    dueMonth.getFullYear(),
    dueMonth.getMonth(),
    clampDay(dueMonth.getFullYear(), dueMonth.getMonth(), paymentDueDay),
  )

  return {
    periodStart: formatDate(periodStart),
    periodEnd: formatDate(periodEnd),
    dueDate: formatDate(dueDate),
  }
}

// 本期应还 = 区间 (periodStart, periodEnd] 内该账户 expense 负腿合计，按分累加避免浮点误差
export function computeCreditReminder(
  account: LedgerAccount,
  transactions: LedgerTransaction[],
  today: Date = new Date(),
): CreditReminder | null {
  const period = creditBillingPeriod(account, today)
  if (!period) return null

  let cents = 0
  for (const tx of transactions) {
    if (tx.type !== 'expense') continue
    const bookedDate = (tx.bookedAt || '').slice(0, 10)
    if (!bookedDate || bookedDate <= period.periodStart || bookedDate > period.periodEnd) continue
    for (const posting of tx.postings) {
      if (String(posting.accountId) === account.id && Number(posting.amount) < 0) {
        cents += Math.round(Math.abs(Number(posting.amount)) * 100)
      }
    }
  }
  const amountDue = cents / 100
  if (amountDue <= 0) return null

  const now = new Date(today.getFullYear(), today.getMonth(), today.getDate())
  const daysUntilDue = Math.round((parseDate(period.dueDate).getTime() - now.getTime()) / DAY_MS)

  return {
    accountId: account.id,
    accountName: account.name,
    amountDue,
    dueDate: period.dueDate,
    daysUntilDue,
    level: daysUntilDue <= 7 ? 'warning' : 'info',
    periodStart: period.periodStart,
    periodEnd: period.periodEnd,
  }
}

// 拉取交易用的时间范围：含起止当天，开始日的开区间排除在 compute 内完成
export function creditReminderQueryRange(
  account: Pick<LedgerAccount, 'billingDay' | 'paymentDueDay'>,
  today: Date = new Date(),
): { startTime: string; endTime: string } | null {
  const period = creditBillingPeriod(account, today)
  if (!period) return null
  return {
    startTime: `${period.periodStart} 00:00:00`,
    endTime: `${period.periodEnd} 23:59:59`,
  }
}

export function reminderBadgeText(reminder: Pick<CreditReminder, 'daysUntilDue'>): string {
  if (reminder.daysUntilDue > 0) return `还有 ${reminder.daysUntilDue} 天还款`
  if (reminder.daysUntilDue === 0) return '今天还款'
  return `已逾期 ${-reminder.daysUntilDue} 天`
}

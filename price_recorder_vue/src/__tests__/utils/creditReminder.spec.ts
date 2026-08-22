import { describe, expect, it } from 'vitest'
import type { LedgerAccount, LedgerTransaction } from '@/api/ledger'
import {
  computeCreditReminder,
  creditBillingPeriod,
  creditReminderQueryRange,
  reminderBadgeText,
} from '@/utils/creditReminder'

function makeAccount(overrides: Partial<LedgerAccount> = {}): LedgerAccount {
  return {
    id: '1',
    name: '招行信用卡',
    type: 'liability',
    subtype: 'credit_card',
    creditLimit: '50000',
    billingDay: 5,
    paymentDueDay: 25,
    ...overrides,
  }
}

function expenseOn(
  accountId: string,
  amount: string,
  bookedAt: string,
): LedgerTransaction {
  return {
    id: 'tx',
    type: 'expense',
    bookedAt,
    remark: '',
    postings: [
      { accountId, amount, sort: 0 },
      { accountId: '999', amount: amount.replace('-', ''), categoryId: '10', sort: 1 },
    ],
  }
}

describe('creditBillingPeriod', () => {
  // 账单日 5 号、还款日 25 号
  it('treats the billing day itself as passed (period ends today)', () => {
    const period = creditBillingPeriod(makeAccount(), new Date(2026, 7, 5))

    expect(period).toEqual({
      periodStart: '2026-07-05',
      periodEnd: '2026-08-05',
      dueDate: '2026-08-25',
    })
  })

  it('uses the previous period on the day before the billing day', () => {
    const period = creditBillingPeriod(makeAccount(), new Date(2026, 7, 4))

    expect(period).toEqual({
      periodStart: '2026-06-05',
      periodEnd: '2026-07-05',
      dueDate: '2026-07-25',
    })
  })

  it('uses the current period on the day after the billing day', () => {
    const period = creditBillingPeriod(makeAccount(), new Date(2026, 7, 6))

    expect(period).toEqual({
      periodStart: '2026-07-05',
      periodEnd: '2026-08-05',
      dueDate: '2026-08-25',
    })
  })

  it('crosses the year boundary when today is early January', () => {
    const period = creditBillingPeriod(makeAccount(), new Date(2026, 0, 4))

    expect(period).toEqual({
      periodStart: '2025-11-05',
      periodEnd: '2025-12-05',
      dueDate: '2025-12-25',
    })
  })

  it('rolls the due date to next month when paymentDueDay <= billingDay', () => {
    const period = creditBillingPeriod(
      makeAccount({ billingDay: 25, paymentDueDay: 5 }),
      new Date(2026, 7, 26),
    )

    expect(period).toEqual({
      periodStart: '2026-07-25',
      periodEnd: '2026-08-25',
      dueDate: '2026-09-05',
    })
  })

  it('clamps the billing day to the end of a short month', () => {
    const period = creditBillingPeriod(makeAccount({ billingDay: 31 }), new Date(2026, 2, 10))

    expect(period?.periodStart).toBe('2026-01-31')
    expect(period?.periodEnd).toBe('2026-02-28')
  })

  it('returns null when billingDay or paymentDueDay is missing or invalid', () => {
    expect(creditBillingPeriod(makeAccount({ billingDay: undefined }))).toBeNull()
    expect(creditBillingPeriod(makeAccount({ paymentDueDay: undefined }))).toBeNull()
    expect(creditBillingPeriod(makeAccount({ billingDay: 0 }))).toBeNull()
    expect(creditBillingPeriod(makeAccount({ paymentDueDay: 32 }))).toBeNull()
  })
})

describe('computeCreditReminder', () => {
  // today = 2026-08-22：本期区间 (2026-07-05, 2026-08-05]，还款日 2026-08-25（还有 3 天）
  const today = new Date(2026, 7, 22)

  it('sums expense legs of the account within (periodStart, periodEnd]', () => {
    const transactions = [
      expenseOn('1', '-100', '2026-08-05 10:00:00'), // 区间结束当天：计入
      expenseOn('1', '-50.5', '2026-07-06 09:00:00'), // 区间内：计入
      expenseOn('1', '-200', '2026-07-05 09:00:00'), // 区间开始当天：开区间排除
      expenseOn('1', '-300', '2026-08-06 09:00:00'), // 区间外：排除
      expenseOn('2', '-400', '2026-08-01 09:00:00'), // 其他账户：排除
    ]

    const reminder = computeCreditReminder(makeAccount(), transactions, today)

    expect(reminder).not.toBeNull()
    expect(reminder?.amountDue).toBeCloseTo(150.5)
    expect(reminder?.periodStart).toBe('2026-07-05')
    expect(reminder?.periodEnd).toBe('2026-08-05')
    expect(reminder?.dueDate).toBe('2026-08-25')
  })

  it('ignores income and transfer transactions', () => {
    const transactions: LedgerTransaction[] = [
      {
        id: 't1',
        type: 'income',
        bookedAt: '2026-08-01 09:00:00',
        remark: '',
        postings: [{ accountId: '1', amount: '500', sort: 0 }],
      },
      {
        id: 't2',
        type: 'transfer',
        bookedAt: '2026-08-02 09:00:00',
        remark: '',
        postings: [
          { accountId: '2', amount: '-800', sort: 0 },
          { accountId: '1', amount: '800', sort: 1 },
        ],
      },
    ]

    expect(computeCreditReminder(makeAccount(), transactions, today)).toBeNull()
  })

  it('returns null when there is nothing to repay this period', () => {
    expect(computeCreditReminder(makeAccount(), [], today)).toBeNull()
  })

  it('returns null for accounts missing billing fields', () => {
    const transactions = [expenseOn('1', '-100', '2026-08-01 10:00:00')]

    expect(
      computeCreditReminder(makeAccount({ billingDay: undefined }), transactions, today),
    ).toBeNull()
    expect(
      computeCreditReminder(makeAccount({ paymentDueDay: undefined }), transactions, today),
    ).toBeNull()
  })

  it('marks level warning when the due date is within 7 days', () => {
    const transactions = [expenseOn('1', '-100', '2026-08-01 10:00:00')]

    // 2026-08-22 → 还款日 2026-08-25，还有 3 天
    const reminder = computeCreditReminder(makeAccount(), transactions, today)
    expect(reminder?.daysUntilDue).toBe(3)
    expect(reminder?.level).toBe('warning')

    // 2026-08-18 → 还有 7 天，仍警示
    const at7 = computeCreditReminder(makeAccount(), transactions, new Date(2026, 7, 18))
    expect(at7?.daysUntilDue).toBe(7)
    expect(at7?.level).toBe('warning')
  })

  it('marks level info when the due date is more than 7 days away', () => {
    const transactions = [expenseOn('1', '-100', '2026-08-01 10:00:00')]

    // 2026-08-17 → 还款日 2026-08-25，还有 8 天
    const reminder = computeCreditReminder(makeAccount(), transactions, new Date(2026, 7, 17))
    expect(reminder?.daysUntilDue).toBe(8)
    expect(reminder?.level).toBe('info')
  })

  it('treats the due day itself as day 0 warning, and past due as overdue warning', () => {
    const transactions = [expenseOn('1', '-100', '2026-08-01 10:00:00')]

    const onDue = computeCreditReminder(makeAccount(), transactions, new Date(2026, 7, 25))
    expect(onDue?.daysUntilDue).toBe(0)
    expect(onDue?.level).toBe('warning')

    const overdue = computeCreditReminder(makeAccount(), transactions, new Date(2026, 7, 27))
    expect(overdue?.daysUntilDue).toBe(-2)
    expect(overdue?.level).toBe('warning')
  })
})

describe('creditReminderQueryRange', () => {
  it('maps the billing period to a full-day query range', () => {
    const range = creditReminderQueryRange(makeAccount(), new Date(2026, 7, 22))

    expect(range).toEqual({
      startTime: '2026-07-05 00:00:00',
      endTime: '2026-08-05 23:59:59',
    })
  })

  it('returns null when billing fields are missing', () => {
    expect(creditReminderQueryRange(makeAccount({ billingDay: undefined }))).toBeNull()
  })
})

describe('reminderBadgeText', () => {
  it('renders countdown, due-today and overdue texts', () => {
    expect(reminderBadgeText({ daysUntilDue: 3 })).toBe('还有 3 天还款')
    expect(reminderBadgeText({ daysUntilDue: 0 })).toBe('今天还款')
    expect(reminderBadgeText({ daysUntilDue: -2 })).toBe('已逾期 2 天')
  })
})

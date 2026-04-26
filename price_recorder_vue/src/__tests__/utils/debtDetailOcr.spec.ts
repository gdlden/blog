import { describe, expect, it } from "vitest";
import type { DebtDetail } from "@/api/debtDetail";
import {
  buildOcrPreviewRows,
  getRowsToCreate,
  getRowsToOverwrite,
} from "@/utils/debtDetailOcr";

describe("debtDetailOcr", () => {
  it("marks period conflicts as skip by default", () => {
    const items: Omit<DebtDetail, "id">[] = [
      {
        debtId: "debt-1",
        postingDate: "2026-01-10",
        principal: "1000",
        interest: "10",
        period: "2026-01",
      },
      {
        debtId: "debt-1",
        postingDate: "2026-02-10",
        principal: "900",
        interest: "9",
        period: "2026-02",
      },
    ];
    const existingDetails: DebtDetail[] = [
      {
        id: "detail-1",
        debtId: "debt-1",
        postingDate: "2026-01-12",
        principal: "1000",
        interest: "11",
        period: "2026-01",
      },
    ];

    expect(buildOcrPreviewRows(items, existingDetails)).toEqual([
      {
        id: "",
        debtId: "debt-1",
        postingDate: "2026-01-10",
        principal: "1000",
        interest: "10",
        period: "2026-01",
        rowId: "2026-01-0",
        isConflict: true,
        conflictDetailId: "detail-1",
        action: "skip",
      },
      {
        id: "",
        debtId: "debt-1",
        postingDate: "2026-02-10",
        principal: "900",
        interest: "9",
        period: "2026-02",
        rowId: "2026-02-1",
        isConflict: false,
        conflictDetailId: "",
        action: "create",
      },
    ]);
  });

  it("splits create and overwrite rows", () => {
    const rows = buildOcrPreviewRows(
      [
        {
          debtId: "debt-1",
          postingDate: "2026-01-10",
          principal: "1000",
          interest: "10",
          period: "2026-01",
        },
        {
          debtId: "debt-1",
          postingDate: "2026-02-10",
          principal: "900",
          interest: "9",
          period: "2026-02",
        },
        {
          debtId: "debt-1",
          postingDate: "2026-03-10",
          principal: "800",
          interest: "8",
          period: "2026-03",
        },
      ],
      [
        {
          id: "detail-1",
          debtId: "debt-1",
          postingDate: "2026-01-12",
          principal: "1000",
          interest: "11",
          period: "2026-01",
        },
        {
          id: "detail-3",
          debtId: "debt-1",
          postingDate: "2026-03-12",
          principal: "800",
          interest: "9",
          period: "2026-03",
        },
      ],
    );
    rows[0].action = "overwrite";
    rows[2].action = "skip";

    expect(getRowsToCreate(rows)).toEqual([
      {
        debtId: "debt-1",
        postingDate: "2026-02-10",
        principal: "900",
        interest: "9",
        period: "2026-02",
      },
    ]);
    expect(getRowsToOverwrite(rows)).toEqual([
      {
        id: "detail-1",
        debtId: "debt-1",
        postingDate: "2026-01-10",
        principal: "1000",
        interest: "10",
        period: "2026-01",
      },
    ]);
  });
});

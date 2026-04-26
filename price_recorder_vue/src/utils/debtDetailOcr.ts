import type { DebtDetail } from "@/api/debtDetail";

export type OcrConflictAction = "create" | "overwrite" | "skip";

export interface DebtDetailOcrPreviewRow extends DebtDetail {
  rowId: string;
  isConflict: boolean;
  conflictDetailId: string;
  action: OcrConflictAction;
}

export function buildOcrPreviewRows(
  items: Omit<DebtDetail, "id">[],
  existingDetails: DebtDetail[],
): DebtDetailOcrPreviewRow[] {
  return items.map((item, index) => {
    const conflict = existingDetails.find((detail) => detail.period === item.period);
    return {
      ...item,
      id: "",
      rowId: `${item.period || "row"}-${index}`,
      isConflict: Boolean(conflict),
      conflictDetailId: conflict?.id || "",
      action: conflict ? "skip" : "create",
    };
  });
}

export function getRowsToCreate(rows: DebtDetailOcrPreviewRow[]): Omit<DebtDetail, "id">[] {
  return rows
    .filter((row) => !row.isConflict && row.action === "create")
    .map(({ id, rowId, isConflict, conflictDetailId, action, ...detail }) => detail);
}

export function getRowsToOverwrite(rows: DebtDetailOcrPreviewRow[]): DebtDetail[] {
  return rows
    .filter((row) => row.isConflict && row.action === "overwrite" && row.conflictDetailId)
    .map(({ rowId, isConflict, conflictDetailId, action, ...detail }) => ({
      ...detail,
      id: conflictDetailId,
    }));
}

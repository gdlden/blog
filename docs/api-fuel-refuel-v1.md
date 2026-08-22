# 加油记录 API — `/fuel/refuel/*`（实付金额 + 附件 OCR 版）

## 概述

加油记录模块近期三项变更，供 App 端对接：

1. **新增实付金额字段** `actualAmount`：应付金额 `amount` 保留（= 油量×单价），实付金额为优惠后实际支付额；统计口径已全部切换为实付。
2. **附件类型体系重建**：`receipt/environment/other` 废弃，新枚举 `station_screen/dashboard/other`，并配套 OCR 识别接口（后端走 Kimi 视觉模型）。
3. **附件数量规则**：`station_screen`、`dashboard` 每条记录各限 1 张，`other` 不限，总数 ≤6。

---

## 数据模型 — RefuelRecord

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | int64 | 响应 | 记录 id（JSON 数字，注意前端如调用 string 入参接口需转字符串） |
| `vehicleId` | int64 | 是 | 车辆 id |
| `refuelTime` | string | 是 | 加油时间，格式 `YYYY-MM-DD HH:mm:ss` |
| `odometer` | string(decimal) | 否 | 总里程 km，不能为负 |
| `volume` | string(decimal) | 是 | 油量 L，必须 > 0 |
| `unitPrice` | string(decimal) | 否 | 单价，不能为负 |
| `amount` | string(decimal) | 见校验规则 | 应付金额 = 油量×单价 |
| **`actualAmount`** | string(decimal) | **是** | **实付金额（新增）。必须非空且 > 0；允许大于 `amount`（如加燃油宝合并结账）；空串报"实付金额不能为空"，0 报"实付金额不能为空"，负数报"实付金额不能为负数"** |
| `station` | string | 否 | 加油站 |
| `isFull` | bool | 否 | 是否加满（影响统计口径） |
| `remark` | string | 否 | 备注 |
| `intervalConsumption` | string(decimal) | 响应 | 区间油耗（后端计算） |
| `attachments` | FuelAttachment[] | 请求 | 附件引用列表，见下文 |
| `attachmentInfos` | FuelAttachmentInfo[] | 响应 | 附件详情（含 url） |

### amount 校验规则（create 与 update 不一致，注意）

- **新建（`/fuel/refuel/save/v1`）**：`amount` 留空时后端按 `油量×单价` 自动计算；手输时必须与 `油量×单价`（保留 2 位）一致，否则报"金额与油量×单价不一致"。
- **编辑（`/fuel/refuel/update/v1`）**：`amount` 一律按 `油量×单价` 重算，忽略入参。
- **`actualAmount` 两种场景都必填**，不会自动计算——客户端应默认取 `amount` 值让用户改。

---

## 附件体系

### FuelAttachment（请求）

| 字段 | 类型 | 说明 |
|---|---|---|
| `fileId` | string | 文件记录 id，先经 `POST /file/upload/raw/v1` 上传获得 |
| `attachType` | string | **新枚举：`station_screen`（加油站屏幕）/ `dashboard`（车辆仪表）/ `other`（其他）** |
| `sort` | int32 | 顺序（后端按提交顺序回填，客户端可不传） |

### 数量与格式限制（create/update 均校验，整组替换语义）

| 规则 | 错误信息 |
|---|---|
| 总数 ≤ 6 张 | `附件数量不能超过 6 张` |
| `attachType` 必须是新枚举之一 | `附件类型不合法: <type>` |
| 文件必须是图片（content-type `image/*` 或常见图片扩展名） | `附件必须是图片: <文件名>` |
| 单张 ≤ 10MB | `附件大小不能超过 10MB: <文件名>` |
| **`station_screen` 最多 1 张** | `加油站屏幕照片只能上传 1 张` |
| **`dashboard` 最多 1 张** | `车辆仪表照片只能上传 1 张` |

### 编辑语义

- 提交 `attachments` 字段 = **整组替换**；不传该字段（null）= 保持现有附件不变。
- 删除记录/车辆的相关约束不变（车辆下有记录时禁止删除）。

---

## OCR 识别接口 — `/fuel/refuel/ocr/v1`

识别加油站屏幕照片（金额/油量/单价）或车辆仪表照片（总里程），用于表单自动回填。**识别与附件存档解耦**：识别用本接口直传原图，附件仍走 `/file/upload/raw/v1` 上传后关联 `fileId`。

### 请求

**方法** `POST`，**路径** `/fuel/refuel/ocr/v1`，**Content-Type** `multipart/form-data`

**请求头**：`Authorization: Bearer <token>`（走 JWT，不在白名单）

| 表单字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `file` | 文件 | 是 | 图片，≤ 8MB，content-type 必须是 `image/*` |
| `attachType` | string | 是 | 仅接受 `station_screen` 或 `dashboard`；其他值（含 `other`）返回 400 |

### 成功响应

**HTTP 200**

```json
{
  "rawText": "金额: 225.00\n油量: 30.50\n单价: 7.38",
  "amount": "225.00",
  "volume": "30.50",
  "unitPrice": "7.38",
  "odometer": ""
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `rawText` | string | 模型原始输出文本（调试用） |
| `amount` | string | 仅 `station_screen` 可能有值：金额 |
| `volume` | string | 仅 `station_screen` 可能有值：油量 |
| `unitPrice` | string | 仅 `station_screen` 可能有值：单价 |
| `odometer` | string | 仅 `dashboard` 可能有值：总里程 |

**容错约定**：识别不到的字段为空字符串；模型输出完全无法解析时返回 200 + 全空字段 + `rawText`，**不报错**。客户端应把"相关字段全空"视为识别失败。

### 错误响应

| 场景 | HTTP | 说明 |
|---|---|---|
| `attachType` 非 `station_screen`/`dashboard` | 400 | reason `FUEL_OCR_BAD_REQUEST` |
| 缺少图片 / 空图片 / 超 8MB / 非图片 | 400 | reason 可能为 `DEBT_DETAIL_OCR_BAD_REQUEST`（复用了通用图片校验，客户端只需判断 400） |
| 模型调用失败（网络/key/模型不可用） | 500 | `fuel ocr failed: ...`，客户端提示"识别失败，请手动填写"即可 |

### 后端实现参考（排查用）

- 模型：Moonshot Kimi，默认 `kimi-k2.6`（环境变量 `OCR_MODEL` 可覆盖）；base URL `https://api.moonshot.cn/v1`。
- API key 环境变量读取顺序：`OCR_API_KEY` → `KIMI_API_KEY` → `MOONSHOT_API_KEY`；未配置时 OCR 接口 500，其余功能不受影响。
- 识别超时 60s；该模型带推理过程，单次调用可能需数秒，**客户端应给足超时并显示"识别中"状态**。
- 注意：债务模块的 OCR（`/debtDetail/ocr/v1`）也走同一 Kimi 配置。

---

## 统计接口口径变更 — `/fuel/stats/v1`

请求/响应结构不变（`GET /fuel/stats/v1?vehicleId=&startTime=&endTime=`），但金额字段语义已切换：

| 字段 | 新口径 |
|---|---|
| `totalAmount` | **实付金额合计**（满油区间内记录的 `actualAmount` 累加，不再用 `amount`） |
| `costPerKm` | **实付口径**每公里成本 = 区间实付合计 ÷ 区间里程 |

其余字段（`totalDistance`、`totalVolume`、`averageConsumption`、`latestConsumption`、`trend`）逻辑不变；trend 点不含金额字段。

---

## 客户端消费逻辑建议（与 Web 端一致）

1. **新建/编辑表单**：
   - 新增"实付金额"输入框，默认跟随 `amount`（用户手改后解除联动；编辑态不联动）。
   - 提交时 `actualAmount` 若为空，兜底取 `amount` 的最终值（后端必填，留空会 400）。
2. **附件上传**：按类型分三个上传入口（加油站屏幕/车辆仪表/其他），入口选择即定死类型；某类型已达上限时隐藏对应入口；传错只能删除重传。
3. **OCR 回填**：上传 `station_screen`/`dashboard` 图片后自动调 OCR 接口；识别期间显示 loading；**只回填表单中的空字段，不覆盖用户已填值**；失败 toast 提示"识别失败，请手动填写"，不阻塞、不影响附件本身。
4. **列表展示**：金额列显示实付金额；当 `actualAmount ≠ amount`（数值比较）时附灰色小字"应付 ¥xxx"对照。

---

## 数据迁移（部署后手动执行）

```sql
-- 1. 历史记录实付金额回填为应付金额
UPDATE refuel_records SET actual_amount = amount WHERE actual_amount = 0 OR actual_amount IS NULL;

-- 2. 旧附件类型归并到 other
UPDATE fuel_attachments SET attach_type = 'other' WHERE attach_type IN ('receipt', 'environment');
```

## 版本兼容性注意

- **旧版客户端不兼容**：新建/编辑记录不传 `actualAmount` 会被 400 拒绝，App 必须与后端同步发版，或后端先行时旧版 App 的加油功能不可用。
- 旧数据的 `attachType` 迁移前为 `receipt`/`environment`，客户端渲染未知类型时按"其他"兜底展示，不要崩溃。

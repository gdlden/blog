# 记账模块 API 对接文档 `/ledger/*`（P1：账户 / 分类 / 复式交易 / 月度统计；P2：分类月度预算 / 余额走势；P3：周期账单）

## 概述

记账模块为全新模块，采用**真复式记账**模型：每笔交易由至少两条腿（posting）组成，腿金额带符号且合计恒为 0。App 端对接要点：

1. **账户**：资产/负债两大类，信用卡字段条件生效；初始余额自动生成期初交易。
2. **分类**：两级结构，支出/收入两个方向，首次拉取自动初始化预置分类。
3. **交易（核心）**：`expense` / `income` / `transfer` 三种用户交易，全部通过 postings 数组表达；`accountId=0` 是系统账户约定值。
4. **月度统计**：只统计收支，转账不计入（信用卡还款是转账，避免重复计支出）。
5. **分类月度预算（P2）**：按 `categoryId + month` 设置支出预算，列表返回含 `used`（含子孙分类聚合的当月已用）。
6. **余额走势（P2）**：单账户或净资产口径的按日收盘余额序列，缺数据日期自动补点。
7. **周期账单（P3）**：月度规则（每月几号 + 金额 + 分类），靠 `apply` 接口**惰性、幂等**生成到期交易；无后台定时任务。

**通用约定**：

- 所有接口走 JWT，请求头携带 `Authorization: Bearer <token>`。
- 金额一律为 string 十进制（如 `"200.00"`），带符号时表示方向（负 = 流出，正 = 流入）。
- 各类 `id` 为 JSON 数字（int64）。
- 时间格式统一为 `YYYY-MM-DD HH:mm:ss`，月份为 `YYYY-MM`。

---

## 数据模型

### 账户 Account

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | int64 | 响应 | 账户 id |
| `name` | string | 是 | 账户名称 |
| `type` | string | 是 | `asset`（资产）/ `liability`（负债）。系统类型 `expense` / `income` / `equity` 用户不可创建 |
| `subtype` | string | 是 | 见下方 subtype 枚举表 |
| `creditLimit` | string(decimal) | 条件 | 信用额度，**仅 `subtype=credit_card` 使用** |
| `billingDay` | int32 | 条件 | 账单日，仅 `subtype=credit_card` 使用 |
| `paymentDueDay` | int32 | 条件 | 还款日，仅 `subtype=credit_card` 使用 |
| `remark` | string | 否 | 备注 |
| `openingBalance` | string(decimal) | 仅 save | 期初余额，见下方"期初余额语义" |
| `sort` | int32 | 响应 | 排序值 |
| `archived` | bool | 响应 | 是否已归档 |
| `balance` | string(decimal) | 响应 | 实时余额（后端聚合全部交易得出） |
| `isSystem` | bool | 响应 | 是否系统账户 |

subtype 枚举（按 type 限定）：

| type | 允许的 subtype |
|---|---|
| `asset` | `cash`（现金）/ `debit_card`（储蓄卡）/ `e_wallet`（电子钱包）/ `prepaid_card`（预付卡）/ `investment`（投资）/ `other` |
| `liability` | `credit_card`（信用卡）/ `huabei_like`（花呗类）/ `loan_payable`（应付借款）/ `loan_receivable`（应收借款）/ `other` |

**期初余额语义**：`openingBalance` 非空且非 0 时，后端自动生成一笔 `type=opening_balance` 交易，两条腿：新账户 +X、期初调整系统账户（equity 类）-X，日期为创建当天。客户端不要手动补这笔交易。

### 分类 Category

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | int64 | 响应 | 分类 id |
| `parentId` | int64 | 否 | 父分类 id；空/0 表示一级分类。最多两级 |
| `name` | string | 是 | 分类名称 |
| `direction` | string | 是 | `expense`（支出）/ `income`（收入） |
| `sort` | int32 | 响应 | 排序值 |
| `isSystem` | bool | 响应 | 是否系统预置（预置分类不可删） |

**初始化**：首次调用 `GET /ledger/category/list/v1` 时后端自动初始化预置分类（支出 12 个、收入 5 个，两个方向各含一个"其它"）。"其它"分类是删除兜底目标，`isSystem=true`，不可删除。

### 交易 Transaction 与腿 Posting

Transaction：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | int64 | 响应 | 交易 id |
| `type` | string | 是 | `expense` / `income` / `transfer` / `opening_balance` / `adjustment`。**用户只能提交前三种**，后两种由系统生成 |
| `bookedAt` | string | 是 | 记账时间，格式 `YYYY-MM-DD HH:mm:ss` |
| `remark` | string | 否 | 备注 |
| `postings` | Posting[] | 是 | 腿数组，规则见下文 |

Posting（腿）：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `accountId` | int64 | 是 | 账户 id。**`0` 是约定值**：`expense` 时解析为系统费用账户，`income` 时解析为系统收入账户；`transfer` 不允许为 0 |
| `amount` | string(decimal) | 是 | 带符号金额，不能为 0 |
| `categoryId` | int64 | 条件 | 分类 id；仅挂账户侧（真实账户腿）不挂分类，分类挂在 `accountId=0` 的系统腿上 |
| `sort` | int32 | 是 | 腿顺序 |

---

## 接口契约

### 1. 账户接口

#### 新建账户 `POST /ledger/account/save/v1`

请求：

```json
{
  "name": "招行信用卡",
  "type": "liability",
  "subtype": "credit_card",
  "creditLimit": "30000.00",
  "billingDay": 5,
  "paymentDueDay": 25,
  "remark": "",
  "openingBalance": "-1200.00"
}
```

响应：

```json
{
  "id": 12,
  "message": "ok"
}
```

注意：`creditLimit` / `billingDay` / `paymentDueDay` 仅在 `subtype=credit_card` 时有效；`openingBalance` 非空非 0 会触发期初交易自动生成（见数据模型章节）。

#### 更新账户 `POST /ledger/account/update/v1`

请求体携带 `id` 及需更新的字段。更新不涉及 `openingBalance` 重算（期初交易只在创建时生成一次）。

#### 删除账户 `POST /ledger/account/delete/v1`

```json
{ "id": 12 }
```

响应：`{ "flag": true }`

**约束：有交易的账户禁止删除，只能归档**。客户端对 `archived` 账户应在记账表单中隐藏，但历史流水仍可查询。

#### 账户列表 `GET /ledger/account/list/v1`

响应：

```json
{
  "list": [
    {
      "id": 3,
      "name": "微信钱包",
      "type": "asset",
      "subtype": "e_wallet",
      "creditLimit": "",
      "billingDay": 0,
      "paymentDueDay": 0,
      "remark": "",
      "sort": 1,
      "archived": false,
      "balance": "1532.50",
      "isSystem": false
    }
  ]
}
```

**list 不含系统账户**（费用/收入/期初调整账户不会出现在列表里），`balance` 为实时聚合值。

---

### 2. 分类接口

#### 新建分类 `POST /ledger/category/save/v1`

```json
{
  "parentId": 0,
  "name": "咖啡",
  "direction": "expense"
}
```

#### 更新分类 `POST /ledger/category/update/v1`

请求体携带 `id` 及需更新的字段。

#### 删除分类 `POST /ledger/category/delete/v1`

```json
{ "id": 21 }
```

响应：`{ "flag": true }`

**约束**：

- 仅可删除叶子分类（有子分类的一级分类不能删）。
- 删除时，引用该分类的账单腿自动迁移到同 `direction` 的"其它"分类。客户端删除前无需手动迁移历史数据。
- `isSystem=true` 的分类（含"其它"）不可删除。

#### 分类列表 `GET /ledger/category/list/v1`

响应：

```json
{
  "list": [
    {
      "id": 1,
      "parentId": 0,
      "name": "餐饮",
      "direction": "expense",
      "sort": 1,
      "isSystem": true
    },
    {
      "id": 15,
      "parentId": 1,
      "name": "早餐",
      "direction": "expense",
      "sort": 1,
      "isSystem": false
    }
  ]
}
```

首次调用自动初始化预置分类。客户端按 `parentId` 组装两级树：`parentId` 为空/0 的是一级分类。

---

### 3. 交易接口（复式核心）

#### postings 规则（重点，这是复式的心脏）

- 一笔交易至少 2 条腿，所有腿 `amount` 带符号，**Σ amount = 0**，每条腿 ≠ 0，全部腿在同一 DB 事务写入。
- `accountId=0` 是约定值：`expense` 交易解析为系统费用账户，`income` 交易解析为系统收入账户；`transfer` 交易不允许出现 `accountId=0`。
- `categoryId` 挂在 `accountId=0` 的系统腿上，且其 `direction` 必须与交易 `type` 匹配；`transfer` 不挂分类。

**支出示例（含多腿拆分）**：微信付款 200 元，其中餐饮 150、日用品 50：

```json
{
  "type": "expense",
  "bookedAt": "2026-08-22 12:30:00",
  "remark": "超市采购",
  "postings": [
    { "accountId": 3, "amount": "-200.00", "sort": 1 },
    { "accountId": 0, "amount": "150.00", "categoryId": 1, "sort": 2 },
    { "accountId": 0, "amount": "50.00", "categoryId": 4, "sort": 3 }
  ]
}
```

结构：付款账户腿为负（流出），一条或多条系统费用腿为正并各挂分类，拆分腿金额合计等于付款金额。

**收入示例（与支出对称）**：工资到账 8000 元进储蓄卡：

```json
{
  "type": "income",
  "bookedAt": "2026-08-15 09:00:00",
  "remark": "8 月工资",
  "postings": [
    { "accountId": 2, "amount": "8000.00", "sort": 1 },
    { "accountId": 0, "amount": "-8000.00", "categoryId": 20, "sort": 2 }
  ]
}
```

**转账示例**：储蓄卡转 500 元到微信钱包：

```json
{
  "type": "transfer",
  "bookedAt": "2026-08-20 18:00:00",
  "remark": "",
  "postings": [
    { "accountId": 2, "amount": "-500.00", "sort": 1 },
    { "accountId": 3, "amount": "500.00", "sort": 2 }
  ]
}
```

转出账户为负、转入账户为正，两腿均不挂分类。信用卡还款同样是 `transfer`（资产账户 → 信用卡负债账户）。

#### 新建交易 `POST /ledger/transaction/save/v1`

请求体如上三例。响应：`{ "id": 88, "message": "ok" }`

#### 更新交易 `POST /ledger/transaction/update/v1`

请求体携带 `id`，**`postings` 为整组替换语义**：提交的新腿数组整体覆盖旧腿。客户端编辑时应先经 get 接口取回完整交易再整体改，不要做增量提交。

#### 删除交易 `POST /ledger/transaction/delete/v1`

```json
{ "id": 88 }
```

响应：`{ "flag": true }`

#### 分页查询 `GET /ledger/transaction/page/v1`

Query 参数：`page`、`pageSize`、`accountId`、`categoryId`、`type`、`startTime`、`endTime`（均可选/按筛选需要传）。

响应：

```json
{
  "page": 1,
  "total": 132,
  "list": [
    {
      "id": 88,
      "type": "expense",
      "bookedAt": "2026-08-22 12:30:00",
      "remark": "超市采购",
      "postings": [
        { "accountId": 3, "amount": "-200.00", "sort": 1 },
        { "accountId": 0, "amount": "150.00", "categoryId": 1, "sort": 2 },
        { "accountId": 0, "amount": "50.00", "categoryId": 4, "sort": 3 }
      ]
    }
  ]
}
```

#### 单条查询 `GET /ledger/transaction/get/v1?id=88`

返回单条完整交易（含全部 postings），用于编辑表单回填。

---

### 4. 月度统计接口

#### `GET /ledger/stats/monthly/v1?month=2026-08`

响应：

```json
{
  "month": "2026-08",
  "totalExpense": "3450.00",
  "totalIncome": "8200.00",
  "expenseByCategory": [
    { "categoryId": 1, "categoryName": "餐饮", "amount": "1200.00" }
  ],
  "incomeByCategory": [
    { "categoryId": 20, "categoryName": "工资", "amount": "8000.00" }
  ]
}
```

**统计口径（务必注意）**：

- **只统计 `expense` / `income` 类型的交易；`transfer` / `opening_balance` / `adjustment` 一律排除。**
- **信用卡还款是 `transfer`，不会计入支出。这条设计的目的是防止重复计支出：刷卡消费时已按 expense 记过一次，还款只是资产和负债账户间的移动。客户端不要把转账类流水自行加进收支报表。**
- 支出按费用腿（`accountId=0` 解析后的系统费用账户腿）的 `categoryId` 分组；收入同理。
- 未挂分类的腿归入"其它"分类。

---

### 5. 预算接口（P2）

预算按「支出方向分类 + 月份」设置，同一分类同一月份只有一条（save 为 upsert 语义）。仅允许 `direction=expense` 的分类。

#### 保存预算 `POST /ledger/budget/save/v1`

```json
{
  "categoryId": 1,
  "month": "2026-08",
  "amount": "1500.00"
}
```

响应：`{ "id": 5, "message": "ok" }`

- `categoryId` + `month` 已存在时更新金额，否则新建。
- `amount` 必须 > 0。
- 对一级分类设预算即可覆盖其全部子分类的支出（见下方 `used` 口径），无需逐个子分类设置。

#### 删除预算 `POST /ledger/budget/delete/v1`

```json
{ "id": 5 }
```

响应：`{ "flag": true }`

#### 预算列表 `GET /ledger/budget/list/v1?month=2026-08`

`month` 为空默认当月。响应：

```json
{
  "list": [
    {
      "id": 5,
      "categoryId": 1,
      "categoryName": "餐饮",
      "amount": "1500.00",
      "used": "1200.00"
    }
  ]
}
```

**`used` 口径**：该分类**及其所有子孙分类**当月费用腿（系统腿）合计，正数。客户端进度条直接算 `used / amount`，不要在本地重新聚合。

---

### 6. 余额走势接口（P2）

#### `GET /ledger/stats/balance-trend/v1?accountId=&startTime=&endTime=`

Query 参数：

| 参数 | 说明 |
|---|---|
| `accountId` | 为空/0 表示**净资产走势**（全部非系统账户合计）；传账户 id 则为该账户余额走势 |
| `startTime` | `YYYY-MM-DD`，缺省 = 近 6 个月 |
| `endTime` | `YYYY-MM-DD`，缺省 = 今天 |

响应：

```json
{
  "points": [
    { "date": "2026-03-01", "balance": "1532.50" },
    { "date": "2026-03-02", "balance": "1532.50" }
  ]
}
```

**口径**：每个自然日一个点（无交易的日期也补点，余额沿用前一收盘值），`balance` 为当日收盘余额。客户端直接按点渲染折线即可，无需补全日期。

---

### 7. 周期账单接口（P3）

周期账单是「每月固定日自动生成一笔 expense/income 交易」的规则。生成是**惰性**的：服务端不跑定时任务，客户端在合适时机（如进入记账页时）调用一次 apply，服务端把所有到期月份补齐。

#### 保存周期账单 `POST /ledger/recurring/save/v1`

```json
{
  "id": 0,
  "accountId": 1,
  "categoryId": 3,
  "type": "expense",
  "amount": "3500.00",
  "remark": "房租",
  "dayOfMonth": 1,
  "startMonth": "2026-08",
  "enabled": true
}
```

响应：`{ "id": 5, "message": "ok" }`

- `id=0` 新建，`id>0` 更新（更新不回退生成进度 `lastGeneratedMonth`）。
- `type` 仅允许 `expense` / `income`（transfer 拒绝）；`amount` 必须为正数十进制字符串。
- `dayOfMonth` 1-31，**短月钳制**：如 30 日在 2 月生成于 28/29 日。
- `categoryId` 必须存在且方向与 `type` 匹配。

#### 删除周期账单 `POST /ledger/recurring/delete/v1`

```json
{ "id": 5 }
```

响应：`{ "flag": true }`

#### 周期账单列表 `GET /ledger/recurring/list/v1`

响应：

```json
{
  "list": [
    {
      "id": 5,
      "accountId": 1,
      "accountName": "招行储蓄卡",
      "categoryId": 3,
      "categoryName": "居住",
      "type": "expense",
      "amount": "3500.00",
      "remark": "房租",
      "dayOfMonth": 1,
      "startMonth": "2026-08",
      "lastGeneratedMonth": "2026-08",
      "enabled": true,
      "nextDate": "2026-09-01"
    }
  ]
}
```

- `nextDate` 为下一次应生成日期（`YYYY-MM-DD`），`enabled=false` 时为空串。

#### 应用周期账单 `POST /ledger/recurring/apply/v1`

请求体为空 `{}`。响应：`{ "created": 3 }`

**apply 语义（幂等）**：遍历全部启用规则，从 `max(startMonth, lastGeneratedMonth 的下一月)` 起逐月检查，应生成日期 ≤ 今天才生成交易并推进 `lastGeneratedMonth`，应生成日期 > 今天即停止该规则。重复调用不会重复生成（第二次返回 `created=0`）。生成的交易走与手工记账相同的复式写入路径：expense → 账户腿（负）+ 费用系统腿（正，挂分类）；income 对称；`bookedAt` = 应生成日，`remark` 沿用规则备注。生成后的交易与普通交易无异，可正常编辑/删除。

---

## 客户端消费逻辑建议

### 1. 记账三 Tab

- **支出 Tab**：选择付款账户 + 金额 + 分类，默认构造两腿（账户负、系统费用腿正）。开启"多腿拆分"后，允许添加多条分类腿，每条各填金额与分类；**提交前校验拆分腿合计 = 付款金额**，不一致时禁止提交并提示差额。方向符号由客户端按模板生成，不要让用户手填正负号。
- **收入 Tab**：与支出对称，收款账户腿为正，系统收入腿为负挂分类。
- **转账 Tab**：选择转出/转入两个账户 + 金额，两腿不带分类；转出与转入不允许选同一账户。
- 三个 Tab 提交后建议刷新账户列表（`balance` 实时变化）。

### 2. 账户页

- 表单按 `subtype` 条件显示字段：仅当 `subtype=credit_card` 时显示信用额度、账单日、还款日三个字段。
- 新建表单提供"初始余额"输入，非 0 时提示用户"将自动生成一笔期初交易"；编辑态不再出现该字段。
- 删除按钮对有交易的账户禁用，改为提供"归档"入口；`archived=true` 的账户灰显或折叠到列表底部。

### 3. 分类页

- 两级展示，一级可展开子分类；新建分类时选择归属方向与父级。
- 删除前提示"引用该分类的账单将自动归入'其它'"；`isSystem=true` 的分类不显示删除按钮。

### 4. 报表页

- 月份切换器传 `month=YYYY-MM` 调用月度统计接口；展示总支出、总收入与两个方向的分类排行（按 `amount` 降序）。
- 排行数据直接渲染 `expenseByCategory` / `incomeByCategory`，不要在本地按流水重新聚合（口径以后端为准，转账已排除）。
- **预算区块（P2）**：随月份切换调用预算列表，渲染 `amount` / `used` 进度条，超支（`used > amount`）用警示色突出。
- **余额走势（P2）**：账户切换器传 `accountId`（含「净资产」选项即不传）；日期范围默认近 6 个月。折线直接渲染 `points`。

### 5. 周期账单页（P3）

- **进入记账/交易列表页时静默调用一次 `POST /ledger/recurring/apply/v1`**，不阻塞首屏、失败不弹错（下次进页会自然重试）；`created > 0` 时刷新交易列表与月度统计。
- 列表展示 `amount` / `accountName` / `categoryName` / `dayOfMonth` / `nextDate` 与启停开关；`nextDate ≤ 今天` 表示有到期未生成账单（通常一次 apply 后即消失）。
- 表单按 `type` 过滤分类（与记账弹窗一致）；`dayOfMonth` 限制 1-31，提示"短月会落到当月最后一天"；`startMonth` 用月份选择器。
- 删除前提示"仅删除规则，已生成的历史交易不受影响"。

### 6. 编辑与删除

- **交易编辑是整组替换语义**：必须先调 `get` 取完整 postings，修改后整组提交 `update`。只提交改动的腿会丢掉其余腿。
- 账户/分类的删除约束（有交易禁删、仅删叶子）由后端兜底报错，客户端按错误提示展示即可。

---

## 版本兼容与注意

- 本模块为全新模块，无历史数据包袱，App 端与后端同步发版即可。
- **开账日期固定为账户创建当天，不可选择过去日期**；`openingBalance` 只在创建时生效一次，后续想调整期初余额请走普通记账。
- **转账不进收支报表**：所有涉及资金在自有账户间移动的场景（信用卡还款、提现、账户互转）都应记为 `transfer`，不要记成支出 + 收入两笔，否则收支统计会虚高。
- 所有接口均走 JWT 鉴权，请求头统一携带 `Authorization: Bearer <token>`。

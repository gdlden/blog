# 函数索引

## price 模块

### biz 层
| 函数 | 文件 | 说明 |
|------|------|------|
| `NewPriceUsecase` | `internal/biz/price.go` | 创建 PriceUscase |
| `CreatePrice` | `internal/biz/price.go` | 创建价格记录 |
| `ListAll` | `internal/biz/price.go` | 查询所有价格记录 |
| `GetPrice` | `internal/biz/price.go` | 按 ID 查询价格记录 |
| `UpdatePrice` | `internal/biz/price.go` | 更新价格记录 |
| `DeletePrice` | `internal/biz/price.go` | 删除价格记录 |

### data 层
| 函数 | 文件 | 说明 |
|------|------|------|
| `NewPriceRepo` | `internal/data/price.go` | 创建 priceRepo |
| `Save` | `internal/data/price.go` | 保存价格记录 |
| `Update` | `internal/data/price.go` | 更新价格记录 |
| `FindByID` | `internal/data/price.go` | 按 ID 查询价格记录 |
| `Delete` | `internal/data/price.go` | 删除价格记录 |
| `ListByHello` | `internal/data/price.go` | (预留) 按条件查询 |
| `ListAll` | `internal/data/price.go` | 查询所有价格记录 |

### service 层
| 函数 | 文件 | 说明 |
|------|------|------|
| `NewPriceService` | `internal/service/price.go` | 创建 PriceService |
| `CreatePrice` | `internal/service/price.go` | POST /price/add |
| `UpdatePrice` | `internal/service/price.go` | PUT /price/{id} |
| `DeletePrice` | `internal/service/price.go` | DELETE /price/{id} |
| `GetPrice` | `internal/service/price.go` | GET /price/{id} |
| `ListPrice` | `internal/service/price.go` | GET /price/list |

# 版本检查 API — `/app/version/v1`

## 概述

查询 App 最新版本信息，用于客户端版本更新提示和 OTA 下载。

---

## 请求

**方法** `GET`

**路径** `/app/version/v1`

**请求头**

| 字段 | 值 | 说明 |
|---|---|---|
| `Authorization` | `Bearer <token>` | 登录 token，与 App 其他接口一致 |

**请求参数** 无

---

## 响应

### 成功响应

**HTTP 状态码** `200`

```json
{
  "code": "0",
  "message": "success",
  "data": {
    "version": "1.0.0",
    "info": [
      "修复bug提升性能",
      "增加彩蛋有趣的功能页面",
      "测试功能"
    ],
    "iosUrl": "itms-apps://itunes.apple.com/cn/app/idxxxxxxxxx?mt=8",
    "androidUrl": "https://download.example.com/app/release-v1.0.0.apk"
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `code` | string | 是 | `"0"` 表示成功，其他值为失败 |
| `message` | string | 是 | 提示信息 |
| `data.version` | string | 是 | 最新版本号，格式 `x.y.z`（三段式，每段数字） |
| `data.info` | string[] | 是 | 更新内容列表，每项一条更新说明 |
| `data.iosUrl` | string | 否 | iOS App Store 地址，为空时 Android 端不影响；iOS 端无此字段则打开 App Store 首页 |
| `data.androidUrl` | string | 否 | Android APK 下载地址，为空时 iOS 端不影响；Android 端无此字段则不显示升级按钮 |

### 无新版本 / 不需要更新

直接返回当前线上版本信息即可，客户端内部通过 `compareVersion()` 比对版本号。

### 错误响应

```json
{
  "code": "1001",
  "message": "系统错误",
  "data": null
}
```

| code | 含义 |
|---|---|
| `0` | 成功 |
| `1001` | 系统内部错误 |
| `1002` | 参数错误 |

---

## 客户端消费逻辑

调用方：`lib/services/common_service.dart` — `getNewVersion()`

```dart
/// 获取APP最新版本号
Future<NewVersionData> getNewVersion() async {
  final raw = await Request.get<Map<String, dynamic>>('/app/version/v1');
  final apiResp = NewVersionRes.fromJson(raw);
  if (apiResp.code == '0' && apiResp.data != null) {
    return apiResp.data!;
  }
  throw Exception(apiResp.message ?? '版本检查失败');
}
```

客户端行为：

1. 调用 `/app/version/v1` 获取最新版本信息
2. 若网络请求失败或 `code != "0"`，**静默跳过**，不弹出任何提示
3. 若成功，用 `compareVersion(serverVersion, localVersion)` 比对：
   - 服务端版本 **大于** 本地版本 → 弹出更新对话框
   - 否则 → 跳过
4. 对话框中的 iOS 下载链接走 `launchUrl` 跳 App Store，Android 链接走 `ota_update` 下载 APK

---

## 建议

1. **缓存**：后端可考虑对 `/app/version/v1` 做 CDN 缓存（如 5 分钟），避免频繁查询
2. **渠道分发**：需要时可以通过 `ANDROID_CHANNEL` 区分渠道，返回不同下载地址
3. **扩展字段**：如果后续需要支持强制更新版本号阈值，可增加 `forceUpdateMinVersion` 字段

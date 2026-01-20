# 本地配置文件说明

## 概述

系统现在支持优先从本地配置文件读取配置。如果本地配置文件不存在，会自动从 Apollo 配置中心读取。

## 配置文件位置

配置文件默认位于项目根目录下的 `config` 目录，也可以通过环境变量 `LOCAL_CONFIG_PATH` 指定自定义路径。

## 配置文件格式

配置文件使用 JSON 格式，文件名为配置 key + `.json`。

## 支持的配置项

### 1. HTTP 客户端配置 (dynamic_http_client_config.json)

文件路径: `config/dynamic_http_client_config.json`

示例配置:

```json
{
  "DialTimeout": "5s",
  "MaxConnsPerHost": 512,
  "MaxIdleConnDuration": "1m",
  "MaxConnDuration": "0s",
  "MaxConnWaitTimeout": "0s",
  "KeepAlive": true,
  "ClientReadTimeout": "0s",
  "ResponseBodyStream": true,
  "WriteTimeout": "0s"
}
```

字段说明:
- `DialTimeout`: TCP 连接建立时的最大等待时间（如 "5s", "1m"）
- `MaxConnsPerHost`: 每个主机允许的最大连接数
- `MaxIdleConnDuration`: 空闲连接最大存活时间（如 "1m", "30s"）
- `MaxConnDuration`: 最大连接存活时间，"0s" 表示不限制
- `MaxConnWaitTimeout`: 等待空闲连接的最大时间，"0s" 表示不限制
- `KeepAlive`: 是否使用长连接
- `ClientReadTimeout`: 读取响应的最大时间，"0s" 表示不限制
- `ResponseBodyStream`: 是否流式读取 body
- `WriteTimeout`: 写入请求超时时间，"0s" 表示不限制

### 2. 缓存配置 (dynamic_cache_config.json)

文件路径: `config/dynamic_cache_config.json`

示例配置:

```json
{
  "NumCounters": 100000,
  "MaxCost": 10000000,
  "BufferItems": 64,
  "BaseTTL": "10s",
  "Jitter": "10s"
}
```

字段说明:
- `NumCounters`: 统计计数器数量
- `MaxCost`: 最大缓存大小（字节），例如 10000000 表示约 10MB
- `BufferItems`: 内部队列缓冲
- `BaseTTL`: 基础 TTL（如 "10s", "1m"）
- `Jitter`: 随机浮动 TTL（如 "10s", "30s"）

### 3. 重试规则配置 (dynamic_retry_rule_config.json)

文件路径: `config/dynamic_retry_rule_config.json`

示例配置:

```json
{
  "rules": [
    {
      "describe": "500错误重试规则",
      "max-retry": 3,
      "retry-interval-time-seconds": 2,
      "reg": [
        {
          "object": "code",
          "regular-expression": "^5\\d{2}$"
        }
      ]
    },
    {
      "describe": "网络超时重试规则",
      "max-retry": 2,
      "retry-interval-time-seconds": 1,
      "reg": [
        {
          "object": "code",
          "regular-expression": "^504$"
        }
      ]
    }
  ]
}
```

字段说明:
- `rules`: 重试规则数组
  - `describe`: 规则描述
  - `max-retry`: 最大重试次数
  - `retry-interval-time-seconds`: 重试间隔时间（秒）
  - `reg`: 匹配规则数组
    - `object`: 匹配对象（如 "code", "msg"）
    - `regular-expression`: 正则表达式

## 使用方式

### 方式一：使用默认路径（推荐）

1. 在项目根目录创建 `config` 目录（如果不存在）
2. 在 `config` 目录下创建对应的 JSON 配置文件
3. 重启应用，配置会自动加载

### 方式二：使用自定义路径

1. 设置环境变量 `LOCAL_CONFIG_PATH` 指向配置目录
   ```bash
   export LOCAL_CONFIG_PATH=/path/to/your/config
   ```
2. 在指定目录下创建对应的 JSON 配置文件
3. 重启应用，配置会自动加载

## 配置优先级

1. **本地配置文件**（如果存在）
2. **Apollo 配置中心**（如果本地文件不存在）

## 注意事项

1. 本地配置文件使用 JSON 格式，确保格式正确
2. Duration 类型的字段使用字符串格式，如 "5s", "1m", "30s" 等
3. 如果本地配置文件存在但格式错误，会记录警告并尝试从 Apollo 读取
4. **本地配置文件不支持热更新**，如需热更新请使用 Apollo
5. 如果配置文件中某个字段缺失，会使用代码中的默认值

## 示例

创建 `config/dynamic_http_client_config.json`:

```json
{
  "DialTimeout": "10s",
  "MaxConnsPerHost": 1024,
  "MaxIdleConnDuration": "2m",
  "MaxConnDuration": "0s",
  "MaxConnWaitTimeout": "0s",
  "KeepAlive": true,
  "ClientReadTimeout": "30s",
  "ResponseBodyStream": true,
  "WriteTimeout": "10s"
}
```

重启应用后，HTTP 客户端会使用上述配置，而不再从 Apollo 读取。


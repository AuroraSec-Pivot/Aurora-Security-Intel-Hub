# Aurora Security Intel Hub (ASIH)
## 配置文件规范 v1.0

### 文档目的
本文档定义 ASIH 配置文件的核心规范（字段类型、默认值、语义约束），作为：
- 程序启动时配置校验（config validation）的依据
- `configs/config.example.yaml` 编写的权威指南
- 版本迭代中配置兼容与变更的基准

### 通用约定
- 时间间隔字段统一使用 **Go duration 格式**：`30s` / `5m` / `2h` / `24h`
- 环境变量注入使用 `${VAR_NAME}` 格式（如企业微信 Webhook URL）
- 未特殊说明的字段，均遵循「合理默认 + 显式覆盖」原则

---

## 1. 顶层结构
配置文件采用 YAML 格式，顶层仅包含以下 5 个核心节点，所有配置均嵌套在对应节点下：

```yaml
# 完整顶层结构示例（带默认值注释）
archive:        # 存储与归档配置（SQLite 为主）
pipeline:       # 运行模式、调度策略、安全护栏
http:           # HTTP 客户端全局配置（超时、重试、UA 等）
notifier:       # 推送渠道配置（MVP 仅支持企业微信）
sources:        # 情报源列表（MVP 仅实现 RSS 类型）
```

---

## 2. archive（归档/存储配置）
用于配置 SQLite 归档数据库，预留扩展其他数据库的接口，核心字段如下：

| 字段                     | 类型    | 必填 | 默认值 | 核心说明                                                                 |
|--------------------------|---------|------|--------|--------------------------------------------------------------------------|
| archive.driver           | string  | 否   | sqlite | 数据库驱动，**当前仅支持 sqlite**（预留 MySQL/PostgreSQL 扩展）           |
| archive.path             | string  | 是   | -      | SQLite DB 文件路径（建议绝对路径，如 `./data/aurora_intel.db`）           |
| archive.wal              | bool    | 否   | true   | 是否启用 WAL 模式（提升并发读写性能，建议保持开启）                       |
| archive.busy_timeout_ms  | int     | 否   | 5000   | SQLite 忙等待超时（毫秒），避免多协程读写锁冲突                          |
| archive.migrate_on_start | bool    | 否   | true   | 启动时自动执行数据库迁移（若实现迁移器，如表结构升级）                     |

### 示例配置
```yaml
archive:
  driver: sqlite
  path: "./data/aurora_intel.db"  # 建议将数据目录加入 .gitignore
  wal: true
  busy_timeout_ms: 5000
  migrate_on_start: true
```

---

## 3. pipeline（运行模式/调度/安全护栏）
控制程序运行模式、调度频率及各类安全护栏（防刷屏、防历史回流等）：

| 字段                                 | 类型     | 必填 | 默认值 | 核心说明                                                                 |
|--------------------------------------|----------|------|--------|--------------------------------------------------------------------------|
| pipeline.mode                        | string   | 否   | once   | 运行模式：`once`（单次执行） / `daemon`（常驻后台）                       |
| pipeline.interval                    | duration | 否   | 30m    | daemon 模式下每轮执行间隔（建议 ≥10m，避免高频请求打爆源站）              |
| pipeline.jitter                      | duration | 否   | 0s     | 执行间隔抖动（如 5m），避免所有源站请求集中在整点/半点                   |
| pipeline.concurrency                 | int      | 否   | 1      | 并发度（MVP 建议 1；V1 支持 worker pool 后可提升）                       |
| pipeline.max_push_per_run            | int      | 否   | 30     | 每轮最多推送条数（防刷屏护栏，超出部分仅归档不推送）                     |
| pipeline.default_push_policy         | string   | 否   | push   | 未显式声明时的默认推送策略：`push`（推送+归档） / `archive_only`（仅归档） |
| pipeline.drop_if_published_before_days | int     | 否   | 30     | 全局历史回流护栏：发布时间早于 N 天的内容，仅入库不推送                  |
| pipeline.dry_run                     | bool     | 否   | false  | 干跑模式：仅抓取+入库，不执行推送（便于灰度测试）                        |

### 关键语义约定
1. 「历史回流护栏」仅作用于**推送层**，不影响入库（需保留完整历史数据用于检索）
2. P2 优先级的源默认采用 `archive_only` 策略（见 `sources` 章节）

### 示例配置
```yaml
pipeline:
  mode: daemon               # 生产环境建议常驻
  interval: 30m              # 每 30 分钟抓取一次
  jitter: 5m                 # 随机偏移 0-5 分钟，避免整点请求
  concurrency: 1             # MVP 阶段单协程执行
  max_push_per_run: 30       # 每轮最多推 30 条
  default_push_policy: push
  drop_if_published_before_days: 30
  dry_run: false             # 测试时可设为 true
```

---

## 4. http（网络请求全局配置）
定义 HTTP 客户端的通用行为，支持全局超时、重试、代理、限流等：

### 基础字段
| 字段                 | 类型     | 必填 | 默认值                          | 核心说明                                                                 |
|----------------------|----------|------|---------------------------------|--------------------------------------------------------------------------|
| http.timeout         | duration | 否   | 15s                             | 单次 HTTP 请求超时（含连接+读取）                                        |
| http.user_agent      | string   | 否   | AuroraSecurityIntelHub/<version> | 全局 User-Agent（建议固定，便于源站识别）                                |
| http.proxy           | string   | 否   | 空                              | HTTP 代理地址（如 `http://127.0.0.1:7890`），按需配置                    |
| http.retry.max       | int      | 否   | 1                               | 请求失败重试次数（MVP 建议 0/1，避免过度重试）                           |
| http.retry.backoff   | duration | 否   | 2s                              | 重试退避时间（简单固定退避，V1 可实现指数退避）                          |

### 限流配置（V1 推荐，MVP 可忽略）
| 字段                          | 类型          | 默认值 | 核心说明                                                                 |
|-------------------------------|---------------|--------|--------------------------------------------------------------------------|
| http.rate_limit.default_rps   | float         | 0      | 全局默认每秒请求数（0 表示不限流）                                       |
| http.rate_limit.per_host      | map[string]float | {}   | 按域名覆盖限流规则（如 `www.reddit.com: 0.1` 表示每秒最多 0.1 次请求）    |

### 示例配置
```yaml
http:
  timeout: 15s
  user_agent: "AuroraSecurityIntelHub/v1.0"
  proxy: ""  # 无代理时留空
  retry:
    max: 1
    backoff: 2s
  # V1 限流配置（MVP 可注释/删除）
  rate_limit:
    default_rps: 0.5  # 全局每秒最多 0.5 次请求
    per_host:
      "www.reddit.com": 0.1  # 针对 Reddit 降低请求频率
      "xz.aliyun.com": 1.0   # 阿里云先知放宽限制
```

---

## 5. notifier（推送渠道配置）
MVP 阶段仅支持**企业微信 Webhook**，后续可扩展钉钉、Slack 等：

| 字段                          | 类型     | 必填 | 默认值 | 核心说明                                                                 |
|-------------------------------|----------|------|--------|--------------------------------------------------------------------------|
| notifier.wecom.webhook_url    | string   | 是   | -      | 企业微信机器人 Webhook URL（建议通过环境变量注入：`${WECOM_WEBHOOK_URL}`） |
| notifier.wecom.template       | string   | 否   | default | 消息模板名（需与 MESSAGE_TEMPLATE.md 定义对齐）                          |
| notifier.wecom.timeout        | duration | 否   | 10s    | 推送请求超时时间                                                         |
| notifier.wecom.retry.max      | int      | 否   | 1      | 推送失败重试次数                                                         |
| notifier.wecom.retry.backoff  | duration | 否   | 2s     | 推送重试退避时间                                                         |

### 关键语义
- 仅当推送**成功**时标记为 `MarkSent`；失败需保留 `pending` 状态，支持后续重试
- Webhook URL 禁止硬编码到配置文件，必须通过环境变量注入（安全合规）

### 示例配置
```yaml
notifier:
  wecom:
    webhook_url: "${WECOM_WEBHOOK_URL}"  # 从环境变量读取
    template: "default"
    timeout: 10s
    retry:
      max: 1
      backoff: 2s
```

---

## 6. sources（情报源列表）
每个 `source` 定义一个独立的情报抓取单元，MVP 仅支持 `type=rss`：

### 核心字段
| 字段                          | 类型          | 必填 | 默认值        | 核心说明                                                                 |
|-------------------------------|---------------|------|---------------|--------------------------------------------------------------------------|
| sources[].source_id           | string        | 是   | -             | 全局唯一 ID（建议语义化，如 `rss_xz_aliyun`）                            |
| sources[].type                | string        | 是   | rss           | 情报源类型（当前仅支持 rss）                                             |
| sources[].url                 | string        | 是   | -             | RSS/Atom 订阅地址（需可直接访问）                                        |
| sources[].interval            | duration      | 否   | pipeline.interval | 该源的抓取周期（优先级高于全局）                                        |
| sources[].priority            | string        | 否   | P1            | 优先级：P0（核心）/ P1（普通）/ P2（低优）                               |
| sources[].fingerprint_fields  | []string      | 否   | [url]         | 生成指纹的字段组合（用于去重，默认按 URL 去重）                          |
| sources[].push_policy         | string        | 否   | 见规则        | 推送策略：`push` / `archive_only`（默认规则见下文）                      |
| sources[].enabled             | bool          | 否   | true          | 是否启用该源                                                            |
| sources[].tags                | []string      | 否   | []            | 自定义标签（如 `web安全`、`漏洞情报`，便于筛选/路由）                    |
| sources[].notes               | map[string]any | 否   | {}            | 源级别附加参数（如覆盖全局历史回流护栏）                                 |

### push_policy 默认规则
1. 若显式配置 `push_policy`，以显式值为准；
2. 未显式配置时：
    - `priority == P2` → `archive_only`（仅归档）
    - 其他优先级 → 继承 `pipeline.default_push_policy`（默认 `push`）

### 源级别护栏（可选）
支持通过 `notes` 覆盖全局历史回流护栏：
```yaml
notes:
  drop_if_published_before_days: 7  # 该源仅推送 7 天内的内容（覆盖全局 30 天）
```

### 示例配置
```yaml
sources:
  - source_id: rss_xz_aliyun
    type: rss
    url: https://xz.aliyun.com/feed
    interval: 30m
    priority: P0
    fingerprint_fields: [url]
    push_policy: push  # 显式声明，优先级最高
    enabled: true
    tags: ["阿里云先知", "漏洞情报", "Web安全"]
    notes:
      drop_if_published_before_days: 7  # 仅推送 7 天内的内容

  - source_id: rss_freebuf
    type: rss
    url: https://www.freebuf.com/feed
    priority: P1
    enabled: true
    tags: ["FreeBuf", "安全资讯"]
    # 未显式声明 push_policy，继承 pipeline.default_push_policy

  - source_id: rss_reddit_netsec
    type: rss
    url: https://www.reddit.com/r/netsec/.rss
    priority: P2
    enabled: true
    tags: ["国外安全", "Reddit"]
    # P2 优先级，默认 push_policy = archive_only
```

---

## 7. 配置兼容性与版本（可选）
为便于版本迭代中的兼容处理，建议在配置文件顶层添加版本标识（MVP 可暂不实现）：
```yaml
config_version: 1  # 配置规范版本，用于程序识别并处理兼容逻辑
```

---

## 8. 安全与密钥管理
1. 敏感信息（如 Webhook URL、API Token）必须通过**环境变量注入**，禁止硬编码到配置文件；
2. 配置文件建议加入 `.gitignore`，仅提交 `config.example.yaml` 作为模板；
3. 生产环境需限制配置文件的读写权限（如 `chmod 600 config.yaml`）。

---

### 总结
1. 配置文件采用分层结构，核心分为 `archive`/`pipeline`/`http`/`notifier`/`sources` 5 个模块，职责清晰；
2. 所有字段均定义了默认值和明确的语义约束，优先遵循「全局默认 + 源级覆盖」原则；
3. 敏感信息通过环境变量注入，推送策略和历史护栏具备分级控制能力，兼顾灵活性与安全性；
4. 格式上统一使用表格梳理字段、代码块展示示例，关键语义和约定突出标注，便于查阅和落地。
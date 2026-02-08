# Aurora Security Intel Hub (ASIH)

安全情报与技术文章统合系统：多源采集 → 归一化 → 去重归档 → 通知推送（Go）。

> 团队：极光实验室  
> 状态：MVP 进行中（以 RSS 源为主）

---

## 目标与范围

Aurora Security Intel Hub（ASIH）面向安全团队，将分散在社区/媒体/漏洞库/公告渠道的安全内容进行统一汇聚与沉淀，并提供可追溯的归档与可控的推送能力。

当前阶段聚焦：
- RSS/Atom 多源采集
- 字段归一化（标题/链接/时间/来源等）
- 指纹去重（幂等入库）
- SQLite 归档与状态机（pending/sent）
- 企业微信（WeCom）Webhook 推送
- 一次性运行（RunOnce）与后续 daemon 定时能力

---

## 核心特性（规划/实现中）

- **多源聚合（RSS/Atom）**：支持配置多个源、分别设置抓取周期与优先级（P0/P1/P2）
- **可配置去重策略**：默认基于 URL 等字段生成 fingerprint，支持按 source 覆盖策略
- **可靠推送语义**：发送成功才标记 sent；失败保留 pending 可重试
- **归档与可追溯**：SQLite 持久化保存条目与原始 payload（JSON）
- **降噪护栏（默认约定）**
  - P2 源默认“仅入库不推送”
  - `published_at < now-30d` 默认“仅入库不推送”（防历史回流刷屏）

---

## 快速开始（MVP 预期）

> 下面命令以未来的 CLI 形态为准；实现完成后会更新为可直接执行的步骤。

### 1) 准备配置

复制示例配置并填写企业微信 Webhook：

- `configs/config.example.yaml` → `configs/config.yaml`
- 设置环境变量：
  - `WECOM_WEBHOOK_URL=...`

### 2) 运行（once）

```bash
./asih run --mode once --config configs/config.yaml
````

### 3) 查看归档（预期）

```bash
./asih entries list --since 7d --status pending
./asih entries list --since 7d --status sent
./asih entries show --id <id>
```

---

## 配置说明（概要）

配置文件建议包含以下段落（详见 `docs/CONFIG_SPEC.md`）：

* `archive`: SQLite 路径、WAL、busy_timeout
* `pipeline`: once/daemon、interval/jitter、并发、每轮推送上限
* `http`: UA、超时、rate limit（per-host）
* `notifier.wecom`: webhook_url、模板
* `sources[]`: 源列表（source_id/type/url/interval/priority/fingerprint_fields/notes）

---

## 默认源清单（当前）

当前已纳入（或计划纳入）的源见 `docs/SOURCES.md`，包含：

* 中文社区/媒体：52pojie、补天、先知、FreeBuf、安圈、看雪
* 英文社区：reddit r/netsec、0x00sec、security.stackexchange（含 web-security tag）
* 漏洞库：Exploit-DB
* 官方公告补充：Cisco PSIRT、CERT-EU Advisories

---

## 架构概览

ASIH 使用流水线（Pipeline）将每轮采集处理为可追溯的条目：

1. **Fetch**：从各 Provider 拉取原始内容（RSS/Atom 等）
2. **Normalize**：字段归一化、URL 规范化、补全来源信息
3. **Dedupe/Archive**：计算 fingerprint 并幂等写入 SQLite（pending）
4. **Send**：通过 Notifier 推送（企业微信 webhook）
5. **MarkSent**：推送成功才标记 sent（失败保留 pending）

状态机（MVP）：

* `pending`：已入库待推送/待重试
* `sent`：已成功推送

（V1 计划加入 `failed` 死信与 retry_count/last_error。）

---

## 开发与贡献（内部）

### 目录结构建议（规划）

```
cmd/asih/                 # main
internal/
  app/                    # 组装依赖、启动 run/daemon
  config/                 # 配置加载与校验
  pipeline/               # RunOnce、调度、并发、限速
  model/                  # Entry/RawItem/Status 等领域模型
  archive/                # sqlite 实现 + migrations
  provider/               # rss/github/html...
  notifier/               # wecom...
  normalize/              # url 规范化、fingerprint
docs/
configs/
```

### 工程约定

* `main` 分支始终可发布（主干开发）
* 代码合并要求：`gofmt` / `go vet` / `golangci-lint` / `go test ./...`
* Provider 必须带 fixture（样例 feed/HTML 快照 + 期望断言）
* 关键语义：**成功才 MarkSent**；失败保留 pending

---

## Roadmap

### MVP

* RunOnce 闭环（Fetch→Normalize→Archive→Send→MarkSent）
* SQLite 幂等去重（fingerprint UNIQUE）
* RSS Provider（多源）
* 企业微信 webhook Notifier
* 最小 CLI（list/show/run）

### V1（稳定性）

* daemon 模式 + interval/jitter
* worker pool 并发 + per-host rate limit
* 重试策略（指数退避）+ `failed` 死信
* HTML Provider（selector 驱动 + fixture 回归）
* 部署模板（systemd/docker）与 RUNBOOK

### V2（增强）

* 检索增强（SQLite FTS5/外部索引）
* Digest（日/周报）
* 相似度归并、更强去重策略
* 知识库同步（Notion/Confluence/GitHub Wiki 等）

---

## 安全与合规提示

* 本项目仅做公开信息聚合与通知转发，不对内容真实性做担保。
* 对于可能存在版权/转载限制的内容，建议仅保留标题、链接与最小摘要，并回链到原站。
* 某些站点可能存在访问频率限制（429/封禁风险），请合理配置 interval、rate limit 与 UA。

---

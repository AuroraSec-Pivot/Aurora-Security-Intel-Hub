Database Schema (SQLite) - Aurora Security Intel Hub

本文档定义 ASIH 的 SQLite 数据库表结构、索引与状态机语义，并给出迁移建议。

目标：

- 支撑“多源采集 → 归一化 → 幂等去重入库 → 推送 → 标记 sent”的闭环

- 重启不重复推送（依赖 fingerprint UNIQUE + status）

- 为 V1 的重试/死信/查询接口预留扩展字段


---
1. 存储引擎与连接参数

- 引擎：SQLite 3

- 建议设置：


- journal_mode = WAL

- busy_timeout = archive.busy_timeout_ms

- foreign_keys = ON（即使当前未使用外键，也建议开启）


---
2. 领域概念与状态机

2.1 Entry（条目）

一条 entry 对应一次“可推送的内容单元”，通常来源于 RSS item。

2.2 状态机（MVP）

- pending：已入库，待推送/待重试

- sent：已成功推送（推送成功后才允许进入该状态）

只有 notifier 返回成功才允许 MarkSent。关键语义：

2.3 V1 扩展（预留）

- failed：重试耗尽进入死信（可人工处理/手动 retry）

- retry_count / last_error / last_attempt_at 等字段用于诊断与退避调度


---
3. 表结构

3.1 entries（核心表）

用于存储归一化后的条目与最小可复现信息，分为「MVP 必选字段」与「V1 预留字段」，按需选择是否提前创建预留字段。

3.1.1 MVP 必选字段定义

字段

类型

必填

说明

id

INTEGER

是

主键，自增

source_id

TEXT

是

来源 ID（如 rss_xz_aliyun）

source_type

TEXT

否

来源类型（MVP 可固定 rss；预留扩展）

title

TEXT

是

标题（归一化后）

url

TEXT

是

规范化后的 canonical URL

url_raw

TEXT

否

原始 URL（可选，便于排错）

author

TEXT

否

作者/发布者（可选）

published_at

TEXT

否

发布时间（ISO-8601 字符串，建议存 UTC）

fetched_at

TEXT

是

抓取时间（ISO-8601，UTC）

fingerprint

TEXT

是

去重指纹（建议 sha256:<hex>）

status

TEXT

是

pending/sent（V1: +failed）

priority

TEXT

否

P0/P1/P2（可选，但建议存，便于分级筛选）

push_policy

TEXT

否

push/archive_only（可选，便于查询筛选）

tags_json

TEXT

否

JSON 数组字符串（如 ["cn","advisory"]，便于标签筛选）

summary

TEXT

否

可选摘要（MVP 可不填）

content_hash

TEXT

否

可选：正文/摘要 hash（用于更强去重/变更检测）

payload_json

TEXT

否

原始/归一化后的 payload（JSON），用于追溯排查

created_at

TEXT

是

入库时间（ISO-8601，UTC）

updated_at

TEXT

是

最近更新时间（ISO-8601，UTC）

3.1.2 V1 预留字段（可选创建）

若不想后续引入迁移框架，可在 MVP 阶段提前创建这些字段，暂不使用；若计划使用迁移框架，建议 V1 再添加。

字段

类型

说明

retry_count

INTEGER

推送失败计数（用于重试退避）

last_error

TEXT

最近一次推送失败原因（截断存储，便于排查）

last_attempt_at

TEXT

最近一次推送尝试时间（ISO-8601，UTC）

sent_at

TEXT

推送成功时间（status=sent 时填写，ISO-8601，UTC）


---
4. 约束与索引（核心）

4.1 去重约束

- fingerprint 必须 唯一：保证幂等入库与重启不重复推送，是核心约束

4.2 推荐索引

索引用于提升查询与筛选性能，MVP 建议全部创建：

- (status, fetched_at)：用于按状态拉取待推送队列（核心索引）

- (source_id, fetched_at)：用于按源查询、统计抓取量

- (published_at)：用于按发布时间过滤（注意处理 NULL 值）

- (priority)：用于分级统计/筛选（可选，建议创建）


---
5. 推荐 DDL（两种版本，二选一）

注意：时间字段存 TEXT（ISO-8601），实现简单；如后续需要更强查询性能，可改为 INTEGER（Unix epoch）。建议统一使用 UTC 存储。

5.1 版本1：MVP 纯核心版（推荐，需后续迁移框架）

仅包含 MVP 必选字段，V1 扩展字段通过迁移脚本添加，适合计划引入迁移框架的场景。

CREATE TABLE IF NOT EXISTS entries (
id            INTEGER PRIMARY KEY AUTOINCREMENT,
source_id     TEXT NOT NULL,
source_type   TEXT,
title         TEXT NOT NULL,
url           TEXT NOT NULL,
url_raw       TEXT,
author        TEXT,
published_at  TEXT,
fetched_at    TEXT NOT NULL,
fingerprint   TEXT NOT NULL,
status        TEXT NOT NULL CHECK (status IN ('pending','sent')),
priority      TEXT,
push_policy   TEXT,
tags_json     TEXT,
summary       TEXT,
content_hash  TEXT,
payload_json  TEXT,
created_at    TEXT NOT NULL,
updated_at    TEXT NOT NULL
);

-- 核心去重索引（必选）
CREATE UNIQUE INDEX IF NOT EXISTS ux_entries_fingerprint ON entries(fingerprint);

-- 推荐查询索引
CREATE INDEX IF NOT EXISTS ix_entries_status_fetched_at ON entries(status, fetched_at);
CREATE INDEX IF NOT EXISTS ix_entries_source_fetched_at ON entries(source_id, fetched_at);
CREATE INDEX IF NOT EXISTS ix_entries_published_at ON entries(published_at);
CREATE INDEX IF NOT EXISTS ix_entries_priority ON entries(priority);

5.2 版本2：MVP+V1预留字段版（无需后续迁移）

一次性创建所有 MVP 字段与 V1 预留字段，暂不使用预留字段，适合不想引入迁移框架、追求轻量的场景。

CREATE TABLE IF NOT EXISTS entries (
id            INTEGER PRIMARY KEY AUTOINCREMENT,
source_id     TEXT NOT NULL,
source_type   TEXT,
title         TEXT NOT NULL,
url           TEXT NOT NULL,
url_raw       TEXT,
author        TEXT,
published_at  TEXT,
fetched_at    TEXT NOT NULL,
fingerprint   TEXT NOT NULL,
status        TEXT NOT NULL CHECK (status IN ('pending','sent')), -- V1 可改为 IN ('pending','sent','failed')
priority      TEXT,
push_policy   TEXT,
tags_json     TEXT,
summary       TEXT,
content_hash  TEXT,
payload_json  TEXT,
created_at    TEXT NOT NULL,
updated_at    TEXT NOT NULL,
-- V1 预留字段（MVP 暂不使用）
retry_count    INTEGER DEFAULT 0,
last_error     TEXT,
last_attempt_at TEXT,
sent_at        TEXT
);

-- 核心去重索引（必选）
CREATE UNIQUE INDEX IF NOT EXISTS ux_entries_fingerprint ON entries(fingerprint);

-- 推荐查询索引
CREATE INDEX IF NOT EXISTS ix_entries_status_fetched_at ON entries(status, fetched_at);
CREATE INDEX IF NOT EXISTS ix_entries_source_fetched_at ON entries(source_id, fetched_at);
CREATE INDEX IF NOT EXISTS ix_entries_published_at ON entries(published_at);
CREATE INDEX IF NOT EXISTS ix_entries_priority ON entries(priority);

-- V1 可添加的补充索引（暂不创建）
-- CREATE INDEX IF NOT EXISTS ix_entries_retry_count ON entries(retry_count);
-- CREATE INDEX IF NOT EXISTS ix_entries_last_attempt_at ON entries(last_attempt_at);


---
6. 关键写入/更新语义（接口级约定）

6.1 UpsertPending（幂等入库）

目的：对同 fingerprint 的条目只保留一份记录，避免重复入库与重复推送。

推荐语义：

- 先计算该条目的 fingerprint（基于 fingerprint_fields 配置）

- 使用 INSERT ... ON CONFLICT(fingerprint) DO NOTHING 实现幂等

- 返回：inserted=true/false（用于统计“新增条目数/去重丢弃数”）

示例（SQLite，适配两种 DDL 版本，预留字段可忽略）：

INSERT INTO entries (
source_id, source_type, title, url, url_raw, author,
published_at, fetched_at, fingerprint, status, priority, push_policy,
tags_json, summary, payload_json, created_at, updated_at
) VALUES (
?, ?, ?, ?, ?, ?,
?, ?, ?, 'pending', ?, ?,
?, ?, ?, ?, ?
)
ON CONFLICT(fingerprint) DO NOTHING;

DO NOTHING建议：如果使用 ，不要更新已有记录，避免“新抓到的噪声数据”覆盖原始有效内容；如确需更新（例如 title 变化），应制定明确的更新策略（V2 再考虑）。

6.2 MarkSent（成功后标记 sent）

触发条件：只有 notifier 推送成功后，才允许调用该接口。

建议语义：

- 仅当当前 status=pending 时，才更新为 sent（避免重复标记）

- 同时更新 updated_at；若使用版本2 DDL，可同步更新 sent_at

示例（适配版本1 DDL）：

UPDATE entries
SET status='sent', updated_at=?
WHERE id=? AND status='pending';

示例（适配版本2 DDL，同步更新 sent_at）：

UPDATE entries
SET status='sent', updated_at=?, sent_at=?
WHERE id=? AND status='pending';

返回影响行数说明：

- 1：成功标记为 sent，符合预期

- 0：表示条目已是 sent 或不存在（不视为错误，但应记录日志用于诊断）


---
7. 查询建议（MVP）

7.1 拉取待推送队列（pending）

按抓取时间升序，获取待推送条目，配合 pipeline.max_push_per_run 控制每轮推送数量：

SELECT *
FROM entries
WHERE status='pending'
ORDER BY fetched_at ASC
LIMIT ?;

7.2 近7天已推送条目（sent）

用于查询近期推送记录，便于审计与回溯：

SELECT *
FROM entries
WHERE status='sent' AND fetched_at >= ? -- ? 传入 7 天前的 UTC 时间（ISO-8601 格式）
ORDER BY fetched_at DESC
LIMIT ?;


---
8. 迁移策略（建议）

8.1 推荐方案（文件式迁移）

采用 migrations/*.sql 文件管理 schema 变更，启动时自动执行，保证幂等性：

- 001_init.sql：执行版本1 DDL，建立 entries 表与核心索引（MVP）

- 002_add_retry_fields.sql：V1 时添加预留字段，更新 status 约束：
  -- 添加 V1 预留字段
  ALTER TABLE entries ADD COLUMN retry_count INTEGER DEFAULT 0;
  ALTER TABLE entries ADD COLUMN last_error TEXT;
  ALTER TABLE entries ADD COLUMN last_attempt_at TEXT;
  ALTER TABLE entries ADD COLUMN sent_at TEXT;

-- 更新 status 约束，支持 failed 状态
ALTER TABLE entries DROP CONSTRAINT IF EXISTS entries_status_check;
ALTER TABLE entries ADD CHECK (status IN ('pending','sent','failed'));

-- 添加 V1 补充索引
CREATE INDEX IF NOT EXISTS ix_entries_retry_count ON entries(retry_count);
CREATE INDEX IF NOT EXISTS ix_entries_last_attempt_at ON entries(last_attempt_at);

8.2 轻量方案（无迁移框架）

直接使用版本2 DDL，一次性创建所有字段，MVP 阶段暂不使用预留字段，V1 时直接启用即可，无需修改表结构。

注意：轻量方案适合小型项目，后续 schema 变更需手动处理，推荐优先使用文件式迁移方案。


---
9. 变更记录

- 2026-02-08：定义 MVP entries schema（pending/sent + fingerprint UNIQUE），并预留 V1 重试/failed 扩展字段；提供两种 DDL 版本，适配不同迁移需求。

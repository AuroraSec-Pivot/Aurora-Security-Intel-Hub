# Sources (Aurora Security Intel Hub)

本文件定义 ASIH 的默认情报源清单与运营约束，用于：
- 配置来源（source_id / url / interval / priority）
- 去重策略（fingerprint_fields）
- 降噪/护栏（P2 默认不推送、历史回流过滤等）

## 全局默认策略（已确认）

1) **P2 源默认“仅入库不推送”**
    - 仍然抓取、归档与去重
    - 需要推送时再做单独路由（独立机器人/群）

2) **历史回流护栏（默认开启）**
    - `published_at < now-30d` 的条目：**仅入库不推送**
    - 适用于可能出现历史内容回流的源（论坛/社区类尤为常见）

> 备注：上述策略属于“推送层策略”，不影响入库与查询。

---

## 字段说明

- **source_id**：唯一稳定标识（全仓库唯一，后续不可随意更改）
- **type**：当前统一为 `rss`
- **url**：RSS/Atom 地址
- **interval**：建议抓取周期（MVP 可先不实现按源 interval 调度，但文档先定口径）
- **priority**：P0/P1/P2（用于推送策略与调度策略）
- **fingerprint_fields**：用于生成 fingerprint 的字段组合（后续可按 source 覆盖默认策略）
    - 可选：`item_id` / `url` / `title` / `published_at`
- **push_policy**：
    - `push`：允许推送（仍受历史回流护栏与每轮上限影响）
    - `archive_only`：仅入库不推送（P2 默认）
- **notes**：风险点、降噪建议、特殊处理

---

## Source List

### P0（高优先 / 默认推送）

#### rss_xz_aliyun
- type: rss
- url: https://xz.aliyun.com/feed
- interval: 30m
- priority: P0
- fingerprint_fields: [url]
- push_policy: push
- notes: 文章类，质量较高；建议 URL normalize。

#### rss_butian_forum
- type: rss
- url: https://forum.butian.net/Rss
- interval: 30m
- priority: P0
- fingerprint_fields: [url, title]
- push_policy: push
- notes:
    - 可能出现历史内容回流；受全局 now-30d 护栏保护。
    - 建议设置每轮推送上限（如 30）避免刷屏。

#### rss_exploit_db
- type: rss
- url: https://www.exploit-db.com/rss.xml
- interval: 6h
- priority: P0
- fingerprint_fields: [url, title]
- push_policy: push
- notes: 漏洞条目更新；频率相对稳定。

#### rss_cisco_psirt
- type: rss
- url: https://sec.cloudapps.cisco.com/security/center/rss.x?i=44
- interval: 6h
- priority: P0
- fingerprint_fields: [url, title]
- push_policy: push
- notes: 官方安全公告，信噪比高。

---

### P1（中优先 / 默认推送，可能需要限流或上限）

#### rss_freebuf
- type: rss
- url: https://www.freebuf.com/feed
- interval: 1h
- priority: P1
- fingerprint_fields: [url]
- push_policy: push
- notes: 内容量较大；建议配合每轮推送上限。

#### rss_anquanke
- type: rss
- url: https://api.anquanke.com/data/v1/rss
- interval: 1h
- priority: P1
- fingerprint_fields: [url]
- push_policy: push
- notes: 文章类；建议 URL 规范化（去参数等按需）。

#### rss_xz_52pojie_forum
- type: rss
- url: https://www.52pojie.cn/forum.php?mod=rss
- interval: 2h
- priority: P1
- fingerprint_fields: [url, title]
- push_policy: push
- notes:
    - 论坛聚合，噪声可能偏高；建议低频与每轮推送上限。
    - 后续可补充“精华/导读”类 RSS 作为替代或并行 feed。

#### rss_kanxue_bbs
- type: rss
- url: https://bbs.kanxue.com/rss.php
- interval: 2h
- priority: P1
- fingerprint_fields: [url, title]
- push_policy: push
- notes: 论坛类；建议低频与上限，必要时分频道路由。

#### rss_cert_eu_advisories
- type: rss
- url: https://cert.europa.eu/publications/security-advisories-rss
- interval: 12h
- priority: P1
- fingerprint_fields: [url, title]
- push_policy: push
- notes: 官方通告汇总，补全“公告类”覆盖面。

---

### P2（低优先 / 默认仅入库不推送）

#### rss_reddit_netsec
- type: rss
- url: https://www.reddit.com/r/netsec/.rss
- interval: 6h
- priority: P2
- fingerprint_fields: [url, title]
- push_policy: archive_only
- notes:
    - 容易触发 429；建议强制 UA、低频、429 退避（backoff）。
    - 如需推送，建议单独路由到独立频道。

#### rss_security_stackexchange_all
- type: rss
- url: https://security.stackexchange.com/feeds
- interval: 12h
- priority: P2
- fingerprint_fields: [url, title]
- push_policy: archive_only
- notes: 问答类；重复与噪声较多，适合仅入库检索。

#### rss_security_stackexchange_web_security
- type: rss
- url: https://security.stackexchange.com/feeds/tag/web-security
- interval: 12h
- priority: P2
- fingerprint_fields: [url, title]
- push_policy: archive_only
- notes: tag 聚合；可作为“Web 安全”专题入库源。

#### rss_0x00sec
- type: rss
- url: https://0x00sec.org/feed.xml
- interval: 12h
- priority: P2
- fingerprint_fields: [url, title]
- push_policy: archive_only
- notes: 社区研究帖质量参差；默认仅入库，后续可人工挑选推送。

---

## 变更记录

- 2026-02-08：初始化第一批 sources（11 + 2），并确认全局策略：
    - P2 默认仅入库不推送
    - now-30d 历史回流护栏（仅入库不推送）

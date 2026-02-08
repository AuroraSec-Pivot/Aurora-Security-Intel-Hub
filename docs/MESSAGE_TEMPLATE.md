Message Template Specification (WeCom) - Aurora Security Intel Hub

本文档定义 ASIH 的企业微信（WeCom）机器人推送消息模板规范，目标：

- 统一消息结构，降低噪声，提高可读性

- 保持可扩展：后续支持摘要/标签/路由/日报等

- 明确截断与护栏策略，避免超长与刷屏

markdown当前实现目标：WeCom  消息（机器人 webhook）。


---
1. 通用原则

1) 最小可用信息集

- 标题（可点击）

- 来源（source_id 或更友好的 source_name）

- 发布时间（published_at，缺失则显示 fetched_at）

- 原文链接（url）

- 可选：标签（tags）、简短摘要（summary）

2) 不在消息中复制整篇内容

- 默认不发送全文，仅发送标题 + 链接 + 元信息

- 如启用摘要，必须截断（见 4.1）

3) 推送语义

- 仅当 webhook 返回成功才MarkSent

- 推送失败保留 pending，可重试


---
2. 数据字段（Entry → Message）

字段

说明

示例

entry.title

标题

Apache Tomcat RCE Analysis

entry.url

原文链接

https://...

entry.source_id

源 ID

rss_xz_aliyun

entry.source_name

源展示名（可选）

阿里云先知

entry.published_at

发布时间（可选）

2026-02-08T10:00:00Z

entry.fetched_at

抓取时间

2026-02-08T10:05:00Z

entry.tags

标签（可选）

["cn","advisory"]

entry.summary

摘要（可选）

...

entry.priority

P0/P1/P2

P0

entry.fingerprint

去重指纹（可选展示/调试）

sha256:...`

source_namesource_idMVP 建议： 先不强制；可以用  显示。


---
3. 模板列表

3.1default（默认模板）

用于 P0/P1 的常规推送，以下提供两个贴合安全团队日常的排版版本（字段一致，仅展示风格不同，可二选一）。

版本1：标题超链接+简洁元信息（推荐，最省空间、重点突出）

**结构（markdown）**

1. 标题（嵌入原文链接，优先级醒目）

2. 一行元信息：来源 + 时间 + 优先级（用符号分隔，简洁直观）

3. 标签（可选，用符号拼接，不占多余行数）

4. 摘要（可选，启用则显示，自动截断）

**示例**

【<priority>】[<title_truncated>](<url>)
来源：<source_name_or_id> | 时间：<time_display>
标签：<tag1> · <tag2> · ...（最多6个）
摘要：<summary_truncated>

**实际渲染效果（企业微信）**

【P0】[Apache Tomcat RCE Analysis](https://xz.aliyun.com/xxx)
来源：rss_xz_aliyun | 时间：2026-02-08 10:00
标签：阿里云先知 · 漏洞情报 · Web安全
摘要：本文详细分析了Apache Tomcat最新RCE漏洞的成因、利用条件及防御方法，影响版本为9.0.xx-9.0.yy，建议相关用户及时升级补丁...

版本2：列表式排版（清晰规整，适合偏好结构化展示的团队）

**结构（markdown）**

1. 优先级标题（醒目）

2. 列表展示所有信息（每条一行，层级清晰，便于快速扫描）

**示例**

【<priority>】安全情报推送
- 标题：[<title_truncated>](<url>)
- 来源：<source_name_or_id>
- 时间：<time_display>
- 标签：<tag1>、<tag2>、...（最多6个）
- 摘要：<summary_truncated>

**实际渲染效果（企业微信）**

【P0】安全情报推送
- 标题：[Apache Tomcat RCE Analysis](https://xz.aliyun.com/xxx)
- 来源：rss_xz_aliyun
- 时间：2026-02-08 10:00
- 标签：阿里云先知、漏洞情报、Web安全
- 摘要：本文详细分析了Apache Tomcat最新RCE漏洞的成因、利用条件及防御方法，影响版本为9.0.xx-9.0.yy，建议相关用户及时升级补丁...

说明：两个版本均符合安全团队日常推送习惯，版本1更简洁省屏，版本2更规整易扫描；MVP可二选一实现，后续可通过配置切换。

时间显示规则

- 若 published_at 存在：显示 published_at

- 否则：显示 fetched_at，并在前缀注明 抓取时间

推荐格式（人类可读）：

- YYYY-MM-DD HH:mm（建议按 Asia/Singapore 或配置时区格式化）


---
3.2 high_priority（高优先级模板，可选）

用于你们后续想把某些源或关键词提升为更醒目的推送（例如 P0 厂商公告、重大漏洞），沿用对应版本的排版，仅增加醒目符号。

**版本1（对应默认模板版本1）示例**

🚨【<priority>】[<title_truncated>](<url>)
来源：<source_name_or_id> | 时间：<time_display>
标签：<tag1> · <tag2> · ...（最多6个）
摘要：<summary_truncated>

**版本2（对应默认模板版本2）示例**

🚨【<priority>】安全情报推送（高优）
- 标题：[<title_truncated>](<url>)
- 来源：<source_name_or_id>
- 时间：<time_display>
- 标签：<tag1>、<tag2>、...（最多6个）
- 摘要：<summary_truncated>

defaultMVP 可以先只实现 ，但建议保留模板名以免后续 breaking change。


---
4. 截断与限制（重要）

企业微信机器人 markdown 有长度限制与显示限制（不同版本可能有差异），因此 ASIH 必须执行“安全截断”。

4.1 建议截断策略（实现建议）

- title：


- 最大 120 字符（超出截断并加 …）

- summary（如果启用）：


- 最大 300 字符（超出截断并加 …）

- 不启用摘要时，不输出 摘要： 行

- tags：


- 最多显示 6 个标签，超出忽略或追加 +N

4.2 URL 处理

- 使用 Normalize 后的 canonical URL 作为最终展示与 fingerprint 基础

- 对于含大量 query 参数的 URL：


- fingerprint 可使用“去参数后 URL”（按 normalize 策略）

- 展示 URL 可保留原始，但建议同样规范化


---
5. 推送与降噪策略（与模板配合）

5.1 P2 默认不推送

- priority == P2 默认 push_policy = archive_only

- 如需推送，建议路由到单独机器人或单独群（后续实现）

5.2 历史回流护栏

- published_at < now-30d → 仅入库不推送

- 对缺失 published_at 的条目：


- 默认允许推送（按 fetched_at）

- 但建议对论坛类源设置更严格规则（可通过 source notes 覆盖）

5.3 每轮推送上限

- 建议 pipeline.max_push_per_run = 30

- 超出的条目：仅入库保持 pending（或延后推送，视实现策略）


---
6. WeCom Markdown 落地实现建议（非强制）

在 notifier.wecom 内建议实现：

- Render(templateName, entry) -> markdown string

- 将“字段清洗/截断”封装为独立函数，保证不同模板复用一致规则

- Send(ctx, msg) 超时可配置；失败保留 pending 可重试


---
7. 后续扩展（V1/V2）

- 按 tags 或关键词路由到不同机器人/群

- Digest（日/周报）模板

- 追加“调试字段”开关（是否展示 fingerprint/source_id）

- 引入摘要生成（可选，本地规则/外部服务）

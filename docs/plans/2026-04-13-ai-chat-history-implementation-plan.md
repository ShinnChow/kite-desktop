# AI 对话历史持久化 Implementation Plan

> **Goal:** 让 AI 聊天的“历史”从前端临时本地缓存升级为桌面端可稳定查询、可跨窗口恢复、可按会话管理的 SQLite 持久化能力，解决当前历史面板为空或不稳定的问题。

> **Tech Stack:** Go, Gin, GORM, SQLite/MySQL/Postgres, React 19, TanStack Query.

---

## 当前排查结论

### 0. 桌面端持久化约束

这个项目当前已经收敛为桌面端产品，持久化设计应以应用内数据库为准，而不是浏览器存储。

本方案后续统一采用以下原则：

- `SQLite` 是桌面端持久化数据的唯一真源
- `localStorage` 不再承载任何“需要重启后仍可靠存在”的业务数据
- `localStorage` 最多只用于短期 UI 状态或一次性迁移辅助标记

因此，AI 对话历史不能继续依赖 `localStorage` 作为正式存储。

### 1. 目前没有对话历史入库

当前仓库里没有 AI 聊天会话或聊天消息的数据库模型，也没有历史查询接口。

- 前端历史只保存在 `localStorage`
  - `ui/src/hooks/use-ai-chat.ts`
  - 使用的 key 前缀是 `ai-chat-history-`
  - 当前桌面端固定写入 key：`ai-chat-history-desktop`
- 前端代码里已经明确写了注释：
  - `ui/src/hooks/use-ai-chat.ts`
  - `// TODO: save in backend.`
- 后端现有 AI 相关接口只有：
  - `POST /api/v1/ai/chat`
  - `POST /api/v1/ai/execute/continue`
  - `POST /api/v1/ai/input/continue`
  - 路由定义见 `internal/server/routes.go`
- 当前数据库里与 AI 相关的持久化只有 `PendingSession`
  - `pkg/model/pending_session.go`
  - 作用是保存“待确认/待补充输入”的临时会话，TTL 15 分钟
  - 不是聊天历史

### 2. 当前“历史”按钮的数据源是前端内存 + localStorage

`useAIChat()` 的 `history` 状态只从 `localStorage` 加载，不会向后端请求：

- 载入：`loadHistoryFromStorage()`
- 保存：`saveHistoryToStorage()`
- 会话更新：`upsertSession()`

在桌面端语境下，这意味着当前历史功能有几个天然问题：

- 无法作为桌面端正式持久化能力使用
- 新窗口/新 webview 不一定能看到同一份历史
- localStorage 被清空后历史全部丢失
- 后端无法做分页、过滤、清理策略
- 无法做真正的“历史记录”查询与审计

### 3. 对“历史面板为空”的判断

基于代码可以确认：

- “历史为空”不是因为后端没查到数据，而是因为当前根本没有后端历史存储链路
- 如果本地 `localStorage` 没写进去、被清掉、或者窗口隔离，就会直接显示空历史

也就是说，当前历史功能本质上还是“浏览器态临时缓存”，不符合桌面端应有的持久化设计。

---

## 目标行为

本功能完成后，AI 聊天历史应满足：

1. 每次对话都有稳定的会话 ID，并在数据库持久化。
2. 历史列表从 SQLite 读取，而不是只读 `localStorage`。
3. 点击某个历史项，可以完整恢复该会话消息。
4. 支持删除历史会话。
5. 支持按集群过滤历史。
6. 新开窗口、应用重启后，历史仍然存在。
7. 当前 `PendingSession` 继续负责“待确认动作恢复”，但不再承担历史能力。
8. AI 历史以数据库为单一真源，前端不再维护独立持久化副本。

---

## 数据模型设计

建议新增两张表，不复用 `PendingSession`。

### Table 1: `ai_chat_sessions`

用途：保存会话元信息，支撑历史列表。

建议字段：

- `id`
- `session_id`
- `title`
- `cluster_name`
- `page`
- `namespace`
- `resource_name`
- `resource_kind`
- `message_count`
- `last_message_at`
- `created_at`
- `updated_at`
- `deleted_at` 可选，若希望支持软删除

约束建议：

- `session_id` 唯一索引
- `(cluster_name, updated_at desc)` 索引，支撑历史列表

### Table 2: `ai_chat_messages`

用途：保存单条消息，支撑恢复整个会话。

建议字段：

- `id`
- `session_id`
- `seq`
- `role`
- `content`
- `thinking`
- `tool_name`
- `tool_call_id`
- `tool_args`
- `tool_result`
- `action_status`
- `input_request`
- `pending_action`
- `created_at`

说明：

- `session_id + seq` 唯一，保证消息顺序稳定
- `tool_args`、`input_request`、`pending_action` 建议复用现有 `JSONField`
- 这样可以完整恢复当前前端 `ChatMessage` 结构，避免历史回放和实时会话渲染分叉

---

## API 设计

建议新增以下接口：

### 1. 获取历史列表

`GET /api/v1/ai/sessions?page=1&pageSize=20&clusterName=xxx`

返回：

- `data`
- `total`
- `page`
- `pageSize`

每项包含：

- `sessionId`
- `title`
- `clusterName`
- `messageCount`
- `createdAt`
- `updatedAt`

### 2. 获取单个会话详情

`GET /api/v1/ai/sessions/:sessionId`

返回：

- 会话元信息
- 按 `seq` 排序后的消息列表

### 3. 创建或更新会话快照

`PUT /api/v1/ai/sessions/:sessionId`

请求体包含：

- 会话元信息
- 当前完整消息数组

说明：

- 先用“整会话快照 upsert”实现，逻辑最稳，前后端最容易收敛
- 后续若要优化，再拆成“追加消息”接口

### 4. 删除历史会话

`DELETE /api/v1/ai/sessions/:sessionId`

说明：

- 删除 `ai_chat_sessions`
- 级联删除 `ai_chat_messages`

---

## 前端改造方案

主要改 `ui/src/hooks/use-ai-chat.ts` 和 `ui/src/components/ai-chat/ai-chatbox.tsx`。

### 1. 历史数据源改成数据库单一真源

- 初始化时调用 `GET /api/v1/ai/sessions`
- 前端不再把历史正式写入 `localStorage`
- `localStorage` 如需保留，只允许用于一次性迁移标记

### 2. 保存时机

建议在以下时机触发 `PUT /api/v1/ai/sessions/:sessionId`：

- 用户发出消息后，立即保存一次
- AI 流式响应结束后，再保存一次
- 执行动作确认完成后，再保存一次
- 输入补充完成后，再保存一次

这样即使应用异常退出，至少能保住用户已发送的问题。

### 3. 加载历史

- 点击历史项时，不再直接从本地 `history` state 读完整消息
- 改为先请求 `GET /api/v1/ai/sessions/:sessionId`
- 成功后填充 `messages` 与 `currentSessionId`

### 4. 删除历史

- 删除时调用后端 `DELETE`
- 成功后刷新列表或本地同步删除

### 5. localStorage 迁移策略

首次接入后端时：

- 若后端历史为空，但本地 `ai-chat-history-desktop` 有数据
- 可执行一次性导入，把旧 `localStorage` 数据写入 SQLite
- 迁移成功后打一个本地标记，例如 `ai-chat-history-migrated-v1=true`

这样不会让已有本地聊天记录直接丢失。

说明：

- 迁移是兼容旧实现的过渡方案，不是长期存储方案
- 迁移完成后，历史读写全部走数据库

---

## 后端改造方案

### Task 1: 新增模型与迁移

**Files:**

- Create: `pkg/model/ai_chat_session.go`
- Create: `pkg/model/ai_chat_message.go`
- Modify: `pkg/model/model.go`

**验收标准：**

- `AutoMigrate` 可创建两张新表
- SQLite/MySQL/Postgres 均可正常启动

### Task 2: 新增 handler 与路由

**Files:**

- Create: `pkg/ai/history_handler.go`
- Modify: `internal/server/routes.go`

**接口：**

- `GET /api/v1/ai/sessions`
- `GET /api/v1/ai/sessions/:sessionId`
- `PUT /api/v1/ai/sessions/:sessionId`
- `DELETE /api/v1/ai/sessions/:sessionId`

**验收标准：**

- 能分页查询会话列表
- 能按 `session_id` 读取完整消息
- 能完整覆盖式保存一次会话快照

### Task 3: 前端 API 封装与移除 localStorage 正式存储

**Files:**

- Create: `ui/src/lib/api/ai-history.ts`
- Modify: `ui/src/hooks/use-ai-chat.ts`

**验收标准：**

- 历史列表来自数据库接口
- 历史详情来自数据库接口
- 新消息会触发会话快照保存
- 正式历史写入不再调用 `localStorage.setItem`

### Task 4: 兼容迁移本地历史

**Files:**

- Modify: `ui/src/hooks/use-ai-chat.ts`

**验收标准：**

- 已有 `localStorage` 历史在升级后可自动导入
- 导入失败不阻塞正常聊天

### Task 5: 测试补齐

**Files:**

- Create: `pkg/ai/history_handler_test.go`
- Create: `pkg/model/ai_chat_session_test.go`
- Create: `ui/src/hooks/use-ai-chat.test.tsx` 或相关集成测试

**验收标准：**

- 后端 CRUD 有测试
- 前端加载/保存/删除/迁移有测试

---

## 为什么建议先做“快照式保存”

当前前端消息结构不只是简单的 `user/assistant` 文本，还包含：

- `thinking`
- `tool` 消息
- `toolArgs`
- `toolResult`
- `inputRequest`
- `pendingAction`
- `actionStatus`

如果一开始就做“增量 append”接口，需要精确处理流式分片、状态更新、工具结果补写，复杂度明显更高。

先做“整会话快照 upsert”有几个好处：

- 前后端收敛快
- 最容易与现有 `ChatMessage` 结构对齐
- 最容易修复“历史按钮为空”的当前问题
- 后续要优化性能，再切增量写入也不晚

---

## 风险与注意事项

### 1. 不要把 `PendingSession` 当历史表复用

它的职责是“待确认动作恢复”，而且有过期时间，不适合作为聊天历史。

### 2. 不要再把业务持久化留在 localStorage

桌面端后续凡是“需要持久化的数据”，都应优先进入 SQLite。

`localStorage` 只适合：

- 临时 UI 偏好
- 一次性迁移标记
- 不重要、可丢失的轻量状态

AI 历史显然不属于这类数据。

### 3. 要考虑消息体积

AI 回复、工具结果、YAML 预览都可能比较长，建议：

- 单条消息内容允许 `text`
- 会话详情接口按需查询
- 历史列表接口只返回摘要，不返回整段消息

### 4. 删除策略要清晰

如果桌面端只面向单用户，本期直接硬删除即可。

### 5. 后续可以再补增强项

本期先不做，但设计上可预留：

- 自动生成标题
- 搜索历史会话
- 导出会话
- 收藏/置顶
- 历史保留天数

---

## 建议实施顺序

1. 后端新增 `ai_chat_sessions` / `ai_chat_messages` 模型与 CRUD。
2. 前端把历史列表、会话详情切到数据库接口。
3. 前端移除历史正式存储的 `localStorage` 写入逻辑。
4. 前端在发送消息和流式结束后执行会话快照保存。
5. 增加 `localStorage` 到数据库的一次性迁移。
6. 补测试并验证：
   - 当前窗口可见历史
   - 新窗口可见历史
   - 应用重启后可见历史
   - 删除后历史面板同步消失

---

## 最终结论

当前 AI 对话历史**没有入库**。

现在的“历史按钮”只是读前端 `localStorage`，不是桌面端正式持久化历史。要彻底解决“历史为空”和“历史不稳定”的问题，应该按本计划补齐 SQLite 会话表、消息表、历史查询接口，并把前端历史能力完全切到数据库。

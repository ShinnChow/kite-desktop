# 桌面版应用数据分析方案与开发计划

> **Goal:** 为 `Kite Desktop` 建立一套适合桌面端场景的数据分析方案，替换当前偏 Web 语义的自动页面统计方式，确保统计结果可用、脱敏、可解释，并与版本检查等其他联网行为明确区分。

> **Scope:** 本文仅讨论桌面版 `Kite Desktop` 的应用数据分析与相关用户设置，不再考虑未来 Web 版本兼容。

> **Current Baseline:** 当前 Umami 通过后端注入脚本接入，统计逻辑集中在 `pkg/utils/utils.go` 与 `internal/server/static.go`。现状具备接入集中、易于切换的优点，但默认自动 pageview 机制对桌面版资源详情路由并不理想，存在 URL 语义过于细粒度、潜在敏感信息暴露面偏大的问题。

> **Related Docs:**
> - [执行 Checklist](./2026-04-18-desktop-analytics-execution-checklist.md)
> - [应用数据分析隐私说明（草案）](../desktop-analytics-privacy-policy.md)
> - [应用数据分析隐私协议（正式版模板）](../desktop-analytics-privacy-notice.md)

---

## 1. 文档定位

本文用于回答三个问题：

1. 桌面版为什么不能继续沿用 Web 风格自动 pageview。
2. 桌面版最终应统计什么，不统计什么。
3. 这套能力应如何拆成可落地的开发任务。

本文是后续编码实施的主计划文档。

---

## 2. 现状结论

### 2.1 当前接入方式

当前应用数据分析通过后端注入 Umami 脚本实现：

- `pkg/utils/utils.go`
- `internal/server/static.go`

运行时开关来自通用设置：

- `pkg/common/common.go`
- `pkg/model/general_setting.go`
- `ui/src/components/settings/general-management.tsx`

### 2.2 当前主要问题

1. 当前模式仍偏向网站 pageview，而不是桌面应用使用分析。
2. 真实路由中含有 `resource / namespace / name`，如果直接统计真实 URL，会带来不必要的数据暴露面。
3. AI sidecar 目前通过 query string 传递 `page`、`namespace`、`resourceName`、`resourceKind`、`sessionId`，若自动采集 URL，会进一步放大敏感信息上送风险。
4. “数据分析”与“版本检查”都是联网行为，但产品语义不同，当前容易被用户混淆。
5. analytics 默认值在后端全局变量、数据库字段 default、前端表单初始值之间不完全一致，后续容易造成认知偏差。

### 2.3 当前仍可保留的优点

1. Umami 接入点集中，迁移成本低。
2. 桌面版已具备显式 runtime 概念，可把 analytics 做成桌面 runtime 的一等能力。
3. 前端已有较清晰的路由与 page context 归一逻辑，可复用为脱敏 page key 模型。

---

## 3. 总体目标

本轮改造后的桌面版数据分析应达到以下状态：

1. 不再自动上报真实 URL。
2. 页面统计只保留脱敏、稳定的页面类型 key。
3. 关键事件只采集对产品判断有价值且不含对象级信息的数据。
4. 应用明确区分：
   - `数据分析`
   - `版本检查`
5. 设置开关、默认值、隐私文案保持一致。

---

## 4. 明确不做

本轮不建议做以下事情：

1. 不做全量行为埋点。
2. 不为每个按钮、每个弹窗、每个 API 请求单独加事件。
3. 不采集对象级资源标识，例如：
   - cluster name
   - namespace
   - resource name
   - AI session id
4. 不采集用户输入内容、Prompt、日志内容、文件路径、下载 URL 等高敏感数据。
5. 不把版本检查结果混入 analytics 事件体系。

---

## 5. 数据分析设计原则

### 5.1 脱敏优先

任何统计上报都不应依赖真实业务对象标识。

允许：

- 页面类型
- 功能入口类型
- 布尔值 / 枚举值 / 稳定分类值

禁止：

- 集群名称
- 命名空间
- 资源名称
- 会话 ID
- 本地目录路径
- 用户输入内容

### 5.2 桌面产品语义优先

桌面版更关心：

- 用户在用哪些模块
- 哪些桌面专属能力被触发
- 更新链路是否被使用
- AI / terminal / 导入等高价值功能是否有使用

而不是：

- 某个真实 URL 被访问了多少次

### 5.3 集中式接入

所有统计上报必须通过统一适配层，不允许页面组件直接调用 Umami 全局对象。

推荐新增：

- `ui/src/lib/analytics.ts`
- `ui/src/lib/analytics-route.ts`

### 5.4 配置与文案一致

用户可见的“启用数据分析”开关、默认值、隐私说明、实际行为必须一致。

---

## 6. 目标统计模型

### 6.1 页面统计

页面统计应从真实路由映射为稳定、脱敏的 page key。

建议 page key 示例：

- `overview`
- `settings`
- `favorites`
- `pods/list`
- `pods/detail`
- `deployments/list`
- `deployments/detail`
- `services/detail`
- `crds/list`
- `crds/detail`
- `ai-chat`

### 6.2 关键事件

建议首批只保留以下事件：

1. `cluster_switch`
2. `ai_chat_open`
3. `kubectl_terminal_open`
4. `kubeconfig_import`
5. `update_check_clicked`
6. `update_download_started`
7. `update_install_started`

建议事件属性只保留非敏感枚举：

- `runtime=desktop`
- `entry=button|shortcut|menu|sidecar`
- `page=overview|pods/detail|settings`
- `mode=file_dialog|text_import`

### 6.3 明确不统计的字段

以下字段不应进入任何 analytics payload：

- `clusterName`
- `namespace`
- `resourceName`
- `resourceKind` 的具体对象实例信息
- `sessionId`
- `prompt`
- `filePath`
- `downloadUrl`

---

## 7. 脚本接入建议

建议将 Umami 脚本收敛为桌面端专用配置，至少包含以下属性：

```html
<script
  defer
  src="https://umami.eryajf.net/script.js"
  data-website-id="8317012e-c8ab-4b59-bc86-2e708ceac202"
  data-auto-track="false"
  data-tag="desktop"
  data-domains="127.0.0.1,localhost"
  data-exclude-search="true"
  data-exclude-hash="true"
  data-do-not-track="true"
></script>
```

设计意图：

1. `data-auto-track="false"`
   - 禁止自动采集真实 URL。
2. `data-tag="desktop"`
   - 让桌面端数据在 Umami 后台具备显式标签。
3. `data-domains="127.0.0.1,localhost"`
   - 明确当前仅统计本地桌面运行态。
4. `data-exclude-search="true"`
   - 为 query string 提供额外兜底。
5. `data-exclude-hash="true"`
   - 禁止 hash 参与统计。
6. `data-do-not-track="true"`
   - 尊重用户的 DNT 偏好。

可选增强：

- 若未来需要更严格控制，可进一步接入 `data-before-send` 做 payload 过滤与白名单处理。

---

## 8. 版本检查与数据分析的边界

版本检查不属于应用数据分析，但同样属于对外联网行为。

当前版本检查相关链路包括：

- `pkg/version/version.go`
- `pkg/version/update_checker.go`
- `ui/src/contexts/runtime-context.tsx`
- `ui/src/hooks/use-desktop-update.ts`

产品层建议：

1. 在设置页将“数据分析”和“版本检查”明确区分为两类能力。
2. 关闭“启用数据分析”后，不应再产生 Umami 请求。
3. 关闭“版本检查”后，应用启动阶段也不应再自动发起 GitHub release 检查。

---

## 9. 默认值策略建议

当前建议统一采用以下桌面端策略：

1. `analytics` 默认关闭
2. `version check` 默认开启

理由：

1. 桌面应用更适合采用保守的默认隐私策略。
2. 更新检查属于产品基本可维护能力，默认开启更符合分发型桌面应用诉求。
3. 两类行为语义明确拆分后，更容易向用户解释。

需要统一的实现位置：

- `pkg/common/common.go`
- `pkg/model/general_setting.go`
- `ui/src/components/settings/general-management.tsx`

---

## 10. 逐文件实施计划

## Task 1: 收敛脚本配置并关闭自动 pageview

**Goal:** 停止自动采集真实 URL，为后续手动上报铺路。

**Files:**
- Modify: `pkg/utils/utils.go`
- Modify: `pkg/utils/utils_test.go`
- Modify: `internal/server/server_test.go`

### 任务说明

- [ ] 把当前 Umami 脚本改成桌面专用配置。
- [ ] 增加 `data-auto-track="false"`。
- [ ] 增加 `data-tag="desktop"`。
- [ ] 增加 `data-domains="127.0.0.1,localhost"`。
- [ ] 恢复 `data-exclude-search="true"`、`data-exclude-hash="true"`、`data-do-not-track="true"`。
- [ ] 更新测试断言，覆盖新属性。

### 验收标准

- [ ] 启动应用后不会因路由切换自动生成真实 URL pageview。
- [ ] 注入脚本属性与桌面方案一致。

---

## Task 2: 新增统一 analytics 适配层

**Goal:** 所有统计调用都经过单一前端封装层。

**Files:**
- Create: `ui/src/lib/analytics.ts`
- Create: `ui/src/lib/analytics-route.ts`
- Test: `ui/src/lib/analytics-route.test.ts`

### 任务说明

- [ ] 在 `analytics.ts` 中封装对 Umami 的访问。
- [ ] 提供 `trackPage()` 与 `trackEvent()` 两个最小接口。
- [ ] 提供安全空实现，确保未注入脚本或 analytics 关闭时不会报错。
- [ ] 在 `analytics-route.ts` 中建立路由到 page key 的映射规则。
- [ ] 对 `ai-chat-box` 做固定 page key 处理，不读取 query 参数。

### 验收标准

- [ ] 页面组件无需直接依赖 `window.umami`。
- [ ] 同一路由模式始终映射到固定 page key。
- [ ] 不会因为 query string 或对象名不同生成新的 page key。

---

## Task 3: 在应用根部接入脱敏 pageview

**Goal:** 用手动 pageview 替代自动 pageview。

**Files:**
- Modify: `ui/src/App.tsx`
- Modify: `ui/src/routes.tsx`
- Test: `ui/src/App.test.tsx` or dedicated analytics bridge tests

### 任务说明

- [ ] 在应用根部监听 React Router 的 location 变化。
- [ ] 将 location 映射为脱敏 page key。
- [ ] 调用 `trackPage(normalizedPageKey)`。
- [ ] 忽略 analytics 关闭状态。
- [ ] 确保 `/ai-chat-box?...` 只记录为 `ai-chat`。

### 验收标准

- [ ] Umami 后台只出现固定 page key，例如 `pods/detail`、`settings`。
- [ ] 不出现 `default`、`prod`、`nginx` 等对象级字符串。

---

## Task 4: 接入首批关键事件

**Goal:** 增加少量高价值、低敏感的数据分析事件。

**Files:**
- Modify: `ui/src/contexts/cluster-context.tsx`
- Modify: `ui/src/contexts/ai-chat-context.tsx`
- Modify: `ui/src/hooks/use-desktop-update.ts`
- Modify: `ui/src/lib/desktop.ts`
- Test: corresponding unit tests

### 任务说明

- [ ] 在 cluster 切换入口增加 `cluster_switch`。
- [ ] 在 AI chat 打开动作增加 `ai_chat_open`。
- [ ] 在 kubectl terminal 打开动作增加 `kubectl_terminal_open`。
- [ ] 在 kubeconfig 导入成功动作增加 `kubeconfig_import`。
- [ ] 在手动检查更新动作增加 `update_check_clicked`。
- [ ] 在开始下载更新动作增加 `update_download_started`。
- [ ] 在开始安装动作增加 `update_install_started`。

### 验收标准

- [ ] 首批事件不超过 7 个。
- [ ] 每个事件 payload 都不含对象级敏感信息。
- [ ] 事件可以回答明确的产品问题。

---

## Task 5: 统一默认值、开关语义与文案

**Goal:** 让 analytics 与 version check 的默认值、产品语义、用户文案保持一致。

**Files:**
- Modify: `pkg/common/common.go`
- Modify: `pkg/model/general_setting.go`
- Modify: `ui/src/components/settings/general-management.tsx`
- Modify: `ui/src/i18n/locales/zh.json`
- Modify: `ui/src/i18n/locales/en.json`
- Doc: `docs/desktop-runtime-contract.md` if user-facing runtime contract wording changes

### 任务说明

- [ ] 统一 `EnableAnalytics` 默认值。
- [ ] 明确 `EnableVersionCheck` 与 analytics 为独立设置。
- [ ] 优化设置页文案，明确“数据分析”和“版本检查”是两类不同联网行为。
- [ ] 若 analytics 开关变更仍需刷新生效，保留当前刷新策略并在文案中说明。

### 验收标准

- [ ] 后端、数据库、前端的默认值一致。
- [ ] 用户能理解关闭 analytics 不等于关闭版本检查。

---

## Task 6: 更新隐私文档与对外说明

**Goal:** 让产品对外说明与实现保持同步。

**Files:**
- Create: `docs/desktop-analytics-privacy-policy.md`
- Modify: `README.md` if later needed

### 任务说明

- [ ] 新增桌面版数据分析隐私说明草案。
- [ ] 明确采集项、不采集项、用户控制方式。
- [ ] 若后续实现与文档不一致，必须同步更新。

### 验收标准

- [ ] 文档可以独立解释 analytics 行为。
- [ ] 文档与最终实现不存在明显冲突。

---

## 11. 推荐实施顺序

建议严格按以下顺序推进：

1. Task 1: 收敛脚本配置
2. Task 2: 新增统一 analytics 适配层
3. Task 3: 接入脱敏 pageview
4. Task 4: 接入首批关键事件
5. Task 5: 统一默认值和文案
6. Task 6: 同步隐私说明

原因：

1. 先止损，先停掉真实 URL 自动上报。
2. 再建立统一入口，防止后续埋点散落。
3. 再补少量高价值事件，避免一开始过度设计。

---

## 12. 最终验收口径

整套方案完成后，应满足以下验收标准：

1. Umami 后台不再出现真实 namespace / resource name / sessionId。
2. 页面统计只出现稳定 page key。
3. analytics 关闭后，应用不再产生 Umami 请求。
4. version check 关闭后，应用不再自动访问 GitHub release API。
5. analytics 默认值在全栈范围一致。
6. 对外隐私文档与实际实现一致。

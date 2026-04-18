# 桌面版应用数据分析执行 Checklist

> **Goal:** 将 [2026-04-18-desktop-analytics-implementation-plan.md](./2026-04-18-desktop-analytics-implementation-plan.md) 收敛为一份可直接排期、执行、验收的开发清单。

> **Usage:** 本文偏执行，不重复解释完整背景。设计原因、原则和边界请参考同目录下的 implementation plan。

> **Related Docs:**
> - [方案与开发计划](./2026-04-18-desktop-analytics-implementation-plan.md)
> - [应用数据分析隐私说明（草案）](../desktop-analytics-privacy-policy.md)
> - [应用数据分析隐私协议（正式版模板）](../desktop-analytics-privacy-notice.md)

---

## 0. 执行前确认

- [ ] 当前项目后续仅保留桌面版，不再兼容 Web 版 analytics 语义。
- [ ] 认可总体方向：
  - [ ] 关闭自动 pageview
  - [ ] 改为手动脱敏 pageview
  - [ ] 只保留少量关键事件
  - [ ] analytics 与 version check 独立开关
- [ ] 认可默认策略：
  - [ ] analytics 默认关闭
  - [ ] version check 默认开启
- [ ] 认可以下字段禁止进入 analytics：
  - [ ] cluster name
  - [ ] namespace
  - [ ] resource name
  - [ ] AI session id
  - [ ] 用户输入内容
  - [ ] 本地路径与下载 URL

---

## 1. 脚本层止损

### 目标

先停掉真实 URL 自动上报风险。

### 文件

- `pkg/utils/utils.go`
- `pkg/utils/utils_test.go`
- `internal/server/server_test.go`

### Checklist

- [ ] 将 Umami 脚本改成桌面专用配置。
- [ ] 加入 `data-auto-track="false"`。
- [ ] 加入 `data-tag="desktop"`。
- [ ] 加入 `data-domains="127.0.0.1,localhost"`。
- [ ] 加入 `data-exclude-search="true"`。
- [ ] 加入 `data-exclude-hash="true"`。
- [ ] 加入 `data-do-not-track="true"`。
- [ ] 若采用 `data-before-send`，补对应行为说明与测试。
- [ ] 更新注入测试断言。

### 验收

- [ ] 启动应用后，路由切换不再自动生成真实 URL pageview。
- [ ] 脚本属性与桌面版方案一致。

---

## 2. 统一 analytics 适配层

### 目标

所有统计调用都必须通过单一前端封装。

### 文件

- `ui/src/lib/analytics.ts`
- `ui/src/lib/analytics-route.ts`
- `ui/src/lib/analytics-route.test.ts`

### Checklist

- [ ] 新增 `trackPage(pageKey)`。
- [ ] 新增 `trackEvent(name, data)`。
- [ ] 提供 `isAnalyticsEnabled()` 或等价安全判断。
- [ ] analytics 关闭时提供安全空实现。
- [ ] 未注入 Umami 时不抛异常。
- [ ] 路由映射器只根据 route pattern 产出 page key。
- [ ] `/ai-chat-box` 固定映射为 `ai-chat`。
- [ ] query string 不参与 page key 生成。

### 验收

- [ ] 页面组件中不再直接调用 `window.umami`。
- [ ] 同一路由模式总是映射到同一个 page key。

---

## 3. 接入手动脱敏 pageview

### 目标

用手动 pageview 替代自动 pageview。

### 文件

- `ui/src/App.tsx`
- `ui/src/routes.tsx`
- `ui/src/App.test.tsx` or dedicated analytics tests

### Checklist

- [ ] 在应用根部监听 `location` 变化。
- [ ] 将真实路由映射为脱敏 page key。
- [ ] 调用 `trackPage()`。
- [ ] 首次启动加载时记录一次 pageview。
- [ ] 后续路由切换时按 page key 去重或按需要重复记录。
- [ ] `settings`、`favorites`、`overview` 等固定页面映射清晰。
- [ ] `:resource` 映射为 `<resource>/list`。
- [ ] `:resource/:namespace/:name` 映射为 `<resource>/detail`。
- [ ] `crds` 单独映射。

### 验收

- [ ] Umami 中只出现 `overview`、`settings`、`pods/list`、`pods/detail` 等固定值。
- [ ] 不出现 `default`、`prod`、`nginx` 等对象级字符串。

---

## 4. 接入首批关键事件

### 目标

补少量真正有用的功能事件。

### 文件

- `ui/src/contexts/cluster-context.tsx`
- `ui/src/contexts/ai-chat-context.tsx`
- `ui/src/hooks/use-desktop-update.ts`
- `ui/src/lib/desktop.ts`
- 对应测试文件

### 首批事件列表

- [ ] `cluster_switch`
- [ ] `ai_chat_open`
- [ ] `kubectl_terminal_open`
- [ ] `kubeconfig_import`
- [ ] `update_check_clicked`
- [ ] `update_download_started`
- [ ] `update_install_started`

### Checklist

- [ ] `cluster_switch` 不带 cluster name。
- [ ] `ai_chat_open` 只带入口与页面类型。
- [ ] `kubectl_terminal_open` 只带入口类型。
- [ ] `kubeconfig_import` 只带导入方式。
- [ ] 更新相关事件不带下载 URL。
- [ ] 所有事件默认增加 `runtime=desktop`。
- [ ] 事件属性仅使用稳定枚举值或布尔值。

### 验收

- [ ] 首批事件数量控制在 7 个以内。
- [ ] 任一事件 payload 中都不含对象级敏感字段。

---

## 5. 收敛 AI sidecar URL 风险

### 目标

确保 AI sidecar 不因 URL 被统计而泄露上下文。

### 文件

- `desktop/bridge.go`
- `ui/src/contexts/ai-chat-context.tsx`
- `ui/src/components/ai-chat/ai-chatbox.tsx`

### Checklist

- [ ] 最低要求：analytics 侧完全忽略 AI sidecar query 参数。
- [ ] 检查 `sessionId` 是否仍必须出现在 URL 中。
- [ ] 如果可以，评估后续将 sidecar 上下文迁移到本地状态而不是 query。
- [ ] 即便暂时保留 query，也要确保 analytics 不会读取该值。

### 验收

- [ ] Umami 后台不会出现 `sessionId=`。
- [ ] Umami 后台不会出现 AI sidecar 的 `namespace` / `resourceName` 查询参数。

---

## 6. 统一默认值与设置语义

### 目标

让默认值、设置项、真实行为完全一致。

### 文件

- `pkg/common/common.go`
- `pkg/model/general_setting.go`
- `ui/src/components/settings/general-management.tsx`
- `ui/src/i18n/locales/zh.json`
- `ui/src/i18n/locales/en.json`

### Checklist

- [ ] 后端默认 `EnableAnalytics` 与产品策略一致。
- [ ] 数据库字段 default 与产品策略一致。
- [ ] 前端表单初始值与产品策略一致。
- [ ] analytics 开关变更后行为立即生效。
- [ ] 设置文案中清晰区分：
  - [ ] 数据分析
  - [ ] 版本检查
- [ ] 如有必要，在 UI 中增加一句简短说明：
  - [ ] “关闭数据分析不会关闭版本检查”

### 验收

- [ ] 后端、数据库、前端默认值一致。
- [ ] 用户不会误解两个开关的含义。

---

## 7. 版本检查边界核对

### 目标

确保版本检查不被误当成 analytics。

### 文件

- `pkg/version/version.go`
- `pkg/version/update_checker.go`
- `ui/src/contexts/runtime-context.tsx`
- `ui/src/hooks/use-desktop-update.ts`

### Checklist

- [ ] 明确启动时是否自动静默检查更新。
- [ ] 明确关闭 version check 后，启动阶段不再访问 GitHub API。
- [ ] About 页面文案与实际行为一致。
- [ ] 不为版本检查结果额外发送 analytics 事件，除非是用户手动触发的功能动作。

### 验收

- [ ] analytics 关闭后，不再发送 Umami 请求。
- [ ] version check 关闭后，不再自动访问 GitHub Releases API。

---

## 8. 测试与验证

### 必做验证

- [ ] Go 测试覆盖新的脚本注入属性。
- [ ] 前端测试覆盖 route -> page key 映射。
- [ ] 前端测试覆盖 analytics 开关关闭后的安全空实现。
- [ ] 前端测试覆盖 AI sidecar page key 不读取 query。
- [ ] 前端测试覆盖 analytics 开关切换后的立即生效行为。

### 手工验证

- [ ] 打开概览页，观察 Umami 是否只出现 `overview`。
- [ ] 打开任意 Pod 详情页，观察是否只出现 `pods/detail`。
- [ ] 打开 AI sidecar，观察是否只出现 `ai-chat`。
- [ ] 关闭 analytics 后，确认不再出现 Umami 请求。
- [ ] 关闭 version check 后，重启应用确认不再自动访问 GitHub API。

---

## 9. 文档同步

### 文件

- `docs/plans/2026-04-18-desktop-analytics-implementation-plan.md`
- `docs/desktop-analytics-privacy-policy.md`
- `docs/desktop-analytics-privacy-notice.md`
- 如有必要：`README.md`

### Checklist

- [ ] 若实现范围变化，更新 implementation plan。
- [ ] 若采集项变化，更新隐私说明与正式隐私协议。
- [ ] 若默认值变化，更新文档中的默认策略描述。
- [ ] 若 analytics 与 version check 的产品语义变化，更新用户文案。

---

## 10. 里程碑建议

### Milestone A: 风险止损

- [ ] 完成脚本层调整
- [ ] 完成 route mapping
- [ ] 完成手动脱敏 pageview

### Milestone B: 数据可用

- [ ] 完成首批关键事件
- [ ] 完成 AI sidecar 风险收敛

### Milestone C: 产品闭环

- [ ] 完成默认值统一
- [ ] 完成设置文案收敛
- [ ] 完成隐私文档同步

---

## 11. 发布前最终检查

- [ ] Umami 后台已无真实 namespace / resource name / sessionId。
- [ ] page key 仅为脱敏稳定值。
- [ ] analytics 开关关闭后完全停报。
- [ ] version check 开关关闭后停止自动检查更新。
- [ ] 对外文案与实现一致。
- [ ] 团队已确认首批事件不会继续扩散成无边界埋点。

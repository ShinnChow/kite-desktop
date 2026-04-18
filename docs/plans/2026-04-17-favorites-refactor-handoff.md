# 收藏功能改造交接记录

日期：2026-04-17

## 1. 背景

当前仓库已经是桌面优先的 `Kite Desktop`，不再适合继续把“收藏”作为纯前端 `localStorage` 功能。

最初分析结论：

- 旧实现完全依赖 `localStorage`
- 收藏主键使用搜索结果里的 `id`
- 搜索结果里的 `id` 实际上是 Kubernetes 资源 `UID`
- `UID` 不稳定，资源重建后收藏会失效
- 桌面版已经有 SQLite、本地桌面用户和偏好接口模式，因此收藏更适合作为桌面本地偏好写入 DB

## 2. 当前已完成的改造

### 2.1 后端

已新增收藏相关模型与接口：

- 新增模型：
  - `pkg/model/favorite_resource.go`
- 已注册到 AutoMigrate：
  - `pkg/model/model.go`
- 新增 handler：
  - `pkg/handlers/favorite_handler.go`
- 已接入路由：
  - `internal/server/routes.go`
- 已更新路由测试：
  - `internal/server/routes_test.go`
- 已新增模型测试：
  - `pkg/model/favorite_resource_test.go`

当前后端设计要点：

- 收藏以独立表存储，而不是塞进 `User` JSON 字段
- 收藏按以下自然键唯一：
  - `user_id`
  - `cluster_name`
  - `resource_type`
  - `namespace`
  - `resource_name`
- 不再使用资源 `UID` 作为收藏身份

当前后端接口：

- `GET /api/v1/preferences/favorites`
- `POST /api/v1/preferences/favorites`
- `POST /api/v1/preferences/favorites/remove`

### 2.2 前端

已完成前端主逻辑改造：

- 收藏 API 已加入：
  - `ui/src/lib/api/core.ts`
- 收藏 hook 已改为 React Query + 后端接口：
  - `ui/src/hooks/use-favorites.ts`
- 搜索弹窗星标逻辑已改为按资源自然键判断：
  - `ui/src/components/global-search.tsx`
- 旧 `ui/src/lib/favorites.ts` 已改为工具函数集合，不再承担主存储职责

当前前端设计要点：

- 收藏主数据源已经切换为后端接口
- 前端 `isFavorite` 判断基于：
  - `resourceType`
  - `namespace`
  - `resourceName`
- 不再依赖搜索结果 `id`

### 2.3 收藏页

已新增独立收藏页：

- 路由：`/favorites`
- 位置：左侧导航 `集群` 分组末尾
- 范围：只展示当前集群收藏

当前页面能力：

- 展示当前集群收藏列表
- 当前集群切换后页面内容同步切换
- 支持本地搜索
- 支持按资源类型筛选
- 支持按命名空间筛选
- 支持点击进入资源详情
- 支持在页面内取消收藏

## 3. 已经确认通过的验证

已执行并通过：

- `go test ./pkg/model ./internal/server ./pkg/handlers`
- `go test ./...`
- `pnpm --dir ui test -- --run src/lib/favorites.test.ts src/hooks/use-favorites.test.tsx src/components/global-search.test.tsx`
- `pnpm --dir ui exec tsc --noEmit`

## 4. 遇到过的问题

### 4.1 曾经尝试过兼容旧 `localStorage` 收藏，但已不再保留

之前一度实现过“自动把旧 `localStorage` 收藏迁移到 SQLite”的逻辑。

问题表现：

- `make dev` 后页面进入错误页
- 用户反馈页面“所有按钮都点不动”
- 截图中出现：
  - `Minified React error #185`

这说明当时不是普通交互失效，而是前端已经进入 React 运行时异常状态。

### 4.2 当前判断

结合现象和改动范围，最可疑的是：

- 收藏 hook 里的自动迁移逻辑
- 自动迁移过程中，query + mutation + 状态更新形成了不稳定的渲染链
- 即使测试没稳定复现，运行时在 Wails dev 环境里仍可能触发异常

## 5. 当前采取的处理

考虑到当前项目还处于初期阶段、历史用户数据几乎可以忽略，当前版本已经做了以下处理：

- 保留 SQLite 作为收藏主存储
- 已完全移除旧 `localStorage` 收藏迁移逻辑
- 已删除仅用于迁移的 `/preferences/favorites/import` 接口
- `ui/src/lib/favorites.ts` 只保留自然键相关工具函数

也就是说，当前状态是：

- 收藏读写已经走 SQLite
- 不再尝试读取、迁移或兼容旧 `localStorage` 收藏数据
- 收藏主流程中已经没有 legacy `localStorage` 依赖

## 6. 当前已知的未完成项

### 6.1 旧收藏兼容路径已移除

当前不再保留这些旧收藏迁移能力：

- 旧 `localStorage` 收藏自动迁移
- migration marker
- 导入接口 `/api/v1/preferences/favorites/import`

这是一个有意的产品取舍，而不是漏做：

1. 当前项目还处于初期
2. 旧收藏历史包袱很小
3. 简化主流程比兼容极少量历史数据更重要

### 6.2 仍然挂靠在本地桌面用户模型上

当前收藏表仍然通过 `user_id` 关联本地桌面用户。

这和当前系统整体状态一致，因为：

- 侧边栏偏好仍挂在本地桌面用户上
- 资源历史也仍使用 `OperatorID -> User`

如果未来要推进“去用户化”改造，收藏这块后续可能还要一起迁出。

## 7. 建议的下一步

如果下次继续做收藏功能，建议按这个顺序推进：

1. 手动验证以下交互：
   - 搜索弹窗打开
   - 点击星标收藏
   - 切换集群后收藏隔离
   - 收藏页筛选与跳转
2. 如果未来推进“去用户化”改造，再评估是否把收藏从本地桌面用户模型中迁出
3. 如果后续收藏数量继续增大，可再评估：
   - 搜索弹窗收藏区展示数量上限
   - 收藏页分组或排序增强

## 8. 调试补充

以后排查前端运行时错误时，可优先尝试打开桌面开发控制台：

- `Cmd + Option + I`
- `Cmd + Option + J`
- `F12`

如果快捷键无效，可直接查看 `make dev` 的终端输出。

## 9. 当前工作树中的相关文件

与本次收藏改造直接相关的主要文件：

- `pkg/model/favorite_resource.go`
- `pkg/model/favorite_resource_test.go`
- `pkg/handlers/favorite_handler.go`
- `pkg/model/model.go`
- `internal/server/routes.go`
- `internal/server/routes_test.go`
- `ui/src/lib/api/core.ts`
- `ui/src/hooks/use-favorites.ts`
- `ui/src/hooks/use-favorites.test.tsx`
- `ui/src/components/global-search.tsx`
- `ui/src/pages/favorites.tsx`
- `ui/src/pages/favorites.test.tsx`
- `ui/src/lib/favorites.ts`
- `ui/src/routes.tsx`
- `ui/src/contexts/sidebar-config-context.tsx`

与本次并行整理但不直接属于收藏逻辑的文档类变更：

- `AGENTS.md`
- `.codex/README.md`
- `.codex/project-context.md`
- `.codex/development-guide.md`
- `docs/plans/2026-04-18-favorites-page-design.md`

## 10. 当前状态总结

当前收藏改造已经进入“SQLite 主存储、无旧收藏兼容路径”的状态。

目前已落地的内容包括：

- 收藏主键从资源 `UID` 切到自然键
- 收藏主存储切换到 SQLite
- 前后端接口全部接通
- 搜索弹窗星标逻辑切到自然键判断
- 当前集群收藏页已落地
- 旧 `localStorage` 收藏逻辑已从主流程移除
- handler 请求语义、主流程测试都已补齐

当前剩余的主要工作已经不是旧逻辑兼容，而是继续围绕桌面端真实使用场景完善收藏体验本身。

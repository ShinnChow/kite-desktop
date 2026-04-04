# Kite 桌面化改造方案

## 目标

在尽量少改现有架构的前提下，把 Kite 改造成一个可用的桌面端版本。

本次方案的核心目标：

- 保留现有 React UI
- 保留现有 Gin API
- 保留现有日志/终端 WebSocket 能力
- 保留现有集群管理与初始化流程

第一阶段的明确非目标：

- 不把系统重写成纯 Wails bindings/events 架构
- 不移除现有 Web/Server 交付模式
- 不对认证模型做超出桌面运行所需的重设计

## 结论摘要

Kite 很适合做桌面化，原因很直接：

- 后端本来就是独立的 Go + Gin 服务
- 前端本来就是独立的 React + Vite SPA
- 前端主要通过相对路径访问 API
- 静态资源已经有标准构建产物

推荐路线：

1. 把当前服务启动逻辑抽成可复用包
2. 保留现有 Gin HTTP 与 WebSocket API
3. 新增 `cmd/desktop`，由 Wails v3 承载桌面壳
4. 桌面模式下在本机 `127.0.0.1:<随机端口>` 启动当前 Gin 服务
5. Wails 主窗口直接加载这个本地地址
6. 只补一层很薄的桌面能力适配层，处理外链、文件选择、路径等原生行为

这条路线的优点是：

- 可以最大化复用现有代码
- 可以绕开第一阶段最昂贵的重写成本
- 能最快做出一个真实可用的桌面版

## 为什么这个项目适合桌面化

当前仓库已经具备比较理想的分层：

- 后端初始化和 Gin 组装集中在 `app.go`
- API 路由集中在 `routes.go`
- 静态资源托管集中在 `static.go`
- 前端是独立的 `ui/` 工程
- Vite 构建后直接输出到 `static/`

相关代码位置：

- `app.go`
- `main.go`
- `routes.go`
- `static.go`
- `ui/vite.config.ts`
- `ui/src/lib/api-client.ts`
- `ui/src/lib/subpath.ts`

这比把一个强耦合的 SSR 项目改成桌面端要轻松很多。

## 现实约束

### Wails v3 的状态

Wails v3 能用，但当前仍然不是完全稳定版。

这意味着：

- 框架本身存在一定不确定性
- 一些 API 或构建行为未来仍可能变化
- 桌面打包和平台兼容性需要单独验证

因此结论是：

- 如果目标是先做出可用桌面版，Wails v3 可以接受
- 如果目标是特别保守的长期稳定桌面产品，需要把框架成熟度风险计入计划

## 当前架构概览

### 后端

当前后端负责：

- 加载环境变量与初始化数据库
- 初始化 RBAC、模板等系统能力
- 当没有集群配置时自动从本地 kubeconfig 导入
- 提供 HTTP JSON API
- 提供日志/终端所需的 WebSocket 接口
- 提供 React SPA 与静态资源托管

当前后端的重要特征：

- 认证依赖 Cookie
- OAuth 回调依赖 HTTP 路由
- 日志和终端依赖 WebSocket
- ClusterManager、Prometheus Client 等都在进程内持有

### 前端

当前前端的主要运行假设：

- 通过 `/api/...` 相对路径访问后端
- WebSocket 地址从 `window.location` 推导
- OAuth 登录通过浏览器跳转完成
- 部分功能使用 `window.open`、`target="_blank"`、`iframe`
- 初始化导入 kubeconfig 时使用浏览器文件选择

这决定了桌面化改造的工作量主要不在页面本身，而在运行时假设的适配。

## 推荐目标架构

### 核心原则

第一阶段保留现有 HTTP + WebSocket 模型。

不建议第一阶段直接改成：

- 每个接口都改走 Wails bindings
- 日志和终端改走 Wails events
- 前端完全假设不存在本地 HTTP 服务

### 推荐运行形态

桌面端运行时建议如下：

- Wails v3 负责桌面壳、窗口、菜单、对话框、系统集成、打包
- Kite 后端继续作为进程内本地服务运行
- React 前端继续作为主 UI

推荐的生产运行流程：

1. Wails 启动
2. 桌面启动逻辑解析应用数据目录
3. 设置桌面模式所需的配置和环境变量
4. Gin 服务绑定到 `127.0.0.1:<随机端口>`
5. Wails 主窗口打开 `http://127.0.0.1:<端口>`
6. 前端继续按原方式访问 HTTP 与 WebSocket

### 为什么第一阶段最适合走 loopback HTTP

优点：

- 现有 Cookie 认证可以继续使用
- 现有 WebSocket 基本可以直接复用
- 代理页面、文件预览下载、iframe 等现有行为更容易保留
- 前端改动最少
- 后端 handler、中间件、路由都能基本不动

代价：

- 桌面应用运行时会占用一个本地端口
- 它不是完全“无本地服务”的纯 WebView 模式

这个代价是可接受的，而且对第一阶段来说是正确的取舍。

## 推荐的代码结构调整

### 当前问题

当前启动逻辑放在根目录 `main` 包里，无法优雅地被桌面入口复用。

### 建议的目录结构

建议拆成：

```text
cmd/
  server/
    main.go
  desktop/
    main.go
internal/
  server/
    app.go
    config.go
    runtime.go
    static.go
```

职责建议：

- `internal/server/app.go`
  负责初始化应用、构建 Gin Engine

- `internal/server/runtime.go`
  负责启动、停止、选择监听端口、暴露 BaseURL

- `cmd/server/main.go`
  保留当前独立服务模式

- `cmd/desktop/main.go`
  启动 Wails，再启动本地服务，再打开窗口

### 推荐抽象

建议最终形成类似这样的接口：

```go
type Runtime struct {
    ClusterManager *cluster.ClusterManager
    Engine         *gin.Engine
    Server         *http.Server
    BaseURL        string
}

func NewRuntime(opts Options) (*Runtime, error)
func (r *Runtime) Start() error
func (r *Runtime) Shutdown(ctx context.Context) error
```

## 桌面端运行时设计

### 数据目录

桌面模式不应该继续使用项目根目录下类似 `dev.db` 这样的开发路径。

建议的桌面数据目录：

- macOS: `~/Library/Application Support/Kite/`
- Linux: `~/.local/share/Kite/`
- Windows: `%AppData%/Kite/`

建议放置的文件：

- `kite.db`
- `logs/`
- `cache/`
- `tmp/`

桌面启动时建议主动设置：

- `DB_DSN=<app-data-dir>/kite.db`
- 必要时设置 `KUBECONFIG=<默认 kubeconfig 路径>`

### 端口策略

建议使用随机空闲的 loopback 端口。

要求：

- 只监听 `127.0.0.1`
- 不监听 `0.0.0.0`
- 启动后把 BaseURL 明确传给前端

### 日志

建议桌面端把运行日志写入应用数据目录，方便调试和排障。

## 前端需要做的适配

前端不需要重写，但需要补一个很薄的运行时适配层。

### 1. 增加桌面运行时适配模块

建议新增：

- `ui/src/lib/runtime.ts`
- `ui/src/lib/desktop.ts`

职责包括：

- 判断当前是否运行在桌面模式
- 获取后端 BaseURL
- 统一处理外链打开方式
- 统一处理文件选择等原生能力
- 按需暴露应用路径信息

### 2. API Base URL 处理

如果 Wails 直接打开本地 loopback URL，那么当前相对路径机制多数情况下还能继续工作。

但仍建议补一个显式运行时对象，例如：

- `window.__KITE_RUNTIME__`

建议字段：

- `mode: "web" | "desktop"`
- `apiBaseUrl`
- `wsBaseUrl`

这样可以减少隐藏耦合，后面也更容易演进。

### 3. 外链与新窗口

当前代码里存在：

- `window.open(...)`
- `<a target="_blank">`

桌面模式下建议明确分流：

- 文档、GitHub、官网等外链，交给系统默认浏览器
- 集群代理类页面，第一阶段可以继续走浏览器
- 第二阶段再考虑是否改成桌面内多窗口

### 4. 文件选择

初始化导入 kubeconfig 当前使用浏览器文件选择。

第一阶段可以有两个方案：

- 保留现有 `<input type="file">`，先验证 Wails WebView 表现
- 更推荐直接加一层原生文件对话框封装

桌面体验更好的做法是：

- 通过 Wails 原生文件对话框选择 kubeconfig
- 在 Go 端读取文件内容
- 再交给现有导入逻辑

### 5. 下载与预览

当前文件下载和预览使用的是浏览器打开新窗口。

桌面第一阶段建议：

- 预览仍走浏览器或本地 URL 打开
- 下载先复用现有能力

第二阶段可以增强为：

- 使用原生保存对话框
- 让用户明确选择保存路径

## 认证方案

认证是桌面化里第二关键的设计点。

### 密码登录与 LDAP 登录

这两类登录方式基本不需要大改，因为它们本身就是普通 HTTP 请求加 Cookie。

### OAuth 登录

第一阶段推荐方案：

- 保留现有 `/api/auth/callback` 回调机制
- 用户点击 OAuth 登录时，通过系统默认浏览器打开授权链接
- 第三方回调到 `http://127.0.0.1:<port>/api/auth/callback`
- 本地 Gin 服务完成登录并写入 Cookie
- 桌面应用再刷新登录态

这个方案的优点：

- 最大化复用当前登录逻辑
- 不需要第一阶段就做自定义协议
- 和原生应用常见的 loopback callback 模式兼容

这要求桌面模式下：

- 后端必须知道当前精确的回调地址
- 建议显式设置：
  - `HOST=http://127.0.0.1:<port>`

### 第二阶段可选增强

后续如果要进一步原生化，可以考虑：

- 自定义协议，例如 `kite://auth/callback`

但这不是第一阶段必须项。

## 集群导入与初始化流程

当前代码已经支持：

- 当数据库里没有集群时，从本地 kubeconfig 自动发现并导入

这对桌面版是一个明显优势。

桌面模式建议：

- 首次启动继续保留自动导入逻辑
- 初始化页面继续保留，用于手动补充和修正
- 增加“原生选择 kubeconfig 文件”作为体验增强

可以进一步优化的点：

- 如果本地没有 kubeconfig，启动时给出更明确的提示

## 打包与发布

### 第一阶段打包目标

目标平台建议先覆盖：

- macOS
- Windows
- Linux

需要补的内容：

- Wails v3 构建配置
- 应用图标
- 应用元信息
- 桌面构建命令

### 发布策略建议

建议先分两步：

1. 先做可运行的内部构建包
2. 再做签名、安装器、完整发布流程

不要把“第一版能跑”和“第一版能正式发行”绑在一起。

### CI 改造建议

当前 CI 主要是面向 Web/Server 形态。

桌面版后续应新增：

- `ui` 构建
- `cmd/desktop` 构建
- 多平台打包
- 基础 smoke test

但第一阶段不建议把完整桌面 CI 作为阻塞项。

## 测试策略

### 后端回归

保留现有 Go 测试体系。

### 前端回归

保留现有 Vitest 体系。

### 桌面端 smoke test

建议最少覆盖：

- 应用可以启动
- 初始化页可以加载
- 本地 kubeconfig 可以自动导入
- 密码登录可用
- 集群列表可用
- 资源列表可用
- 日志 WebSocket 可用
- 终端 WebSocket 可用
- 至少一个 OAuth Provider 可用
- 外链打开行为正确
- 下载/预览不会卡死

### 桌面端 E2E

如果后续桌面版成为主要交付形态，再补专门的桌面端 E2E。第一阶段不建议把它放到关键路径上。

## 分阶段实施方案

## Phase 0：技术验证

目标：

- 证明 Wails 壳可以拉起 Kite 的本地服务并打开页面

任务：

- 创建一个最小 Wails v3 原型
- 在桌面进程内启动 Gin 服务
- 让主窗口打开本地 URL

产出：

- 可以启动的验证版桌面壳

预估工时：

- 1 到 2 天

## Phase 1：服务启动逻辑抽离

目标：

- 让当前服务端既能独立运行，也能被桌面入口复用

任务：

- 把启动逻辑从根目录 `main` 中抽出
- 引入可复用的 server runtime 包
- 保持原有 Web/Server 构建方式不变
- 保持现有 API 和静态资源行为不变

产出：

- `cmd/server` 可正常工作
- 现有 Web 版功能不回退

预估工时：

- 2 到 4 天

## Phase 2：第一个可用桌面版

目标：

- 产出一个团队内部可用的桌面版本

任务：

- 新增 `cmd/desktop`
- 桌面模式启动 loopback 后端
- 主窗口加载本地 URL
- 接入桌面数据目录与数据库路径
- 注入运行时配置
- 修正外链、登录、基础文件选择等问题
- 验证初始化、登录、集群、日志、终端

产出：

- 能给开发者或内部测试人员使用的桌面包

预估工时：

- 4 到 7 天

## Phase 3：桌面体验补强

目标：

- 解决明显的“浏览器味道过重”的交互问题

任务：

- 加原生文件选择
- 优化下载和导出
- 增加错误提示与桌面菜单
- 可选增加打开数据目录、打开日志目录等能力

产出：

- 内测质量的桌面体验

预估工时：

- 4 到 8 天

## Phase 4：打包与发布流程

目标：

- 形成可重复的桌面构建和打包流程

任务：

- 增加 Wails 打包配置
- 增加图标和元信息
- 增加发布命令
- 增加桌面构建 CI
- 如有需要，加入签名和 notarization 方案

产出：

- 可重复生成的桌面安装包或分发包

预估工时：

- 4 到 10 天

## 成本预估

如果目标是“先做出一个可用桌面版”：

- 预估人力：2 到 4 周
- 风险等级：中等
- 重写比例：低到中

如果目标是“做成一个较成熟的跨平台桌面产品”：

- 预估人力：4 到 8 周
- 风险等级：中到中高

不确定性的主要来源：

- Wails v3 本身的成熟度
- OAuth 在桌面端的兼容性
- 多平台打包与签名
- 下载、预览、外链、代理页面等边角行为

## 主要风险

### 1. Wails v3 成熟度风险

影响：

- API 或构建行为可能变化
- 各平台表现不完全一致

缓解：

- 第一阶段尽量少做 Wails 专属抽象
- 架构上保守，优先复用现有服务模型

### 2. OAuth 桌面登录风险

影响：

- 登录链路可能成为第一个实质性阻塞点

缓解：

- 第一阶段采用 loopback callback
- 显式设置 `HOST`
- 尽早拿一个真实 Provider 联调

### 3. 浏览器行为迁移风险

影响：

- 新窗口、下载、预览、iframe 等行为可能体验不一致

缓解：

- 尽早补统一运行时封装
- 优先修外链和文件导入导出

### 4. 打包链路复杂度

影响：

- 最后的发布环节容易超出预期

缓解：

- 把“能用”与“能正式发”拆成两个里程碑

## 第一阶段不要重写的部分

以下能力第一阶段应尽量保持原样：

- Gin 路由体系
- 现有 REST handlers
- 现有 WebSocket handlers
- React 路由结构
- React Query 数据访问模型
- 初始化流程
- ClusterManager 主体逻辑

这是控制成本的关键原则。

## 第一阶段的明确交付物

在“第一个可用桌面版”阶段，仓库里应至少新增：

- 可复用的 server bootstrap 包
- `cmd/server`
- `cmd/desktop`
- Wails v3 配置
- 桌面图标资源
- 前端桌面运行时适配层
- 桌面数据目录处理逻辑
- 桌面端构建与启动说明

## 推荐的下一步动作

建议按这个顺序推进：

1. 先抽服务启动逻辑
2. 再加桌面入口
3. 先让主界面跑起来
4. 再修认证、外链、文件选择这些边缘能力
5. 最后再做打包发布

这个顺序的好处是：

- 最快看到真实结果
- 最容易控制风险
- 最不容易把项目带入重写泥潭

## 本方案参考的 Wails 官方资料

- 状态页：`https://v3alpha.wails.io/status/`
- Gin Routing：`https://v3alpha.wails.io/guides/gin-routing/`
- Gin Services：`https://v3alpha.wails.io/guides/gin-services/`
- Browser Integration：`https://v3alpha.wails.io/features/browser/integration/`
- Window Options：`https://v3alpha.wails.io/features/windows/options/`
- File Dialogs：`https://v3alpha.wails.io/features/dialogs/file/`
- Custom URL Protocols：`https://v3alpha.wails.io/guides/distribution/custom-protocols/`
- Installation：`https://v3alpha.wails.io/getting-started/installation/`

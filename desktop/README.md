# Kite Desktop

这是基于 Wails v3 的桌面端子工程。

当前实现状态：

- 已接入现有 Kite 后端
- 启动时会在本机 `127.0.0.1` 随机端口拉起内嵌 Gin 服务
- Wails 主窗口直接加载本地 Kite Web UI
- 默认数据目录会落到用户配置目录下的 `Kite`

## 本地运行

推荐直接从仓库根目录执行：

```bash
make desktop-dev
```

如果 `wails3` 不在 `PATH` 中，可以手动指定：

```bash
make desktop-dev WAILS3=/path/to/wails3
```

或者在 `desktop/` 目录里直接运行：

```bash
cd desktop
wails3 dev -config ./build/config.yml
```

## 构建

推荐：

```bash
make desktop-build
```

如果需要手动指定 `wails3`：

```bash
make desktop-build WAILS3=/path/to/wails3
```

或者在 `desktop/` 目录里直接运行：

```bash
cd desktop
wails3 build
```

构建产物默认在：

- `desktop/bin/kite`

## 数据目录

桌面模式默认会使用系统用户配置目录：

- macOS: `~/Library/Application Support/Kite/`
- Linux: `~/.config/Kite/` 或系统对应 `UserConfigDir`
- Windows: `%AppData%/Kite/`

默认 SQLite 路径：

- `Kite/kite.db`

如果你显式设置了 `DB_DSN`，则优先使用外部配置。

## 当前状态

- 当前窗口内容仍然是现有 Web UI，不是重写后的原生桌面界面
- 外链和 `target="_blank"` / `window.open` 已通过桌面宿主统一接管
- 初始化页面的 kubeconfig 导入已支持原生文件选择器
- OAuth 虽然已经具备桌面端 loopback 回调基础，但还没有做完整联调

## 下一步

后续会继续补：

- 更完整的原生文件选择与下载保存
- OAuth 桌面登录联调
- 更完整的桌面菜单与应用元信息

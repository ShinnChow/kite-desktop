<div align="center">

<img src="./docs/src/banner.svg" alt="Kite Desktop Banner" width="900">

<h1>Kite Desktop</h1>

[![Auth](https://img.shields.io/badge/Auth-eryajf-ff69b4)](https://github.com/eryajf)
[![Go Version](https://img.shields.io/github/go-mod/go-version/eryajf/kite-desktop)](https://github.com/eryajf/kite-desktop)
[![Gin Version](https://img.shields.io/badge/Gin-1.6.3-brightgreen)](https://github.com/eryajf/kite-desktop)
[![Gorm Version](https://img.shields.io/badge/Gorm-1.24.5-brightgreen)](https://github.com/eryajf/kite-desktop)
[![GitHub Pull Requests](https://img.shields.io/github/stars/eryajf/kite-desktop)](https://github.com/eryajf/kite-desktop/stargazers)
[![HitCount](https://views.whatilearened.today/views/github/eryajf/kite-desktop.svg)](https://github.com/eryajf/kite-desktop)
[![GitHub license](https://img.shields.io/github/license/eryajf/kite-desktop)](https://github.com/eryajf/kite-desktop/blob/main/LICENSE)
[![Commits](https://img.shields.io/github/commit-activity/m/eryajf/kite-desktop?color=ffff00)](https://github.com/eryajf/kite-desktop/commits/main)

<p> 🪁 A Wails v3-based desktop tool for multi-cluster K8S management 🪁</p>

<img src="https://cdn.jsdelivr.net/gh/eryajf/tu@main/img/image_20240420_214408.gif" width="800"  height="3">
</div><br>

<p align="center">
  <a href="" rel="noopener">
 <img src="https://cdn.jsdelivr.net/gh/eryajf/tu/img/image_20260415_222836.png" alt="Project logo"></a>
</p>

## Acknowledgement

This project is based on the original open source project [Kite](https://github.com/kite-org/kite).

First, thanks to the original Kite authors and all contributors. The upstream project already provided a very solid foundation, including Kubernetes resource management, cluster management workflows, backend capabilities, and the overall product direction. The desktop transformation in this repository is built directly on top of those results.

## Why This Repository Exists

`Kite Desktop` is not a simple mirror of the original repository, nor is it just a thin shell around it.

This project is the result of a substantial desktop-oriented rework based on the original Kite. The goal is to gradually reshape what was originally more Web / Server oriented into a truly installable, distributable, locally usable desktop Kubernetes management tool. At the same time, the project will explore deeper integration with AI capabilities.

## Tech Stack

The current desktop edition is built on the following core stack:

- `Go` for backend logic and Kubernetes integration
- `React` for the application UI
- `Wails v3` for desktop runtime, native windowing, system integration, and desktop packaging

Among them, `Wails v3` is the key infrastructure behind this transformation. Many future desktop capabilities will be built on top of it, such as:

- native window behavior adaptations
- local file access
- system file pickers
- external link handling with the system browser
- desktop package building and release workflows

## Project Direction

From now on, this repository will gradually separate from the original Kite repository and continue evolving independently around desktop use cases.

That means:

- desktop-native capabilities will continue to be enhanced
- interaction flows and feature boundaries will be adjusted for desktop usage scenarios
- parts that are no longer suitable for desktop will be trimmed, refactored, or replaced
- new capabilities with stronger desktop value will be introduced
- a dedicated release, installation, and upgrade system for the desktop app will be built

## Development

Install dependencies:

```bash
make deps
```

Run the desktop app in development mode:

```bash
make dev
```

Build the desktop app:

```bash
make build
```

## Release Targets

The project will primarily be delivered as desktop installation packages, with gradual support for the following platforms:

- macOS Intel
- macOS Apple Silicon
- Windows x64
- Windows ARM64

## Analytics Privacy

If you need the privacy notice for desktop analytics, see:

- [Desktop Analytics Privacy Notice](./docs/desktop-analytics-privacy-notice.md)

## License

This repository is licensed under `AGPL-3.0-only`. See [LICENSE](./LICENSE) for details.
> Note: this repository is a deeply modified derivative of the upstream `Kite` project. It may still contain code inherited from upstream under `Apache-2.0`, along with the corresponding attribution obligations. See [NOTICE](./NOTICE) and [licenses/Apache-2.0.txt](./licenses/Apache-2.0.txt).

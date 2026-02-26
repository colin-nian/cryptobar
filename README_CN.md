<p align="center">
  <img src="assets/logo.png" width="128" height="128" alt="CryptoBar Logo">
</p>

<h1 align="center">CryptoBar</h1>

<p align="center">
  <b>macOS 菜单栏实时加密货币价格追踪工具</b>
</p>

<p align="center">
  简体中文 · <a href="README.md">English</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/平台-macOS-blue?style=flat-square&logo=apple" alt="Platform">
  <img src="https://img.shields.io/badge/语言-Go-00ADD8?style=flat-square&logo=go" alt="Language">
  <img src="https://img.shields.io/badge/协议-MIT-green?style=flat-square" alt="License">
  <a href="https://github.com/colin-nian/cryptobar/releases/latest"><img src="https://img.shields.io/github/v/release/colin-nian/cryptobar?style=flat-square&color=orange" alt="Release"></a>
</p>

---

CryptoBar 是一个轻量级的 macOS 菜单栏应用，可在菜单栏实时显示加密货币价格。支持多个交易所数据源、币种 Logo 图标显示以及原生设置窗口，静默运行在你的菜单栏中。

## 功能特性

- **实时价格** — 通过 WebSocket 推送实时加密货币价格
- **多数据源** — 支持 Binance、HTX（火币）、Gate.io，一键切换
- **币种 Logo** — 在菜单栏直接显示币种图标（彩色或灰白）
- **24h 涨跌** — 可选显示 24 小时价格变动百分比
- **价格提醒** — 设置价格上限/下限提醒，通过 macOS 原生通知推送
- **原生设置** — 多标签页设置窗口（个性化、数据源、隐私、捐赠）
- **国际化** — 支持 English、简体中文、繁體中文、日本語、한국어
- **轻量运行** — 无 Dock 图标，完全在菜单栏运行
- **自定义** — 字体大小、显示模式（Logo/文字/混合）、紧凑名称

## 快速开始

> **下载即用，无需编译。**

1. **[下载 CryptoBar.app](https://github.com/colin-nian/cryptobar/releases/latest)** — 从 Release 页面下载最新版 `.zip`
2. 解压后将 `CryptoBar.app` 拖入 `/Applications`（应用程序）文件夹
3. 启动应用 — CryptoBar 会出现在菜单栏中

> **首次启动提示：** macOS 可能会对未签名应用弹出安全警告，请前往 **系统设置 → 隐私与安全性**，点击 **仍要打开** 即可。

## 从源码构建

**环境要求：** Go 1.23+、macOS 12+、Xcode 命令行工具

```bash
git clone https://github.com/colin-nian/cryptobar.git
cd cryptobar
make app
```

构建完成后应用位于 `build/CryptoBar.app`。

然后安装到应用程序：

```bash
make install
```

## 使用方法

1. 启动 `CryptoBar.app` — 它会出现在菜单栏
2. 点击菜单栏图标查看详细价格信息
3. 点击 **设置** 自定义显示、切换数据源或更改语言
4. 点击币种名称可以切换其在菜单栏中的显示/隐藏
5. 使用 **选择币种** 添加或移除追踪的加密货币

## 配置说明

配置文件保存在 `~/.cryptobar/config.json`，支持以下设置：

| 设置项 | 说明 |
|--------|------|
| 字体大小 | 小 (10)、中 (12)、大 (14) |
| 币种显示 | Logo 图标、文字符号、Logo + 文字 |
| Logo 颜色 | 彩色 或 灰白 |
| 显示 24h 涨跌 | 开启/关闭价格变动百分比 |
| 紧凑名称 | 仅显示币种代号 |
| 语言 | 英语、简中、繁中、日语、韩语 |
| 数据源 | Binance、HTX、Gate.io |

## 支持的交易所

| 交易所 | WebSocket | REST API | 状态 |
|--------|-----------|----------|------|
| Binance | `wss://stream.binance.com` | `api.binance.com` | ✅ |
| HTX（火币） | `wss://api.huobi.pro/ws` | `api.huobi.pro` | ✅ |
| Gate.io | `wss://api.gateio.ws/ws/v4/` | `api.gateio.ws` | ✅ |

## 技术栈

- **Go** — 核心应用逻辑
- **Objective-C / Cocoa** — 原生 macOS UI（NSWindow、NSTabView、NSStatusBar）
- **CGO** — Go 与原生 macOS API 的桥接
- **WebSocket** — 通过 gorilla/websocket 实时价格推送
- **Core Image** — 灰白色 Logo 渲染

## 开源协议

MIT License。详见 [LICENSE](LICENSE)。

## 捐赠

如果你觉得 CryptoBar 好用，欢迎支持开发：

**USDT (TRC20):** `TV75kwC1n7yA33kMi9yYw7EVgybPua4fvQ`

> 请确保使用 TRC20 网络发送 USDT，使用其他网络可能导致资金永久丢失。

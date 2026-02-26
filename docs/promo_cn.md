# 中文推广文案

---

## V2EX（/t/share 或 /t/macos）

**标题：** 分享一个自制的 macOS 菜单栏加密货币实时价格工具 — CryptoBar

**正文：**

之前一直用 CoinTick 看币价，但它经常显示问号、价格延迟、数据不准，忍无可忍自己写了一个。

**CryptoBar** 是一款完全免费、开源的 macOS 菜单栏加密货币价格追踪工具。

**核心特性：**

- **实时价格** — 通过 WebSocket 从交易所推送，不是轮询，延迟极低
- **支持三大交易所** — Binance、HTX（火币）、Gate.io，一键切换
- **1900+ 币种** — 原生搜索窗口，快速查找任意交易对
- **菜单栏显示币种 Logo** — 支持彩色/灰白模式
- **24 小时涨跌幅** 显示
- **价格提醒** — 原生 macOS 通知
- **5 种语言** — 简中、繁中、英语、日语、韩语
- **纯原生** — Go + Objective-C/Cocoa 开发，不是 Electron，体积 ~6MB
- **隐私友好** — 无账号、无埋点、无追踪，所有数据本地存储

**安装方式：**

```
brew tap colin-nian/tap
brew install cryptobar
```

或者直接从 [GitHub Releases](https://github.com/colin-nian/cryptobar/releases/latest) 下载 zip 解压到 Applications。

**项目地址：** https://github.com/colin-nian/cryptobar  
**主页：** https://colin-nian.github.io/cryptobar/

欢迎 Star、提 Issue、PR，也欢迎各种反馈！

---

## 少数派 / 掘金（技术分享风格）

**标题：** 用 Go + Objective-C 从零打造 macOS 菜单栏加密货币行情工具

**正文：**

### 背景

作为一个长期关注加密货币的开发者，我一直在用 CoinTick 之类的菜单栏工具看币价。但这些工具要么更新慢（轮询间隔长），要么数据不准（经常显示问号），要么不支持自定义。于是决定自己动手做一个。

### CryptoBar 是什么

CryptoBar 是一款开源的 macOS 菜单栏应用，实时展示加密货币价格。核心卖点：

1. **WebSocket 实时推送** — 连接交易所 WebSocket API，价格变动毫秒级更新
2. **多交易所支持** — Binance、HTX、Gate.io，国内用户也能用
3. **1900+ 币种** — 原生搜索窗口快速选币
4. **菜单栏 Logo 图标** — 彩色/灰白可选
5. **5 语言国际化** — 简/繁中、英、日、韩
6. **完全原生** — 非 Electron，体积仅 ~6MB

### 技术架构

| 层级 | 技术 | 作用 |
|------|------|------|
| 核心逻辑 | Go | WebSocket 连接、数据处理、配置管理、i18n |
| UI 层 | Objective-C / Cocoa | NSStatusBar、NSWindow、NSTableView |
| 桥接 | CGO | Go ↔ ObjC 数据传递，JSON 序列化 |
| 菜单栏库 | menuet (fork) | 自定义 TitleSegment 支持图文混排 |

### 几个有趣的技术点

**1. CGO 跨语言数据传递**

Go 和 ObjC 之间通过 JSON 字符串传递复杂数据。遇到一个坑：ObjC 的 `BOOL` 序列化为 `1`/`0`，但 Go 的 `json.Unmarshal` 期望 `true`/`false`，需要自定义解析。

**2. 菜单栏图文混排**

使用 `NSTextAttachment` 将币种 Logo 嵌入 `NSAttributedString`，实现菜单栏中图标 + 价格的紧凑展示。还支持通过 Core Image 实时转灰白。

**3. 1900+ 币种图标覆盖**

通过 CryptoCompare API 获取全量币种图标映射，本地 JSON 缓存（7 天刷新），加上 CoinCap CDN 作为兜底，覆盖绝大多数币种。`NSImage` 层也加了 nil 安全检查，防止个别缺失图标导致崩溃。

**4. 原生多标签页设置窗口**

纯 ObjC 实现的 `NSTabView` 设置窗口，包括个性化、数据源配置、隐私、捐赠四个标签页，每个数据源支持连通性测试。

### 安装与使用

```
brew tap colin-nian/tap
brew install cryptobar
```

或直接下载：https://github.com/colin-nian/cryptobar/releases/latest

### 开源地址

GitHub: https://github.com/colin-nian/cryptobar  
主页: https://colin-nian.github.io/cryptobar/  
License: MIT

欢迎 Star 和反馈！如果对 Go + CGO + Cocoa 的架构设计感兴趣，欢迎交流。

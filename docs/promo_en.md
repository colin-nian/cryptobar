# English Promotion Posts

---

## Reddit r/macapps

**Title:** I built a free, open-source menu bar crypto price tracker for macOS

**Body:**

Hey everyone! I was frustrated with existing menu bar crypto trackers — slow updates, broken prices, no customization — so I built my own.

**CryptoBar** is a free, open-source macOS menu bar app that shows real-time cryptocurrency prices with WebSocket streaming.

**What makes it different:**

- **Real-time prices** — WebSocket streaming from Binance, HTX (Huobi), or Gate.io. No slow polling.
- **1900+ coins** — Search and pick any trading pair with a native search window.
- **Coin logos** in the menu bar — color or grayscale mode.
- **24h price change** display.
- **Price alerts** with native macOS notifications.
- **5 languages** — English, Chinese (Simplified/Traditional), Japanese, Korean.
- **Native macOS app** — Built with Go + Objective-C/Cocoa. No Electron. Lightweight (~6MB).
- **Privacy-friendly** — No accounts, no analytics, no tracking. Your data stays local.

**Install:**

```
brew tap colin-nian/tap
brew install cryptobar
```

Or download from [GitHub Releases](https://github.com/colin-nian/cryptobar/releases/latest).

📸 Screenshots: https://colin-nian.github.io/cryptobar/

Would love to hear your feedback!

---

## Reddit r/golang

**Title:** Show r/golang: CryptoBar — a macOS menu bar app built with Go + CGO + Cocoa

**Body:**

I built a macOS menu bar cryptocurrency price tracker using Go as the primary language, with CGO bridging to Objective-C/Cocoa for native UI.

**Tech stack:**

- **Go** — Core logic, WebSocket handling, config management, i18n
- **CGO → Objective-C** — Native NSWindow, NSTableView, NSTabView for settings and coin selector
- **WebSocket** (gorilla/websocket) — Real-time price streaming from Binance, HTX, Gate.io
- **Cocoa/AppKit** — NSStatusBar menu, NSAttributedString with NSTextAttachment for inline coin logos
- Forked `github.com/caseymrm/menuet` to add custom title segments with image+text rendering and grayscale support via Core Image

**Interesting challenges solved:**

1. **Go ↔ ObjC data passing** — JSON serialization across the CGO boundary, including handling ObjC BOOL (1/0) vs Go bool (true/false)
2. **Thread safety** — Go goroutines updating the menu bar title while ObjC runs on the main thread
3. **Icon caching** — Fetching 1900+ coin icons from CryptoCompare API with local JSON cache and NSImage cache with nil-safety
4. **i18n without frameworks** — Simple Go map-based localization passed to ObjC via JSON

The app runs as a status-bar-only app (no Dock icon), under 6MB. 

GitHub: https://github.com/colin-nian/cryptobar  
Homepage: https://colin-nian.github.io/cryptobar/

Happy to discuss the Go+CGO architecture if anyone is interested!

---

## Hacker News

**Title:** Show HN: CryptoBar – Real-time crypto prices in your macOS menu bar (Go + Cocoa)

**Body:**

CryptoBar is a free, open-source macOS menu bar app that displays real-time cryptocurrency prices via WebSocket streaming.

Built with Go and native Objective-C/Cocoa (no Electron). Supports Binance, HTX, and Gate.io as data sources. Features include a searchable coin selector (1900+ coins), inline coin logos, price alerts, and 5-language i18n.

Install: `brew tap colin-nian/tap && brew install cryptobar`

https://github.com/colin-nian/cryptobar

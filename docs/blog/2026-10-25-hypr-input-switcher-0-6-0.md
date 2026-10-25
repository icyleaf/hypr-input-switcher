---
slug: hypr-input-switcher-0-6-0
title: "Hypr Input Switcher 0.6.0: Multi-Engine Support, Per-App Memory, and Configuration v3"
authors: [icyleaf]
tags: [release, announcement, fcitx5, hyprland]
date: 2026-10-25T10:00:00+09:00
---

We are excited to announce the release of **Hypr Input Switcher v0.6.0**! 🚀

This release marks our largest architectural update to date. It introduces native multi-engine fcitx5 support, per-application input method memory, first-class Wayland layer-shell overlay rules, and **Configuration Specification v3**.

<!-- truncate -->

## Highlights in v0.6.0

### 1. Multi-Engine Input Method Switching

Prior to v0.6.0, `hypr-input-switcher` was tightly coupled to Rime: non-English input methods were assumed to be Rime schemas, and switching to English deactivated the fcitx5 controller globally. 

In environments running multiple fcitx5 engines—such as [`msime`](https://msime.app/) for Chinese alongside `rime` (for Japanese or dialect schemas) and `keyboard-us` for English—this caused state mismatch issues:
- Other engines were incorrectly classified as English.
- Attempting to switch to non-Rime input methods failed because the switcher attempted to set non-existent Rime schemas.
- Deactivating the controller for English caused disruptive state teardowns across window focus changes.

In **v0.6.0**, the switching architecture is completely decoupled:
- **Direct Engine Switching**: Input method labels now map directly to fcitx5 input engines (e.g. `keyboard-us`, `msime`, `rime`, `mozc`). The switcher communicates directly via `SetCurrentIM`.
- **Non-Disruptive English**: Switching to `english` selects the configured English layout (`keyboard-us`) directly without deactivating the fcitx5 controller.

---

### 2. Per-App Input Method Memory

A long-standing pain point was the "English infection" problem: when navigating between windows using the `keep` rule, the switcher passively inherited whichever input method the previous window had left active. If you switched from a chat application (Chinese) to an IDE (English) and back to the chat application, your chat window was forced into English.

Starting in **v0.6.0**:
- The switcher maintains an internal `appHistory` cache keyed by window class (`clientInfo.Class`).
- When switching to a window governed by a `keep` rule (or falling back to a `keep` default), the switcher restores the **last-used input method for that specific application**.
- Explicit target rules (e.g. `input_method: chinese`) continue to take immediate precedence and update the application's remembered state.

---

### 3. Configuration Specification v3 & `profile` Abstraction

To support multi-engine setups cleanly, v0.6.0 introduces **Configuration Specification v3**. 

Instead of fragmenting configuration across separate top-level `input_methods`, `rime_schemas`, `display_names`, and `icons` sections, v3 unifies them into cohesive `input_methods` entities with a generic **`profile`** abstraction:

```yaml
version: 3
description: Hyprland Input Method Switcher Configuration

input_methods:
  english:
    engine: keyboard-us
    display_name: English
    icon: 🇺🇸

  chinese:
    engine: msime
    display_name: 中文
    icon: 🇨🇳

  japanese:
    engine: rime
    profile: jaroomaji
    display_name: 日本語
    icon: 🇯🇵
```

#### Flexible Shorthand Syntax
For rapid configuration, shorthand strings are fully supported:
- **`engine:profile`**: `japanese: rime:jaroomaji`
- **Plain `engine`**: `english: keyboard-us` or `chinese: msime`

#### 100% Backward Compatibility
If you have an existing `version: 2` configuration (used across releases v0.1.0 through v0.5.1), **no manual file edits are required**. The switcher transparently normalizes older configurations in memory upon loading.

---

### 4. Desktop Overlay Support (`layer_rules`)

Desktop overlays—such as app launchers (e.g. Omarchy Menu, Rofi, Walker) or floating widgets (calculators, emoji pickers)—run as Wayland layer surfaces (`zwlr_layer_shell_v1`). They never appear in `hyprctl clients` and cannot be matched by standard window rules.

v0.6.0 introduces `layer_rules` matched by layer namespace:

```yaml
layer_rules:
  - namespace: omarchy-menu
    input_method: english
  - namespace: icyleaf-qalculator
    input_method: english
```

When an overlay opens, the switcher activates the corresponding input method. Once closed, input method focus cleanly returns to the underlying active window.

---

### 5. Multi-Backend Notifications & Embedded Icons

- **17 Embedded SVG Icons**: Flag and layout icons are now bundled directly within the binary and automatically unpacked to `~/.local/share/hypr-input-switcher/icons/`.
- **Fallback Notification Daemon Support**: Notification delivery automatically selects available backends with automatic failover across `notify-send`, `hyprctl`, `dunstify`, `swaync-client`, and `mako`.

---

## Upgrading to v0.6.0

### Arch Linux / Omarchy
```bash
# Update via AUR or pacman
yay -Syu hypr-input-switcher
```

### From Source
```bash
git clone https://github.com/icyleaf/hypr-input-switcher.git
cd hypr-input-switcher
go build -o ~/.local/bin/hypr-input-switcher ./cmd/hypr-input-switcher
```

### Migrating Your Configuration
Existing configurations continue to work seamlessly. If you wish to migrate your configuration file to the modern v3 format, use the built-in migration utility:

```bash
# Migrate config file to v3
hypr-input-switcher migrate ~/.config/hypr-input-switcher/config.yaml
```

Check out the [**Configuration v3 Guide**](/docs/configuration/v3) for complete documentation and examples.

---

## Detailed Changelog

### Features
- **Engine**: Support arbitrary fcitx5 input method engines (`msime`, `rime`, `keyboard-us`, `mozc`) via D-Bus `SetCurrentIM`.
- **Memory**: Implement per-application input method memory for `keep` and unvisited windows.
- **Config**: Introduce Configuration Specification v3 with cohesive `input_methods` entities and generic `profile` abstraction.
- **Config**: Add support for shorthand formats (`engine:profile` and string engine names).
- **Migration**: Provide bidirectional normalization and automated migration from v1/v2 to v3.
- **Layer Shell**: Support `layer_rules` targeting Wayland desktop overlays by namespace.
- **Assets**: Embed 17 SVG icons with automatic runtime extraction.

### Bug Fixes
- **Engine**: Fix bug where non-Rime input methods (e.g. `msime`) were misidentified as English.
- **Engine**: Fix bug where switching to non-Rime engines attempted to set invalid Rime schemas.
- **State**: Fix "English infection" where navigating through code editors reset subsequent `keep` windows to English.
- **Daemon**: Fix controller teardown when switching to English layout.

### Documentation
- **Docs**: Restructure configuration documentation into versioned guides (`v3`, `v2`, `v1`).
- **ADR**: Add ADR-0001 covering multi-engine switching and per-app memory design.
- **Blog**: Launch official blog and release announcement series.

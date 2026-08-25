---
sidebar_position: 4
---

# Omarchy Plugin (GUI)

[**Hypr Input Switcher for Omarchy**](https://github.com/icyleaf/omarchy-hypr-input-switcher) is the official visual rule management overlay plugin designed specifically for **[Omarchy](https://omarchy.org/) 4.0+ (Quickshell)**.

It allows you to inspect, configure, and persist automatic input method switching rules for `hypr-input-switcher` in real time with **100% Pure Quickshell & JavaScript (Zero Python Dependency)**.

![Omarchy Plugin Preview](https://raw.githubusercontent.com/icyleaf/omarchy-hypr-input-switcher/main/preview.png)

---

## ✨ Features

- 🪟 **Visual Rule Management**: Easily view, add, edit, reorder, and remove window rules (Class, Regex matching, Window Title, and Target Input Method).
- 🎯 **Pure Native Window Picker**: Click-to-pick any running window on screen (chaining `slurp` + `hyprctl`) directly in Quickshell to auto-fill window `class` and `title`.
- ⌨️ **Input Method Management**: Dedicated tab to manage supported input methods, display names, engine backends, and inline Rime schemas (e.g., `rime_frost`, `jaroomaji`).
- 🔍 **System Diagnostics Banner**: Live status indicator displaying binary installation, version, daemon process state (with a one-click `▶ Start Service` button), and configuration file health.
- ↕️ **Rule Priority Reordering**: Move rules up and down to match first-fit execution precedence in `hypr-input-switcher`.
- ⚙️ **Default Input Method Switching**: Switch global fallback input method behavior (`english`, `chinese`, `keep`, etc.) on the fly.
- 💾 **Native Atomic YAML Sync**: Powered by a bundled pure JavaScript YAML engine and `Quickshell.Io.FileView` for instant atomic saves and hot reload with zero external runtime dependencies.
- 🎨 **Native Omarchy 4.0 Styling**: Full integration with Omarchy `qs.Commons` and `qs.Ui` theme tokens, rounded cards, keyboard focus handling, and backdrop dismissal.

---

## Prerequisites

Before installing the plugin, ensure you have:

1. **Omarchy 4.0+** running `omarchy-shell` (Quickshell).
2. **`hypr-input-switcher`** installed on your system (see [Installation](installation.md)).
3. **`slurp`** (installed by default on Omarchy) for visual window picking.

---

## Installation

### Method 1: Using Omarchy CLI (Recommended)

You can install the plugin directly with the `omarchy` package/plugin manager:

```bash
omarchy plugin install https://github.com/icyleaf/omarchy-hypr-input-switcher
```

### Method 2: Manual Clone

Alternatively, you can clone the repository directly into your Omarchy user plugins directory:

```bash
git clone https://github.com/icyleaf/omarchy-hypr-input-switcher.git ~/.config/omarchy/plugins/icyleaf.hypr-input-switcher
```

Once installed, Omarchy will automatically detect and load the plugin. If needed, you can force a plugin rescan:

```bash
omarchy-shell shell rescanPlugins
# or restart the shell
omarchy restart shell
```

---

## Usage

### Opening the Overlay

You can trigger the Hypr Input Switcher configuration overlay in several ways:

1. **Using Omarchy Shell Command**:
   ```bash
   omarchy-shell overlay toggle icyleaf.hypr-input-switcher
   ```

2. **Binding to a Hyprland Shortcut**:
   Add a keybinding in your Hyprland configuration (`~/.config/hypr/hyprland.conf` or `~/.config/omarchy/hypr/bindings.conf`):
   ```ini
   # Toggle Hypr Input Switcher settings overlay
   bind = SUPER SHIFT, I, exec, omarchy-shell overlay toggle icyleaf.hypr-input-switcher
   ```

---

## Key Features Breakdown

### 🎯 Native Window Picker

Instead of manually running `hyprctl activewindow` to find the exact class name or title:
1. Click the **"Pick Window"** crosshair icon in the rule dialog.
2. The screen enters interactive selection mode via `slurp`.
3. Click any target window.
4. The window's `class` and `title` are automatically populated into the form.

### 🔍 System Diagnostics & Service Control

The plugin actively monitors the health of `hypr-input-switcher`:
- **Binary Status**: Confirms whether the `hypr-input-switcher` executable is present in `$PATH` and reports its version.
- **Daemon State**: Checks if the background daemon process is running.
- **One-Click Start**: If the daemon is stopped, a convenient `▶ Start Service` button lets you launch it immediately.
- **Config File Health**: Verifies that `~/.config/hypr-input-switcher/config.yaml` is readable and valid.

### ↕️ First-Fit Rule Priority

`hypr-input-switcher` evaluates window rules in the order they appear in `config.yaml`. The Omarchy plugin lets you reorder rules with up/down arrows so more specific regex rules take precedence over broader matches.

### 💾 Atomic YAML Sync

The plugin reads and writes directly to `~/.config/hypr-input-switcher/config.yaml`. Because `hypr-input-switcher` supports config hot-reloading (via `--watch` or file change detection), changes made in the GUI take effect immediately without requiring a daemon restart.

---

## Source Repository

For source code, issue tracking, and contributions, visit:
- **GitHub**: [icyleaf/omarchy-hypr-input-switcher](https://github.com/icyleaf/omarchy-hypr-input-switcher)

---
sidebar_position: 1
---

# Introduction

Welcome to **Hypr Input Switcher** - a smart, performance-focused input method switcher designed specifically for Hyprland.

## What is Hypr Input Switcher?

Hypr Input Switcher is a lightweight, intelligent application that automatically switches your input methods based on the active window in Hyprland. No more manual switching between English and other languages when moving between different applications!

## Key Features

### 🚀 **Zero Configuration**
- Works out of the box with intelligent defaults
- Automatic detection of common applications
- No complex setup required

### ⚡ **Lightning Fast**
- Built with Go for maximum performance
- Switches input methods in under 100ms
- Minimal memory footprint (< 5MB)

### 🎯 **Smart Detection**
- Automatically detects application context
- Supports regex and string matching
- Customizable rules for any application

### 🔧 **Highly Configurable**
- YAML-based configuration
- Hot-reload support
- Extensive customization options

### 🔔 **Rich Notifications**
- Multiple notification backends
- Emoji and icon support
- Visual feedback for every switch

### 🛠 **Developer Friendly**
- Comprehensive logging
- Debugging tools
- Extensive documentation

## How It Works

```mermaid
graph LR
    A[Window Change] --> B[Detect Application]
    B --> C[Match Rules]
    C --> D[Switch Input Method]
    D --> E[Show Notification]
    E --> F[Ready for Next Change]
```

1. **Monitor Window Changes**: Listens to Hyprland window events
2. **Detect Application**: Identifies the active application
3. **Match Rules**: Applies your configured rules
4. **Switch Input Method**: Changes to the appropriate input method
5. **Notify User**: Shows visual feedback (optional)

## Supported Systems

- **Window Manager**: Hyprland (required)
- **Input Methods**: Fcitx5, Rime
- **Notifications**: hyprctl, notify-send, dunstify, swaync-client, mako
- **Platforms**: Linux (x86_64, ARM64)

## Quick Example

```yaml
# ~/.config/hypr-input-switcher/config.yaml
client_rules:
  - class: firefox          # Firefox browser
    input_method: chinese   # Use Chinese input
  - class: kitty           # Terminal
    input_method: english   # Use English input
  - class: code            # VS Code
    input_method: english   # Use English input

default_input_method: english
```

## Ready to Get Started?

Choose your installation method and get up and running in less than 30 seconds:

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
  <TabItem value="aur" label="🏠 Arch Linux (AUR)" default>
    ```bash
    paru -S hypr-input-switcher-bin
    ```
  </TabItem>
  <TabItem value="github" label="📦 GitHub Releases">
    ```bash
    # Download latest release
    wget https://github.com/icyleaf/hypr-input-switcher/releases/latest/download/hypr-input-switcher_Linux_x86_64.tar.gz

    # Extract and install
    tar -xzf hypr-input-switcher_Linux_x86_64.tar.gz
    sudo cp hypr-input-switcher /usr/local/bin/
    ```
  </TabItem>
  <TabItem value="source" label="🔨 From Source">
    ```bash
    git clone https://github.com/icyleaf/hypr-input-switcher.git
    cd hypr-input-switcher
    make build
    sudo make install
    ```
  </TabItem>
</Tabs>

Then add to your Hyprland config:

```ini
# ~/.config/hypr/hyprland.conf
exec-once = hypr-input-switcher
```

That's it! Your input methods will now switch automatically based on the active application.

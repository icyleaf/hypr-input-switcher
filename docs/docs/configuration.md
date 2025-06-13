---
sidebar_position: 3
---

# Configuration

Learn how to configure Hypr Input Switcher to match your workflow and preferences.

## Configuration File

The configuration file is located at `~/.config/hypr-input-switcher/config.yaml`. If it doesn't exist, you can create one or copy from the default template:

```bash
mkdir -p ~/.config/hypr-input-switcher
cp /usr/share/hypr-input-switcher/default.yaml ~/.config/hypr-input-switcher/config.yaml
```

## Basic Configuration

Here's a minimal configuration example:

```yaml
# Default input method when no rules match
default_input_method: english

# Input method definitions
input_methods:
  english:
    name: "English (US)"
    fcitx_name: "keyboard-us"
  chinese:
    name: "中文 (Rime)"
    fcitx_name: "rime"

# Application-specific rules
client_rules:
  - class: firefox
    input_method: chinese
  - class: kitty
    input_method: english
  - class: code
    input_method: english

# Notification settings
notifications:
  enabled: true
  backend: hyprctl
  show_method_name: true
  show_emoji: true
```

## Input Methods

Define all the input methods you want to use:

```yaml
input_methods:
  # English keyboard
  english:
    name: "English (US)"
    fcitx_name: "keyboard-us"
    emoji: "🇺🇸"

  # Chinese Rime
  chinese:
    name: "中文 (Rime)"
    fcitx_name: "rime"
    emoji: "🇨🇳"

  # Japanese Mozc
  japanese:
    name: "日本語 (Mozc)"
    fcitx_name: "mozc"
    emoji: "🇯🇵"

  # Korean Hangul
  korean:
    name: "한국어 (Hangul)"
    fcitx_name: "hangul"
    emoji: "🇰🇷"
```

### Finding Fcitx5 Input Method Names

To find the correct `fcitx_name` for your input methods:

```bash
# List all available input methods
fcitx5-remote -l

# Example output:
# keyboard-us
# rime
# mozc
# hangul
```

## Client Rules

Client rules determine which input method to use for specific applications:

### Basic Rules

```yaml
client_rules:
  # Browser applications - use Chinese
  - class: firefox
    input_method: chinese
  - class: chromium
    input_method: chinese
  - class: google-chrome
    input_method: chinese

  # Development tools - use English
  - class: code
    input_method: english
  - class: nvim
    input_method: english
  - class: jetbrains-.*
    input_method: english
    regex: true

  # Terminal applications - use English
  - class: kitty
    input_method: english
  - class: alacritty
    input_method: english
  - class: foot
    input_method: english
```

### Advanced Rules

#### Using Regex Patterns

```yaml
client_rules:
  # Match any JetBrains IDE
  - class: jetbrains-.*
    input_method: english
    regex: true

  # Match terminals
  - class: (kitty|alacritty|foot|wezterm)
    input_method: english
    regex: true

  # Match browsers
  - class: (firefox|chromium|chrome|edge)
    input_method: chinese
    regex: true
```

#### Title-based Rules

```yaml
client_rules:
  # Switch based on window title
  - title: ".*GitHub.*"
    input_method: english
    regex: true

  # Specific application with title
  - class: code
    title: ".*\\.md.*"  # Markdown files
    input_method: chinese
    regex: true
```

#### Multiple Conditions

```yaml
client_rules:
  # All conditions must match
  - class: firefox
    title: ".*Bilibili.*"
    input_method: chinese

  # Priority-based matching (first match wins)
  - class: code
    title: ".*README.*"
    input_method: chinese
    priority: 10

  - class: code
    input_method: english
    priority: 1
```

### Rule Properties

| Property | Type | Description | Example |
|----------|------|-------------|---------|
| `class` | string | Window class name | `firefox`, `code` |
| `title` | string | Window title | `GitHub`, `*.md` |
| `input_method` | string | Target input method | `english`, `chinese` |
| `regex` | boolean | Enable regex matching | `true`, `false` |
| `priority` | number | Rule priority (higher = first) | `10`, `1` |
| `enabled` | boolean | Enable/disable rule | `true`, `false` |

## Window Detection

### Finding Window Information

Use these commands to find window class and title:

```bash
# Get active window info (Hyprland)
hyprctl activewindow

# Monitor window events
hyprctl monitors

# List all windows
hyprctl clients
```

### Common Application Classes

Here are some common application class names:

#### Browsers
- `firefox` - Mozilla Firefox
- `chromium` - Chromium
- `google-chrome` - Google Chrome
- `microsoft-edge` - Microsoft Edge

#### Development
- `code` - Visual Studio Code
- `jetbrains-idea` - IntelliJ IDEA
- `jetbrains-pycharm` - PyCharm
- `nvim` - Neovim (in terminal)

#### Terminals
- `kitty` - Kitty terminal
- `alacritty` - Alacritty terminal
- `foot` - Foot terminal
- `wezterm` - WezTerm

#### Communication
- `discord` - Discord
- `telegram-desktop` - Telegram
- `slack` - Slack
- `teams` - Microsoft Teams

## Notifications

Configure how you want to be notified of input method changes:

```yaml
notifications:
  enabled: true
  backend: hyprctl
  duration: 2000  # milliseconds
  show_method_name: true
  show_emoji: true
  position: top-right

  # Notification backends (in order of preference)
  backends:
    - hyprctl
    - notify-send
    - dunstify
    - swaync-client
    - mako
```

### Backend Options

| Backend | Description | Requirements |
|---------|-------------|--------------|
| `hyprctl` | Hyprland native notifications | Hyprland |
| `notify-send` | Standard desktop notifications | libnotify |
| `dunstify` | Dunst notification daemon | dunst |
| `swaync-client` | SwayNC notifications | swaync |
| `mako` | Mako notification daemon | mako |

### Notification Customization

```yaml
notifications:
  templates:
    switch: "🔄 Switched to {method_name}"
    error: "❌ Failed to switch input method"

  # Custom per-method notifications
  method_notifications:
    chinese:
      emoji: "🇨🇳"
      message: "中文输入法已启用"
    english:
      emoji: "🇺🇸"
      message: "English input enabled"
```

## Advanced Settings

### Performance Tuning

```yaml
# Performance settings
performance:
  window_polling_interval: 100  # milliseconds
  debounce_delay: 50           # milliseconds
  max_history_size: 100        # entries

# Logging configuration
logging:
  level: info  # debug, info, warn, error
  file: ~/.local/share/hypr-input-switcher/app.log
  max_size: 10MB
  max_files: 5
```

### Startup Behavior

```yaml
# Startup configuration
startup:
  restore_last_method: true
  check_dependencies: true
  auto_configure: true

# Shutdown behavior
shutdown:
  save_state: true
  restore_original_method: false
```

## Hot Reload

The configuration file supports hot reloading. Changes are automatically detected and applied without restarting the application:

```bash
# Force reload configuration
hypr-input-switcher --reload

# Check configuration validity
hypr-input-switcher --check-config
```

## Configuration Validation

Validate your configuration before applying:

```bash
# Check syntax and logic
hypr-input-switcher --validate-config

# Test rules against current windows
hypr-input-switcher --test-rules

# Dry run (show what would happen)
hypr-input-switcher --dry-run
```

## Example Configurations

### Minimal Setup

```yaml
default_input_method: english

input_methods:
  english:
    name: "English"
    fcitx_name: "keyboard-us"
  chinese:
    name: "中文"
    fcitx_name: "rime"

client_rules:
  - class: firefox
    input_method: chinese
  - class: kitty
    input_method: english
```

### Advanced Setup

```yaml
# Advanced configuration example
default_input_method: english

input_methods:
  english:
    name: "English (US)"
    fcitx_name: "keyboard-us"
    emoji: "🇺🇸"
  chinese:
    name: "中文 (Rime)"
    fcitx_name: "rime"
    emoji: "🇨🇳"
  japanese:
    name: "日本語 (Mozc)"
    fcitx_name: "mozc"
    emoji: "🇯🇵"

client_rules:
  # High priority rules
  - class: code
    title: ".*\\.md"
    input_method: chinese
    priority: 10
    regex: true

  # Development tools
  - class: jetbrains-.*
    input_method: english
    regex: true

  # Browsers for different purposes
  - class: firefox
    title: ".*GitHub.*"
    input_method: english
    regex: true

  - class: firefox
    input_method: chinese

  # Communication
  - class: discord
    input_method: chinese
  - class: telegram-desktop
    input_method: chinese

  # Terminals always English
  - class: (kitty|alacritty|foot)
    input_method: english
    regex: true

notifications:
  enabled: true
  backend: hyprctl
  duration: 1500
  show_method_name: true
  show_emoji: true

performance:
  window_polling_interval: 50
  debounce_delay: 30

logging:
  level: info
  file: ~/.local/share/hypr-input-switcher/app.log
```

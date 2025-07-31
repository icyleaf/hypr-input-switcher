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
# Rime schema mapping
rime_schemas:
  chinese: rime_frost
  japanese: jaroomaji

# Notification settings
notifications:
  enabled: true
  duration: 2000
  show_on_switch: true
  show_app_name: true
  icon_path: "~/.local/share/hypr-input-switcher/icons"
  methods:
    - notify-send
    - dunstify
    - hyprctl
    - swaync-client
    - mako

# Display names for input methods
display_names:
  english: English
  chinese: 中文
  japanese: 日本語

# Application-specific rules
client_rules:
  - class: firefox
    input_method: chinese
  - class: kitty
    input_method: english
  - class: code
    input_method: english
```

## Icons Configuration

Configure how icons are displayed in notifications:

### Icon Path Settings

```yaml
notifications:
  enabled: true
  icon_path: "~/.local/share/hypr-input-switcher/icons"  # Path to icon files
  methods:
    - notify-send    # Supports image icons
    - dunstify       # Supports image icons
    - hyprctl        # Only supports emoji
    - swaync-client  # Only supports emoji
    - mako           # Supports image icons
```

### Icon Configuration Options

You have multiple ways to configure icons:

#### 1. Automatic Icon Detection (Recommended)

The application will automatically look for icon files in the `icon_path` directory:

```yaml
notifications:
  icon_path: "~/.local/share/hypr-input-switcher/icons"

# No icons section needed - will automatically find:
# - english: us.svg, en.png, english.svg, etc.
# - chinese: cn.svg, zh.png, chinese.svg, etc.
# - japanese: jp.svg, ja.png, japanese.svg, etc.
```

#### 2. Custom Icon Files

Specify custom icon files (relative to `icon_path`):

```yaml
notifications:
  icon_path: "~/.local/share/hypr-input-switcher/icons"

icons:
  english: "us.svg"       # Will look for us.svg, us.png, etc. in icon_path
  chinese: "cn.png"       # Will look for cn.png, cn.svg, etc. in icon_path
  japanese: "jp.svg"      # Will look for jp.svg, jp.png, etc. in icon_path
```

#### 3. Mixed Configuration (Files + Emoji)

Combine icon files and emoji:

```yaml
notifications:
  icon_path: "~/.local/share/hypr-input-switcher/icons"

icons:
  english: "us.svg"       # Use icon file
  chinese: "🇨🇳"          # Use emoji directly
  japanese: "jp.png"      # Use icon file
  korean: "🇰🇷"           # Use emoji directly
```

#### 4. Absolute Paths

Use absolute paths for custom icon locations:

```yaml
icons:
  english: "/usr/share/pixmaps/flags/us.png"
  chinese: "/home/user/my-icons/china.svg"
  japanese: "🇯🇵"  # Mix with emoji
```

#### 5. System Icon Names

Use system icon names (for icon themes):

```yaml
icons:
  english: "preferences-desktop-locale"
  chinese: "input-keyboard"
  japanese: "applications-education-language"
```

### Icon File Search Priority

The application searches for icon files in this order:

1. **Custom icons from config**: If specified in `icons` section
2. **Method name**: `{method}.png`, `{method}.svg`, etc.
3. **Language code**: `en.png`, `zh.svg`, `ja.png`, etc.
4. **Country code**: `us.png`, `cn.svg`, `jp.png`, etc.
5. **Emoji fallback**: Built-in emoji if no files found

### Supported Icon Formats

- **PNG**: `.png`
- **SVG**: `.svg` (recommended for scalability)
- **JPEG**: `.jpg`, `.jpeg`
- **ICO**: `.ico`
- **GIF**: `.gif`
- **BMP**: `.bmp`

### Notification Method Compatibility

| Method | Image Files | Emoji | System Icons |
|--------|-------------|-------|--------------|
| `notify-send` | ✅ | ✅ | ✅ |
| `dunstify` | ✅ | ✅ | ✅ |
| `mako` | ✅ | ✅ | ✅ |
| `hyprctl` | ❌ | ✅ | ❌ |
| `swaync-client` | ❌ | ✅ | ❌ |

## Rime Schemas

Configure Rime input method schemas:

```yaml
rime_schemas:
  chinese: rime_frost     # Use rime_frost schema for Chinese
  japanese: jaroomaji     # Use jaroomaji schema for Japanese
  korean: hangul          # Use hangul schema for Korean
```

## Display Names

Customize how input method names appear in notifications:

```yaml
display_names:
  english: "English (US)"
  chinese: "中文简体"
  japanese: "日本語"
  korean: "한국어"
  german: "Deutsch"
  french: "Français"
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

## Notifications

Configure notification appearance and behavior:

```yaml
notifications:
  enabled: true
  duration: 2000                    # Duration in milliseconds
  show_on_switch: true             # Show notification when switching
  show_app_name: true              # Include app name in notification
  icon_path: "~/.local/share/hypr-input-switcher/icons"

  # Notification methods (in priority order)
  methods:
    - notify-send                   # Primary choice
    - dunstify                      # Secondary choice
    - hyprctl                       # Hyprland native
    - swaync-client                 # Sway notification center
    - mako                          # Mako daemon

  # Force specific method (optional)
  force_method: ""                  # Leave empty for auto-detection

  # Disable specific methods
  disabled_methods: []              # e.g., ["hyprctl", "mako"]
```

### Notification Method Configuration

#### notify-send (Recommended)
```yaml
notifications:
  methods:
    - notify-send
  # Works with most desktop environments
  # Supports images, emoji, and system icons
```

#### Dunstify
```yaml
notifications:
  methods:
    - dunstify
  # Works with Dunst notification daemon
  # Supports images, emoji, and system icons
```

#### Hyprctl (Hyprland Only)
```yaml
notifications:
  methods:
    - hyprctl
  # Native Hyprland notifications
  # Only supports emoji and text
```

## Example Configurations

### Minimal Setup with Icons

```yaml
notifications:
  enabled: true
  icon_path: "~/.local/share/hypr-input-switcher/icons"

display_names:
  english: English
  chinese: 中文

# Icons will be automatically detected from icon_path
# Looks for: us.svg, cn.svg, en.png, zh.png, etc.

client_rules:
  - class: firefox
    input_method: chinese
  - class: kitty
    input_method: english
```

### Advanced Setup with Custom Icons

```yaml
rime_schemas:
  chinese: rime_frost
  japanese: jaroomaji

notifications:
  enabled: true
  duration: 1500
  show_on_switch: true
  show_app_name: true
  icon_path: "~/.local/share/hypr-input-switcher/icons"
  methods:
    - notify-send
    - dunstify
    - hyprctl
  force_method: ""
  disabled_methods: []

display_names:
  english: "English (US)"
  chinese: "中文简体"
  japanese: "日本語"
  korean: "한국어"

icons:
  english: "us.svg"              # Use custom icon file
  chinese: "cn.svg"              # Use custom icon file
  japanese: "🇯🇵"                # Use emoji
  korean: "kr.png"               # Use custom icon file
  german: "/usr/share/flags/de.svg"  # Use absolute path

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

  # Browsers
  - class: firefox
    title: ".*GitHub.*"
    input_method: english
    regex: true

  - class: firefox
    input_method: chinese

  # Communication apps
  - class: discord
    input_method: chinese
  - class: telegram-desktop
    input_method: chinese

  # Terminals always English
  - class: (kitty|alacritty|foot)
    input_method: english
    regex: true
```

### Emoji-Only Setup

```yaml
notifications:
  enabled: true
  methods:
    - hyprctl          # Hyprland native (emoji only)
    - notify-send      # Fallback

display_names:
  english: English
  chinese: 中文
  japanese: 日本語

icons:
  english: "🇺🇸"
  chinese: "🇨🇳"
  japanese: "🇯🇵"
  korean: "🇰🇷"
  german: "🇩🇪"
  french: "🇫🇷"

client_rules:
  - class: firefox
    input_method: chinese
  - class: kitty
    input_method: english
```

## Icon Management

### Embedded Icons

The application comes with embedded country flag icons that are automatically extracted on first run:

```bash
# Icons are automatically extracted to:
~/.local/share/hypr-input-switcher/icons/

# Available icons: us.svg, cn.svg, jp.svg, kr.svg, de.svg, fr.svg, etc.
```

### Custom Icon Directory

You can use any directory for icons:

```yaml
notifications:
  icon_path: "/usr/share/pixmaps/flags"  # System-wide icons
  # or
  icon_path: "~/Pictures/input-method-icons"  # Personal icons
```

### Icon Debugging

To troubleshoot icon issues:

```bash
# Run with debug logging
hypr-input-switcher --log-level debug

# Check icon status
hypr-input-switcher status

# Test notification with icon
notify-send -i ~/.local/share/hypr-input-switcher/icons/us.svg "Test" "Icon test"
```

## Hot Reload

The configuration file supports hot reloading. Changes are automatically detected and applied without restarting the application:

```bash
# Watch mode (auto-reload on config changes)
hypr-input-switcher --watch

# Force reload configuration
killall -HUP
```

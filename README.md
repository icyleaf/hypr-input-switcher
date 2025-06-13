<div align="center">
  <img src="docs/static/img/logo.svg" alt="Hypr Input Switcher" width="64" height="64">
  <h1>Hypr Input Switcher</h1>
</div>

# Hyprland Input Method Switcher

A smart input method switcher for Hyprland that automatically switches input methods based on the active window. Built with Go for performance and reliability.

## Features

- 🚀 **Automatic Input Method Switching**: Seamlessly switches input methods based on active application windows
- ⚙️ **Flexible Configuration**: YAML-based configuration with hot-reload support
- 🔔 **Rich Notification System**: Multiple notification backends with emoji support
- 🎯 **Pattern Matching**: Supports regex and string matching for window classes and titles
- 📁 **Hot Configuration Reload**: Watch configuration file changes and reload without restart
- 🛠 **Multiple Input Methods**: Full support for Fcitx5, Rime
- 📊 **Comprehensive Logging**: Configurable log levels with multiple output options
- 🌐 **Cross-Platform Icons**: Support for both emoji and traditional desktop icons

## Prerequisites

### Hyprland Requirements

This application is specifically designed for **Hyprland** and requires:

- **Hyprland** (any recent version)
- **hyprctl** command-line tool (comes with Hyprland)
- One of the supported input method frameworks:
  - **Fcitx5** (recommended)
  - **Rime** (can work with Fcitx5)

### Hyprland Configuration

To get the best experience, add the following to your Hyprland configuration (`~/.config/hypr/hyprland.conf`):

```ini
# Input method environment variables
env = GTK_IM_MODULE,fcitx
env = QT_IM_MODULE,fcitx
env = XMODIFIERS,@im=fcitx
env = SDL_IM_MODULE,fcitx
env = GLFW_IM_MODULE,ibus

# Auto-start input method
exec-once = fcitx5 -d  # For Fcitx5

# Auto-start hypr-input-switcher
exec-once = hypr-input-switcher

# Optional: Hyprland window rules for better input method handling
windowrulev2 = float,class:^(fcitx5-config-qt)$
windowrulev2 = float,class:^(org.fcitx.*)$
```

### Input Method Setup

#### For Fcitx5 (Recommended)

1. **Install Fcitx5:**
   ```bash
   # Arch Linux
   sudo pacman -S fcitx5 fcitx5-gtk fcitx5-qt fcitx5-configtool

   # Ubuntu/Debian
   sudo apt install fcitx5 fcitx5-frontend-gtk3 fcitx5-frontend-qt5 fcitx5-config-qt

   # Fedora
   sudo dnf install fcitx5 fcitx5-gtk fcitx5-qt fcitx5-configtool
   ```

2. **Install Input Methods:**
   ```bash
   # Chinese (Rime)
   sudo pacman -S fcitx5-rime  # Arch
   sudo apt install fcitx5-rime  # Ubuntu

   # Japanese
   sudo pacman -S fcitx5-mozc  # Arch
   sudo apt install fcitx5-mozc  # Ubuntu

   # Korean
   sudo pacman -S fcitx5-hangul  # Arch
   sudo apt install fcitx5-hangul  # Ubuntu
   ```

3. **Configure Fcitx5:**
   ```bash
   # Run configuration tool
   fcitx5-configtool

   # Add input methods you need
   # Set keyboard layout and input methods
   ```

## Supported Notification Systems

- **notify-send** (libnotify - GNOME, KDE, etc.)
- **dunstify** (Dunst notification daemon)
- **hyprctl** (Hyprland native notifications)
- **swaync-client** (Sway Notification Center)
- **mako** (Mako for Sway/Wayland)

## Installation

### Official Releases

Download the latest stable release from the [GitHub Releases page](https://github.com/icyleaf/hypr-input-switcher/releases).

```bash
# Download and install the latest release
curl -s https://api.github.com/repos/icyleaf/hypr-input-switcher/releases/latest \
  | grep "browser_download_url.*Linux_x86_64.tar.gz" \
  | cut -d : -f 2,3 \
  | tr -d \" \
  | wget -qi -

# Extract and install
tar -xzf hypr-input-switcher_Linux_x86_64.tar.gz
sudo cp hypr-input-switcher /usr/local/bin/
sudo chmod +x /usr/local/bin/hypr-input-switcher
```

### Development Snapshots

For testing the latest development version, you can download snapshot builds from GitHub Actions:

1. **Visit the [Actions page](https://github.com/icyleaf/hypr-input-switcher/actions/workflows/snapshot.yml)**
2. **Click on the latest successful build from the `main` branch**
3. **Download the artifact named `hypr-input-switcher-snapshot-YYYY-MM-DD`**
4. **Extract and install:**
   ```bash
   # Unzip the downloaded artifact
   unzip hypr-input-switcher-snapshot-*.zip

   # Install
   cd hypr-input-switcher-snapshot-*
   sudo cp hypr-input-switcher_linux_amd64_v1/hypr-input-switcher /usr/local/bin/
   sudo chmod +x /usr/local/bin/hypr-input-switcher
   ```

> **Note**: Snapshot builds require a GitHub account to download artifacts. These are automatically built from the latest `main` branch and may contain unstable features.

### Package Managers (WIP)

#### Arch Linux (AUR)

```bash
# Install from AUR using your preferred AUR helper
paru -S hypr-input-switcher-bin

# Or manually
git clone https://aur.archlinux.org/hypr-input-switcher-bin.git
cd hypr-input-switcher-bin
makepkg -si
```

#### APT (Debian/Ubuntu)

```bash
# Download the .deb package from releases
wget https://github.com/icyleaf/hypr-input-switcher/releases/latest/download/hypr-input-switcher_amd64.deb

# Install
sudo dpkg -i hypr-input-switcher_amd64.deb
sudo apt-get install -f  # Fix dependencies if needed
```

#### RPM (Fedora/RHEL/CentOS)

```bash
# Download the .rpm package from releases
wget https://github.com/icyleaf/hypr-input-switcher/releases/latest/download/hypr-input-switcher_x86_64.rpm

# Install
sudo rpm -i hypr-input-switcher_x86_64.rpm
```

### From Source

1. **Clone the repository:**
   ```bash
   git clone https://github.com/icyleaf/hypr-input-switcher.git
   cd hypr-input-switcher
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Build and install:**
   ```bash
   # System-wide installation
   make build
   sudo make install

   # Or user installation
   make install-dev
   ```

### Auto-start with Hyprland

Add to your Hyprland config (`~/.config/hypr/hyprland.conf`):

```ini
# Auto-start hypr-input-switcher
exec-once = hypr-input-switcher

# Or with custom config
exec-once = hypr-input-switcher --config=~/.config/hypr-input-switcher/my-config.yaml

# Or with debug logging
exec-once = hypr-input-switcher --log-level=debug --watch
```

## Quick Start

1. **Ensure Hyprland is running:**
   ```bash
   # Check if Hyprland is active
   echo $HYPRLAND_INSTANCE_SIGNATURE

   # Test hyprctl command
   hyprctl version
   ```

2. **Run with default settings:**
   ```bash
   hypr-input-switcher
   ```

3. **Run with custom configuration:**
   ```bash
   hypr-input-switcher --config=/path/to/your/config.yaml
   ```

4. **Enable hot-reload for development:**
   ```bash
   hypr-input-switcher --watch --log-level=debug
   ```

## Configuration

### Default Configuration Location

- **User config**: `~/.config/hypr-input-switcher/config.yaml`
- **System config**: `/etc/hypr-input-switcher/config.yaml`

If no configuration file exists, a default one will be created automatically.

### Hyprland Window Class Detection

To find the correct window class names for your applications:

```bash
# Get current active window info
hyprctl activewindow

# List all windows with their classes
hyprctl clients

# Monitor window changes in real-time
hyprctl monitors
```

### Example Configuration

```yaml
version: 2
description: Hyprland Input Method Switcher Configuration

# Input method definitions
input_methods:
  english: keyboard-us
  chinese: rime
  japanese: rime

# Application-specific rules (use hyprctl to find class names)
client_rules:
  - class: firefox                    # Firefox browser
    input_method: chinese
  - class: google-chrome              # Chrome browser
    input_method: chinese
  - class: code                       # VS Code
    input_method: english
  - class: kitty                      # Kitty terminal
    input_method: english
  - class: org.wezfurlong.wezterm     # WezTerm terminal
    input_method: english
  - class: "^(org.telegram.desktop)$" # Telegram (regex)
    input_method: chinese
  - class: firefox                    # Firefox with specific title
    title: ".*GitHub.*"
    input_method: english

# Default input method when no rules match
default_input_method: english

# Fcitx5 configuration
fcitx5:
  enabled: true
  rime_input_method: rime

# Rime schema mappings
rime_schemas:
  chinese: rime_frost      # Your Rime schema name
  japanese: jaroomaji      # Japanese input schema

# Notification settings
notifications:
  enabled: true
  duration: 2000
  show_on_switch: true
  show_app_name: true
  methods:
    - hyprctl          # Use Hyprland's native notifications first
    - notify-send      # Fallback to libnotify
    - dunstify         # Dunst notifications
  force_method: ""          # Force specific method
  disabled_methods: []      # Disable specific methods

# Display names for input methods
display_names:
  english: English
  chinese: 中文
  japanese: 日本語

# Icons (supports emoji and icon names)
icons:
  english: "🇺🇸"   # US flag
  chinese: "🇨🇳"   # Chinese flag
  japanese: "🇯🇵"  # Japanese flag
```

## Hyprland Integration

### Window Information

The application uses `hyprctl` to get window information:

```bash
# Current active window
hyprctl activewindow -j

# All windows
hyprctl clients -j

# Monitor information
hyprctl monitors -j

# Workspace information
hyprctl workspaces -j
```

### Supported Window Properties

You can match windows based on:

- **class**: Application class name (most common)
- **title**: Window title (optional, supports regex)
- **address**: Window address (advanced usage)

### Finding Window Classes

Use these commands to identify window classes:

```bash
# Method 1: Get active window
hyprctl activewindow | grep "class:"

# Method 2: List all windows
hyprctl clients | grep -E "(class|title):"

# Method 3: Real-time monitoring
watch -n 1 'hyprctl activewindow | grep -E "(class|title):"'
```

Common application classes:
- **Browsers**: `firefox`, `google-chrome`, `chromium`
- **Terminals**: `kitty`, `alacritty`, `org.wezfurlong.wezterm`
- **Editors**: `code`, `nvim`, `emacs`
- **Communication**: `org.telegram.desktop`, `discord`, `slack`

## Command Line Usage

```bash
# Basic usage
hypr-input-switcher

# Specify configuration file
hypr-input-switcher --config=/path/to/config.yaml

# Enable hot-reload
hypr-input-switcher --watch

# Set log level
hypr-input-switcher --log-level=debug

# Force stdout logging
hypr-input-switcher --log-stdout

# Show version
hypr-input-switcher version

# Combine options
hypr-input-switcher --config=./my-config.yaml --watch --log-level=debug
```

### Environment Variables

All command line options can be set via environment variables:

```bash
export HYPR_INPUT_SWITCHER_CONFIG="/path/to/config.yaml"
export HYPR_INPUT_SWITCHER_LOG_LEVEL="debug"
export HYPR_INPUT_SWITCHER_WATCH="true"
export HYPR_INPUT_SWITCHER_LOG_STDOUT="true"

hypr-input-switcher
```

## Project Structure

```
hypr-input-switcher/
├── cmd/
│   └── hypr-input-switcher/
│       └── main.go              # Application entry point
├── internal/
│   ├── app/
│   │   └── app.go              # Main application logic
│   ├── config/
│   │   ├── config.go           # Configuration structures
│   │   └── manager.go          # Config management with hot-reload
│   ├── inputmethod/
│   │   ├── switcher.go         # Input method switching logic
│   │   ├── fcitx5.go          # Fcitx5 backend implementation
│   │   └── rime.go            # Rime backend implementation
│   └── notification/
│       └── notifier.go         # Multi-backend notification system
├── pkg/
│   └── logger/
│       └── logger.go           # Logging utilities
├── configs/
│   └── default.yaml            # Default configuration template
├── Makefile                    # Build and installation targets
├── go.mod                      # Go module definition
└── README.md                   # This file
```

## Advanced Features

### Hot Configuration Reload

Enable configuration hot-reload to modify settings without restarting:

```bash
hypr-input-switcher --watch
```

The application will automatically detect changes to the configuration file and reload settings in real-time.

### Pattern Matching

The application supports both regex and string matching for window classes and titles:

```yaml
client_rules:
  # Exact string match
  - class: firefox
    input_method: chinese

  # Regex pattern
  - class: "^(code|codium)$"
    input_method: english

  # Match both class and title
  - class: firefox
    title: ".*GitHub.*"
    input_method: english

  # Match terminals
  - class: "^(kitty|alacritty|wezterm)$"
    input_method: english
```

### Custom Notification Methods

Configure notification priority and methods:

```yaml
notifications:
  methods:
    - hyprctl          # Hyprland native (recommended for Hyprland)
    - swaync-client    # SwayNC
    - dunstify         # Dunst
    - notify-send      # libnotify fallback
  force_method: "hyprctl"        # Force Hyprland notifications
  disabled_methods: ["mako"]     # Disable specific methods
```

## Development

### Running from Source

```bash
# Build and run with development config
make run-dev

# Build only
make build

# Run tests
make test

# Clean build artifacts
make clean
```

### Debugging

Enable debug logging to troubleshoot issues:

```bash
hypr-input-switcher --log-level=debug --log-stdout
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Setup

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Troubleshooting

### Common Issues

1. **"Hyprland is not running"**: Ensure you're running this inside a Hyprland session
2. **"hyprctl command not found"**: Install Hyprland or ensure it's in your PATH
3. **Configuration file not found**: The application will create a default config if none exists
4. **Input method not switching**:
   - Check if the input method names match your system configuration
   - Verify Fcitx5 is running: `ps aux | grep fcitx5`
   - Test manual switching: `fcitx5-remote -s rime`
5. **Notifications not showing**:
   - Verify your notification daemon is running
   - Test with: `hyprctl notify 1 2000 "rgb(ff1ea3)" "Test notification"`
6. **Window class not matching**: Use `hyprctl activewindow` to get correct class names

### Getting Help

- Check if Hyprland is running: `hyprctl version`
- Check the debug logs: `hypr-input-switcher --log-level=debug`
- Verify your configuration: `hypr-input-switcher --config=/path/to/config.yaml --log-level=debug`
- Test window detection: `watch -n 1 'hyprctl activewindow | grep class'`
- Open an issue on GitHub with logs and configuration details

### Hyprland-specific Debugging

```bash
# Check Hyprland environment
echo $HYPRLAND_INSTANCE_SIGNATURE

# Monitor window changes
hyprctl --batch "clients; activewindow"

# Test input method switching manually
fcitx5-remote -s keyboard-us  # Switch to English
fcitx5-remote -s rime         # Switch to Chinese
fcitx5-remote -c              # Get current input method

# Test notifications
hyprctl notify 1 3000 "rgb(ff1ea3)" "Test notification"
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Hyprland community for the excellent window manager
- Fcitx5 and Rime developers for robust input method frameworks
- All contributors and users who help improve this project

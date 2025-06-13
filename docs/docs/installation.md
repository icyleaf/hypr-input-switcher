---
sidebar_position: 2
---

# Installation

This guide covers all the ways to install Hypr Input Switcher on your system.

## Prerequisites

Before installing, make sure you have:

- **Hyprland** window manager installed and running
- **Fcitx5** input method framework
- One or more input methods configured (e.g., fcitx5-rime, fcitx5-mozc)

## Installation Methods

### Arch Linux (AUR)

The easiest way to install on Arch Linux is through the AUR:

```bash
# Using paru
paru -S hypr-input-switcher-bin

# Using yay
yay -S hypr-input-switcher-bin

# Manual installation
git clone https://aur.archlinux.org/hypr-input-switcher-bin.git
cd hypr-input-switcher-bin
makepkg -si
```

### Pre-built Binaries

Download the latest pre-built binary from GitHub releases:

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
  <TabItem value="x86_64" label="x86_64 (Intel/AMD)" default>
    ```bash
    # Download the latest release
    wget https://github.com/icyleaf/hypr-input-switcher/releases/latest/download/hypr-input-switcher_Linux_x86_64.tar.gz

    # Extract the archive
    tar -xzf hypr-input-switcher_Linux_x86_64.tar.gz

    # Install the binary
    sudo cp hypr-input-switcher /usr/local/bin/
    sudo chmod +x /usr/local/bin/hypr-input-switcher

    # Install configuration files
    mkdir -p ~/.config/hypr-input-switcher
    cp configs/default.yaml ~/.config/hypr-input-switcher/config.yaml
    ```
  </TabItem>
  <TabItem value="arm64" label="ARM64">
    ```bash
    # Download the latest release
    wget https://github.com/icyleaf/hypr-input-switcher/releases/latest/download/hypr-input-switcher_Linux_arm64.tar.gz

    # Extract the archive
    tar -xzf hypr-input-switcher_Linux_arm64.tar.gz

    # Install the binary
    sudo cp hypr-input-switcher /usr/local/bin/
    sudo chmod +x /usr/local/bin/hypr-input-switcher

    # Install configuration files
    mkdir -p ~/.config/hypr-input-switcher
    cp configs/default.yaml ~/.config/hypr-input-switcher/config.yaml
    ```
  </TabItem>
</Tabs>

### Package Managers

#### Debian/Ubuntu (APT)

```bash
# Download the .deb package
wget https://github.com/icyleaf/hypr-input-switcher/releases/latest/download/hypr-input-switcher_amd64.deb

# Install the package
sudo dpkg -i hypr-input-switcher_amd64.deb

# Fix dependencies if needed
sudo apt-get install -f
```

#### Fedora/RHEL/CentOS (RPM)

```bash
# Download the .rpm package
wget https://github.com/icyleaf/hypr-input-switcher/releases/latest/download/hypr-input-switcher_x86_64.rpm

# Install the package
sudo rpm -i hypr-input-switcher_x86_64.rpm

# Or using dnf
sudo dnf install hypr-input-switcher_x86_64.rpm
```

### From Source

#### Prerequisites

- Go 1.21 or later
- Git
- Make

#### Build and Install

```bash
# Clone the repository
git clone https://github.com/icyleaf/hypr-input-switcher.git
cd hypr-input-switcher

# Install dependencies
go mod download

# Build the binary
make build

# Install system-wide
sudo make install

# Or install to local bin
make install-local
```

#### Development Build

```bash
# Build with debugging symbols
make build-debug

# Run tests
make test

# Build documentation
make docs-build
```

## Verification

After installation, verify that everything is working:

```bash
# Check if the binary is installed
which hypr-input-switcher

# Check version
hypr-input-switcher --version

# Test the configuration
hypr-input-switcher --check-config
```

## Configuration

### Hyprland Integration

Add the following to your Hyprland configuration:

```ini
# ~/.config/hypr/hyprland.conf

# Start automatically with Hyprland
exec-once = hypr-input-switcher

# Optional: Add keybindings for manual control
bind = SUPER, I, exec, hypr-input-switcher --toggle
bind = SUPER_SHIFT, I, exec, hypr-input-switcher --next
```

### Initial Configuration

Create a basic configuration file:

```bash
# Create config directory
mkdir -p ~/.config/hypr-input-switcher

# Copy default config
cp /usr/share/hypr-input-switcher/default.yaml ~/.config/hypr-input-switcher/config.yaml

# Edit the configuration
$EDITOR ~/.config/hypr-input-switcher/config.yaml
```

### Service Integration (Optional)

For automatic startup, you can create a systemd user service:

```ini
# ~/.config/systemd/user/hypr-input-switcher.service
[Unit]
Description=Hypr Input Switcher
After=graphical-session.target

[Service]
Type=simple
ExecStart=/usr/local/bin/hypr-input-switcher --watch
Restart=on-failure
RestartSec=1

[Install]
WantedBy=default.target
```

Enable and start the service:

```bash
# Reload systemd
systemctl --user daemon-reload

# Enable the service
systemctl --user enable hypr-input-switcher.service

# Start the service
systemctl --user start hypr-input-switcher.service
```

## Troubleshooting

### Common Issues

#### Binary not found

Make sure `/usr/local/bin` is in your PATH:

```bash
echo $PATH
export PATH="$PATH:/usr/local/bin"
```

#### Permission denied

Ensure the binary has execute permissions:

```bash
sudo chmod +x /usr/local/bin/hypr-input-switcher
```

#### Fcitx5 not detected

Verify Fcitx5 is running:

```bash
# Check if fcitx5 is running
pgrep fcitx5

# Check available input methods
fcitx5-remote -l
```

### Getting Help

If you encounter issues:

1. Enable debug logging: `hypr-input-switcher --debug`
1. Check existing [GitHub issues](https://github.com/icyleaf/hypr-input-switcher/issues)
1. Create a new issue with logs and system information

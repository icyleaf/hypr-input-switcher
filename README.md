# Hypr Smart Input Go

Hypr Smart Input Go is a Go implementation of an input method switcher designed for the Hyprland window manager. This application allows users to seamlessly switch between different input methods based on the active window, providing a smooth and efficient typing experience.

## Features

- **Dynamic Input Method Switching**: Automatically switches input methods based on the active application window.
- **Configuration Management**: Loads configuration settings from a YAML file, allowing for easy customization.
- **Notification System**: Displays notifications when the input method is switched, enhancing user awareness.
- **Compatibility**: Supports multiple input methods, including Fcitx5 and Rime.

## Project Structure

```
hypr-input-switcher
├── cmd
│   └── hypr-smart-input
│       └── main.go          # Entry point of the application
├── internal
│   ├── config
│   │   ├── config.go       # Configuration structure and loading
│   │   ├── migration.go     # Configuration migration logic
│   │   └── validation.go    # Configuration validation logic
│   ├── inputmethod
│   │   ├── fcitx5.go        # Fcitx5 input method interactions
│   │   ├── rime.go          # Rime input method management
│   │   └── switcher.go      # Input method switching logic
│   ├── notification
│   │   ├── notification.go   # Notification system implementation
│   │   └── backends.go      # Notification backends
│   ├── window
│   │   ├── hyprland.go      # Hyprland window manager interactions
│   │   └── monitor.go       # Window monitoring logic
│   └── app
│       └── app.go           # Main application logic
├── pkg
│   ├── logger
│   │   └── logger.go        # Logging utility
│   └── utils
│       └── utils.go         # Utility functions
├── configs
│   └── default.yaml         # Default configuration settings
├── scripts
│   ├── build.sh             # Build automation script
│   └── install.sh           # Installation script
├── go.mod                   # Go module definition
├── go.sum                   # Module dependency checksums
├── Makefile                 # Build and installation commands
└── README.md                # Project documentation
```

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/your-repo/hypr-input-switcher.git
   cd hypr-input-switcher
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Build the application:
   ```
   make build
   ```

4. Run the application:
   ```
   ./cmd/hypr-smart-input/hypr-smart-input
   ```

## Configuration

The application uses a YAML configuration file located at `configs/default.yaml`. You can customize the input methods and other settings according to your preferences.

## Usage

Once the application is running, it will monitor the active window and switch input methods based on the defined rules in the configuration file. Notifications will be displayed whenever the input method is changed.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.

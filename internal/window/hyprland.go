package window

import (
	"encoding/json"
	"os/exec"
	"strings"
)

// type ClientInfo struct {
// 	Class string `json:"class"`
// 	Title string `json:"title"`
// }

func GetCurrentClient() (*ClientInfo, error) {
	cmd := exec.Command("hyprctl", "activewindow", "-j")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var clientInfo ClientInfo
	if err := json.Unmarshal(output, &clientInfo); err != nil {
		return nil, err
	}

	return &clientInfo, nil
}

func GetCurrentInputMethod() (string, error) {
	cmd := exec.Command("fcitx5-remote", "-n")
	output, err := cmd.Output()
	if err != nil {
		return "unknown", nil
	}

	return strings.TrimSpace(string(output)), nil
}

# Multi-Engine Input Method Switching and Per-App Memory

## Context
Previously, `hypr-input-switcher` assumed `rime` was the only non-English input method backend and equated English with deactivating the fcitx5 controller. When users configured multiple independent input methods (such as `msime` alongside `rime`), state detection treated `msime` as English, and switching to non-Rime methods failed because the daemon tried to set non-existent Rime schemas. Furthermore, the `keep` rule passively inherited the input method from the previously focused window, causing focus transitions through English-enforced windows (like IDEs or browsers) to "infect" all subsequent windows with English.

## Decision
1. **Decouple from Rime**: `input_methods` labels now map directly to fcitx5 input method names (e.g. `msime`, `rime`, `keyboard-us`). The switcher selects engines via fcitx5's `SetCurrentIM(name)`, only invoking Rime schema switching when the selected backend is specifically `rime`.
2. **Switch to English Layout**: Map the `english` label to `keyboard-us` (or configured English layout) rather than deactivating the fcitx5 controller, avoiding unwanted side-effects on controller state.
3. **Per-App Input Method Memory**: Maintain a cache of the last-used Input Method Label keyed by window class (`clientInfo.Class`). When focusing an application with a `keep` rule or no explicit rule, the daemon restores that application's remembered input method rather than inheriting whatever was left by the previous window.
4. **Cold-Start Fallback**: If an application has no explicit rule and has not been visited before (no memory entry), fall back to `default_input_method` (e.g. `chinese`). Only if `default_input_method` itself is `keep` will it passively inherit the active input method.

## Consequences
- Arbitrary fcitx5 engines (MSIME, Rime, Mozc, etc.) are supported seamlessly.
- Navigating between code editors and messaging apps preserves the intended typing mode in each.
- Memory overhead is negligible (proportional to unique application classes opened).

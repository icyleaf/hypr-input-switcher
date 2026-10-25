# hypr-input-switcher

A lightweight daemon that switches input methods based on active window focus in Hyprland.

## Language

**Input Method (IM)**:
A specific input method engine registered in fcitx5, such as `keyboard-us`, `msime`, or `rime`.
_Avoid_: Schema, layout

**Input Method Label**:
A high-level semantic identifier configured by the user (e.g. `english`, `chinese`, `japanese`) mapped to a concrete Input Method.
_Avoid_: Engine name, IM ID

**Rime Schema**:
An internal sub-configuration of the `rime` input method engine representing a particular language or typing system (e.g. `jaroomaji`, `rime_frost`).
_Avoid_: Input method, engine

**Profile**:
A unified sub-mode or schema identifier configured under an input method entity (e.g. `jaroomaji` for Rime, or a specific typing mode for MSIME).
_Avoid_: Engine sub-config, dialect

**Application Rule**:
A pattern-matching rule based on Hyprland window properties (`class` and optional `title`) that dictates the target Input Method Label.
_Avoid_: Window hook, focus handler

**Per-App Memory**:
The cache maintained by the daemon tracking the last active Input Method Label for each application window class.
_Avoid_: Global state, window stack

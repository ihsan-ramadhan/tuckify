# tuckify-gui

Desktop GUI for tuckify, built with Wails v2.

## Requirements

- Go 1.22+
- Node.js 18+ (for the Vite frontend)
- Linux: `webkit2gtk-4.1` (not 4.0 — most modern distros ship 4.1 only)
  - Arch/CachyOS: `pacman -S webkit2gtk-4.1`
  - Debian/Ubuntu: `apt install libwebkit2gtk-4.1-dev`

## Build tags

Since this system only has `webkit2gtk-4.1` (the older `4.0` is deprecated on most
distros), always pass `-tags webkit2_41` to wails commands:

```bash
wails dev -tags webkit2_41
wails build -tags webkit2_41
```

## Development

```bash
wails dev -tags webkit2_41
```

## Build

```bash
wails build -tags webkit2_41
```

Output binary lands in `gui/build/bin/`.

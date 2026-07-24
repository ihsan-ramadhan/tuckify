<div align="center">

# tuckify

**Your Downloads folder is a graveyard. tuckify cleans it automatically.**

Organize files by rule, schedule it, manage multiple folders — without touching them again.

</div>

---

## Install

### Linux & macOS

**CLI:**
```bash
curl -fsSL https://raw.githubusercontent.com/ihsan-ramadhan/tuckify/main/install.sh | sh
```
Installs to `~/.local/bin/tuckify`. Make sure `~/.local/bin` is in your `PATH`.

**GUI:**
```bash
curl -fsSL https://raw.githubusercontent.com/ihsan-ramadhan/tuckify/main/install.sh | sh -s -- --gui
```
- On **macOS**, this downloads and installs `tuckify-gui.app` directly into your `/Applications` folder.
- On **Linux**, this installs the GUI binary to `~/.local/bin/tuckify-gui`, registers the application icon, and creates a desktop shortcut (`~/.local/share/applications/tuckify.desktop`) so it appears in your application launcher/dock.

### Windows

**CLI:**
1. Download `tuckify-windows-amd64.exe` from [Releases](https://github.com/ihsan-ramadhan/tuckify/releases)
2. Rename to `tuckify.exe` and add it to your `PATH`.

**GUI:**
1. Download `tuckify-gui-windows-amd64.exe` from [Releases](https://github.com/ihsan-ramadhan/tuckify/releases)
2. Run the executable to launch the GUI.

### Build from source

Requires Go 1.22+

**CLI:**
```bash
git clone https://github.com/ihsan-ramadhan/tuckify
cd tuckify
go build -o tuckify .
```

**GUI:**
Requires Wails CLI and Node.js:
```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0

# Build GUI binary (output inside build/bin/)
wails build -tags desktop
```

---

## Quick Start

**1. Create a config at `~/.tuckify/rules.toml`:**

```bash
# Linux & macOS
mkdir -p ~/.tuckify
curl -fsSL https://raw.githubusercontent.com/ihsan-ramadhan/tuckify/main/rules.example.toml -o ~/.tuckify/rules.toml
```

```powershell
# Windows (PowerShell)
New-Item -ItemType Directory -Path "$HOME\.tuckify" -Force
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ihsan-ramadhan/tuckify/main/rules.example.toml" -OutFile "$HOME\.tuckify\rules.toml"
```

**2. Run once:**

```bash
tuckify run [folders...]
tuckify run ~/Downloads ~/Desktop
tuckify run                             # runs on all locations defined in rules
tuckify run ~/Downloads --dry-run        # preview without moving
```

---

## Usage

### Commands

| Command | Description |
|---|---|
| `run [folders...]` | Organize files in one or more folders. If no folders are specified, it automatically processes all unique directories defined in rule `locations`. |
| `undo` | Undo the last `tuckify run` (restores moved files) |
| `schedule` | Save a named schedule (accepts one or more folders, use `--start` to run as service, `--run` to run in foreground) |
| `list` | Show all saved schedules and their status |
| `edit` | Update an existing schedule's cron, folder, or config |
| `start` | Activate a saved schedule as a background service |
| `stop` | Deactivate a service (keeps it in the list) |
| `restart` | Stop then start a service (picks up config changes) |
| `logs` | Show service logs (`-f` to follow, `-n` for line count) |
| `delete` | Remove a schedule from the list and stop its service |
| `startup` | Install all saved schedules as system services (survives reboot) |
| `unstartup` | Remove all tuckify system services |
| `uninstall` | Remove binary, services, and optionally config |

### Schedule lifecycle

```bash
# save a schedule and start it as a background service immediately
tuckify schedule downloads ~/Downloads ~/Desktop --cron "0 9 * * *" --start

# or save a schedule first
tuckify schedule downloads ~/Downloads ~/Desktop --cron "0 9 * * *"

# and activate it as a background service later
tuckify start downloads

# check status (shows SERVICE status column)
tuckify list

# install all saved schedules as system services (survives reboot)
tuckify startup

# remove schedule and stop its service
tuckify delete downloads
```

### Cron Expression

```
┌─ minute (0-59)
│ ┌─ hour (0-23)
│ │ ┌─ day of month (1-31)
│ │ │ ┌─ month (1-12)
│ │ │ │ ┌─ day of week (0-6, 0=Sunday)
│ │ │ │ │
* * * * *

Examples:
  "0 9 * * *"      every day at 09:00
  "0 */2 * * *"    every 2 hours
  "0 9 * * 1"      every Monday at 09:00
  "*/30 * * * *"   every 30 minutes
```

---

## Config Reference

Default path: `~/.tuckify/rules.toml`

### Rule Matchers
A rule can use one or more matchers (all are **case-insensitive**):

| Matcher | Example | Matches |
|---|---|---|
| `extensions` | `[".pdf", ".docx"]` | File extensions |
| `filename_patterns` | `["*Modul*", "Invoice_*"]` | Glob patterns (`*` matches any characters) |
| `filename_regex` | `["^invoice_\\d{4}\\.pdf$"]` | Go regular expressions |
| `locations` | `["~/Downloads", "~/Desktop/Inbox"]` | List of folders this rule applies to. If omitted, applies anywhere. |

### Rule Actions
* `action = "move"` (default): Moves the file to `destination`.
* `action = "copy"`: Copies the file to `destination`.
* `action = "delete"`: Deletes the file (no `destination` required).

### Rule Settings
* `conflict_strategy`: How to handle files that already exist at the destination.
  * `"rename"` (default): Appends suffix `_1`, `_2`, etc.
  * `"skip"`: Leaves the file in source, does not move.
  * `"overwrite"`: Overwrites the destination file.
  * `"delete_duplicate"`: Deletes source if SHA-256 matches destination, otherwise falls back to `"rename"`.
  * `"ask"`: Prompts interactively: `[O]verwrite`, `[S]kip`, `[R]ename`.

### Size and Age Filters
* Size: `min_size` / `max_size` (e.g. `"50B"`, `"10KB"`, `"100MB"`, `"1GB"`)
* Age: `min_age` / `max_age` (e.g. `"30d"`, `"1h"`, `"2w"`, `"1y"`)

### Template Tokens
Can be used in `destination` and `rename` paths:
* `{year}`, `{month}`, `{day}`, `{hour}`, `{minute}`, `{second}` (e.g. `~/Archive/{year}`)
* `{base}`: The original filename without extension.
* `{ext}`: The original extension.
* Modifiers (e.g. `{base:slug}`, `{base:lower}`, `{base:upper}`).

See [`rules.example.toml`](rules.example.toml) for complete examples.

### Behavior

- Rules execute **top to bottom**; a file matches the **first matching rule only**.
- Files without extensions can be matched via `filename_patterns` or `filename_regex`.
- Missing destination folders are created automatically.
- When `--recursive` / `-r` is used, empty source subdirectories are automatically cleaned up.
- Deletion rules require interactive confirmation during manual runs, bypassable via `--yes` / `-y`.

---

## Uninstall

```bash
tuckify uninstall
# or if you only installed the GUI:
tuckify-gui uninstall
```

Removes the binary, system services, and desktop integration files (Linux `.desktop` and icons). Prompts whether to also delete `~/.tuckify/`.

---

## License

MIT

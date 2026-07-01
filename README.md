<div align="center">

# tuckify

**Your Downloads folder is a graveyard. tuckify cleans it automatically.**

Organize files by rule, schedule it, manage multiple folders — without touching them again.

</div>

---

## Install

### Linux & macOS

```bash
curl -fsSL https://raw.githubusercontent.com/ihsan-ramadhan/tuckify/main/install.sh | sh
```

Installs to `~/.local/bin/tuckify`. Make sure `~/.local/bin` is in your `PATH`.

### Windows

1. Download `tuckify-windows-amd64.exe` from [Releases](https://github.com/ihsan-ramadhan/tuckify/releases)
2. Rename to `tuckify.exe`
3. Add its folder to your `PATH`

### Build from source

Requires Go 1.22+

```bash
git clone https://github.com/ihsan-ramadhan/tuckify
cd tuckify
go build -o tuckify .
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

**2. Run:**

```bash
# organize once
tuckify run ~/Downloads

# preview without moving files
tuckify run ~/Downloads --dry-run

# save a schedule
tuckify schedule downloads ~/Downloads --cron "0 9 * * *"

# activate as a background service
tuckify start downloads

# or save + test interactively in one command
tuckify schedule downloads ~/Downloads --cron "0 9 * * *" --run

# check status
tuckify list
```

---

## Usage

```
tuckify run <folder> [--dry-run] [--config <path>]
tuckify schedule <name> <folder> --cron "<expr>" [--run] [--config <path>]
tuckify list
tuckify start <name>
tuckify stop <name>
tuckify restart <name>
tuckify logs <name> [-f] [-n <lines>]
tuckify delete <name>
tuckify startup
tuckify unstartup
tuckify uninstall
```

### Commands

| Command | Description |
|---|---|
| `run` | Organize files once |
| `schedule` | Save a named schedule (`--run` to also start interactively) |
| `list` | Show all saved schedules and their status |
| `start` | Activate a saved schedule as a background service |
| `stop` | Deactivate a service (keeps it in the list) |
| `restart` | Stop then start a service (picks up config changes) |
| `logs` | Show service logs (`-f` to follow, `-n` for line count) |
| `delete` | Remove a schedule from the list and stop its service |
| `startup` | Install all saved schedules as system services (survives reboot) |
| `unstartup` | Remove all tuckify system services |
| `uninstall` | Remove binary, services, and optionally config |

### Workflow

```
# 1. save a schedule
tuckify schedule downloads ~/Downloads --cron "0 9 * * *"
# saved schedule "downloads"
#   run 'tuckify start downloads' to activate as a background service

# 2. activate it as a background service
tuckify start downloads

# 3. check status
tuckify list
#  NAME               │ STATUS   │ SAVED  │ CRON           │ FOLDER
# ────────────────────┼──────────┼────────┼────────────────┼──────────────────────
#  downloads          │ online   │ yes    │ 0 9 * * *      │ /home/user/Downloads

# 4. survive reboots
tuckify startup

# 5. remove
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

```toml
[settings]
conflict_strategy = "rename"   # "rename" | "skip" | "overwrite"

[[rule]]
name        = "By extension"
extensions  = [".pdf", ".docx"]
destination = "~/Documents"

[[rule]]
name              = "By filename"
filename_patterns = ["*Modul*", "Invoice_*"]
destination       = "~/Documents/Sorted"
```

A rule can have `extensions`, `filename_patterns`, or both — a file matches if either condition is met.

Filename patterns use glob syntax (`*` matches any characters, case-insensitive):
- `"*Modul*"` — any file containing "Modul"
- `"Invoice_*"` — any file starting with "Invoice_"
- `"*_2024.*"` — any file with "_2024" before the extension

### Behavior

- Rules run **top to bottom**, file matches **first rule only**
- Extension matching is **case-insensitive** (`.PDF` == `.pdf`)
- Filename pattern matching is **case-insensitive**
- Files without an extension can match via `filename_patterns`
- Missing destination folders are created automatically
- Default conflict strategy `rename`: appends `_1`, `_2`, etc.

---

## Uninstall

```bash
tuckify uninstall
```

Removes the binary and services. Prompts whether to also delete `~/.tuckify/`.

---

## License

MIT

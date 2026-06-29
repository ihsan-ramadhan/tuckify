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

# schedule it (saves to list + runs interactively)
tuckify schedule downloads ~/Downloads --cron "0 9 * * *"

# activate as a background service
tuckify start downloads

# check status
tuckify list
```

---

## Usage

```
tuckify run <folder> [--dry-run] [--config <path>]
tuckify schedule <name> <folder> --cron "<expr>" [--config <path>]
tuckify list
tuckify start <name>
tuckify stop <name>
tuckify delete <name>
tuckify startup
tuckify unstartup
tuckify uninstall
```

### Commands

| Command | Description |
|---|---|
| `run` | Organize files once |
| `schedule` | Save a named schedule and run it interactively |
| `list` | Show all saved schedules and their status |
| `start` | Activate a saved schedule as a background service |
| `stop` | Deactivate a service (keeps it in the list) |
| `delete` | Remove a schedule from the list and stop its service |
| `startup` | Install all saved schedules as system services (survives reboot) |
| `unstartup` | Remove all tuckify system services |
| `uninstall` | Remove binary, services, and optionally config |

### Workflow

```
# 1. define a schedule (auto-saved to list)
tuckify schedule downloads ~/Downloads --cron "0 9 * * *"

# 2. activate it
tuckify start downloads

# 3. check
tuckify list
# NAME                 STATUS     CRON            FOLDER
# downloads            online     0 9 * * *       /home/user/Downloads

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
name        = "Rule name (optional, for logging)"
extensions  = [".pdf", ".docx"]
destination = "~/Documents"
```

### Behavior

- Rules run **top to bottom**, file matches **first rule only**
- Extension matching is **case-insensitive** (`.PDF` == `.pdf`)
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

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

**2. Run once:**

```bash
tuckify run ~/Downloads
tuckify run ~/Downloads --dry-run   # preview without moving
```

---

## Usage

### Commands

| Command | Description |
|---|---|
| `run` | Organize files once |
| `schedule` | Save a named schedule (`--run` to also start interactively) |
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
# save a schedule
tuckify schedule downloads ~/Downloads --cron "0 9 * * *"

# activate as a background service
tuckify start downloads

# check status
tuckify list

# install all saved schedules as system services (survives reboot)
tuckify startup

# remove
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
```

Removes the binary and services. Prompts whether to also delete `~/.tuckify/`.

---

## License

MIT

# tuckify

> automatic file organizer with scheduling — cross-platform CLI

tuckify organizes files in a folder based on rules you define. Run it once, schedule it, or register it as a startup service.

---

## Install

### Linux & macOS

```bash
curl -fsSL https://raw.githubusercontent.com/ihsan-ramadhan/tuckify/main/install.sh | sh
```

Installs the binary to `~/.local/bin/tuckify`. Make sure `~/.local/bin` is in your `PATH`.

### Windows

1. Download the latest `tuckify-windows-amd64.exe` from the [Releases](https://github.com/ihsan-ramadhan/tuckify/releases) page.
2. Rename the binary to `tuckify.exe`.
3. Add the folder containing `tuckify.exe` to your system's `PATH` environment variable.

### Build from source

Requires Go 1.22+

```bash
git clone https://github.com/ihsan-ramadhan/tuckify
cd tuckify
go build -o tuckify .
```

---

## Quick Start

**1. Create a config file at `~/.tuckify/rules.toml`:**

### Linux & macOS
```bash
mkdir -p ~/.tuckify
curl -fsSL https://raw.githubusercontent.com/ihsan-ramadhan/tuckify/main/rules.example.toml -o ~/.tuckify/rules.toml
```

### Windows (PowerShell)
```powershell
New-Item -ItemType Directory -Path "$HOME\.tuckify" -Force
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ihsan-ramadhan/tuckify/main/rules.example.toml" -OutFile "$HOME\.tuckify\rules.toml"
```

### Windows (CMD)
```cmd
mkdir %USERPROFILE%\.tuckify
curl -fsSL https://raw.githubusercontent.com/ihsan-ramadhan/tuckify/main/rules.example.toml -o %USERPROFILE%\.tuckify\rules.toml
```

**2. Run:**

### Linux & macOS
```bash
# organize once (replace ~/Downloads with your target folder path)
tuckify run ~/Downloads

# preview without moving files
tuckify run ~/Downloads --dry-run

# run on a schedule (long-running process)
tuckify schedule ~/Downloads --cron "0 9 * * *"

# register as a startup service
tuckify init ~/Downloads --cron "0 9 * * *"
```

### Windows (PowerShell & CMD)
```powershell
# organize once (replace "$HOME\Downloads" with your target folder path)
tuckify run "$HOME\Downloads"

# preview without moving files
tuckify run "$HOME\Downloads" --dry-run

# run on a schedule (long-running process)
tuckify schedule "$HOME\Downloads" --cron "0 9 * * *"

# register as a startup service
tuckify init "$HOME\Downloads" --cron "0 9 * * *"
```

---

## Usage

```
tuckify run <folder> [--config <path>] [--dry-run]
tuckify schedule <folder> --cron "<expr>" [--config <path>]
tuckify init <folder> --cron "<expr>" [--config <path>]
tuckify uninit
tuckify uninstall
tuckify --help
tuckify --version
```

### Commands

| Command | Description |
|---|---|
| `run` | Organize files once |
| `schedule` | Run organizer on a cron schedule (stays running) |
| `init` | Register tuckify as a startup service (systemd/launchd/schtasks) |
| `uninit` | Remove tuckify from startup |
| `uninstall` | Remove binary, service, and optionally config |

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

- Rules execute **top to bottom**, file matches the **first rule only**
- Extension matching is **case-insensitive** (`.PDF` == `.pdf`)
- Missing destination folders are created automatically
- Default conflict strategy `rename`: appends suffix `_1`, `_2`, etc.

---

## Uninstall

```bash
tuckify uninstall
```

Removes the binary and service. Prompts whether to also delete `~/.tuckify/`.

---

## License

MIT

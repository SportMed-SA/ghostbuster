# 👻 Ghostbuster

Ghostbuster is a simple command-line tool that helps you keep translation files clean and complete. It scans nested JSON translations, checks your frontend source code for referenced keys, reports what is safe to clean up, and identifies missing translations.

> [!IMPORTANT]
> Ghostbuster relies on static analysis of source files with prefixed translation keys (e.g. `_globalTranslation.key`). It may not catch dynamically constructed keys or keys used in non-standard ways. These scenarios are currently not supported, so please be careful when using Ghostbuster in such cases and review the detected unused keys before removal.

## Why Ghostbuster? 👣

- Fast cleanup of stale translation keys.
- Detection of referenced keys that are missing globally or from individual translation files.
- Safer maintenance with backup and restore support.
- Friendly interactive mode for local/manual use.
- Scriptable command mode for CI pipelines.
- Designed to keep large i18n files from becoming haunted by dead entries.

## Installation 📦

### Option 1: Install from source (recommended)

From the repository root, run:

```bash
go install ./cmd/ghostbuster
```

This installs the `ghostbuster` binary into your Go bin directory (usually `%USERPROFILE%\\go\\bin` on Windows).
Make sure that directory is in your `PATH`.

### Option 2: Build a local binary

```bash
go build -o ghostbuster ./cmd/ghostbuster
```

Then run it from the current directory:

```bash
./ghostbuster
```

### Option 3: Run without installing

```bash
go run ./cmd/ghostbuster
```

## Quick Start

### Interactive mode 🕯️

Run the root command with no arguments:

```bash
ghostbuster
```

This opens the step-by-step Terminal UI (TUI), where you can choose all operations interactively, without needing to remember flags or subcommands. The main menu includes:

- `detect`
- `hunt`
- `restore`
- `help`

Each flow ends with a command preview and confirmation before execution, so nothing runs unexpectedly.

Note: in non-interactive contexts, bare `ghostbuster` returns an explicit non-zero error. Use subcommands directly for CI/automation.

### Command (CI/script) mode ⚙️

Detect unused and missing keys:

```bash
ghostbuster detect --translations ./public/assets/i18n --source ./src
```

Remove unused keys:

```bash
ghostbuster hunt --translations ./public/assets/i18n --source ./src
```

Restore from backup:

```bash
ghostbuster restore --translations ./public/assets/i18n
```

## Commands

### `detect`

Find translation keys that exist in JSON files but are not referenced in source files, referenced keys that do not exist in any translation file, and referenced keys that are missing from individual translation files. The command exits with code `2` when it finds any of these issues, making it suitable for CI audits before release.

```bash
ghostbuster detect \
	--translations ./public/assets/i18n \
	--source ./src \
	--format text
```

### `hunt`

Find and remove unused translation keys. Missing referenced keys and incomplete translation files are reported but are never created or otherwise modified automatically. The command exits with code `2` if it removes unused keys or reports missing translations.

By default, Ghostbuster creates timestamped backups before writing changes, so you can safely roll back if needed:

```text
<file>.bak.<unix-nano>
```

Example:

```bash
ghostbuster hunt \
	--translations ./public/assets/i18n \
	--source ./src
```

Disable backup creation (not recommended unless you have another rollback strategy):

```bash
ghostbuster hunt --translations ./public/assets/i18n --source ./src --no-backup
```

### `restore`

Restore translation files from backups created by `hunt`.

The restore flow checks for available backups per translation file and restores from the best available candidate.

```bash
ghostbuster restore --translations ./public/assets/i18n --format text
```

## Typical workflow 🔎

1. Run `detect` to review unused and missing keys.
2. Run `hunt` to remove them (with timestamped backups).
3. If needed, run `restore` to roll back translation files.
4. Commit clean translation files once everything looks good.

## Example

This repository includes an example project written in Angular with `ngx-translate`. If you want to try out Ghostbuster without setting up your own project, feel free to clone this repo and run the commands in the `example` folder:.

## Help

```bash
ghostbuster --help
ghostbuster detect --help
ghostbuster hunt --help
ghostbuster restore --help
```

# ghostbuster

Ghostbuster is a simple command-line tool to find unused translation keys in your project.

## Interactive mode

Run `ghostbuster` without arguments to open the interactive TUI (powered by go-huh).
The interactive flow supports `detect`, `hunt`, `restore`, and `help`, shows a command preview, and asks for confirmation before execution.

In non-interactive contexts, bare `ghostbuster` exits with an explicit error. Use subcommands directly for automation/CI.

## Commands

### Detect unused keys

```bash
ghostbuster detect --translations ./public/assets/i18n --source ./src
```

### Hunt (remove) unused keys

By default, hunt creates timestamped backup files (`.bak.<unix-nano>`) before writing changes.

```bash
ghostbuster hunt --translations ./public/assets/i18n --source ./src
```

Skip backup creation:

```bash
ghostbuster hunt --translations ./public/assets/i18n --source ./src --no-backup
```

### Restore from backup

Restore command looks for backup files created by `hunt` (`.bak` and timestamped `.bak.*`) and restores each translation file from the best available backup.

```bash
ghostbuster restore --translations ./public/assets/i18n
```

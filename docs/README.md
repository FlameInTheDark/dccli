# dccli Documentation

Command reference for dccli - Discord CLI Tool.

## Quick Reference

- [Command Reference](commands.md) - Complete list of all commands

## Configuration

Config file location:
- **Linux/macOS**: `~/.dccli/config.yaml`
- **Windows**: `%USERPROFILE%\.dccli\config.yaml`

Example config:
```yaml
current-bot: mybot
bots:
  - name: mybot
    bot:
      token: your_bot_token_here
```

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output format: `table`, `json`, `yaml` |
| `--bot` | `-b` | Bot name to use |
| `--token` | `-t` | Bot token (overrides config) |

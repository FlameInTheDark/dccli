# dccli - Discord CLI Tool

A command-line interface for managing Discord bots and guilds.

## Installation

### Binary (One-liner)

**Linux / macOS**:
```bash
curl -sL https://github.com/FlameInTheDark/dccli/releases/latest/download/dccli-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64 -o dccli && chmod +x dccli && sudo mv dccli /usr/local/bin/
```

**Windows (PowerShell)**:
```powershell
iwr -useb https://github.com/FlameInTheDark/dccli/releases/latest/download/dccli-windows-amd64.exe -OutFile dccli.exe
```

### From Source

Using Go:
```bash
go install github.com/FlameInTheDark/dccli/cmd/dccli@latest
```

## Quick Start

1. **Add your bot:**
   ```bash
   dccli config bot add mybot YOUR_BOT_TOKEN
   ```

2. **List your guilds:**
   ```bash
   dccli guilds list
   ```

3. **Get help:**
   ```bash
   dccli --help
   dccli guilds --help
   ```

## Configuration

Config file location:
- **Linux/macOS**: `~/.dccli/config.yaml`
- **Windows**: `%USERPROFILE%\.dccli\config.yaml`

Config format:
```yaml
current-bot: mybot
bots:
  - name: mybot
    bot:
      token: your_bot_token_here
```

Environment variables:
- `DCLI_OUTPUT` - Default output format (`table`, `json`, `yaml`)
- `DCLI_BOT` - Default bot name to use
- `DCLI_TOKEN` - Bot token (overrides config)

## Command Overview

| Command | Description |
|---------|-------------|
| `config` | Manage bot configurations |
| `guilds` | Guild management |
| `channels` | Channel operations |
| `messages` | Message operations |
| `roles` | Role management |
| `members` | Member management |
| `users` | User/bot info |
| `webhooks` | Webhook management |
| `emoji` | Emoji management |
| `stickers` | Sticker management |
| `events` | Scheduled events |
| `automod` | Auto-moderation rules |
| `voice` | Voice regions |
| `invites` | Invite management |
| `applications` | Application commands |
| `completion` | Shell completion scripts |

## Global Flags

| Flag | Description |
|------|-------------|
| `-o, --output` | Output format: `table`, `json`, `yaml` |
| `-b, --bot` | Bot name to use |
| `-t, --token` | Bot token (overrides config) |

## Examples

```bash
# List guilds
dccli guilds list

# Get guild details
dccli guilds describe GUILD_ID

# List channels
dccli channels list --guild GUILD_ID

# Send a message
dccli messages send CHANNEL_ID --content "Hello!"

# List members
dccli members list --guild GUILD_ID

# Output as JSON
dccli -o json guilds list
```

## Documentation

See [docs/README.md](docs/README.md) for detailed command reference.

## License

MIT

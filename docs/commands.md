# Command Reference

## Global Flags

| Flag | Short | Description | Environment Variable |
|------|-------|-------------|---------------------|
| `--output` | `-o` | Output format: `table`, `json`, `yaml` | `DCLI_OUTPUT` |
| `--bot` | `-b` | Bot name to use | `DCLI_BOT` |
| `--token` | `-t` | Bot token (overrides config) | `DCLI_TOKEN` |
| `--quiet` | `-q` | Suppress status messages | |

---

## Config Commands

Manage bot configurations.

### config bot add
Add a bot to configuration.

```bash
dccli config bot add <name> <token>
```

### config bot set
Set the current bot.

```bash
dccli config bot set <name>
```

### config bot list
List configured bots.

```bash
dccli config bot list [--tokens]
```

### config bot remove
Remove a bot from configuration.

```bash
dccli config bot remove <name> [--force]
```

### config bot edit
Edit a bot's token.

```bash
dccli config bot edit <name> --token <new-token>
```

### config validate
Validate configuration file.

```bash
dccli config validate
```

---

## Guild Commands

### guilds list
List guilds the bot is in.

```bash
dccli guilds list [--limit 10] [--before ID] [--after ID]
```

### guilds describe
Get detailed guild information.

```bash
dccli guilds describe <guild-id>
```

### guilds edit
Edit guild properties.

```bash
dccli guilds edit <guild-id> [--name <name>] [--region <region>] [--verification <level>]
```

### guilds leave
Make the bot leave a guild.

```bash
dccli guilds leave <guild-id> [--force]
```

### guilds channels
Manage guild channels (list, create, edit, delete).

```bash
dccli guilds channels list <guild-id>
dccli guilds channels create <guild-id> --name <name> [--type <type>]
dccli guilds channels edit <channel-id> [--name <name>]
dccli guilds channels delete <channel-id> [--force]
```

### guilds roles
Manage guild roles (list, create, edit, delete, assign).

```bash
dccli guilds roles list <guild-id>
dccli guilds roles create <guild-id> --name <name> [--color <hex>]
dccli guilds roles edit <role-id> --guild <guild-id> [--name <name>]
dccli guilds roles delete <role-id> --guild <guild-id> [--force]
dccli guilds roles assign <role-id> --guild <guild-id> --user <user-id>
dccli guilds roles remove <role-id> --guild <guild-id> --user <user-id>
```

### guilds members
Manage guild members (list, kick, ban, timeout).

```bash
dccli guilds members list <guild-id> [--limit 100]
dccli guilds members get <user-id> --guild <guild-id>
dccli guilds members kick <user-id> --guild <guild-id> [--force]
dccli guilds members ban <user-id> --guild <guild-id> [--reason <reason>] [--force]
dccli guilds members unban <user-id> --guild <guild-id> [--force]
dccli guilds members timeout <user-id> --guild <guild-id> --duration <duration>
```

### guilds invites
Manage guild invites.

```bash
dccli guilds invites list <guild-id>
dccli guilds invites create <channel-id> [--max-age <seconds>] [--max-uses <count>]
dccli guilds invites delete <invite-code> [--force]
```

---

## Channel Commands

### channels list
List channels in a guild.

```bash
dccli channels list --guild <guild-id>
```

### channels get
Get channel details.

```bash
dccli channels get <channel-id>
```

### channels create
Create a new channel.

```bash
dccli channels create --guild <guild-id> --name <name> [--type <type>] [--parent <parent-id>]
```

### channels edit
Edit channel settings.

```bash
dccli channels edit <channel-id> [--name <name>] [--topic <topic>] [--nsfw <bool>]
```

### channels delete
Delete a channel.

```bash
dccli channels delete <channel-id> [--force]
```

### channels messages
Manage channel messages.

```bash
dccli channels messages list <channel-id> [--limit 50]
dccli channels messages get <channel-id> <message-id>
dccli channels messages send <channel-id> --content <text> [--tts]
dccli channels messages edit <channel-id> <message-id> --content <text>
dccli channels messages delete <channel-id> <message-id> [--force]
dccli channels messages bulk-delete <channel-id> --messages <id1,id2,...> [--force]
```

### channels webhooks
Manage channel webhooks.

```bash
dccli channels webhooks list <channel-id>
dccli channels webhooks create <channel-id> --name <name> [--avatar <file>]
dccli channels webhooks edit <webhook-id> [--name <name>]
dccli channels webhooks delete <webhook-id> [--force]
```

---

## Message Commands

### messages get
Get a specific message.

```bash
dccli messages get <channel-id> <message-id>
```

### messages listen
Listen for new messages in a channel.

```bash
dccli messages listen <channel-id> [--mentions] [--users <id1,id2...>] [--limit <count>] [--format <text|json>]
```
Prints new messages to stdout in the format: `UserName (user_id): message-text`.
Attachments are listed as URLs.
If `--mentions` is used, only messages that mention the bot will be displayed.
If `--users` is used, only messages from the specified user IDs will be displayed.
If `--limit` is used, the command will exit after receiving the specified number of matching messages.
If `--format json` is used, messages will be printed as JSON objects.

### messages send
Send a message.

```bash
dccli messages send <channel-id> --content <text> [--tts] [--embed-file <file>]
```

### messages edit
Edit a message.

```bash
dccli messages edit <channel-id> <message-id> --content <text> [--embed-file <file>]
```

### messages delete
Delete a message.

```bash
dccli messages delete <channel-id> <message-id> [--force]
```

### messages reactions
Manage message reactions.

```bash
dccli messages reactions list <channel-id> <message-id>
dccli messages reactions add <channel-id> <message-id> <emoji>
dccli messages reactions remove <channel-id> <message-id> <emoji> [--user <user-id>]
```

---

## Role Commands

### roles list
List roles in a guild.

```bash
dccli roles list --guild <guild-id>
```

### roles get
Get role details.

```bash
dccli roles get <role-id> --guild <guild-id>
```

### roles create
Create a role.

```bash
dccli roles create --guild <guild-id> --name <name> [--color <hex>] [--hoist] [--mentionable]
```

### roles edit
Edit a role.

```bash
dccli roles edit <role-id> --guild <guild-id> [--name <name>] [--color <hex>] [--permissions <perms>]
```

### roles delete
Delete a role.

```bash
dccli roles delete <role-id> --guild <guild-id> [--force]
```

### roles assign
Assign a role to a member.

```bash
dccli roles assign <role-id> --guild <guild-id> --user <user-id>
```

### roles remove
Remove a role from a member.

```bash
dccli roles remove <role-id> --guild <guild-id> --user <user-id>
```

---

## Member Commands

### members list
List members in a guild.

```bash
dccli members list --guild <guild-id> [--limit 100] [--after <user-id>]
```

### members get
Get member details.

```bash
dccli members get <user-id> --guild <guild-id>
```

### members ban
Ban a member.

```bash
dccli members ban <user-id> --guild <guild-id> [--reason <reason>] [--delete-days <days>] [--force]
```

### members unban
Unban a user.

```bash
dccli members unban <user-id> --guild <guild-id> [--force]
```

### members kick
Kick a member.

```bash
dccli members kick <user-id> --guild <guild-id> [--reason <reason>] [--force]
```

### members timeout
Timeout a member.

```bash
dccli members timeout <user-id> --guild <guild-id> --duration <duration>
```

### members nick
Set a member's nickname.

```bash
dccli members nick <user-id> --guild <guild-id> --nick <nickname>
```

---

## User Commands

### users get
Get current user/bot info.

```bash
dccli users get
```

### users guilds
List user guilds.

```bash
dccli users guilds [--limit 10]
```

### users connections
Get user connections.

```bash
dccli users connections
```

---

## Webhook Commands

### webhooks list
List webhooks.

```bash
dccli webhooks list --channel <channel-id>
dccli webhooks list --guild <guild-id>
```

### webhooks get
Get webhook details.

```bash
dccli webhooks get <webhook-id>
```

### webhooks create
Create a webhook.

```bash
dccli webhooks create --channel <channel-id> --name <name> [--avatar <file>]
```

### webhooks edit
Edit a webhook.

```bash
dccli webhooks edit <webhook-id> --channel <channel-id> [--name <name>] [--avatar <file>]
```

### webhooks delete
Delete a webhook.

```bash
dccli webhooks delete <webhook-id> [--force]
```

### webhooks execute
Execute a webhook.

```bash
dccli webhooks execute <webhook-id> --content <text> [--username <name>] [--avatar-url <url>]
```

---

## Emoji Commands

### emoji list
List emojis in a guild.

```bash
dccli emoji list --guild <guild-id>
```

### emoji get
Get emoji details.

```bash
dccli emoji get <emoji-id> --guild <guild-id>
```

### emoji create
Create an emoji.

```bash
dccli emoji create --guild <guild-id> --name <name> --image <file> [--roles <role-ids>]
```

### emoji edit
Edit an emoji.

```bash
dccli emoji edit <emoji-id> --guild <guild-id> [--name <name>] [--roles <role-ids>]
```

### emoji delete
Delete an emoji.

```bash
dccli emoji delete <emoji-id> --guild <guild-id> [--force]
```

---

## Sticker Commands

### stickers list
List stickers in a guild.

```bash
dccli stickers list --guild <guild-id>
```

### stickers get
Get sticker details.

```bash
dccli stickers get <sticker-id> --guild <guild-id>
```

### stickers create
Create a sticker.

```bash
dccli stickers create --guild <guild-id> --name <name> --image <file> [--tags <tags>] [--description <desc>]
```

### stickers edit
Edit a sticker.

```bash
dccli stickers edit <sticker-id> --guild <guild-id> [--name <name>] [--tags <tags>]
```

### stickers delete
Delete a sticker.

```bash
dccli stickers delete <sticker-id> --guild <guild-id> [--force]
```

---

## Event Commands

### events list
List scheduled events.

```bash
dccli events list --guild <guild-id>
```

### events get
Get event details.

```bash
dccli events get <event-id> --guild <guild-id>
```

### events create
Create a scheduled event.

```bash
dccli events create --guild <guild-id> --name <name> --start <time> [--end <time>] [--type <type>]
```

### events edit
Edit a scheduled event.

```bash
dccli events edit <event-id> --guild <guild-id> [--name <name>] [--start <time>] [--end <time>]
```

### events delete
Delete a scheduled event.

```bash
dccli events delete <event-id> --guild <guild-id> [--force]
```

### events users
List users subscribed to an event.

```bash
dccli events users <event-id> --guild <guild-id>
```

---

## AutoMod Commands

### automod list
List auto-moderation rules.

```bash
dccli automod list --guild <guild-id>
```

### automod get
Get rule details.

```bash
dccli automod get <rule-id> --guild <guild-id>
```

### automod create
Create a rule.

```bash
dccli automod create --guild <guild-id> --name <name> --type <type> --actions <json> [--keywords <json>]
```

### automod edit
Edit a rule.

```bash
dccli automod edit <rule-id> --guild <guild-id> [--name <name>] [--actions <json>] [--enabled <bool>]
```

### automod delete
Delete a rule.

```bash
dccli automod delete <rule-id> --guild <guild-id> [--force]
```

---

## Voice Commands

### voice regions
List voice regions.

```bash
dccli voice regions
```

### voice join
Join a voice channel.

```bash
dccli voice join <channel-id> [guild-id]
```

### voice leave
Leave a voice channel.

```bash
dccli voice leave <guild-id>
```

### voice record
Record audio from a voice channel.

```bash
dccli voice record <channel-id> [guild-id] [--out <output-path>]
```
Records audio to OGG files. Each user's audio is saved to a separate file with the naming convention:
`<timestamp>_<SSRC>_<output>.ogg`

### voice play
Play an audio file in a voice channel.

```bash
dccli voice play <channel-id> [guild-id] --file <file-path>
```
Supports any audio format that ffmpeg can decode (MP3, WAV, OGG, FLAC, etc.).
The bot will join, play the file, and then leave the channel.

---

## Invite Commands

### invites get
Get invite details.

```bash
dccli invites get <invite-code>
```

### invites list
List invites.

```bash
dccli invites list --guild <guild-id>
dccli invites list --channel <channel-id>
```

### invites delete
Delete an invite.

```bash
dccli invites delete <invite-code> [--force]
```

---

## Application Commands

### applications commands list
List application commands.

```bash
dccli applications commands list [--guild <guild-id>]
```

### applications commands get
Get command details.

```bash
dccli applications commands get --command <command-id> [--guild <guild-id>]
```

### applications commands create
Create a command.

```bash
dccli applications commands create --name <name> --description <desc> [--guild <guild-id>] [--options-file <file>]
```

### applications commands edit
Edit a command.

```bash
dccli applications commands edit --command <command-id> [--guild <guild-id>] [--name <name>] [--description <desc>]
```

### applications commands delete
Delete a command.

```bash
dccli applications commands delete --command <command-id> [--guild <guild-id>]
```

### applications commands delete-all
Delete all commands.

```bash
dccli applications commands delete-all [--guild <guild-id>] [--force]
```

### applications commands bulk-overwrite
Overwrite all commands from JSON file.

```bash
dccli applications commands bulk-overwrite --file <json-file> [--guild <guild-id>]
```

---

## Completion Commands

Generate shell completion scripts for bash, zsh, fish, and PowerShell.

### completion bash

Generate bash completion script.

```bash
# Print to stdout
dccli completion bash

# Save to file
dccli completion bash -o dccli-completion.bash

# Install system-wide (Linux)
dccli completion bash | sudo tee /etc/bash_completion.d/dccli > /dev/null

# Install for current user (Linux/macOS)
mkdir -p ~/.local/share/bash-completion/completions
dccli completion bash > ~/.local/share/bash-completion/completions/dccli
```

### completion zsh

Generate zsh completion script.

```bash
# Print to stdout
dccli completion zsh

# Save to file
dccli completion zsh -o _dccli

# Install system-wide
dccli completion zsh | sudo tee /usr/local/share/zsh/site-functions/_dccli > /dev/null

# Install for Oh-My-Zsh
mkdir -p ~/.oh-my-zsh/completions
dccli completion zsh > ~/.oh-my-zsh/completions/_dccli
```

### completion fish

Generate fish completion script.

```bash
# Print to stdout
dccli completion fish

# Save to file
dccli completion fish -o dccli.fish

# Install for current user
mkdir -p ~/.config/fish/completions
dccli completion fish > ~/.config/fish/completions/dccli.fish
```

### completion powershell

Generate PowerShell completion script.

```bash
# Print to stdout
dccli completion powershell

# Save to file
dccli completion powershell -o dccli-completion.ps1

# Add to PowerShell profile
Add-Content $PROFILE (dccli completion powershell)
```

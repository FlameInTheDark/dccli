package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
)

func CompletionCommand() *cli.Command {
	return &cli.Command{
		Name:        "completion",
		Usage:       "Generate shell completion scripts",
		Description: "Generate shell completion scripts for bash, zsh, fish, or powershell",
		Commands: []*cli.Command{
			completionBashCommand(),
			completionZshCommand(),
			completionFishCommand(),
			completionPowerShellCommand(),
		},
	}
}

func completionBashCommand() *cli.Command {
	return &cli.Command{
		Name:        "bash",
		Usage:       "Generate bash completion script",
		Description: "Generate a bash completion script for dccli",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file (default: stdout)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			script := generateBashCompletion()
			return writeCompletion(script, c.String("output"))
		},
	}
}

func completionZshCommand() *cli.Command {
	return &cli.Command{
		Name:        "zsh",
		Usage:       "Generate zsh completion script",
		Description: "Generate a zsh completion script for dccli",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file (default: stdout)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			script := generateZshCompletion()
			return writeCompletion(script, c.String("output"))
		},
	}
}

func completionFishCommand() *cli.Command {
	return &cli.Command{
		Name:        "fish",
		Usage:       "Generate fish completion script",
		Description: "Generate a fish completion script for dccli",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file (default: stdout)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			script := generateFishCompletion()
			return writeCompletion(script, c.String("output"))
		},
	}
}

func completionPowerShellCommand() *cli.Command {
	return &cli.Command{
		Name:        "powershell",
		Usage:       "Generate PowerShell completion script",
		Description: "Generate a PowerShell completion script for dccli",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file (default: stdout)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			script := generatePowerShellCompletion()
			return writeCompletion(script, c.String("output"))
		},
	}
}

func writeCompletion(script, outputPath string) error {
	if outputPath == "" {
		fmt.Print(script)
		return nil
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	_, err = io.WriteString(file, script)
	if err != nil {
		return fmt.Errorf("failed to write completion script: %w", err)
	}

	fmt.Printf("Completion script written to: %s\n", outputPath)
	return nil
}

func generateBashCompletion() string {
	return `# dccli bash completion script
# Source this file: source <(dccli completion bash)
# Or add to ~/.bashrc: eval "$(dccli completion bash)"

_dccli_complete() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Main commands
    local commands="applications guilds channels messages roles members webhooks users emoji stickers events automod voice invites config completion version help"

    # Global flags
    local global_flags="--output -o --bot -b --token -t --help -h --version -v"

    case "${COMP_CWORD}" in
        1)
            COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
            return 0
            ;;
        2)
            case "${prev}" in
                applications)
                    COMPREPLY=( $(compgen -W "commands" -- ${cur}) )
                    ;;
                guilds)
                    COMPREPLY=( $(compgen -W "list describe edit leave channels roles members invites" -- ${cur}) )
                    ;;
                channels)
                    COMPREPLY=( $(compgen -W "list describe create edit delete typing permissions invites webhooks messages pins" -- ${cur}) )
                    ;;
                messages)
                    COMPREPLY=( $(compgen -W "send edit delete get react unreact reactions pin unpin crosspost reply purge" -- ${cur}) )
                    ;;
                roles)
                    COMPREPLY=( $(compgen -W "list create edit delete assign unassign" -- ${cur}) )
                    ;;
                members)
                    COMPREPLY=( $(compgen -W "list describe edit kick ban unban bans timeout untimeout prune nick add-role remove-role" -- ${cur}) )
                    ;;
                webhooks)
                    COMPREPLY=( $(compgen -W "list create edit delete send" -- ${cur}) )
                    ;;
                users)
                    COMPREPLY=( $(compgen -W "me get avatar dm" -- ${cur}) )
                    ;;
                emoji)
                    COMPREPLY=( $(compgen -W "list create delete edit" -- ${cur}) )
                    ;;
                stickers)
                    COMPREPLY=( $(compgen -W "list get create delete edit" -- ${cur}) )
                    ;;
                events)
                    COMPREPLY=( $(compgen -W "list create edit delete users start end" -- ${cur}) )
                    ;;
                automod)
                    COMPREPLY=( $(compgen -W "rules create edit delete" -- ${cur}) )
                    ;;
                voice)
                    COMPREPLY=( $(compgen -W "regions join leave" -- ${cur}) )
                    ;;
                invites)
                    COMPREPLY=( $(compgen -W "list get create delete" -- ${cur}) )
                    ;;
                config)
                    COMPREPLY=( $(compgen -W "bot validate" -- ${cur}) )
                    ;;
                completion)
                    COMPREPLY=( $(compgen -W "bash zsh fish powershell" -- ${cur}) )
                    ;;
                *)
                    COMPREPLY=( $(compgen -W "${global_flags}" -- ${cur}) )
                    ;;
            esac
            return 0
            ;;
        *)
            COMPREPLY=( $(compgen -W "${global_flags}" -- ${cur}) )
            return 0
            ;;
    esac
}

complete -F _dccli_complete dccli
`
}

func generateZshCompletion() string {
	return `#compdef dccli
# dccli zsh completion script
# Add to fpath: dccli completion zsh > "${fpath[1]}/_dccli"
# Or add to ~/.zshrc: eval "$(dccli completion zsh)"

_dccli() {
    local curcontext="$curcontext" state line
    typeset -A opt_args

    _arguments -C \
        '(-o --output)'{-o,--output}'[Output format (table|json|yaml)]:format:(table json yaml)' \
        '(-b --bot)'{-b,--bot}'[Bot name to use]:bot:' \
        '(-t --token)'{-t,--token}'[Bot token]:token:' \
        '(-h --help)'{-h,--help}'[Show help]' \
        '(-v --version)'{-v,--version}'[Show version]' \
        '1: :_dccli_commands' \
        '*::arg:->args'

    case "$line[1]" in
        applications)
            _dccli_applications
            ;;
        guilds)
            _dccli_guilds
            ;;
        channels)
            _dccli_channels
            ;;
        messages)
            _dccli_messages
            ;;
        roles)
            _dccli_roles
            ;;
        members)
            _dccli_members
            ;;
        webhooks)
            _dccli_webhooks
            ;;
        users)
            _dccli_users
            ;;
        emoji)
            _dccli_emoji
            ;;
        stickers)
            _dccli_stickers
            ;;
        events)
            _dccli_events
            ;;
        automod)
            _dccli_automod
            ;;
        voice)
            _dccli_voice
            ;;
        invites)
            _dccli_invites
            ;;
        config)
            _dccli_config
            ;;
        completion)
            _dccli_completion
            ;;
    esac
}

_dccli_commands() {
    local commands=(
        "applications:Applications manipulation"
        "guilds:Guild management"
        "channels:Channel management"
        "messages:Message operations"
        "roles:Role management"
        "members:Member management"
        "webhooks:Webhook management"
        "users:User operations"
        "emoji:Emoji management"
        "stickers:Sticker management"
        "events:Scheduled events"
        "automod:AutoMod rules"
        "voice:Voice operations"
        "invites:Invite management"
        "config:Configuration"
        "completion:Shell completion"
        "version:Show version"
        "help:Show help"
    )
    _describe -t commands 'dccli commands' commands
}

_dccli_applications() {
    local subcmds=(
        "commands:Application commands"
    )
    _describe -t commands 'applications subcommands' subcmds
}

_dccli_guilds() {
    local subcmds=(
        "list:List guilds"
        "describe:Show guild details"
        "edit:Edit guild"
        "leave:Leave guild"
        "channels:List channels"
        "roles:List roles"
        "members:List members"
        "invites:List invites"
    )
    _describe -t commands 'guilds subcommands' subcmds
}

_dccli_channels() {
    local subcmds=(
        "list:List channels"
        "describe:Show channel details"
        "create:Create channel"
        "edit:Edit channel"
        "delete:Delete channel"
        "typing:Send typing indicator"
        "permissions:Manage permissions"
        "invites:List invites"
        "webhooks:List webhooks"
        "messages:List messages"
        "pins:List pinned messages"
    )
    _describe -t commands 'channels subcommands' subcmds
}

_dccli_messages() {
    local subcmds=(
        "send:Send message"
        "edit:Edit message"
        "delete:Delete message"
        "get:Get message"
        "react:Add reaction"
        "unreact:Remove reaction"
        "reactions:List reactions"
        "pin:Pin message"
        "unpin:Unpin message"
        "crosspost:Crosspost message"
        "reply:Reply to message"
        "purge:Bulk delete messages"
    )
    _describe -t commands 'messages subcommands' subcmds
}

_dccli_roles() {
    local subcmds=(
        "list:List roles"
        "create:Create role"
        "edit:Edit role"
        "delete:Delete role"
        "assign:Assign role"
        "unassign:Unassign role"
    )
    _describe -t commands 'roles subcommands' subcmds
}

_dccli_members() {
    local subcmds=(
        "list:List members"
        "describe:Show member details"
        "edit:Edit member"
        "kick:Kick member"
        "ban:Ban member"
        "unban:Unban member"
        "bans:List bans"
        "timeout:Timeout member"
        "untimeout:Remove timeout"
        "prune:Prune members"
        "nick:Set nickname"
        "add-role:Add role"
        "remove-role:Remove role"
    )
    _describe -t commands 'members subcommands' subcmds
}

_dccli_webhooks() {
    local subcmds=(
        "list:List webhooks"
        "create:Create webhook"
        "edit:Edit webhook"
        "delete:Delete webhook"
        "send:Send webhook message"
    )
    _describe -t commands 'webhooks subcommands' subcmds
}

_dccli_users() {
    local subcmds=(
        "me:Show current user"
        "get:Get user"
        "avatar:Get user avatar"
        "dm:Send DM"
    )
    _describe -t commands 'users subcommands' subcmds
}

_dccli_emoji() {
    local subcmds=(
        "list:List emoji"
        "create:Create emoji"
        "delete:Delete emoji"
        "edit:Edit emoji"
    )
    _describe -t commands 'emoji subcommands' subcmds
}

_dccli_stickers() {
    local subcmds=(
        "list:List stickers"
        "get:Get sticker"
        "create:Create sticker"
        "delete:Delete sticker"
        "edit:Edit sticker"
    )
    _describe -t commands 'stickers subcommands' subcmds
}

_dccli_events() {
    local subcmds=(
        "list:List events"
        "create:Create event"
        "edit:Edit event"
        "delete:Delete event"
        "users:List event users"
        "start:Start event"
        "end:End event"
    )
    _describe -t commands 'events subcommands' subcmds
}

_dccli_automod() {
    local subcmds=(
        "rules:List rules"
        "create:Create rule"
        "edit:Edit rule"
        "delete:Delete rule"
    )
    _describe -t commands 'automod subcommands' subcmds
}

_dccli_voice() {
    local subcmds=(
        "regions:List voice regions"
        "join:Join voice channel"
        "leave:Leave voice channel"
    )
    _describe -t commands 'voice subcommands' subcmds
}

_dccli_invites() {
    local subcmds=(
        "list:List invites"
        "get:Get invite"
        "create:Create invite"
        "delete:Delete invite"
    )
    _describe -t commands 'invites subcommands' subcmds
}

_dccli_config() {
    local subcmds=(
        "bot:Bot management"
        "validate:Validate config"
    )
    _describe -t commands 'config subcommands' subcmds
}

_dccli_config_bot() {
    local subcmds=(
        "add:Add bot"
        "remove:Remove bot"
        "set:Set current bot"
        "list:List bots"
        "edit:Edit bot"
    )
    _describe -t commands 'bot subcommands' subcmds
}

_dccli_completion() {
    local subcmds=(
        "bash:Generate bash completion"
        "zsh:Generate zsh completion"
        "fish:Generate fish completion"
        "powershell:Generate PowerShell completion"
    )
    _describe -t commands 'completion subcommands' subcmds
}

_dccli "$@"
`
}

func generateFishCompletion() string {
	return `# dccli fish completion script
# Save to: ~/.config/fish/completions/dccli.fish

# Disable file completions for dccli
complete -c dccli -f

# Global flags
complete -c dccli -s o -l output -d "Output format (table|json|yaml)" -a "table json yaml"
complete -c dccli -s b -l bot -d "Bot name to use"
complete -c dccli -s t -l token -d "Bot token"
complete -c dccli -s h -l help -d "Show help"
complete -c dccli -s v -l version -d "Show version"

# Main commands
complete -c dccli -n "__fish_use_subcommand" -a "applications" -d "Applications manipulation"
complete -c dccli -n "__fish_use_subcommand" -a "guilds" -d "Guild management"
complete -c dccli -n "__fish_use_subcommand" -a "channels" -d "Channel management"
complete -c dccli -n "__fish_use_subcommand" -a "messages" -d "Message operations"
complete -c dccli -n "__fish_use_subcommand" -a "roles" -d "Role management"
complete -c dccli -n "__fish_use_subcommand" -a "members" -d "Member management"
complete -c dccli -n "__fish_use_subcommand" -a "webhooks" -d "Webhook management"
complete -c dccli -n "__fish_use_subcommand" -a "users" -d "User operations"
complete -c dccli -n "__fish_use_subcommand" -a "emoji" -d "Emoji management"
complete -c dccli -n "__fish_use_subcommand" -a "stickers" -d "Sticker management"
complete -c dccli -n "__fish_use_subcommand" -a "events" -d "Scheduled events"
complete -c dccli -n "__fish_use_subcommand" -a "automod" -d "AutoMod rules"
complete -c dccli -n "__fish_use_subcommand" -a "voice" -d "Voice operations"
complete -c dccli -n "__fish_use_subcommand" -a "invites" -d "Invite management"
complete -c dccli -n "__fish_use_subcommand" -a "config" -d "Configuration"
complete -c dccli -n "__fish_use_subcommand" -a "completion" -d "Shell completion"
complete -c dccli -n "__fish_use_subcommand" -a "version" -d "Show version"
complete -c dccli -n "__fish_use_subcommand" -a "help" -d "Show help"

# applications subcommands
complete -c dccli -n "__fish_seen_subcommand_from applications" -a "commands" -d "Application commands"

# guilds subcommands
complete -c dccli -n "__fish_seen_subcommand_from guilds" -a "list" -d "List guilds"
complete -c dccli -n "__fish_seen_subcommand_from guilds" -a "describe" -d "Show guild details"
complete -c dccli -n "__fish_seen_subcommand_from guilds" -a "edit" -d "Edit guild"
complete -c dccli -n "__fish_seen_subcommand_from guilds" -a "leave" -d "Leave guild"
complete -c dccli -n "__fish_seen_subcommand_from guilds" -a "channels" -d "List channels"
complete -c dccli -n "__fish_seen_subcommand_from guilds" -a "roles" -d "List roles"
complete -c dccli -n "__fish_seen_subcommand_from guilds" -a "members" -d "List members"
complete -c dccli -n "__fish_seen_subcommand_from guilds" -a "invites" -d "List invites"

# channels subcommands
complete -c dccli -n "__fish_seen_subcommand_from channels" -a "list" -d "List channels"
complete -c dccli -n "__fish_seen_subcommand_from channels" -a "describe" -d "Show channel details"
complete -c dccli -n "__fish_seen_subcommand_from channels" -a "create" -d "Create channel"
complete -c dccli -n "__fish_seen_subcommand_from channels" -a "edit" -d "Edit channel"
complete -c dccli -n "__fish_seen_subcommand_from channels" -a "delete" -d "Delete channel"
complete -c dccli -n "__fish_seen_subcommand_from channels" -a "typing" -d "Send typing indicator"
complete -c dccli -n "__fish_seen_subcommand_from channels" -a "permissions" -d "Manage permissions"
complete -c dccli -n "__fish_seen_subcommand_from channels" -a "invites" -d "List invites"
complete -c dccli -n "__fish_seen_subcommand_from channels" -a "webhooks" -d "List webhooks"
complete -c dccli -n "__fish_seen_subcommand_from channels" -a "messages" -d "List messages"
complete -c dccli -n "__fish_seen_subcommand_from channels" -a "pins" -d "List pinned messages"

# messages subcommands
complete -c dccli -n "__fish_seen_subcommand_from messages" -a "send" -d "Send message"
complete -c dccli -n "__fish_seen_subcommand_from messages" -a "edit" -d "Edit message"
complete -c dccli -n "__fish_seen_subcommand_from messages" -a "delete" -d "Delete message"
complete -c dccli -n "__fish_seen_subcommand_from messages" -a "get" -d "Get message"
complete -c dccli -n "__fish_seen_subcommand_from messages" -a "react" -d "Add reaction"
complete -c dccli -n "__fish_seen_subcommand_from messages" -a "unreact" -d "Remove reaction"
complete -c dccli -n "__fish_seen_subcommand_from messages" -a "reactions" -d "List reactions"
complete -c dccli -n "__fish_seen_subcommand_from messages" -a "pin" -d "Pin message"
complete -c dccli -n "__fish_seen_subcommand_from messages" -a "unpin" -d "Unpin message"
complete -c dccli -n "__fish_seen_subcommand_from messages" -a "crosspost" -d "Crosspost message"
complete -c dccli -n "__fish_seen_subcommand_from messages" -a "reply" -d "Reply to message"
complete -c dccli -n "__fish_seen_subcommand_from messages" -a "purge" -d "Bulk delete messages"

# roles subcommands
complete -c dccli -n "__fish_seen_subcommand_from roles" -a "list" -d "List roles"
complete -c dccli -n "__fish_seen_subcommand_from roles" -a "create" -d "Create role"
complete -c dccli -n "__fish_seen_subcommand_from roles" -a "edit" -d "Edit role"
complete -c dccli -n "__fish_seen_subcommand_from roles" -a "delete" -d "Delete role"
complete -c dccli -n "__fish_seen_subcommand_from roles" -a "assign" -d "Assign role"
complete -c dccli -n "__fish_seen_subcommand_from roles" -a "unassign" -d "Unassign role"

# members subcommands
complete -c dccli -n "__fish_seen_subcommand_from members" -a "list" -d "List members"
complete -c dccli -n "__fish_seen_subcommand_from members" -a "describe" -d "Show member details"
complete -c dccli -n "__fish_seen_subcommand_from members" -a "edit" -d "Edit member"
complete -c dccli -n "__fish_seen_subcommand_from members" -a "kick" -d "Kick member"
complete -c dccli -n "__fish_seen_subcommand_from members" -a "ban" -d "Ban member"
complete -c dccli -n "__fish_seen_subcommand_from members" -a "unban" -d "Unban member"
complete -c dccli -n "__fish_seen_subcommand_from members" -a "bans" -d "List bans"
complete -c dccli -n "__fish_seen_subcommand_from members" -a "timeout" -d "Timeout member"
complete -c dccli -n "__fish_seen_subcommand_from members" -a "untimeout" -d "Remove timeout"
complete -c dccli -n "__fish_seen_subcommand_from members" -a "prune" -d "Prune members"
complete -c dccli -n "__fish_seen_subcommand_from members" -a "nick" -d "Set nickname"
complete -c dccli -n "__fish_seen_subcommand_from members" -a "add-role" -d "Add role"
complete -c dccli -n "__fish_seen_subcommand_from members" -a "remove-role" -d "Remove role"

# webhooks subcommands
complete -c dccli -n "__fish_seen_subcommand_from webhooks" -a "list" -d "List webhooks"
complete -c dccli -n "__fish_seen_subcommand_from webhooks" -a "create" -d "Create webhook"
complete -c dccli -n "__fish_seen_subcommand_from webhooks" -a "edit" -d "Edit webhook"
complete -c dccli -n "__fish_seen_subcommand_from webhooks" -a "delete" -d "Delete webhook"
complete -c dccli -n "__fish_seen_subcommand_from webhooks" -a "send" -d "Send webhook message"

# users subcommands
complete -c dccli -n "__fish_seen_subcommand_from users" -a "me" -d "Show current user"
complete -c dccli -n "__fish_seen_subcommand_from users" -a "get" -d "Get user"
complete -c dccli -n "__fish_seen_subcommand_from users" -a "avatar" -d "Get user avatar"
complete -c dccli -n "__fish_seen_subcommand_from users" -a "dm" -d "Send DM"

# emoji subcommands
complete -c dccli -n "__fish_seen_subcommand_from emoji" -a "list" -d "List emoji"
complete -c dccli -n "__fish_seen_subcommand_from emoji" -a "create" -d "Create emoji"
complete -c dccli -n "__fish_seen_subcommand_from emoji" -a "delete" -d "Delete emoji"
complete -c dccli -n "__fish_seen_subcommand_from emoji" -a "edit" -d "Edit emoji"

# stickers subcommands
complete -c dccli -n "__fish_seen_subcommand_from stickers" -a "list" -d "List stickers"
complete -c dccli -n "__fish_seen_subcommand_from stickers" -a "get" -d "Get sticker"
complete -c dccli -n "__fish_seen_subcommand_from stickers" -a "create" -d "Create sticker"
complete -c dccli -n "__fish_seen_subcommand_from stickers" -a "delete" -d "Delete sticker"
complete -c dccli -n "__fish_seen_subcommand_from stickers" -a "edit" -d "Edit sticker"

# events subcommands
complete -c dccli -n "__fish_seen_subcommand_from events" -a "list" -d "List events"
complete -c dccli -n "__fish_seen_subcommand_from events" -a "create" -d "Create event"
complete -c dccli -n "__fish_seen_subcommand_from events" -a "edit" -d "Edit event"
complete -c dccli -n "__fish_seen_subcommand_from events" -a "delete" -d "Delete event"
complete -c dccli -n "__fish_seen_subcommand_from events" -a "users" -d "List event users"
complete -c dccli -n "__fish_seen_subcommand_from events" -a "start" -d "Start event"
complete -c dccli -n "__fish_seen_subcommand_from events" -a "end" -d "End event"

# automod subcommands
complete -c dccli -n "__fish_seen_subcommand_from automod" -a "rules" -d "List rules"
complete -c dccli -n "__fish_seen_subcommand_from automod" -a "create" -d "Create rule"
complete -c dccli -n "__fish_seen_subcommand_from automod" -a "edit" -d "Edit rule"
complete -c dccli -n "__fish_seen_subcommand_from automod" -a "delete" -d "Delete rule"

# voice subcommands
complete -c dccli -n "__fish_seen_subcommand_from voice" -a "regions" -d "List voice regions"
complete -c dccli -n "__fish_seen_subcommand_from voice" -a "join" -d "Join voice channel"
complete -c dccli -n "__fish_seen_subcommand_from voice" -a "leave" -d "Leave voice channel"

# invites subcommands
complete -c dccli -n "__fish_seen_subcommand_from invites" -a "list" -d "List invites"
complete -c dccli -n "__fish_seen_subcommand_from invites" -a "get" -d "Get invite"
complete -c dccli -n "__fish_seen_subcommand_from invites" -a "create" -d "Create invite"
complete -c dccli -n "__fish_seen_subcommand_from invites" -a "delete" -d "Delete invite"

# config subcommands
complete -c dccli -n "__fish_seen_subcommand_from config" -a "bot" -d "Bot management"
complete -c dccli -n "__fish_seen_subcommand_from config" -a "validate" -d "Validate config"

# config bot subcommands
complete -c dccli -n "__fish_seen_subcommand_from config; and __fish_seen_subcommand_from bot" -a "add" -d "Add bot"
complete -c dccli -n "__fish_seen_subcommand_from config; and __fish_seen_subcommand_from bot" -a "remove" -d "Remove bot"
complete -c dccli -n "__fish_seen_subcommand_from config; and __fish_seen_subcommand_from bot" -a "set" -d "Set current bot"
complete -c dccli -n "__fish_seen_subcommand_from config; and __fish_seen_subcommand_from bot" -a "list" -d "List bots"
complete -c dccli -n "__fish_seen_subcommand_from config; and __fish_seen_subcommand_from bot" -a "edit" -d "Edit bot"

# completion subcommands
complete -c dccli -n "__fish_seen_subcommand_from completion" -a "bash" -d "Generate bash completion"
complete -c dccli -n "__fish_seen_subcommand_from completion" -a "zsh" -d "Generate zsh completion"
complete -c dccli -n "__fish_seen_subcommand_from completion" -a "fish" -d "Generate fish completion"
complete -c dccli -n "__fish_seen_subcommand_from completion" -a "powershell" -d "Generate PowerShell completion"
`
}

func generatePowerShellCompletion() string {
	return `# dccli PowerShell completion script
# Add to $PROFILE: dccli completion powershell | Out-String | Invoke-Expression
# Or save to file: dccli completion powershell > $PROFILE
dccli.ps1

using namespace System.Management.Automation
using namespace System.Management.Automation.Language

Register-ArgumentCompleter -Native -CommandName 'dccli' -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    $commandElements = $commandAst.CommandElements
    $command = @(
        for ($i = 1; $i -lt $commandElements.Count; $i++) {
            $element = $commandElements[$i]
            if ($element -isnot [StringLiteralExpressionAst] -and
                $element -isnot [ConstantExpressionAst]) {
                break
            }
            $element.Value
        }
    ) -join ';'

    $completions = @(switch ($command) {
        '' {
            [CompletionResult]::new('applications', 'applications', [CompletionResultType]::ParameterValue, 'Applications manipulation')
            [CompletionResult]::new('guilds', 'guilds', [CompletionResultType]::ParameterValue, 'Guild management')
            [CompletionResult]::new('channels', 'channels', [CompletionResultType]::ParameterValue, 'Channel management')
            [CompletionResult]::new('messages', 'messages', [CompletionResultType]::ParameterValue, 'Message operations')
            [CompletionResult]::new('roles', 'roles', [CompletionResultType]::ParameterValue, 'Role management')
            [CompletionResult]::new('members', 'members', [CompletionResultType]::ParameterValue, 'Member management')
            [CompletionResult]::new('webhooks', 'webhooks', [CompletionResultType]::ParameterValue, 'Webhook management')
            [CompletionResult]::new('users', 'users', [CompletionResultType]::ParameterValue, 'User operations')
            [CompletionResult]::new('emoji', 'emoji', [CompletionResultType]::ParameterValue, 'Emoji management')
            [CompletionResult]::new('stickers', 'stickers', [CompletionResultType]::ParameterValue, 'Sticker management')
            [CompletionResult]::new('events', 'events', [CompletionResultType]::ParameterValue, 'Scheduled events')
            [CompletionResult]::new('automod', 'automod', [CompletionResultType]::ParameterValue, 'AutoMod rules')
            [CompletionResult]::new('voice', 'voice', [CompletionResultType]::ParameterValue, 'Voice operations')
            [CompletionResult]::new('invites', 'invites', [CompletionResultType]::ParameterValue, 'Invite management')
            [CompletionResult]::new('config', 'config', [CompletionResultType]::ParameterValue, 'Configuration')
            [CompletionResult]::new('completion', 'completion', [CompletionResultType]::ParameterValue, 'Shell completion')
            [CompletionResult]::new('version', 'version', [CompletionResultType]::ParameterValue, 'Show version')
            [CompletionResult]::new('help', 'help', [CompletionResultType]::ParameterValue, 'Show help')
            break
        }
        'applications' {
            [CompletionResult]::new('commands', 'commands', [CompletionResultType]::ParameterValue, 'Application commands')
            break
        }
        'guilds' {
            [CompletionResult]::new('list', 'list', [CompletionResultType]::ParameterValue, 'List guilds')
            [CompletionResult]::new('describe', 'describe', [CompletionResultType]::ParameterValue, 'Show guild details')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit guild')
            [CompletionResult]::new('leave', 'leave', [CompletionResultType]::ParameterValue, 'Leave guild')
            [CompletionResult]::new('channels', 'channels', [CompletionResultType]::ParameterValue, 'List channels')
            [CompletionResult]::new('roles', 'roles', [CompletionResultType]::ParameterValue, 'List roles')
            [CompletionResult]::new('members', 'members', [CompletionResultType]::ParameterValue, 'List members')
            [CompletionResult]::new('invites', 'invites', [CompletionResultType]::ParameterValue, 'List invites')
            break
        }
        'channels' {
            [CompletionResult]::new('list', 'list', [CompletionResultType]::ParameterValue, 'List channels')
            [CompletionResult]::new('describe', 'describe', [CompletionResultType]::ParameterValue, 'Show channel details')
            [CompletionResult]::new('create', 'create', [CompletionResultType]::ParameterValue, 'Create channel')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit channel')
            [CompletionResult]::new('delete', 'delete', [CompletionResultType]::ParameterValue, 'Delete channel')
            [CompletionResult]::new('typing', 'typing', [CompletionResultType]::ParameterValue, 'Send typing indicator')
            [CompletionResult]::new('permissions', 'permissions', [CompletionResultType]::ParameterValue, 'Manage permissions')
            [CompletionResult]::new('invites', 'invites', [CompletionResultType]::ParameterValue, 'List invites')
            [CompletionResult]::new('webhooks', 'webhooks', [CompletionResultType]::ParameterValue, 'List webhooks')
            [CompletionResult]::new('messages', 'messages', [CompletionResultType]::ParameterValue, 'List messages')
            [CompletionResult]::new('pins', 'pins', [CompletionResultType]::ParameterValue, 'List pinned messages')
            break
        }
        'messages' {
            [CompletionResult]::new('send', 'send', [CompletionResultType]::ParameterValue, 'Send message')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit message')
            [CompletionResult]::new('delete', 'delete', [CompletionResultType]::ParameterValue, 'Delete message')
            [CompletionResult]::new('get', 'get', [CompletionResultType]::ParameterValue, 'Get message')
            [CompletionResult]::new('react', 'react', [CompletionResultType]::ParameterValue, 'Add reaction')
            [CompletionResult]::new('unreact', 'unreact', [CompletionResultType]::ParameterValue, 'Remove reaction')
            [CompletionResult]::new('reactions', 'reactions', [CompletionResultType]::ParameterValue, 'List reactions')
            [CompletionResult]::new('pin', 'pin', [CompletionResultType]::ParameterValue, 'Pin message')
            [CompletionResult]::new('unpin', 'unpin', [CompletionResultType]::ParameterValue, 'Unpin message')
            [CompletionResult]::new('crosspost', 'crosspost', [CompletionResultType]::ParameterValue, 'Crosspost message')
            [CompletionResult]::new('reply', 'reply', [CompletionResultType]::ParameterValue, 'Reply to message')
            [CompletionResult]::new('purge', 'purge', [CompletionResultType]::ParameterValue, 'Bulk delete messages')
            break
        }
        'roles' {
            [CompletionResult]::new('list', 'list', [CompletionResultType]::ParameterValue, 'List roles')
            [CompletionResult]::new('create', 'create', [CompletionResultType]::ParameterValue, 'Create role')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit role')
            [CompletionResult]::new('delete', 'delete', [CompletionResultType]::ParameterValue, 'Delete role')
            [CompletionResult]::new('assign', 'assign', [CompletionResultType]::ParameterValue, 'Assign role')
            [CompletionResult]::new('unassign', 'unassign', [CompletionResultType]::ParameterValue, 'Unassign role')
            break
        }
        'members' {
            [CompletionResult]::new('list', 'list', [CompletionResultType]::ParameterValue, 'List members')
            [CompletionResult]::new('describe', 'describe', [CompletionResultType]::ParameterValue, 'Show member details')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit member')
            [CompletionResult]::new('kick', 'kick', [CompletionResultType]::ParameterValue, 'Kick member')
            [CompletionResult]::new('ban', 'ban', [CompletionResultType]::ParameterValue, 'Ban member')
            [CompletionResult]::new('unban', 'unban', [CompletionResultType]::ParameterValue, 'Unban member')
            [CompletionResult]::new('bans', 'bans', [CompletionResultType]::ParameterValue, 'List bans')
            [CompletionResult]::new('timeout', 'timeout', [CompletionResultType]::ParameterValue, 'Timeout member')
            [CompletionResult]::new('untimeout', 'untimeout', [CompletionResultType]::ParameterValue, 'Remove timeout')
            [CompletionResult]::new('prune', 'prune', [CompletionResultType]::ParameterValue, 'Prune members')
            [CompletionResult]::new('nick', 'nick', [CompletionResultType]::ParameterValue, 'Set nickname')
            [CompletionResult]::new('add-role', 'add-role', [CompletionResultType]::ParameterValue, 'Add role')
            [CompletionResult]::new('remove-role', 'remove-role', [CompletionResultType]::ParameterValue, 'Remove role')
            break
        }
        'webhooks' {
            [CompletionResult]::new('list', 'list', [CompletionResultType]::ParameterValue, 'List webhooks')
            [CompletionResult]::new('create', 'create', [CompletionResultType]::ParameterValue, 'Create webhook')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit webhook')
            [CompletionResult]::new('delete', 'delete', [CompletionResultType]::ParameterValue, 'Delete webhook')
            [CompletionResult]::new('send', 'send', [CompletionResultType]::ParameterValue, 'Send webhook message')
            break
        }
        'users' {
            [CompletionResult]::new('me', 'me', [CompletionResultType]::ParameterValue, 'Show current user')
            [CompletionResult]::new('get', 'get', [CompletionResultType]::ParameterValue, 'Get user')
            [CompletionResult]::new('avatar', 'avatar', [CompletionResultType]::ParameterValue, 'Get user avatar')
            [CompletionResult]::new('dm', 'dm', [CompletionResultType]::ParameterValue, 'Send DM')
            break
        }
        'emoji' {
            [CompletionResult]::new('list', 'list', [CompletionResultType]::ParameterValue, 'List emoji')
            [CompletionResult]::new('create', 'create', [CompletionResultType]::ParameterValue, 'Create emoji')
            [CompletionResult]::new('delete', 'delete', [CompletionResultType]::ParameterValue, 'Delete emoji')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit emoji')
            break
        }
        'stickers' {
            [CompletionResult]::new('list', 'list', [CompletionResultType]::ParameterValue, 'List stickers')
            [CompletionResult]::new('get', 'get', [CompletionResultType]::ParameterValue, 'Get sticker')
            [CompletionResult]::new('create', 'create', [CompletionResultType]::ParameterValue, 'Create sticker')
            [CompletionResult]::new('delete', 'delete', [CompletionResultType]::ParameterValue, 'Delete sticker')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit sticker')
            break
        }
        'events' {
            [CompletionResult]::new('list', 'list', [CompletionResultType]::ParameterValue, 'List events')
            [CompletionResult]::new('create', 'create', [CompletionResultType]::ParameterValue, 'Create event')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit event')
            [CompletionResult]::new('delete', 'delete', [CompletionResultType]::ParameterValue, 'Delete event')
            [CompletionResult]::new('users', 'users', [CompletionResultType]::ParameterValue, 'List event users')
            [CompletionResult]::new('start', 'start', [CompletionResultType]::ParameterValue, 'Start event')
            [CompletionResult]::new('end', 'end', [CompletionResultType]::ParameterValue, 'End event')
            break
        }
        'automod' {
            [CompletionResult]::new('rules', 'rules', [CompletionResultType]::ParameterValue, 'List rules')
            [CompletionResult]::new('create', 'create', [CompletionResultType]::ParameterValue, 'Create rule')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit rule')
            [CompletionResult]::new('delete', 'delete', [CompletionResultType]::ParameterValue, 'Delete rule')
            break
        }
        'voice' {
            [CompletionResult]::new('regions', 'regions', [CompletionResultType]::ParameterValue, 'List voice regions')
            [CompletionResult]::new('join', 'join', [CompletionResultType]::ParameterValue, 'Join voice channel')
            [CompletionResult]::new('leave', 'leave', [CompletionResultType]::ParameterValue, 'Leave voice channel')
            break
        }
        'invites' {
            [CompletionResult]::new('list', 'list', [CompletionResultType]::ParameterValue, 'List invites')
            [CompletionResult]::new('get', 'get', [CompletionResultType]::ParameterValue, 'Get invite')
            [CompletionResult]::new('create', 'create', [CompletionResultType]::ParameterValue, 'Create invite')
            [CompletionResult]::new('delete', 'delete', [CompletionResultType]::ParameterValue, 'Delete invite')
            break
        }
        'config' {
            [CompletionResult]::new('bot', 'bot', [CompletionResultType]::ParameterValue, 'Bot management')
            [CompletionResult]::new('validate', 'validate', [CompletionResultType]::ParameterValue, 'Validate config')
            break
        }
        'config;bot' {
            [CompletionResult]::new('add', 'add', [CompletionResultType]::ParameterValue, 'Add bot')
            [CompletionResult]::new('remove', 'remove', [CompletionResultType]::ParameterValue, 'Remove bot')
            [CompletionResult]::new('set', 'set', [CompletionResultType]::ParameterValue, 'Set current bot')
            [CompletionResult]::new('list', 'list', [CompletionResultType]::ParameterValue, 'List bots')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit bot')
            break
        }
        'completion' {
            [CompletionResult]::new('bash', 'bash', [CompletionResultType]::ParameterValue, 'Generate bash completion')
            [CompletionResult]::new('zsh', 'zsh', [CompletionResultType]::ParameterValue, 'Generate zsh completion')
            [CompletionResult]::new('fish', 'fish', [CompletionResultType]::ParameterValue, 'Generate fish completion')
            [CompletionResult]::new('powershell', 'powershell', [CompletionResultType]::ParameterValue, 'Generate PowerShell completion')
            break
        }
    })

    # Global flags
    if ($wordToComplete -match '^-') {
        $completions += @(
            [CompletionResult]::new('-o', '-o', [CompletionResultType]::ParameterName, 'Output format')
            [CompletionResult]::new('--output', '--output', [CompletionResultType]::ParameterName, 'Output format')
            [CompletionResult]::new('-b', '-b', [CompletionResultType]::ParameterName, 'Bot name')
            [CompletionResult]::new('--bot', '--bot', [CompletionResultType]::ParameterName, 'Bot name')
            [CompletionResult]::new('-t', '-t', [CompletionResultType]::ParameterName, 'Bot token')
            [CompletionResult]::new('--token', '--token', [CompletionResultType]::ParameterName, 'Bot token')
            [CompletionResult]::new('-h', '-h', [CompletionResultType]::ParameterName, 'Show help')
            [CompletionResult]::new('--help', '--help', [CompletionResultType]::ParameterName, 'Show help')
            [CompletionResult]::new('-v', '-v', [CompletionResultType]::ParameterName, 'Show version')
            [CompletionResult]::new('--version', '--version', [CompletionResultType]::ParameterName, 'Show version')
        )
    }

    $completions | Where-Object { $_.CompletionText -like "$wordToComplete*" }
}
`
}

package commands

import (
	"github.com/urfave/cli/v3"
)

func AppCommandsCommand() *cli.Command {
	return &cli.Command{
		Name:  "commands",
		Usage: "Application commands",
		Commands: []*cli.Command{
			ListAppCommandsCommand(),
			CreateAppCmdCommand(),
			EditAppCmdCommand(),
			RemoveAppCmdCommand(),
			DeleteAllAppCmdCommand(),
			DescribeAppCmdCommand(),
			BulkOverwriteAppCmdCommand(),
			AppCommandPermissionsCommand(),
		},
	}
}

func GuildsCommand() *cli.Command {
	return &cli.Command{
		Name:  "guilds",
		Usage: "Guilds",
		Commands: []*cli.Command{
			GuildListCommand(),
			DescribeGuildCommand(),
			LeaveGuildCommand(),
			EditGuildCommand(),
			GuildChannelsCommand(),
			GuildRolesCommand(),
			GuildMembersCommand(),
			GuildInvitesCommand(),
		},
	}
}

func ApplicationsCommands() *cli.Command {
	return &cli.Command{
		Name:  "applications",
		Usage: "Applications manipulation",
		Commands: []*cli.Command{
			AppCommandsCommand(),
		},
	}
}

func ConfigCommands() *cli.Command {
	return &cli.Command{
		Name:        "config",
		Usage:       "Change cli settings",
		Description: "Manage bot configurations and settings",
		Commands: []*cli.Command{
			BotCommand(),
			ConfigValidateCommand(),
		},
	}
}

func ChannelsRootCommand() *cli.Command {
	return ChannelsCommand()
}

func MessagesRootCommand() *cli.Command {
	return MessagesCommand()
}

func RolesRootCommand() *cli.Command {
	return RolesCommand()
}

func MembersRootCommand() *cli.Command {
	return MembersCommand()
}

func WebhooksRootCommand() *cli.Command {
	return WebhooksCommand()
}

func UsersRootCommand() *cli.Command {
	return UsersCommand()
}

func EmojiRootCommand() *cli.Command {
	return EmojiCommand()
}

func StickersRootCommand() *cli.Command {
	return StickersCommand()
}

func EventsRootCommand() *cli.Command {
	return EventsCommand()
}

func AutoModRootCommand() *cli.Command {
	return AutoModCommand()
}

// VoiceRegionsRootCommand returns the voice regions command
func VoiceRegionsRootCommand() *cli.Command {
	return VoiceRegionsCommand()
}

func InvitesRootCommand() *cli.Command {
	return InvitesCommand()
}

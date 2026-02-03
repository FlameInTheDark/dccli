package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/cmd/dccli/commands"
	"github.com/FlameInTheDark/dccli/pkg/utils"
)

var build = "develop"

func getVersionInfo() string {
	goVersion := runtime.Version()
	discordgoVersion := discordgo.VERSION

	var buildInfo string
	if info, ok := debug.ReadBuildInfo(); ok {
		buildInfo = fmt.Sprintf("Module: %s\n", info.Main.Path)
		for _, dep := range info.Deps {
			if dep.Path == "github.com/bwmarrin/discordgo" {
				discordgoVersion = dep.Version
				break
			}
		}
	}

	return fmt.Sprintf(`Version:    %s
Go Version: %s
DiscordGo:  %s
Build:      %s
%s`, build, goVersion, discordgoVersion, build, buildInfo)
}

func main() {
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Print(getVersionInfo())
	}

	app := &cli.Command{
		Name:        "dccli",
		Version:     build,
		Usage:       "Discord CLI Tool - Manage Discord bots and guilds from the command line",
		Description: "A powerful CLI for Discord API operations including guild management, messaging, roles, and more.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Usage:       "Output format (table|json|yaml)",
				Value:       "table",
				Sources:     cli.EnvVars("DCLI_OUTPUT"),
				DefaultText: "table",
			},
			&cli.StringFlag{
				Name:    "bot",
				Aliases: []string{"b"},
				Usage:   "Bot name to use (overrides current from config)",
				Sources: cli.EnvVars("DCLI_BOT"),
			},
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "Bot token (overrides config entirely)",
				Sources: cli.EnvVars("DCLI_TOKEN"),
			},
		},
		Commands: []*cli.Command{
			commands.ApplicationsCommands(),
			commands.GuildsCommand(),
			commands.ChannelsRootCommand(),
			commands.MessagesRootCommand(),
			commands.RolesRootCommand(),
			commands.MembersRootCommand(),
			commands.WebhooksRootCommand(),
			commands.UsersRootCommand(),
			commands.EmojiRootCommand(),
			commands.StickersRootCommand(),
			commands.EventsRootCommand(),
			commands.AutoModRootCommand(),
			commands.VoiceRootCommand(),
			commands.InvitesRootCommand(),
			commands.ConfigCommands(),
			commands.CompletionCommand(),
		},
	}

	err := app.Run(context.Background(), os.Args)
	utils.HandleError(err)
}

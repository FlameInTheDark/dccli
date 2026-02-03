package commands

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/pkg/dprint"
	"github.com/FlameInTheDark/dccli/pkg/utils"
)

func GuildListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Returns a list of guilds",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "before",
				Usage: "Before guild ID",
			},
			&cli.StringFlag{
				Name:  "after",
				Usage: "After guild ID",
			},
			&cli.IntFlag{
				Name:  "limit",
				Usage: "Limit number of guilds. Max 100. Default is 10",
				Value: 10,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			guilds, err := cliCtx.Client.GuildList(int(c.Int("limit")), c.String("before"), c.String("after"))
			if err != nil {
				return utils.DiscordErrorf("failed to list guilds: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				if c.String("after") != "" {
					fmt.Println("Results after: ", c.String("after"))
				}
				if c.String("before") != "" {
					fmt.Println("Results before: ", c.String("before"))
				}

				data := [][]string{}
				for _, g := range guilds {
					data = append(data, []string{g.ID, g.Name})
				}
				header := []string{"ID", "Name"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(guilds); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func DescribeGuildCommand() *cli.Command {
	return &cli.Command{
		Name:      "describe",
		Usage:     "Describe guild",
		ArgsUsage: "[guild ID]",
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			g, err := cliCtx.Client.GuildDescribe(c.Args().Get(0))
			if err != nil {
				return utils.DiscordErrorf("failed to describe guild: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				err = dprint.PrintTemplate(dprint.GuildDescribeTemplate, g)
				if err != nil {
					return err
				}
			} else {
				if err := output.Print(g); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func LeaveGuildCommand() *cli.Command {
	return &cli.Command{
		Name:      "leave",
		Usage:     "Make the bot leave a guild",
		ArgsUsage: "[guild-id]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "force",
				Usage: "Skip confirmation prompt",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("guild ID is required")
			}
			guildID := c.Args().First()

			// Get guild info for confirmation
			guild, err := cliCtx.Client.GuildDescribe(guildID)
			if err != nil {
				return utils.DiscordErrorf("failed to get guild info: %w", err)
			}

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to leave guild: %s (ID: %s)\n", guild.Name, guildID)
				fmt.Print("Are you sure? (yes/no): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if response != "yes\n" && response != "yes\r\n" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			// Leave the guild
			if err := cliCtx.Client.LeaveGuild(guildID); err != nil {
				return utils.DiscordErrorf("failed to leave guild: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully left guild: %s\n", guild.Name)
			} else {
				result := map[string]interface{}{
					"success":  true,
					"guild_id": guildID,
					"name":     guild.Name,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func EditGuildCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit guild properties",
		ArgsUsage: "[guild-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "New guild name",
			},
			&cli.StringFlag{
				Name:  "region",
				Usage: "New voice region",
			},
			&cli.IntFlag{
				Name:  "verification-level",
				Usage: "Verification level (0=None, 1=Low, 2=Medium, 3=High, 4=VeryHigh)",
			},
			&cli.IntFlag{
				Name:  "default-message-notifications",
				Usage: "Default message notifications (0=AllMessages, 1=OnlyMentions)",
			},
			&cli.IntFlag{
				Name:  "explicit-content-filter",
				Usage: "Explicit content filter (0=Disabled, 1=MembersWithoutRoles, 2=AllMembers)",
			},
			&cli.StringFlag{
				Name:  "afk-channel",
				Usage: "AFK channel ID",
			},
			&cli.IntFlag{
				Name:  "afk-timeout",
				Usage: "AFK timeout in seconds (60, 300, 900, 1800, 3600)",
			},
			&cli.StringFlag{
				Name:  "icon",
				Usage: "Path to icon image file",
			},
			&cli.StringFlag{
				Name:  "splash",
				Usage: "Path to splash image file",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("guild ID is required")
			}
			guildID := c.Args().First()

			// Build guild params
			params := &discordgo.GuildParams{}

			if name := c.String("name"); name != "" {
				params.Name = name
			}
			if region := c.String("region"); region != "" {
				params.Region = region
			}
			if level := c.Int("verification-level"); c.IsSet("verification-level") {
				vl := discordgo.VerificationLevel(level)
				params.VerificationLevel = &vl
			}
			if notif := c.Int("default-message-notifications"); c.IsSet("default-message-notifications") {
				params.DefaultMessageNotifications = notif
			}
			if filter := c.Int("explicit-content-filter"); c.IsSet("explicit-content-filter") {
				params.ExplicitContentFilter = filter
			}
			if afkChannel := c.String("afk-channel"); afkChannel != "" {
				params.AfkChannelID = afkChannel
			}
			if afkTimeout := c.Int("afk-timeout"); c.IsSet("afk-timeout") {
				params.AfkTimeout = afkTimeout
			}

			// Handle icon file
			if iconPath := c.String("icon"); iconPath != "" {
				iconData, err := os.ReadFile(iconPath)
				if err != nil {
					return utils.ValidationErrorf("failed to read icon file: %w", err)
				}
				params.Icon = "data:image/png;base64," + base64.StdEncoding.EncodeToString(iconData)
			}

			// Handle splash file
			if splashPath := c.String("splash"); splashPath != "" {
				splashData, err := os.ReadFile(splashPath)
				if err != nil {
					return utils.ValidationErrorf("failed to read splash file: %w", err)
				}
				params.Splash = "data:image/png;base64," + base64.StdEncoding.EncodeToString(splashData)
			}

			updatedGuild, err := cliCtx.Client.EditGuild(guildID, params)
			if err != nil {
				return utils.DiscordErrorf("failed to edit guild: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Guild updated successfully!\n")
				fmt.Printf("ID: %s\n", updatedGuild.ID)
				fmt.Printf("Name: %s\n", updatedGuild.Name)
				fmt.Printf("Region: %s\n", updatedGuild.Region)
			} else {
				result := map[string]interface{}{
					"success": true,
					"guild":   updatedGuild,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// GuildChannelsCommand manages guild channels
func GuildChannelsCommand() *cli.Command {
	return &cli.Command{
		Name:      "channels",
		Usage:     "Manage guild channels",
		ArgsUsage: "[guild-id]",
		Commands: []*cli.Command{
			ChannelsListCommand(),
			ChannelsGetCommand(),
			ChannelsCreateCommand(),
			ChannelsEditCommand(),
			ChannelsDeleteCommand(),
		},
	}
}

// GuildRolesCommand manages guild roles
func GuildRolesCommand() *cli.Command {
	return &cli.Command{
		Name:      "roles",
		Usage:     "Manage guild roles",
		ArgsUsage: "[guild-id]",
		Commands: []*cli.Command{
			RolesListCommand(),
			RolesGetCommand(),
			RolesCreateCommand(),
			RolesEditCommand(),
			RolesDeleteCommand(),
		},
	}
}

// GuildMembersCommand manages guild members
func GuildMembersCommand() *cli.Command {
	return &cli.Command{
		Name:      "members",
		Usage:     "Manage guild members",
		ArgsUsage: "[guild-id]",
		Commands: []*cli.Command{
			MembersListCommand(),
			MembersGetCommand(),
			MembersBanCommand(),
			MembersUnbanCommand(),
			MembersKickCommand(),
		},
	}
}

// GuildInvitesCommand manages guild invites
func GuildInvitesCommand() *cli.Command {
	return &cli.Command{
		Name:      "invites",
		Usage:     "Manage guild invites",
		ArgsUsage: "[guild-id]",
		Commands: []*cli.Command{
			ListInvitesCommand(),
			CreateInviteCommand(),
			DeleteInviteCommand(),
		},
	}
}

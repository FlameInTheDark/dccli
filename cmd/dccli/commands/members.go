package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/pkg/dprint"
	"github.com/FlameInTheDark/dccli/pkg/utils"
)

func MembersCommand() *cli.Command {
	return &cli.Command{
		Name:  "members",
		Usage: "Member management",
		Commands: []*cli.Command{
			MembersListCommand(),
			MembersGetCommand(),
			MembersBanCommand(),
			MembersUnbanCommand(),
			MembersKickCommand(),
			MembersTimeoutCommand(),
			MembersUntimeoutCommand(),
			MembersNickCommand(),
			MembersAddRoleCommand(),
			MembersRemoveRoleCommand(),
		},
	}
}

func MembersListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List members in a guild",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "limit",
				Usage: "Maximum number of members to return (1-1000)",
				Value: 100,
			},
			&cli.StringFlag{
				Name:  "after",
				Usage: "Get members after this user ID",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			guildID := c.String("guild")

			members, err := cliCtx.Client.GetGuildMembers(guildID, int(c.Int("limit")), c.String("after"))
			if err != nil {
				return utils.DiscordErrorf("failed to list members: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				data := [][]string{}
				for _, m := range members {
					username := m.User.Username
					if m.User.Discriminator != "0" {
						username += "#" + m.User.Discriminator
					}
					nick := m.Nick
					if nick == "" {
						nick = "-"
					}
					data = append(data, []string{
						m.User.ID,
						username,
						nick,
						m.JoinedAt.Format("2006-01-02"),
					})
				}
				header := []string{"ID", "Username", "Nickname", "Joined"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(members); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func MembersGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get member details",
		ArgsUsage: "[user-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("user ID is required")
			}
			userID := c.Args().First()
			guildID := c.String("guild")

			member, err := cliCtx.Client.GetGuildMember(guildID, userID)
			if err != nil {
				return utils.DiscordErrorf("failed to get member: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("User ID: %s\n", member.User.ID)
				fmt.Printf("Username: %s\n", member.User.Username)
				if member.Nick != "" {
					fmt.Printf("Nickname: %s\n", member.Nick)
				}
				fmt.Printf("Joined: %s\n", member.JoinedAt.Format("2006-01-02 15:04:05"))
				if member.JoinedAt.IsZero() {
					// Handle cases where joined at might be zero (shouldn't happen usually)
				}
				fmt.Printf("Roles: %d\n", len(member.Roles))
				if len(member.Roles) > 0 {
					fmt.Printf("Role IDs: %v\n", member.Roles)
				}
			} else {
				if err := output.Print(member); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func MembersBanCommand() *cli.Command {
	return &cli.Command{
		Name:      "ban",
		Usage:     "Ban a member from the guild",
		ArgsUsage: "[user-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "reason",
				Usage: "Ban reason",
			},
			&cli.IntFlag{
				Name:  "delete-days",
				Usage: "Number of days of messages to delete (0-7)",
				Value: 0,
			},
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
				return utils.ValidationError("user ID is required")
			}
			userID := c.Args().First()
			guildID := c.String("guild")
			deleteDays := int(c.Int("delete-days"))
			reason := c.String("reason")

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to ban user: %s\n", userID)
				fmt.Print("Are you sure? (yes/no): ")
				var response string
				fmt.Scanln(&response)
				if response != "yes" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := cliCtx.Client.GuildBanCreate(guildID, userID, deleteDays, reason); err != nil {
				return utils.DiscordErrorf("failed to ban member: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully banned user: %s\n", userID)
				if reason != "" {
					fmt.Printf("Reason: %s\n", reason)
				}
			} else {
				result := map[string]interface{}{
					"success": true,
					"user_id": userID,
					"reason":  reason,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func MembersUnbanCommand() *cli.Command {
	return &cli.Command{
		Name:      "unban",
		Usage:     "Unban a user from the guild",
		ArgsUsage: "[user-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
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
				return utils.ValidationError("user ID is required")
			}
			userID := c.Args().First()
			guildID := c.String("guild")

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to unban user: %s\n", userID)
				fmt.Print("Are you sure? (yes/no): ")
				var response string
				fmt.Scanln(&response)
				if response != "yes" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := cliCtx.Client.GuildBanDelete(guildID, userID); err != nil {
				return utils.DiscordErrorf("failed to unban user: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully unbanned user: %s\n", userID)
			} else {
				result := map[string]interface{}{
					"success": true,
					"user_id": userID,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func MembersKickCommand() *cli.Command {
	return &cli.Command{
		Name:      "kick",
		Usage:     "Kick a member from the guild",
		ArgsUsage: "[user-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "reason",
				Usage: "Kick reason",
			},
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
				return utils.ValidationError("user ID is required")
			}
			userID := c.Args().First()
			guildID := c.String("guild")

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to kick user: %s\n", userID)
				fmt.Print("Are you sure? (yes/no): ")
				var response string
				fmt.Scanln(&response)
				if response != "yes" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := cliCtx.Client.GuildMemberDelete(guildID, userID); err != nil {
				return utils.DiscordErrorf("failed to kick member: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully kicked user: %s\n", userID)
				if reason := c.String("reason"); reason != "" {
					fmt.Printf("Reason: %s\n", reason)
				}
			} else {
				result := map[string]interface{}{
					"success": true,
					"user_id": userID,
					"reason":  c.String("reason"),
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func MembersTimeoutCommand() *cli.Command {
	return &cli.Command{
		Name:      "timeout",
		Usage:     "Timeout a member",
		ArgsUsage: "[user-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "duration",
				Usage:    "Duration (e.g. 10m, 1h, 24h). Max 28 days.",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("user ID is required")
			}
			userID := c.Args().First()
			guildID := c.String("guild")
			durationStr := c.String("duration")

			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				return utils.ValidationErrorf("invalid duration: %w", err)
			}

			until := time.Now().Add(duration)
			if err := cliCtx.Client.GuildMemberTimeout(guildID, userID, &until); err != nil {
				return utils.DiscordErrorf("failed to timeout member: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully timed out user: %s for %s\n", userID, durationStr)
			} else {
				result := map[string]interface{}{
					"success":  true,
					"user_id":  userID,
					"duration": durationStr,
					"until":    until,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func MembersUntimeoutCommand() *cli.Command {
	return &cli.Command{
		Name:      "untimeout",
		Usage:     "Remove timeout from a member",
		ArgsUsage: "[user-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("user ID is required")
			}
			userID := c.Args().First()
			guildID := c.String("guild")

			if err := cliCtx.Client.GuildMemberTimeout(guildID, userID, nil); err != nil {
				return utils.DiscordErrorf("failed to remove timeout: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully removed timeout from user: %s\n", userID)
			} else {
				result := map[string]interface{}{
					"success": true,
					"user_id": userID,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func MembersNickCommand() *cli.Command {
	return &cli.Command{
		Name:      "nick",
		Usage:     "Set nickname for a member",
		ArgsUsage: "[user-id] [nickname]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "nick",
				Usage: "New nickname (leave empty to reset)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("user ID is required")
			}
			userID := c.Args().First()
			guildID := c.String("guild")
			nick := c.String("nick")

			// If nickname provided as argument
			if c.NArg() > 1 {
				nick = c.Args().Get(1)
			}

			if err := cliCtx.Client.GuildMemberNickname(guildID, userID, nick); err != nil {
				return utils.DiscordErrorf("failed to set nickname: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				if nick == "" {
					fmt.Printf("Successfully reset nickname for user: %s\n", userID)
				} else {
					fmt.Printf("Successfully set nickname to '%s' for user: %s\n", nick, userID)
				}
			} else {
				result := map[string]interface{}{
					"success":  true,
					"user_id":  userID,
					"nickname": nick,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func MembersAddRoleCommand() *cli.Command {
	return &cli.Command{
		Name:      "add-role",
		Usage:     "Add a role to a member",
		ArgsUsage: "[user-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "role",
				Usage:    "Role ID",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("user ID is required")
			}
			userID := c.Args().First()
			guildID := c.String("guild")
			roleID := c.String("role")

			if err := cliCtx.Client.GuildMemberRoleAdd(guildID, userID, roleID); err != nil {
				return utils.DiscordErrorf("failed to add role: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully added role %s to user %s\n", roleID, userID)
			} else {
				result := map[string]interface{}{
					"success": true,
					"user_id": userID,
					"role_id": roleID,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func MembersRemoveRoleCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove-role",
		Usage:     "Remove a role from a member",
		ArgsUsage: "[user-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "role",
				Usage:    "Role ID",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("user ID is required")
			}
			userID := c.Args().First()
			guildID := c.String("guild")
			roleID := c.String("role")

			if err := cliCtx.Client.GuildMemberRoleRemove(guildID, userID, roleID); err != nil {
				return utils.DiscordErrorf("failed to remove role: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully removed role %s from user %s\n", roleID, userID)
			} else {
				result := map[string]interface{}{
					"success": true,
					"user_id": userID,
					"role_id": roleID,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

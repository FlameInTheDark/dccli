package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/pkg/dprint"
	"github.com/FlameInTheDark/dccli/pkg/utils"
)

func InvitesCommand() *cli.Command {
	return &cli.Command{
		Name:  "invites",
		Usage: "Invite management",
		Commands: []*cli.Command{
			InvitesGetCommand(),
			ListInvitesCommand(),
			CreateInviteCommand(),
			DeleteInviteCommand(),
		},
	}
}

// InvitesGetCommand gets invite details
func InvitesGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get invite details",
		ArgsUsage: "[invite-code]",
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("invite code is required")
			}
			inviteCode := c.Args().First()

			invite, err := cliCtx.Client.GetInvite(inviteCode)
			if err != nil {
				return utils.DiscordErrorf("failed to get invite: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Code: %s\n", invite.Code)
				fmt.Printf("URL: https://discord.gg/%s\n", invite.Code)
				if invite.Guild != nil {
					fmt.Printf("Guild: %s (%s)\n", invite.Guild.Name, invite.Guild.ID)
				}
				if invite.Channel != nil {
					fmt.Printf("Channel: %s (%s)\n", invite.Channel.Name, invite.Channel.ID)
				}
				if invite.Inviter != nil {
					fmt.Printf("Inviter: %s#%s (%s)\n", invite.Inviter.Username, invite.Inviter.Discriminator, invite.Inviter.ID)
				}
				fmt.Printf("Uses: %d\n", invite.Uses)
				fmt.Printf("Max Uses: %d\n", invite.MaxUses)
				if invite.MaxAge > 0 {
					fmt.Printf("Max Age: %d seconds\n", invite.MaxAge)
				} else {
					fmt.Printf("Max Age: Never expires\n")
				}
				fmt.Printf("Temporary: %v\n", invite.Temporary)
				fmt.Printf("Created At: %s\n", invite.CreatedAt.Format("2006-01-02 15:04:05"))
			} else {
				if err := output.Print(invite); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// ListInvitesCommand lists invites
func ListInvitesCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List invites for a guild or channel",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "guild",
				Usage: "Guild ID to list invites from",
			},
			&cli.StringFlag{
				Name:  "channel",
				Usage: "Channel ID to list invites from",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			var invites []*discordgo.Invite

			if guildID := c.String("guild"); guildID != "" {
				invites, err = cliCtx.Client.GetGuildInvites(guildID)
			} else if channelID := c.String("channel"); channelID != "" {
				invites, err = cliCtx.Client.GetChannelInvites(channelID)
			} else {
				return utils.ValidationError("either --guild or --channel is required")
			}

			if err != nil {
				return utils.DiscordErrorf("failed to list invites: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				data := [][]string{}
				for _, invite := range invites {
					uses := fmt.Sprintf("%d", invite.Uses)
					maxUses := fmt.Sprintf("%d", invite.MaxUses)
					if invite.MaxUses == 0 {
						maxUses = "∞"
					}
					expires := "Never"
					if invite.MaxAge > 0 {
						expires = fmt.Sprintf("%ds", invite.MaxAge)
					}
					inviter := "-"
					if invite.Inviter != nil {
						inviter = invite.Inviter.Username
					}
					data = append(data, []string{
						invite.Code,
						inviter,
						uses,
						maxUses,
						expires,
					})
				}
				header := []string{"Code", "Inviter", "Uses", "Max Uses", "Expires"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(invites); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// CreateInviteCommand creates a new invite
func CreateInviteCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a new invite for a channel",
		ArgsUsage: "[channel-id]",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "max-age",
				Usage: "Max age in seconds (0 = never expires)",
				Value: 86400, // 24 hours default
			},
			&cli.IntFlag{
				Name:  "max-uses",
				Usage: "Max number of uses (0 = unlimited)",
				Value: 0,
			},
			&cli.BoolFlag{
				Name:  "temporary",
				Usage: "Grant temporary membership",
			},
			&cli.BoolFlag{
				Name:  "unique",
				Usage: "Create a unique invite",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("channel ID is required")
			}
			channelID := c.Args().First()

			maxAge := int(c.Int("max-age"))
			maxUses := int(c.Int("max-uses"))
			temporary := c.Bool("temporary")
		
unique := c.Bool("unique")

			invite, err := cliCtx.Client.CreateChannelInvite(channelID, maxAge, maxUses, temporary, unique)
			if err != nil {
				return utils.DiscordErrorf("failed to create invite: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Invite created successfully!\n")
				fmt.Printf("Code: %s\n", invite.Code)
				fmt.Printf("URL: https://discord.gg/%s\n", invite.Code)
				if invite.Channel != nil {
					fmt.Printf("Channel: %s\n", invite.Channel.Name)
				}
				if invite.MaxAge > 0 {
					fmt.Printf("Expires in: %d seconds\n", invite.MaxAge)
				} else {
					fmt.Printf("Expires: Never\n")
				}
				if invite.MaxUses > 0 {
					fmt.Printf("Max uses: %d\n", invite.MaxUses)
				} else {
					fmt.Printf("Max uses: Unlimited\n")
				}
			} else {
				result := map[string]interface{}{
					"success": true,
					"invite":  invite,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// DeleteInviteCommand deletes an invite
func DeleteInviteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete an invite",
		ArgsUsage: "[invite-code]",
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
				return utils.ValidationError("invite code is required")
			}
			inviteCode := c.Args().First()

			// Get invite info for confirmation
			invite, err := cliCtx.Client.GetInvite(inviteCode)
			if err != nil {
				return utils.DiscordErrorf("failed to get invite info: %w", err)
			}

			// Confirmation prompt
			if !c.Bool("force") {
				channelName := "Unknown"
				if invite.Channel != nil {
					channelName = invite.Channel.Name
				}
				fmt.Printf("You are about to delete invite: %s (Channel: %s)\n", inviteCode, channelName)
				fmt.Print("Are you sure? (yes/no): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if response != "yes\n" && response != "yes\r\n" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := cliCtx.Client.DeleteInvite(inviteCode); err != nil {
				return utils.DiscordErrorf("failed to delete invite: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully deleted invite: %s\n", inviteCode)
			} else {
				result := map[string]interface{}{
					"success":     true,
					"invite_code": inviteCode,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

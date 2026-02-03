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

func RolesCommand() *cli.Command {
	return &cli.Command{
		Name:  "roles",
		Usage: "Role management",
		Commands: []*cli.Command{
			RolesListCommand(),
			RolesGetCommand(),
			RolesCreateCommand(),
			RolesEditCommand(),
			RolesDeleteCommand(),
			RolesAssignCommand(),
			RolesRemoveCommand(),
		},
	}
}

func RolesListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List roles in a guild",
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

			guildID := c.String("guild")

			roles, err := cliCtx.Client.GetGuildRoles(guildID)
			if err != nil {
				return utils.DiscordErrorf("failed to list roles: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				data := [][]string{}
				for _, role := range roles {
					color := fmt.Sprintf("#%06X", role.Color)
					hoist := "No"
					if role.Hoist {
						hoist = "Yes"
					}
					data = append(data, []string{
						role.ID,
						role.Name,
						color,
						hoist,
					})
				}
				header := []string{"ID", "Name", "Color", "Hoisted"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(roles); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func RolesGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get role details",
		ArgsUsage: "[role-id]",
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
				return utils.ValidationError("role ID is required")
			}
			roleID := c.Args().First()
			guildID := c.String("guild")

			role, err := cliCtx.Client.GetGuildRole(guildID, roleID)
			if err != nil {
				return utils.DiscordErrorf("failed to get role: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("ID: %s\n", role.ID)
				fmt.Printf("Name: %s\n", role.Name)
				fmt.Printf("Color: #%06X\n", role.Color)
				fmt.Printf("Hoist: %v\n", role.Hoist)
				fmt.Printf("Mentionable: %v\n", role.Mentionable)
				fmt.Printf("Position: %d\n", role.Position)
				fmt.Printf("Permissions: %d\n", role.Permissions)
			} else {
				if err := output.Print(role); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func RolesCreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new role",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Role name",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "color",
				Usage: "Role color (hex integer, e.g., 0x3498db)",
			},
			&cli.BoolFlag{
				Name:  "hoist",
				Usage: "Display role separately in member list",
			},
			&cli.BoolFlag{
				Name:  "mentionable",
				Usage: "Allow anyone to mention this role",
			},
			&cli.IntFlag{
				Name:  "permissions",
				Usage: "Role permissions (bitwise value)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			guildID := c.String("guild")

			data := &discordgo.RoleParams{
				Name: c.String("name"),
			}

			if c.IsSet("color") {
				color := c.Int("color")
				data.Color = &color
			}
			if c.IsSet("hoist") {
				hoist := c.Bool("hoist")
				data.Hoist = &hoist
			}
			if c.IsSet("mentionable") {
				mentionable := c.Bool("mentionable")
				data.Mentionable = &mentionable
			}
			if c.IsSet("permissions") {
				perms := int64(c.Int("permissions"))
				data.Permissions = &perms
			}

			role, err := cliCtx.Client.CreateGuildRole(guildID, data)
			if err != nil {
				return utils.DiscordErrorf("failed to create role: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Role created successfully!\n")
				fmt.Printf("ID: %s\n", role.ID)
				fmt.Printf("Name: %s\n", role.Name)
				fmt.Printf("Color: #%06X\n", role.Color)
			} else {
				result := map[string]interface{}{
					"success": true,
					"role":    role,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// RolesEditCommand edits an existing role
func RolesEditCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit a role",
		ArgsUsage: "[role-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "New role name",
			},
			&cli.IntFlag{
				Name:  "color",
				Usage: "New role color (hex integer)",
			},
			&cli.BoolFlag{
				Name:  "hoist",
				Usage: "Display role separately in member list",
			},
			&cli.BoolFlag{
				Name:  "mentionable",
				Usage: "Allow anyone to mention this role",
			},
			&cli.IntFlag{
				Name:  "permissions",
				Usage: "Role permissions (bitwise value)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("role ID is required")
			}
			roleID := c.Args().First()
			guildID := c.String("guild")

			data := &discordgo.RoleParams{}

			if name := c.String("name"); name != "" {
				data.Name = name
			}
			if c.IsSet("color") {
				color := c.Int("color")
				data.Color = &color
			}
			if c.IsSet("hoist") {
				hoist := c.Bool("hoist")
				data.Hoist = &hoist
			}
			if c.IsSet("mentionable") {
				mentionable := c.Bool("mentionable")
				data.Mentionable = &mentionable
			}
			if c.IsSet("permissions") {
				perms := int64(c.Int("permissions"))
				data.Permissions = &perms
			}

			role, err := cliCtx.Client.EditGuildRole(guildID, roleID, data)
			if err != nil {
				return utils.DiscordErrorf("failed to edit role: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Role updated successfully!\n")
				fmt.Printf("ID: %s\n", role.ID)
				fmt.Printf("Name: %s\n", role.Name)
			} else {
				result := map[string]interface{}{
					"success": true,
					"role":    role,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// RolesDeleteCommand deletes a role
func RolesDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a role",
		ArgsUsage: "[role-id]",
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
				return utils.ValidationError("role ID is required")
			}
			roleID := c.Args().First()
			guildID := c.String("guild")

			// Get role info for confirmation
			role, err := cliCtx.Client.GetGuildRole(guildID, roleID)
			if err != nil {
				return utils.DiscordErrorf("failed to get role info: %w", err)
			}

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to delete role: %s (ID: %s)\n", role.Name, roleID)
				fmt.Print("Are you sure? (yes/no): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if response != "yes\n" && response != "yes\r\n" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := cliCtx.Client.DeleteGuildRole(guildID, roleID); err != nil {
				return utils.DiscordErrorf("failed to delete role: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully deleted role: %s\n", role.Name)
			} else {
				result := map[string]interface{}{
					"success": true,
					"role_id": roleID,
					"name":    role.Name,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// RolesAssignCommand assigns a role to a user
func RolesAssignCommand() *cli.Command {
	return &cli.Command{
		Name:      "assign",
		Usage:     "Assign a role to a user",
		ArgsUsage: "[user-id] [role-id]",
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

			if c.NArg() < 2 {
				return utils.ValidationError("user ID and role ID are required")
			}
			userID := c.Args().Get(0)
			roleID := c.Args().Get(1)
			guildID := c.String("guild")

			if err := cliCtx.Client.GuildMemberRoleAdd(guildID, userID, roleID); err != nil {
				return utils.DiscordErrorf("failed to assign role: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully assigned role %s to user %s\n", roleID, userID)
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

// RolesRemoveCommand removes a role from a user
func RolesRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove a role from a user",
		ArgsUsage: "[user-id] [role-id]",
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

			if c.NArg() < 2 {
				return utils.ValidationError("user ID and role ID are required")
			}
			userID := c.Args().Get(0)
			roleID := c.Args().Get(1)
			guildID := c.String("guild")

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
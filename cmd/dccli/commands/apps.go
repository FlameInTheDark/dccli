package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/pkg/discord"
	"github.com/FlameInTheDark/dccli/pkg/dprint"
	"github.com/FlameInTheDark/dccli/pkg/utils"
)

func ListAppCommandsCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Returns a list of application commands",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "guild",
				Usage: "Guild ID",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			var commands []discord.DiscordCommand
			if c.String("guild") == "" {
				commands, err = cliCtx.Client.GetCurrentAppGlobalCommands()
				if err != nil {
					return utils.DiscordErrorf("failed to get global commands: %w", err)
				}
			} else {
				id, err := cliCtx.Client.GetCurrentAppID()
				if err != nil {
					return utils.DiscordErrorf("failed to get application ID: %w", err)
				}
				commands, err = cliCtx.Client.GetComands(id, c.String("guild"))
				if err != nil {
					return utils.DiscordErrorf("failed to get guild commands: %w", err)
				}
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				data := [][]string{}
				for _, cmd := range commands {
					data = append(data, []string{cmd.ID, cmd.Name, strconv.Itoa(cmd.OptionsCount)})
				}
				header := []string{"ID", "Name", "Options"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(commands); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func RemoveAppCmdCommand() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Remove an application command",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "guild",
				Usage: "Guild ID",
			},
			&cli.StringFlag{
				Name:  "command",
				Usage: "Command ID",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			err = cliCtx.Client.RemoveCurrentAppCommand(c.String("guild"), c.String("command"))
			if err != nil {
				return utils.DiscordErrorf("failed to remove command: %w", err)
			}

			output := cliCtx.GetOutputManager()
			if output.GetFormat() == dprint.FormatTable {
				// Simple success message for table format
				fmt.Println("Command deleted successfully!")
			} else {
				// Structured output for JSON/YAML
				result := map[string]interface{}{
					"success":    true,
					"command_id": c.String("command"),
					"guild_id":   c.String("guild"),
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func DescribeAppCmdCommand() *cli.Command {
	return &cli.Command{
		Name:  "describe",
		Usage: "Describe command",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "guild",
				Usage: "Guild ID",
			},
			&cli.StringFlag{
				Name:  "command",
				Usage: "Command ID",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			cmd, err := cliCtx.Client.DescribeCurrentAppCommand(c.String("guild"), c.String("command"))
			if err != nil {
				return utils.DiscordErrorf("failed to describe command: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				// Use template for table format
				err = dprint.PrintTemplate(dprint.CommandDescribeTemplate, cmd)
				if err != nil {
					return err
				}

				data := [][]string{}
				for _, o := range cmd.Options {
					var required string
					if o.Required {
						required = "YES"
					}
					data = append(data, []string{o.Name, required, o.Description})
				}

				header := []string{"Name", "Required", "Description"}
				dprint.Table(header, data)
			} else {
				// Output structured data for JSON/YAML
				if err := output.Print(cmd); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func CreateAppCmdCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new application command",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Command name (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "description",
				Usage:    "Command description (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "guild",
				Usage: "Guild ID (optional, creates guild-specific command if provided)",
			},
			&cli.IntFlag{
				Name:  "type",
				Usage: "Command type: 1=slash (default), 2=user, 3=message",
				Value: 1,
			},
			&cli.StringFlag{
				Name:  "options-file",
				Usage: "JSON file containing command options",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			cmd := &discordgo.ApplicationCommand{
				Name:        c.String("name"),
				Description: c.String("description"),
				Type:        discordgo.ApplicationCommandType(c.Int("type")),
			}

			// Load options from file if provided
			if optionsFile := c.String("options-file"); optionsFile != "" {
				data, err := os.ReadFile(optionsFile)
				if err != nil {
					return utils.ValidationErrorf("failed to read options file: %w", err)
				}
				if err := json.Unmarshal(data, &cmd.Options); err != nil {
					return utils.ValidationErrorf("failed to parse options file: %w", err)
				}
			}

			guildID := c.String("guild")
			createdCmd, err := cliCtx.Client.CreateCurrentAppCommand(guildID, cmd)
			if err != nil {
				return utils.DiscordErrorf("failed to create command: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Command created successfully!\n")
				fmt.Printf("ID: %s\n", createdCmd.ID)
				fmt.Printf("Name: %s\n", createdCmd.Name)
				fmt.Printf("Description: %s\n", createdCmd.Description)
				fmt.Printf("Type: %d\n", createdCmd.Type)
				if guildID != "" {
					fmt.Printf("Guild: %s\n", guildID)
				} else {
					fmt.Printf("Scope: Global\n")
				}
			} else {
				result := map[string]interface{}{
					"success":     true,
					"id":          createdCmd.ID,
					"name":        createdCmd.Name,
					"description": createdCmd.Description,
					"type":        createdCmd.Type,
					"guild_id":    guildID,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func EditAppCmdCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit an existing application command",
		ArgsUsage: "[command-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "guild",
				Usage: "Guild ID (optional)",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "New command name",
			},
			&cli.StringFlag{
				Name:  "description",
				Usage: "New command description",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("command ID is required")
			}
			commandID := c.Args().First()

			// Get existing command first to preserve unmodified fields
			guildID := c.String("guild")
			existingCmd, err := cliCtx.Client.DescribeCurrentAppCommand(guildID, commandID)
			if err != nil {
				return utils.DiscordErrorf("failed to get existing command: %w", err)
			}

			// Build update command
			cmd := &discordgo.ApplicationCommand{
				ID:   commandID,
				Name: existingCmd.Name,
			}

			if name := c.String("name"); name != "" {
				cmd.Name = name
			}
			if desc := c.String("description"); desc != "" {
				cmd.Description = desc
			}

			updatedCmd, err := cliCtx.Client.EditCurrentAppCommand(guildID, commandID, cmd)
			if err != nil {
				return utils.DiscordErrorf("failed to edit command: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Command updated successfully!\n")
				fmt.Printf("ID: %s\n", updatedCmd.ID)
				fmt.Printf("Name: %s\n", updatedCmd.Name)
				fmt.Printf("Description: %s\n", updatedCmd.Description)
			} else {
				result := map[string]interface{}{
					"success":     true,
					"id":          updatedCmd.ID,
					"name":        updatedCmd.Name,
					"description": updatedCmd.Description,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func DeleteAllAppCmdCommand() *cli.Command {
	return &cli.Command{
		Name:  "delete-all",
		Usage: "Delete all application commands",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "guild",
				Usage: "Guild ID (optional, deletes guild commands if provided)",
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

			guildID := c.String("guild")

			// Get count of commands to be deleted
			var commands []discord.DiscordCommand
			if guildID == "" {
				commands, err = cliCtx.Client.GetCurrentAppGlobalCommands()
			} else {
				appID, err := cliCtx.Client.GetCurrentAppID()
				if err != nil {
					return utils.DiscordErrorf("failed to get application ID: %w", err)
				}
				commands, err = cliCtx.Client.GetComands(appID, guildID)
			}
			if err != nil {
				return utils.DiscordErrorf("failed to get commands: %w", err)
			}

			if len(commands) == 0 {
				fmt.Println("No commands to delete.")
				return nil
			}

			// Confirmation prompt
			if !c.Bool("force") {
				scope := "global"
				if guildID != "" {
					scope = fmt.Sprintf("guild %s", guildID)
				}
				fmt.Printf("You are about to delete %d %s command(s).\n", len(commands), scope)
				fmt.Print("Are you sure? (yes/no): ")
				var response string
				fmt.Scanln(&response)
				if response != "yes" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			// Delete all commands
			if err := cliCtx.Client.DeleteAllCurrentAppCommands(guildID); err != nil {
				return utils.DiscordErrorf("failed to delete commands: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully deleted %d command(s).\n", len(commands))
			} else {
				result := map[string]interface{}{
					"success":  true,
					"deleted":  len(commands),
					"guild_id": guildID,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func BulkOverwriteAppCmdCommand() *cli.Command {
	return &cli.Command{
		Name:  "bulk-overwrite",
		Usage: "Overwrite all commands with those from a JSON file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Usage:    "JSON file containing array of commands",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "guild",
				Usage: "Guild ID (optional, overwrites guild commands if provided)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			// Read and parse the JSON file
			data, err := os.ReadFile(c.String("file"))
			if err != nil {
				return utils.ValidationErrorf("failed to read file: %w", err)
			}

			var commands []*discordgo.ApplicationCommand
			if err := json.Unmarshal(data, &commands); err != nil {
				return utils.ValidationErrorf("failed to parse JSON: %w", err)
			}

			guildID := c.String("guild")
			createdCmds, err := cliCtx.Client.BulkOverwriteCurrentAppCommands(guildID, commands)
			if err != nil {
				return utils.DiscordErrorf("failed to overwrite commands: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				scope := "global"
				if guildID != "" {
					scope = fmt.Sprintf("guild %s", guildID)
				}
				fmt.Printf("Successfully overwrote %s commands with %d command(s).\n", scope, len(createdCmds))
				for _, cmd := range createdCmds {
					fmt.Printf("  - %s (ID: %s)\n", cmd.Name, cmd.ID)
				}
			} else {
				result := map[string]interface{}{
					"success":  true,
					"count":    len(createdCmds),
					"guild_id": guildID,
					"commands": createdCmds,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func AppCommandPermissionsCommand() *cli.Command {
	return &cli.Command{
		Name:  "permissions",
		Usage: "Manage command permissions",
		Commands: []*cli.Command{
			GetCommandPermissionsCommand(),
			SetCommandPermissionsCommand(),
		},
	}
}

func GetCommandPermissionsCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get permissions for a command",
		ArgsUsage: "[command-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID (required)",
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
				return utils.ValidationError("command ID is required")
			}
			commandID := c.Args().First()
			guildID := c.String("guild")

			perms, err := cliCtx.Client.GetCurrentAppCommandPermissions(guildID, commandID)
			if err != nil {
				return utils.DiscordErrorf("failed to get command permissions: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Command ID: %s\n", perms.ID)
				fmt.Printf("Application ID: %s\n", perms.ApplicationID)
				fmt.Printf("Guild ID: %s\n", perms.GuildID)
				fmt.Printf("Permissions:\n")
				if len(perms.Permissions) == 0 {
					fmt.Println("  No permissions set")
				} else {
					data := [][]string{}
					for _, p := range perms.Permissions {
						typeStr := "Unknown"
						switch p.Type {
						case discordgo.ApplicationCommandPermissionTypeRole:
							typeStr = "Role"
						case discordgo.ApplicationCommandPermissionTypeUser:
							typeStr = "User"
						case discordgo.ApplicationCommandPermissionTypeChannel:
							typeStr = "Channel"
						}
						data = append(data, []string{p.ID, typeStr, strconv.FormatBool(p.Permission)})
					}
					header := []string{"ID", "Type", "Permission"}
					dprint.Table(header, data)
				}
			} else {
				if err := output.Print(perms); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func SetCommandPermissionsCommand() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Set permissions for a command",
		ArgsUsage: "[command-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "permissions",
				Usage:    "JSON string or file path containing permissions array",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "from-file",
				Usage: "Treat --permissions as a file path",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("command ID is required")
			}
			commandID := c.Args().First()
			guildID := c.String("guild")

			var permissionsData []byte
			if c.Bool("from-file") {
				permissionsData, err = os.ReadFile(c.String("permissions"))
				if err != nil {
					return utils.ValidationErrorf("failed to read permissions file: %w", err)
				}
			} else {
				permissionsData = []byte(c.String("permissions"))
			}

			var permissionsList discordgo.ApplicationCommandPermissionsList
			if err := json.Unmarshal(permissionsData, &permissionsList); err != nil {
				return utils.ValidationErrorf("failed to parse permissions JSON: %w", err)
			}

			perms, err := cliCtx.Client.SetCurrentAppCommandPermissions(guildID, commandID, &permissionsList)
			if err != nil {
				return utils.DiscordErrorf("failed to set command permissions: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Permissions updated successfully!\n")
				fmt.Printf("Command ID: %s\n", perms.ID)
				fmt.Printf("Total permissions: %d\n", len(perms.Permissions))
			} else {
				result := map[string]interface{}{
					"success":     true,
					"command_id":  perms.ID,
					"guild_id":    perms.GuildID,
					"permissions": perms.Permissions,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
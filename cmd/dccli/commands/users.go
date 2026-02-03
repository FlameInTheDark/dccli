package commands

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/pkg/dprint"
	"github.com/FlameInTheDark/dccli/pkg/utils"
)

func UsersCommand() *cli.Command {
	return &cli.Command{
		Name:  "users",
		Usage: "User & Bot operations",
		Commands: []*cli.Command{
			UsersGetCommand(),
			UsersGuildsCommand(),
			UsersConnectionsCommand(),
		},
	}
}

func UsersGetCommand() *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Get current user/bot info",
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			user, err := cliCtx.Client.GetCurrentUser()
			if err != nil {
				return utils.DiscordErrorf("failed to get user info: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("ID: %s\n", user.ID)
				fmt.Printf("Username: %s\n", user.Username)
				fmt.Printf("Discriminator: %s\n", user.Discriminator)
				if user.Avatar != "" {
					fmt.Printf("Avatar: %s\n", user.Avatar)
				}
				if user.Bot {
					fmt.Printf("Type: Bot\n")
				} else {
					fmt.Printf("Type: User\n")
				}
				if user.Email != "" {
					fmt.Printf("Email: %s\n", user.Email)
				}
				if user.Verified {
					fmt.Printf("Verified: Yes\n")
				}
			} else {
				if err := output.Print(user); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func UsersGuildsCommand() *cli.Command {
	return &cli.Command{
		Name:  "guilds",
		Usage: "List user guilds",
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

func UsersConnectionsCommand() *cli.Command {
	return &cli.Command{
		Name:  "connections",
		Usage: "Get user connections",
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			connections, err := cliCtx.Client.GetUserConnections()
			if err != nil {
				return utils.DiscordErrorf("failed to get user connections: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				if len(connections) == 0 {
					fmt.Println("No connections found")
					return nil
				}

				data := [][]string{}
				for _, conn := range connections {
					data = append(data, []string{
						conn.ID,
						conn.Name,
						conn.Type,
					})
				}
				header := []string{"ID", "Name", "Type"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(connections); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
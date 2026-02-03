package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/pkg/dprint"
	"github.com/FlameInTheDark/dccli/pkg/utils"
)

func AutoModCommand() *cli.Command {
	return &cli.Command{
		Name:  "automod",
		Usage: "Auto-moderation rule management",
		Commands: []*cli.Command{
			AutoModListCommand(),
			AutoModGetCommand(),
			AutoModCreateCommand(),
			AutoModEditCommand(),
			AutoModDeleteCommand(),
		},
	}
}

// AutoModListCommand lists auto-moderation rules in a guild
func AutoModListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List auto-moderation rules in a guild",
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

			rules, err := cliCtx.Client.GetAutoModRules(guildID)
			if err != nil {
				return utils.DiscordErrorf("failed to list auto-mod rules: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				data := [][]string{}
				for _, rule := range rules {
					enabled := "No"
					if rule.Enabled != nil && *rule.Enabled {
						enabled = "Yes"
					}
					data = append(data, []string{
						rule.ID,
						rule.Name,
						autoModRuleTypeToString(rule.EventType),
						enabled,
						fmt.Sprintf("%d", len(rule.Actions)),
					})
				}
				header := []string{"ID", "Name", "Type", "Enabled", "Actions"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(rules); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// AutoModGetCommand gets auto-moderation rule details
func AutoModGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get auto-moderation rule details",
		ArgsUsage: "[rule-id]",
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
				return utils.ValidationError("rule ID is required")
			}
			ruleID := c.Args().First()
			guildID := c.String("guild")

			rule, err := cliCtx.Client.GetAutoModRule(guildID, ruleID)
			if err != nil {
				return utils.DiscordErrorf("failed to get auto-mod rule: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("ID: %s\n", rule.ID)
				fmt.Printf("Name: %s\n", rule.Name)
				if rule.CreatorID != "" {
					fmt.Printf("Creator ID: %s\n", rule.CreatorID)
				}
				fmt.Printf("Event Type: %s\n", autoModRuleTypeToString(rule.EventType))
				fmt.Printf("Trigger Type: %s\n", autoModTriggerTypeToString(rule.TriggerType))
				if rule.Enabled != nil {
					fmt.Printf("Enabled: %v\n", *rule.Enabled)
				}
				if rule.ExemptRoles != nil && len(*rule.ExemptRoles) > 0 {
					fmt.Printf("Exempt Roles: %v\n", *rule.ExemptRoles)
				}
				if rule.ExemptChannels != nil && len(*rule.ExemptChannels) > 0 {
					fmt.Printf("Exempt Channels: %v\n", *rule.ExemptChannels)
				}
				fmt.Printf("Actions (%d):\n", len(rule.Actions))
				for i, action := range rule.Actions {
					fmt.Printf("  %d. %s\n", i+1, autoModActionTypeToString(action.Type))
					if action.Metadata != nil {
						if action.Metadata.ChannelID != "" {
							fmt.Printf("     Channel: %s\n", action.Metadata.ChannelID)
						}
					}
				}
			} else {
				if err := output.Print(rule); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// AutoModCreateCommand creates a new auto-moderation rule
func AutoModCreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new auto-moderation rule",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Rule name (1-100 characters)",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "type",
				Usage:    "Rule trigger type: 1=Keyword, 3=Spam, 4=KeywordPreset, 5=MentionSpam",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "actions",
				Usage:    `Actions as JSON array, e.g., '[{"type":1},{"type":2,"metadata":{"duration_seconds":60}}]'`,
				Required: true,
			},
			&cli.StringFlag{
				Name:  "keywords",
				Usage: `Keywords for keyword filter (JSON array of strings), e.g., '["bad", "word"]'`,
			},
			&cli.StringFlag{
				Name:  "presets",
				Usage: `Keyword presets for KeywordPreset filter (JSON array), e.g., '[1, 2]' (1=Profanity, 2=SexualContent, 3=Slurs)`,
			},
			&cli.IntFlag{
				Name:  "mention-limit",
				Usage: "Mention limit for MentionSpam filter",
			},
			&cli.StringSliceFlag{
				Name:  "exempt-roles",
				Usage: "Role IDs exempt from this rule",
			},
			&cli.StringSliceFlag{
				Name:  "exempt-channels",
				Usage: "Channel IDs exempt from this rule",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			guildID := c.String("guild")
			name := c.String("name")
			triggerType := discordgo.AutoModerationRuleTriggerType(c.Int("type"))

			// Validate trigger type
			if triggerType != discordgo.AutoModerationEventTriggerKeyword &&
				triggerType != discordgo.AutoModerationEventTriggerSpam &&
				triggerType != discordgo.AutoModerationEventTriggerKeywordPreset {
				return utils.ValidationError("invalid trigger type: must be 1 (Keyword), 2 (Harmful Link), 3 (Spam), or 4 (KeywordPreset)")
			}

			// Parse actions
			var actions []discordgo.AutoModerationAction
			if err := json.Unmarshal([]byte(c.String("actions")), &actions); err != nil {
				return utils.ValidationErrorf("failed to parse actions JSON: %w", err)
			}

			// Build trigger metadata based on type
			var triggerMetadata *discordgo.AutoModerationTriggerMetadata

			switch triggerType {
			case discordgo.AutoModerationEventTriggerKeyword:
				if c.String("keywords") == "" {
					return utils.ValidationError("keywords are required for Keyword filter type")
				}
				var keywords []string
				if err := json.Unmarshal([]byte(c.String("keywords")), &keywords); err != nil {
					return utils.ValidationErrorf("failed to parse keywords JSON: %w", err)
				}
				triggerMetadata = &discordgo.AutoModerationTriggerMetadata{
					KeywordFilter: keywords,
				}

			case discordgo.AutoModerationEventTriggerKeywordPreset:
				if c.String("presets") == "" {
					return utils.ValidationError("presets are required for KeywordPreset filter type")
				}
				var presets []discordgo.AutoModerationKeywordPreset
				if err := json.Unmarshal([]byte(c.String("presets")), &presets); err != nil {
					return utils.ValidationErrorf("failed to parse presets JSON: %w", err)
				}
				triggerMetadata = &discordgo.AutoModerationTriggerMetadata{
					Presets: presets,
				}
			}

			enabled := true
			exemptRoles := c.StringSlice("exempt-roles")
			exemptChannels := c.StringSlice("exempt-channels")

			params := &discordgo.AutoModerationRule{
				Name:            name,
				EventType:       discordgo.AutoModerationEventMessageSend,
				TriggerType:     triggerType,
				TriggerMetadata: triggerMetadata,
				Actions:         actions,
				Enabled:         &enabled,
				ExemptRoles:     &exemptRoles,
				ExemptChannels:  &exemptChannels,
			}

			rule, err := cliCtx.Client.CreateAutoModRule(guildID, params)
			if err != nil {
				return utils.DiscordErrorf("failed to create auto-mod rule: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Auto-moderation rule created successfully!\n")
				fmt.Printf("ID: %s\n", rule.ID)
				fmt.Printf("Name: %s\n", rule.Name)
				fmt.Printf("Type: %s\n", autoModTriggerTypeToString(rule.TriggerType))
			} else {
				result := map[string]interface{}{
					"success": true,
					"rule":    rule,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// AutoModEditCommand edits an auto-moderation rule
func AutoModEditCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit an auto-moderation rule",
		ArgsUsage: "[rule-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "New rule name",
			},
			&cli.StringFlag{
				Name:  "actions",
				Usage: `Actions as JSON array, e.g., '[{"type":1},{"type":2,"metadata":{"duration_seconds":60}}]'`,
			},
			&cli.BoolFlag{
				Name:  "enabled",
				Usage: "Enable or disable the rule",
			},
			&cli.StringSliceFlag{
				Name:  "exempt-roles",
				Usage: "Role IDs exempt from this rule",
			},
			&cli.StringSliceFlag{
				Name:  "exempt-channels",
				Usage: "Channel IDs exempt from this rule",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("rule ID is required")
			}
			ruleID := c.Args().First()
			guildID := c.String("guild")

			// Get existing rule first
			existingRule, err := cliCtx.Client.GetAutoModRule(guildID, ruleID)
			if err != nil {
				return utils.DiscordErrorf("failed to get auto-mod rule: %w", err)
			}

			params := &discordgo.AutoModerationRule{
				EventType:       existingRule.EventType,
				TriggerType:     existingRule.TriggerType,
				TriggerMetadata: existingRule.TriggerMetadata,
			}

			// Update name if provided
			if newName := c.String("name"); newName != "" {
				params.Name = newName
			} else {
				params.Name = existingRule.Name
			}

			// Update actions if provided
			if actionsStr := c.String("actions"); actionsStr != "" {
				var actions []discordgo.AutoModerationAction
				if err := json.Unmarshal([]byte(actionsStr), &actions); err != nil {
					return utils.ValidationErrorf("failed to parse actions JSON: %w", err)
				}
				params.Actions = actions
			} else {
				params.Actions = existingRule.Actions
			}

			// Update enabled status if provided
			if c.IsSet("enabled") {
				enabled := c.Bool("enabled")
				params.Enabled = &enabled
			} else {
				params.Enabled = existingRule.Enabled
			}

			// Update exempt roles if provided
			if exemptRoles := c.StringSlice("exempt-roles"); len(exemptRoles) > 0 {
				params.ExemptRoles = &exemptRoles
			} else {
				params.ExemptRoles = existingRule.ExemptRoles
			}

			// Update exempt channels if provided
			if exemptChannels := c.StringSlice("exempt-channels"); len(exemptChannels) > 0 {
				params.ExemptChannels = &exemptChannels
			} else {
				params.ExemptChannels = existingRule.ExemptChannels
			}

			rule, err := cliCtx.Client.EditAutoModRule(guildID, ruleID, params)
			if err != nil {
				return utils.DiscordErrorf("failed to edit auto-mod rule: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Auto-moderation rule updated successfully!\n")
				fmt.Printf("ID: %s\n", rule.ID)
				fmt.Printf("Name: %s\n", rule.Name)
				if rule.Enabled != nil {
					fmt.Printf("Enabled: %v\n", *rule.Enabled)
				}
			} else {
				result := map[string]interface{}{
					"success": true,
					"rule":    rule,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// AutoModDeleteCommand deletes an auto-moderation rule
func AutoModDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete an auto-moderation rule",
		ArgsUsage: "[rule-id]",
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
				return utils.ValidationError("rule ID is required")
			}
			ruleID := c.Args().First()
			guildID := c.String("guild")

			// Get rule info for confirmation
			rule, err := cliCtx.Client.GetAutoModRule(guildID, ruleID)
			if err != nil {
				return utils.DiscordErrorf("failed to get auto-mod rule info: %w", err)
			}

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to delete auto-mod rule: %s (ID: %s)\n", rule.Name, ruleID)
				fmt.Print("Are you sure? (yes/no): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if response != "yes\n" && response != "yes\r\n" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := cliCtx.Client.DeleteAutoModRule(guildID, ruleID); err != nil {
				return utils.DiscordErrorf("failed to delete auto-mod rule: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully deleted auto-mod rule: %s\n", rule.Name)
			} else {
				result := map[string]interface{}{
					"success": true,
					"rule_id": ruleID,
					"name":    rule.Name,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// Helper functions

func autoModRuleTypeToString(eventType discordgo.AutoModerationRuleEventType) string {
	switch eventType {
	case discordgo.AutoModerationEventMessageSend:
		return "Message Send"
	default:
		return "Unknown"
	}
}

func autoModTriggerTypeToString(triggerType discordgo.AutoModerationRuleTriggerType) string {
	switch triggerType {
	case discordgo.AutoModerationEventTriggerKeyword:
		return "Keyword"
	case discordgo.AutoModerationEventTriggerHarmfulLink:
		return "Harmful Link"
	case discordgo.AutoModerationEventTriggerSpam:
		return "Spam"
	case discordgo.AutoModerationEventTriggerKeywordPreset:
		return "Keyword Preset"
	default:
		return "Unknown"
	}
}

func autoModActionTypeToString(actionType discordgo.AutoModerationActionType) string {
	switch actionType {
	case discordgo.AutoModerationRuleActionBlockMessage:
		return "Block Message"
	case discordgo.AutoModerationRuleActionSendAlertMessage:
		return "Send Alert"
	case discordgo.AutoModerationRuleActionTimeout:
		return "Timeout User"
	default:
		return "Unknown"
	}
}
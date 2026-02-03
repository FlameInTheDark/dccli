package discord

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"

	"github.com/FlameInTheDark/dccli/pkg/voice"
)

var (
	ErrCannotGetID = errors.New("cannot get application ID")
)

type DiscordClient struct {
	token    string
	session  *discordgo.Session
	voiceMgr *voice.ConnectionManager
}

func NewClient(token string) (*DiscordClient, error) {
	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	// Set intents to listen for messages and guild events
	sess.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentsMessageContent

	err = sess.Open()
	if err != nil {
		return nil, err
	}

	client := &DiscordClient{
		session: sess,
		token:   token,
	}
	client.voiceMgr = voice.NewConnectionManager(sess)

	return client, nil
}

// JoinVoiceChannel joins a voice channel in a guild
func (c *DiscordClient) JoinVoiceChannel(guildID, channelID string) (*voice.Connection, error) {
	return c.voiceMgr.Join(guildID, channelID)
}

// LeaveVoiceChannel leaves a voice channel in a guild
func (c *DiscordClient) LeaveVoiceChannel(guildID string) error {
	return c.voiceMgr.Leave(guildID)
}

// GetVoiceConnection returns the voice connection for a guild
func (c *DiscordClient) GetVoiceConnection(guildID string) (*voice.Connection, bool) {
	return c.voiceMgr.GetConnection(guildID)
}

// Session returns the underlying discordgo session
func (c *DiscordClient) Session() *discordgo.Session {
	return c.session
}

type DiscordApplication struct {
	ID        string
	Name      string
	NCommands int
}

func (c *DiscordClient) GetApplications() ([]DiscordApplication, error) {
	apps, err := c.session.Applications()
	if err != nil {
		return nil, err
	}
	var applications []DiscordApplication
	for _, app := range apps {
		var cmdsCount int
		cmds, err := c.session.ApplicationCommands(app.ID, "")
		if err == nil {
			cmdsCount = len(cmds)
		}
		applications = append(applications, DiscordApplication{
			ID:        app.ID,
			Name:      app.Name,
			NCommands: cmdsCount,
		})
	}
	return applications, nil
}

type DiscordCommand struct {
	ID           string
	Name         string
	OptionsCount int
}

func (c *DiscordClient) GetComands(appID, guildID string) ([]DiscordCommand, error) {
	cmds, err := c.session.ApplicationCommands(appID, guildID)
	if err != nil {
		return nil, err
	}
	var commands []DiscordCommand
	for _, cmd := range cmds {
		commands = append(commands, DiscordCommand{
			ID:           cmd.ID,
			Name:         cmd.Name,
			OptionsCount: len(cmd.Options),
		})
	}
	return commands, nil
}

func (c *DiscordClient) GetCurrentAppGlobalCommands() ([]DiscordCommand, error) {
	return c.GetComands(c.session.State.User.ID, "")
}

func (c *DiscordClient) GetCurrentAppID() (string, error) {
	if c.session.State != nil {
		if c.session.State.User != nil {
			return c.session.State.User.ID, nil
		}
	}
	return "", ErrCannotGetID
}

func (c *DiscordClient) RemoveCurrentAppCommand(guildID, commandID string) error {
	id, err := c.GetCurrentAppID()
	if err != nil {
		return err
	}
	return c.RemoveGuildCommand(id, commandID, guildID)
}

func (c *DiscordClient) RemoveGuildCommand(appID, commandID, guildID string) error {
	return c.session.ApplicationCommandDelete(appID, guildID, commandID)
}

type DiscordCommandDescription struct {
	ID      string
	Name    string
	Options []DiscordCommandOption
}

type DiscordCommandOption struct {
	Name        string
	Required    bool
	Description string
}

func (c *DiscordClient) DescribeCommand(appID, commandID, guildID string) (*DiscordCommandDescription, error) {
	cmd, err := c.session.ApplicationCommand(appID, guildID, commandID)
	if err != nil {
		return nil, err
	}
	var command DiscordCommandDescription
	command.ID = cmd.ID
	command.Name = cmd.Name

	for _, o := range cmd.Options {
		command.Options = append(command.Options, DiscordCommandOption{
			Name:        o.Name,
			Required:    o.Required,
			Description: o.Description,
		})
	}

	return &command, nil
}

func (c *DiscordClient) DescribeCurrentAppCommand(guildID, commandID string) (*DiscordCommandDescription, error) {
	id, err := c.GetCurrentAppID()
	if err != nil {
		return nil, err
	}

	return c.DescribeCommand(id, commandID, guildID)
}

func (c *DiscordClient) CreateCommand(appID string, cmd *discordgo.ApplicationCommand) (*discordgo.ApplicationCommand, error) {
	return c.session.ApplicationCommandCreate(appID, "", cmd)
}

func (c *DiscordClient) CreateGuildCommand(appID, guildID string, cmd *discordgo.ApplicationCommand) (*discordgo.ApplicationCommand, error) {
	return c.session.ApplicationCommandCreate(appID, guildID, cmd)
}

func (c *DiscordClient) CreateCurrentAppCommand(guildID string, cmd *discordgo.ApplicationCommand) (*discordgo.ApplicationCommand, error) {
	id, err := c.GetCurrentAppID()
	if err != nil {
		return nil, err
	}
	return c.session.ApplicationCommandCreate(id, guildID, cmd)
}

func (c *DiscordClient) EditCommand(appID, guildID, commandID string, cmd *discordgo.ApplicationCommand) (*discordgo.ApplicationCommand, error) {
	return c.session.ApplicationCommandEdit(appID, guildID, commandID, cmd)
}

func (c *DiscordClient) EditCurrentAppCommand(guildID, commandID string, cmd *discordgo.ApplicationCommand) (*discordgo.ApplicationCommand, error) {
	id, err := c.GetCurrentAppID()
	if err != nil {
		return nil, err
	}
	return c.session.ApplicationCommandEdit(id, guildID, commandID, cmd)
}

func (c *DiscordClient) DeleteAllCommands(appID, guildID string) error {
	cmds, err := c.session.ApplicationCommands(appID, guildID)
	if err != nil {
		return err
	}

	for _, cmd := range cmds {
		if err := c.session.ApplicationCommandDelete(appID, guildID, cmd.ID); err != nil {
			return errors.Wrapf(err, "failed to delete command %s", cmd.ID)
		}
	}
	return nil
}

func (c *DiscordClient) DeleteAllCurrentAppCommands(guildID string) error {
	id, err := c.GetCurrentAppID()
	if err != nil {
		return err
	}
	return c.DeleteAllCommands(id, guildID)
}

func (c *DiscordClient) BulkOverwriteCommands(appID, guildID string, cmds []*discordgo.ApplicationCommand) ([]*discordgo.ApplicationCommand, error) {
	return c.session.ApplicationCommandBulkOverwrite(appID, guildID, cmds)
}

func (c *DiscordClient) BulkOverwriteCurrentAppCommands(guildID string, cmds []*discordgo.ApplicationCommand) ([]*discordgo.ApplicationCommand, error) {
	id, err := c.GetCurrentAppID()
	if err != nil {
		return nil, err
	}
	return c.session.ApplicationCommandBulkOverwrite(id, guildID, cmds)
}

func (c *DiscordClient) GetCommandPermissions(appID, guildID, commandID string) (*discordgo.GuildApplicationCommandPermissions, error) {
	return c.session.ApplicationCommandPermissions(appID, guildID, commandID)
}

func (c *DiscordClient) GetCurrentAppCommandPermissions(guildID, commandID string) (*discordgo.GuildApplicationCommandPermissions, error) {
	id, err := c.GetCurrentAppID()
	if err != nil {
		return nil, err
	}
	return c.session.ApplicationCommandPermissions(id, guildID, commandID)
}

func (c *DiscordClient) SetCommandPermissions(appID, guildID, commandID string, permissions *discordgo.ApplicationCommandPermissionsList) (*discordgo.GuildApplicationCommandPermissions, error) {
	err := c.session.ApplicationCommandPermissionsEdit(appID, guildID, commandID, permissions)
	if err != nil {
		return nil, err
	}
	return c.session.ApplicationCommandPermissions(appID, guildID, commandID)
}

func (c *DiscordClient) SetCurrentAppCommandPermissions(guildID, commandID string, permissions *discordgo.ApplicationCommandPermissionsList) (*discordgo.GuildApplicationCommandPermissions, error) {
	id, err := c.GetCurrentAppID()
	if err != nil {
		return nil, err
	}
	err = c.session.ApplicationCommandPermissionsEdit(id, guildID, commandID, permissions)
	if err != nil {
		return nil, err
	}
	return c.session.ApplicationCommandPermissions(id, guildID, commandID)
}

type DiscordGuild struct {
	ID   string
	Name string
}

func (c *DiscordClient) GuildList(limit int, beforeID, afterID string) ([]DiscordGuild, error) {
	guilds, err := c.session.UserGuilds(limit, beforeID, afterID, false)
	if err != nil {
		return nil, err
	}
	var result []DiscordGuild

	for _, g := range guilds {
		result = append(result, DiscordGuild{
			ID:   g.ID,
			Name: g.Name,
		})
	}
	return result, nil
}

type DiscordGuildDescription struct {
	ID            string
	Name          string
	OwnerID       string
	OwnerName     string
	OwnerNickname string
	EmojisCount   int
	RolesCount    int
}

func (c *DiscordClient) GuildDescribe(id string) (*DiscordGuildDescription, error) {
	g, err := c.session.Guild(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get guild")
	}

	own, err := c.session.GuildMember(g.ID, g.OwnerID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get guild owner")
	}

	return &DiscordGuildDescription{
		ID:            g.ID,
		Name:          g.Name,
		OwnerID:       g.OwnerID,
		OwnerName:     own.User.Username + "#" + own.User.Discriminator,
		OwnerNickname: own.Nick,
		EmojisCount:   len(g.Emojis),
		RolesCount:    len(g.Roles),
	}, nil
}

func (c *DiscordClient) LeaveGuild(guildID string) error {
	return c.session.GuildLeave(guildID)
}

func (c *DiscordClient) EditGuild(guildID string, params *discordgo.GuildParams) (*discordgo.Guild, error) {
	return c.session.GuildEdit(guildID, params)
}

func (c *DiscordClient) GetGuildChannels(guildID string) ([]*discordgo.Channel, error) {
	return c.session.GuildChannels(guildID)
}

func (c *DiscordClient) GetChannel(channelID string) (*discordgo.Channel, error) {
	return c.session.Channel(channelID)
}

func (c *DiscordClient) CreateGuildChannel(guildID string, data *discordgo.GuildChannelCreateData) (*discordgo.Channel, error) {
	return c.session.GuildChannelCreateComplex(guildID, *data)
}

func (c *DiscordClient) EditChannel(channelID string, data *discordgo.ChannelEdit) (*discordgo.Channel, error) {
	return c.session.ChannelEdit(channelID, data)
}

func (c *DiscordClient) DeleteChannel(channelID string) error {
	_, err := c.session.ChannelDelete(channelID)
	return err
}

func (c *DiscordClient) GetGuildRoles(guildID string) ([]*discordgo.Role, error) {
	return c.session.GuildRoles(guildID)
}

func (c *DiscordClient) GetGuildRole(guildID, roleID string) (*discordgo.Role, error) {
	roles, err := c.session.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role.ID == roleID {
			return role, nil
		}
	}
	return nil, fmt.Errorf("role not found")
}

func (c *DiscordClient) CreateGuildRole(guildID string, params *discordgo.RoleParams) (*discordgo.Role, error) {
	return c.session.GuildRoleCreate(guildID, params)
}

func (c *DiscordClient) EditGuildRole(guildID, roleID string, params *discordgo.RoleParams) (*discordgo.Role, error) {
	return c.session.GuildRoleEdit(guildID, roleID, params)
}

func (c *DiscordClient) DeleteGuildRole(guildID, roleID string) error {
	return c.session.GuildRoleDelete(guildID, roleID)
}

func (c *DiscordClient) GetGuildMembers(guildID string, limit int, after string) ([]*discordgo.Member, error) {
	return c.session.GuildMembers(guildID, after, limit)
}

func (c *DiscordClient) GetGuildMember(guildID, userID string) (*discordgo.Member, error) {
	return c.session.GuildMember(guildID, userID)
}

func (c *DiscordClient) GuildBanCreate(guildID, userID string, deleteDays int, reason string) error {
	return c.session.GuildBanCreateWithReason(guildID, userID, reason, deleteDays)
}

func (c *DiscordClient) GuildBanDelete(guildID, userID string) error {
	return c.session.GuildBanDelete(guildID, userID)
}

func (c *DiscordClient) GuildMemberDelete(guildID, userID string) error {
	return c.session.GuildMemberDelete(guildID, userID)
}

func (c *DiscordClient) GetGuildInvites(guildID string) ([]*discordgo.Invite, error) {
	return c.session.GuildInvites(guildID)
}

func (c *DiscordClient) CreateChannelInvite(channelID string, maxAge, maxUses int, temporary, unique bool) (*discordgo.Invite, error) {
	return c.session.ChannelInviteCreate(channelID, discordgo.Invite{
		MaxAge:    maxAge,
		MaxUses:   maxUses,
		Temporary: temporary,
		Unique:    unique,
	})
}

func (c *DiscordClient) GetInvite(inviteCode string) (*discordgo.Invite, error) {
	return c.session.Invite(inviteCode)
}

func (c *DiscordClient) DeleteInvite(inviteCode string) error {
	_, err := c.session.InviteDelete(inviteCode)
	return err
}

func (c *DiscordClient) GetChannelMessages(channelID string, limit int, before, after, around string) ([]*discordgo.Message, error) {
	return c.session.ChannelMessages(channelID, limit, before, after, around)
}

func (c *DiscordClient) GetChannelMessage(channelID, messageID string) (*discordgo.Message, error) {
	return c.session.ChannelMessage(channelID, messageID)
}

func (c *DiscordClient) SendChannelMessage(channelID string, msg *discordgo.MessageSend) (*discordgo.Message, error) {
	return c.session.ChannelMessageSendComplex(channelID, msg)
}

func (c *DiscordClient) EditChannelMessage(channelID, messageID string, msg *discordgo.MessageEdit) (*discordgo.Message, error) {
	return c.session.ChannelMessageEditComplex(msg)
}

func (c *DiscordClient) DeleteChannelMessage(channelID, messageID string) error {
	return c.session.ChannelMessageDelete(channelID, messageID)
}

func (c *DiscordClient) BulkDeleteMessages(channelID string, messageIDs []string) error {
	return c.session.ChannelMessagesBulkDelete(channelID, messageIDs)
}

func (c *DiscordClient) AddReaction(channelID, messageID, emoji string) error {
	return c.session.MessageReactionAdd(channelID, messageID, emoji)
}

func (c *DiscordClient) RemoveReaction(channelID, messageID, emoji, userID string) error {
	if userID == "" {
		return c.session.MessageReactionRemove(channelID, messageID, emoji, "@me")
	}
	return c.session.MessageReactionRemove(channelID, messageID, emoji, userID)
}

func (c *DiscordClient) GuildMemberRoleAdd(guildID, userID, roleID string) error {
	return c.session.GuildMemberRoleAdd(guildID, userID, roleID)
}

func (c *DiscordClient) GuildMemberRoleRemove(guildID, userID, roleID string) error {
	return c.session.GuildMemberRoleRemove(guildID, userID, roleID)
}

func (c *DiscordClient) GuildMemberTimeout(guildID, userID string, until *time.Time) error {
	return c.session.GuildMemberTimeout(guildID, userID, until)
}

func (c *DiscordClient) GuildMemberNickname(guildID, userID, nick string) error {
	return c.session.GuildMemberNickname(guildID, userID, nick)
}

func (c *DiscordClient) GetChannelWebhooks(channelID string) ([]*discordgo.Webhook, error) {
	return c.session.ChannelWebhooks(channelID)
}

func (c *DiscordClient) GetGuildWebhooks(guildID string) ([]*discordgo.Webhook, error) {
	return c.session.GuildWebhooks(guildID)
}

func (c *DiscordClient) GetWebhook(webhookID string) (*discordgo.Webhook, error) {
	return c.session.Webhook(webhookID)
}

func (c *DiscordClient) CreateWebhook(channelID, name, avatar string) (*discordgo.Webhook, error) {
	return c.session.WebhookCreate(channelID, name, avatar)
}

func (c *DiscordClient) EditWebhook(webhookID, name, avatar, channelID string) (*discordgo.Webhook, error) {
	return c.session.WebhookEdit(webhookID, name, avatar, channelID)
}

func (c *DiscordClient) DeleteWebhook(webhookID string) error {
	return c.session.WebhookDelete(webhookID)
}

func (c *DiscordClient) ExecuteWebhook(webhookID, token string, wait bool, params *discordgo.WebhookParams) (*discordgo.Message, error) {
	return c.session.WebhookExecute(webhookID, token, wait, params)
}

func (c *DiscordClient) GetCurrentUser() (*discordgo.User, error) {
	return c.session.User("@me")
}

func (c *DiscordClient) GetUserConnections() ([]*discordgo.UserConnection, error) {
	return c.session.UserConnections()
}

func (c *DiscordClient) GetGuildEmojis(guildID string) ([]*discordgo.Emoji, error) {
	return c.session.GuildEmojis(guildID)
}

func (c *DiscordClient) GetGuildEmoji(guildID, emojiID string) (*discordgo.Emoji, error) {
	return c.session.GuildEmoji(guildID, emojiID)
}

func (c *DiscordClient) CreateGuildEmoji(guildID string, params *discordgo.EmojiParams) (*discordgo.Emoji, error) {
	return c.session.GuildEmojiCreate(guildID, params)
}

func (c *DiscordClient) EditGuildEmoji(guildID, emojiID string, params *discordgo.EmojiParams) (*discordgo.Emoji, error) {
	return c.session.GuildEmojiEdit(guildID, emojiID, params)
}

func (c *DiscordClient) DeleteGuildEmoji(guildID, emojiID string) error {
	return c.session.GuildEmojiDelete(guildID, emojiID)
}

func (c *DiscordClient) GetGuildStickers(guildID string) ([]*discordgo.Sticker, error) {
	// Stickers are available through Guild function in this version
	guild, err := c.session.Guild(guildID)
	if err != nil {
		return nil, err
	}
	return guild.Stickers, nil
}

func (c *DiscordClient) GetGuildSticker(guildID, stickerID string) (*discordgo.Sticker, error) {
	stickers, err := c.GetGuildStickers(guildID)
	if err != nil {
		return nil, err
	}
	for _, sticker := range stickers {
		if sticker.ID == stickerID {
			return sticker, nil
		}
	}
	return nil, fmt.Errorf("sticker not found")
}

func (c *DiscordClient) CreateGuildSticker(guildID, name, description, tags, filePath string) (*discordgo.Sticker, error) {
	// Sticker creation not directly supported in this version of discordgo
	return nil, fmt.Errorf("sticker creation not supported in this version")
}

func (c *DiscordClient) EditGuildSticker(guildID, stickerID, name, description, tags string) (*discordgo.Sticker, error) {
	// Sticker editing not directly supported in this version of discordgo
	return nil, fmt.Errorf("sticker editing not supported in this version")
}

func (c *DiscordClient) DeleteGuildSticker(guildID, stickerID string) error {
	// Sticker deletion not directly supported in this version of discordgo
	return fmt.Errorf("sticker deletion not supported in this version")
}

func (c *DiscordClient) GetGuildEvents(guildID string) ([]*discordgo.GuildScheduledEvent, error) {
	return c.session.GuildScheduledEvents(guildID, false)
}

func (c *DiscordClient) GetGuildEvent(guildID, eventID string) (*discordgo.GuildScheduledEvent, error) {
	return c.session.GuildScheduledEvent(guildID, eventID, false)
}

func (c *DiscordClient) CreateGuildEvent(guildID string, params *discordgo.GuildScheduledEventParams) (*discordgo.GuildScheduledEvent, error) {
	return c.session.GuildScheduledEventCreate(guildID, params)
}

func (c *DiscordClient) EditGuildEvent(guildID, eventID string, params *discordgo.GuildScheduledEventParams) (*discordgo.GuildScheduledEvent, error) {
	return c.session.GuildScheduledEventEdit(guildID, eventID, params)
}

func (c *DiscordClient) DeleteGuildEvent(guildID, eventID string) error {
	return c.session.GuildScheduledEventDelete(guildID, eventID)
}

func (c *DiscordClient) GetGuildEventUsers(guildID, eventID string, limit int, before, after string) ([]*discordgo.GuildScheduledEventUser, error) {
	return c.session.GuildScheduledEventUsers(guildID, eventID, limit, false, before, after)
}

func (c *DiscordClient) GetVoiceRegions() ([]*discordgo.VoiceRegion, error) {
	return c.session.VoiceRegions()
}

func (c *DiscordClient) GetChannelInvites(channelID string) ([]*discordgo.Invite, error) {
	return c.session.ChannelInvites(channelID)
}

func (c *DiscordClient) GetAutoModRules(guildID string) ([]*discordgo.AutoModerationRule, error) {
	return c.session.AutoModerationRules(guildID)
}

func (c *DiscordClient) GetAutoModRule(guildID, ruleID string) (*discordgo.AutoModerationRule, error) {
	return c.session.AutoModerationRule(guildID, ruleID)
}

func (c *DiscordClient) CreateAutoModRule(guildID string, rule *discordgo.AutoModerationRule) (*discordgo.AutoModerationRule, error) {
	return c.session.AutoModerationRuleCreate(guildID, rule)
}

func (c *DiscordClient) EditAutoModRule(guildID, ruleID string, rule *discordgo.AutoModerationRule) (*discordgo.AutoModerationRule, error) {
	return c.session.AutoModerationRuleEdit(guildID, ruleID, rule)
}

func (c *DiscordClient) DeleteAutoModRule(guildID, ruleID string) error {
	return c.session.AutoModerationRuleDelete(guildID, ruleID)
}

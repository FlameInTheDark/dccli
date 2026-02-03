package voice

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Connection manages a Discord voice connection
type Connection struct {
	session    *discordgo.Session
	voice      *discordgo.VoiceConnection
	guildID    string
	channelID  string
	mu         sync.RWMutex
	connected  bool
	connecting chan struct{}
}

// NewConnection creates a new voice connection manager from an existing VoiceConnection
func NewConnection(voice *discordgo.VoiceConnection) *Connection {
	return &Connection{
		voice:      voice,
		guildID:    voice.GuildID,
		channelID:  voice.ChannelID,
		connecting: make(chan struct{}),
		connected:  voice.Ready,
	}
}

// Join connects to a voice channel
func (c *Connection) Join() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	voice, err := c.session.ChannelVoiceJoin(c.guildID, c.channelID, false, false)
	if err != nil {
		return fmt.Errorf("failed to join voice channel: %w", err)
	}

	c.voice = voice
	c.connected = true

	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("voice connection timeout")
		case <-ticker.C:
			if voice.Ready {
				return nil
			}
		}
	}
}

// Leave disconnects from the voice channel
func (c *Connection) Leave() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected || c.voice == nil {
		return nil
	}

	err := c.voice.Disconnect()
	if err != nil {
		return fmt.Errorf("failed to disconnect from voice channel: %w", err)
	}

	c.connected = false
	c.voice = nil
	return nil
}

// IsConnected returns whether the bot is connected to a voice channel
func (c *Connection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.voice != nil && c.voice.Ready
}

// GetVoiceConnection returns the underlying discordgo voice connection
func (c *Connection) GetVoiceConnection() *discordgo.VoiceConnection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.voice
}

// GetGuildID returns the guild ID
func (c *Connection) GetGuildID() string {
	return c.guildID
}

// GetChannelID returns the channel ID
func (c *Connection) GetChannelID() string {
	return c.channelID
}

// ConnectionManager manages multiple voice connections across guilds
type ConnectionManager struct {
	session     *discordgo.Session
	connections map[string]*Connection // guildID -> Connection
	mu          sync.RWMutex
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(session *discordgo.Session) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*Connection),
		session:     session,
	}
}

// Join connects to a voice channel in a guild
func (m *ConnectionManager) Join(guildID, channelID string) (*Connection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, exists := m.connections[guildID]; exists {
		if conn.IsConnected() && conn.GetChannelID() == channelID {
			return conn, nil
		}
		conn.Leave()
		delete(m.connections, guildID)
	}

	conn := NewConnection(&discordgo.VoiceConnection{
		GuildID:   guildID,
		ChannelID: channelID,
	})
	conn.session = m.session
	if err := conn.Join(); err != nil {
		return nil, err
	}

	m.connections[guildID] = conn
	return conn, nil
}

// Leave disconnects from a voice channel in a guild
func (m *ConnectionManager) Leave(guildID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exists := m.connections[guildID]
	if !exists {
		return fmt.Errorf("not connected to any voice channel in this guild")
	}

	err := conn.Leave()
	delete(m.connections, guildID)
	return err
}

// GetConnection returns the connection for a guild
func (m *ConnectionManager) GetConnection(guildID string) (*Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, exists := m.connections[guildID]
	return conn, exists
}

// IsConnected checks if connected to a voice channel in a guild
func (m *ConnectionManager) IsConnected(guildID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, exists := m.connections[guildID]
	if !exists {
		return false
	}
	return conn.IsConnected()
}

// Close disconnects from all voice channels
func (m *ConnectionManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, conn := range m.connections {
		conn.Leave()
	}
	m.connections = make(map[string]*Connection)
}

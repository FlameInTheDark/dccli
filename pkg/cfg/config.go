package cfg

import (
	"errors"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	filePath = "/.dccli"
	fileName = "/config.yaml"
)

var (
	ErrNotConfigured     = errors.New("bot not configured")
	ErrBotConfigNotFound = errors.New("bot config not found")
)

// BotConfig represents a bot configuration from a config file
type BotConfig struct {
	Name string `yaml:"name"`
	Bot  Bot    `yaml:"bot"`
}

// Bot contains bot configurations
type Bot struct {
	Token string `yaml:"token"`
}

// Config represents a CLI configuration file
type Config struct {
	CurrentBot string      `yaml:"current-bot"`
	Bots       []BotConfig `yaml:"bots"`
}

// GetBotByName returns bot by its config name
func (c *Config) GetBotByName(name string) (*BotConfig, error) {
	for i, bot := range c.Bots {
		if bot.Name == name {
			return &c.Bots[i], nil
		}
	}
	return nil, ErrBotConfigNotFound
}

// GetCurrent returns current configured bot.
// If current bot is not found, change current to first from the configured bot list
func (c *Config) GetCurrent() (*BotConfig, error) {
	if len(c.Bots) == 0 {
		return nil, ErrNotConfigured
	}

	bot, err := c.GetBotByName(c.CurrentBot)
	if err != nil {
		log.Printf("Bot %s not found in config. Changing to first from list", c.CurrentBot)
		c.CurrentBot = c.Bots[0].Name
		SaveConfig(c)
		return &c.Bots[0], nil
	}
	return bot, nil
}

func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Unable to get home directory")
	}
	if _, err := os.Stat(home + filePath); os.IsNotExist(err) {
		err := os.Mkdir(home+filePath, os.ModePerm)
		if err != nil {
			log.Fatal("Unable to create config directory: ", err)
		}
		return nil, ErrNotConfigured
	}
	if _, err := os.Stat(home + filePath + fileName); os.IsNotExist(err) {
		return nil, ErrNotConfigured
	}

	file, err := os.ReadFile(home + filePath + fileName)
	if err != nil {
		log.Fatal("Unable to open config file: ", err)
	}

	var config *Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Unable to parse config")
	}

	return config, nil
}

func SaveConfig(config *Config) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Unable to get home directory")
	}
	file, err := os.OpenFile(home+filePath+fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatal("Unable to open config file: ", err)
	}
	defer file.Close()

	out, err := yaml.Marshal(config)
	if err != nil {
		log.Fatal("Unable to parse config")
	}

	_, err = file.Write(out)
	if err != nil {
		log.Fatal("Unable to write config")
	}
}

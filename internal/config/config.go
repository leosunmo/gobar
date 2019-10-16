package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/flopp/go-findfont"
	"gopkg.in/ini.v1"
)

// Config for the bar and its blocks
var Config BarConfig

// BarConfig is both the Global config and the Blocklet configs
type BarConfig struct {
	Global       GlobalConfig
	BlockConfigs []BlockConfig
}

// GlobalConfig is a global section of the ini conf file that
// applies as defaults to all blocks or sets generic settings
type GlobalConfig struct {
	Command    string
	SepWidth   int
	Markup     string
	Color      string
	LabelColor string
	IconFont   string
	LabelFont  string
	ValueFont  string
}

// BlockConfig is the configuration for a single Blocklet
type BlockConfig struct {
	Command    string
	Markup     string
	Color      string
	LabelColor string
	IconFont   string
	LabelFont  string
	ValueFont  string
}

// ReadConfig reads the block config and returns BarConfig
func ReadConfig(configFile string) error {
	if configFile == "" {
		configFile = "i3xrocks.conf"
	}
	cfg, err := ini.Load(configFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	Config.readSections(cfg)

	return nil
}

func findFont(fontName string) (string, error) {
	fontPath, err := findfont.Find(fontName)
	if err != nil {
		return "", err
	}
	return fontPath, nil
}

func (bc *BarConfig) readSections(cfg *ini.File) error {
	var err error
	// Read the global section first
	gs := cfg.Section("DEFAULT")
	err = gs.MapTo(bc.Global)
	if err != nil {
		return fmt.Errorf("failed to map global config to struct, err %s", err.Error())
	}
	for _, s := range cfg.Sections() {
		if s.Name() != "DEFAULT" { // lazy but it works
			blockConfig := BlockConfig{}
			err = s.MapTo(blockConfig)
			if err != nil {
				return fmt.Errorf("failed to map block config to struct, err %s", err.Error())
			}
			// Check if the command is exectuable and if we need to use the global path
			if ok := isExecutable(blockConfig.Command); !ok {
				// command isn't excutable or doesn't exist, even after using global path prefix
				fmt.Printf("Failed execute %s", blockConfig.Command)
			}
			bc.BlockConfigs = append(bc.BlockConfigs, blockConfig)
		}
	}
	return nil
}

func findExecutable(command string) string {
	_, err := os.Stat(command)
	if os.IsNotExist(err) {
		// file doesn't exist, check global config path prefix
		if Config.Global.Command != "" {
			command = strings.Replace(Config.Global.Command, "$BLOCK_NAME", command, 1)
		}
		_, err = os.Stat(command)
		if os.IsNotExist(err) {
			return ""
		}
	}
	return command
}

func isExecutable(command string) bool {
	info, err := os.Stat(command)
	if !os.IsNotExist(err) {
		if info.Mode() == 0100755 { // It's executable by somebody
			return true
		}
	}
	return false
}

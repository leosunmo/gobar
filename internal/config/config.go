package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/flopp/go-findfont"
	"gopkg.in/ini.v1"
)

const (
	Setuid uint32 = 1 << (12 - 1 - iota)
	Setgid
	Sticky
	UserRead
	UserWrite
	UserExecute
	GroupRead
	GroupWrite
	GroupExecute
	OtherRead
	OtherWrite
	OtherExecute
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
	Command    string `ini:"command"`
	SepWidth   int    `ini:"separator_block_width"`
	Markup     string `ini:"markup"`
	Color      string `ini:"color"`
	LabelColor string `ini:"label_color"`
	IconFont   string `ini:"icon_font"`
	LabelFont  string `ini:"label_font"`
	ValueFont  string `ini:"value_font"`
	Spacer     string `ini:"spacer"`
}

// BlockConfig is the configuration for a single Blocklet
type BlockConfig struct {
	Command    string `ini:"command"`
	Interval   string `ini:"interval"`
	Markup     string `ini:"markup"`
	Color      string `ini:"color"`
	LabelColor string `ini:"label_color"`
	IconFont   string `ini:"icon_font"`
	LabelFont  string `ini:"label_font"`
	ValueFont  string `ini:"value_font"`
	blockName  string
	blockCmd   string
	envVars    map[string]string
}

// ReadConfig reads the block config and returns BarConfig
func ReadConfig(configFile string) error {
	if configFile == "" {
		configFile = "i3xrocks.conf"
	}
	cfg, err := ini.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}
	err = Config.readSections(cfg)
	if err != nil {
		return err
	}
	return nil
}

func (bc *BarConfig) readSections(cfg *ini.File) error {
	var err error
	// Set defaults
	bc.SetDefaults()

	// Read the global section first
	gs := cfg.Section(ini.DEFAULT_SECTION)
	if len(gs.Keys()) > 0 {
		err = gs.MapTo(&bc.Global)
		if err != nil {
			return fmt.Errorf("failed to map global config to struct, err %s", err.Error())
		}
	}
	for _, s := range cfg.Sections() {
		if s.Name() != ini.DEFAULT_SECTION {
			blockConfig := BlockConfig{}
			err = s.MapTo(&blockConfig)
			if err != nil {
				return fmt.Errorf("failed to map block config to struct, err %s", err.Error())
			}
			// Grab the values and send them to the shell as env vars
			blockConfig.envVars = s.KeysHash()
			// Save the name of the section to the block
			blockConfig.blockName = s.Name()
			// If the 'command' section is empty, use the title as the command
			if blockConfig.Command == "" {
				blockConfig.blockCmd = blockConfig.blockName
			} else {
				blockConfig.blockCmd = blockConfig.Command
			}
			// Find the blocks executable by using global command
			blockConfig.blockCmd = findExecutable(blockConfig.blockCmd)
			// Check if the command is exectuable and if we need to use the global path
			if ok := isExecutable(blockConfig.blockCmd); !ok {
				// command isn't excutable or doesn't exist, even after using global path prefix
				fmt.Printf("%s doesn't exist or isn't executable\n", blockConfig.blockCmd)
			}

			// Append to our list of Blocklets
			bc.BlockConfigs = append(bc.BlockConfigs, blockConfig)
		}
	}
	return nil
}

func (bc *BarConfig) SetDefaults() {
	bc.Global.Spacer = " " // Figure out how to set Pango as a spacer, like <span size='xx-small'> </span>

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
		if uint32(info.Mode().Perm())&UserExecute != 0 {
			return true // It's executable by somebody
		}
	}
	return false
}

func findFont(fontName string) (string, error) {
	fontPath, err := findfont.Find(fontName)
	if err != nil {
		return "", err
	}
	return fontPath, nil
}

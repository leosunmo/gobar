package blocks

import (
	"fmt"
	"time"

	"barista.run/bar"
	"barista.run/modules/shell"
	"barista.run/outputs"
	"github.com/leosunmo/gobar/internal/config"
)

type blocks map[string]*bar.Module

// PrintBlockConfig prints all the blocklet configs
func PrintBlockConfig() {
	cfg := config.Config
	for _, b := range cfg.BlockConfigs {
		fmt.Printf("Block [%s]\n", b.Command)
		block := shell.New("/bin/sh", "-c", s.Name()).
			Every(time.Second).
			Output(func(count string) bar.Output {
				return outputs.Textf("%s chrome procs", count)
			})
		blocks = append(blocks, block)

	}
	for _, k := range s.Keys() {
		fmt.Printf("\t%s = %s\n", k.Name(), k.Value())
	}
}

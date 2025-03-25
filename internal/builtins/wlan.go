package builtins

import (
	"github.com/leosunmo/barista/bar"
	"github.com/leosunmo/barista/colors"
	"github.com/leosunmo/barista/modules/wlan"
	"github.com/leosunmo/barista/outputs"
	"github.com/leosunmo/barista/pango"
	"github.com/leosunmo/gobar/internal/utils"
)

// NewWlan returns a Wlan module that selects the active wifi with default icon
func NewWlan() (bar.Module, error) {
	icon := pango.Icon("symbol-wifi").Color(colors.Scheme("dim-icon"))

	return NewWlanWithIcon(icon)
}

// NewWlanWithIcon returns a Wlan module that selects the active wifi with the provided Pango Icon
func NewWlanWithIcon(icon *pango.Node) (bar.Module, error) {
	w := wlan.Any()

	mod := w.Output(func(info wlan.Info) bar.Output {
		s := pango.Textf(" %s", utils.Truncate(info.SSID, 20))

		return outputs.Pango(icon, s)
	})

	return mod, nil
}

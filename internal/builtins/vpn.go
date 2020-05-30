package builtins

import (
	"barista.run/bar"
	"barista.run/colors"
	"barista.run/modules/vpn"
	"barista.run/outputs"
	"barista.run/pango"
	"github.com/leosunmo/gobar/internal/utils"
)

// NewVPN returns a Wlan module that selects the active wifi with default icon
func NewVPN(iface string) (bar.Module, error) {
	icon := pango.Icon("material-security").Color(colors.Scheme("dim-icon"))

	return NewVPNWithIcon(icon, iface)
}

// NewVPNWithIcon returns a Wlan module that selects the active wifi with the provided Pango Icon
func NewVPNWithIcon(icon *pango.Node, iface string) (bar.Module, error) {
	v := &vpn.Module{}
	if iface == "" {
		v = vpn.DefaultInterface()
	} else {
		v = vpn.New(iface)
	}
	mod := v.Output(func(state vpn.State) bar.Output {
		s := &pango.Node{}
		switch state {
		case vpn.Connected:
			s = pango.Textf(" %s", utils.Truncate("VPN", 10))
		case vpn.Waiting:
			s = pango.Textf(" %s", utils.Truncate("...", 10))
		case vpn.Disconnected:
			s = pango.Textf(" %s", utils.Truncate("OFF", 10))
		default:
			return nil
		}

		return outputs.Pango(icon, s)
	})

	return mod, nil
}

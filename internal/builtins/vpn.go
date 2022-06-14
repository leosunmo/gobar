package builtins

import (
	"time"

	"barista.run/bar"
	"barista.run/colors"
	"barista.run/modules/shell"
	"barista.run/outputs"
	"barista.run/pango"
	"github.com/leosunmo/gobar/internal/utils"
)

var VPNCommand = `C_ALL=C nmcli -t connection show --active | awk -F ':' '
{ if($3 == "vpn") {
    vpn_name=$1
  } else if ($3 == "tun"){
    tun_name=$1
  } else if ($3 == "tap"){
    tun_name=$1
  }
}
END{if (vpn_name) {printf("%s", vpn_name)}  else if(tun_name) {printf("%s", tun_name)}}'`

// NewVPN returns a Shell module that displays the current VPN status
func NewVPN() (bar.Module, error) {
	icon := pango.Icon("fa-lock").Color(colors.Scheme("dim-icon"))
	return NewVPNWithIcon(icon)
}

// NewVPNWithIcon returns a Shell module that displays the current VPN status with the provided Pango Icon
func NewVPNWithIcon(icon *pango.Node) (bar.Module, error) {
	s := shell.New("sh", "-c", VPNCommand).Every(time.Second * 1)

	mod := s.Output(func(vpnName string) bar.Output {
		if vpnName != "" {
			s := pango.Textf(" %s", utils.Truncate(vpnName, 20))

			return outputs.Pango(icon, s)
		}
		return nil
	})

	return mod, nil
}

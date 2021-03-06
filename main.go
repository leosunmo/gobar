package main

import (
	"fmt"
	"os/user"
	"path/filepath"
	"time"

	"barista.run"
	"barista.run/bar"
	"barista.run/base/click"
	"barista.run/base/watchers/netlink"
	"barista.run/colors"
	"barista.run/format"
	"barista.run/group/collapsing"
	"barista.run/modules/battery"
	"barista.run/modules/clock"
	"barista.run/modules/cputemp"
	"barista.run/modules/meminfo"
	"barista.run/modules/netspeed"
	"barista.run/modules/sysinfo"
	"barista.run/outputs"
	"barista.run/pango"
	"barista.run/pango/icons/fontawesome"
	"barista.run/pango/icons/material"
	"barista.run/pango/icons/mdi"
	"barista.run/pango/icons/typicons"

	"github.com/leosunmo/gobar/internal/builtins"
	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/martinlindhe/unit"
)

var spacer = pango.Text(" ").XXSmall()

func truncate(in string, l int) string {
	if len([]rune(in)) <= l {
		return in
	}
	return string([]rune(in)[:l-1]) + "⋯"
}

var startTaskManager = click.RunLeft("gnome-system-monitor")

func home(path string) string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(usr.HomeDir, path)
}

// func calendarNotifyHandler(e calendar.Event) func(bar.Event) {
// 	notifyBody := e.Start.Format("15:04")
// 	if !e.End.Equal(e.Start) {
// 		notifyBody += " - " + e.End.Format("15:04")
// 	}
// 	if e.Location != "" {
// 		notifyBody += "\n" + e.Location
// 	}
// 	return click.RunLeft("notify-send", e.Summary, notifyBody)
// }

// func setupOauthEncryption() error {
// 	const service = "barista-sample-bar"
// 	var username string
// 	if u, err := user.Current(); err == nil {
// 		username = u.Username
// 	} else {
// 		username = fmt.Sprintf("user-%d", os.Getuid())
// 	}
// 	var secretBytes []byte
// 	// IMPORTANT: The oauth tokens used by some modules are very sensitive, so
// 	// we encrypt them with a random key and store that random key using
// 	// libsecret (gnome-keyring or equivalent). If no secret provider is
// 	// available, there is no way to store tokens (since the version of
// 	// sample-bar used for setup-oauth will have a different key from the one
// 	// running in i3bar). See also https://github.com/zalando/go-keyring#linux.
// 	secret, err := keyring.Get(service, username)
// 	if err == nil {
// 		secretBytes, err = base64.RawURLEncoding.DecodeString(secret)
// 	}
// 	if err != nil {
// 		secretBytes = make([]byte, 64)
// 		_, err := rand.Read(secretBytes)
// 		if err != nil {
// 			return err
// 		}
// 		secret = base64.RawURLEncoding.EncodeToString(secretBytes)
// 		keyring.Set(service, username, secret)
// 	}
// 	oauth.SetEncryptionKey(secretBytes)
// 	return nil
// }

func main() {
	material.Load(home("/.config/regolith/styles/fonts/material-design-icons"))
	mdi.Load(home("/.config/regolith/styles/fonts/MaterialDesign-Webfont"))
	typicons.Load(home("/.config/regolith/styles/fonts/typicons.font"))
	fontawesome.Load(home("/.config/regolith/styles/fonts/Font-Awesome"))

	colors.LoadBarConfig()
	bg := colors.Scheme("background")
	fg := colors.Scheme("statusline")
	if fg != nil && bg != nil {
		iconColor := fg.Colorful().BlendHcl(bg.Colorful(), 0.5).Clamped()
		colors.Set("dim-icon", iconColor)
		_, _, v := fg.Colorful().Hsv()
		if v < 0.3 {
			v = 0.3
		}
		colors.Set("bad", colorful.Hcl(40, 1.0, v).Clamped())
		colors.Set("degraded", colorful.Hcl(90, 1.0, v).Clamped())
		colors.Set("good", colorful.Hcl(120, 1.0, v).Clamped())
	}

	// if err := setupOauthEncryption(); err != nil {
	// 	panic(fmt.Sprintf("Could not setup oauth token encryption: %v", err))
	// }

	localtime := clock.Local().
		Output(time.Second, func(now time.Time) bar.Output {
			return outputs.Pango(
				pango.Icon("material-today").Color(colors.Scheme("dim-icon")),
				spacer,
				now.Format("Mon Jan 2 "),
				pango.Icon("material-access-time").Color(colors.Scheme("dim-icon")),
				spacer,
				now.Format("15:04:05"),
				spacer,
			).OnClick(click.RunLeft("gsimplecal"))
		})

	buildBattOutput := func(i battery.Info, disp *pango.Node) *bar.Segment {
		if i.Status == battery.Disconnected || i.Status == battery.Unknown {
			return nil
		}
		iconName := "battery"
		if i.Status == battery.Charging {
			iconName += "-charging"
		}
		tenth := i.RemainingPct() / 10
		switch {
		case tenth == 0:
			iconName += "-outline"
		case tenth < 10:
			iconName += fmt.Sprintf("-%d0", tenth)
		}
		out := outputs.Pango(pango.Icon("mdi-"+iconName).Color(colors.Scheme("dim-icon")), disp)
		switch {
		case i.RemainingPct() <= 5:
			out.Urgent(true)
		case i.RemainingPct() <= 15:
			out.Color(colors.Scheme("bad"))
		case i.RemainingPct() <= 25:
			out.Color(colors.Scheme("degraded"))
		}
		return out
	}
	var showBattPct, showBattTime func(battery.Info) bar.Output

	batt := battery.All()
	showBattPct = func(i battery.Info) bar.Output {
		out := buildBattOutput(i, pango.Textf("%d%%", i.RemainingPct()))
		if out == nil {
			return nil
		}
		return out.OnClick(click.Left(func() {
			batt.Output(showBattTime)
		}))
	}
	showBattTime = func(i battery.Info) bar.Output {
		rem := i.RemainingTime()
		out := buildBattOutput(i, pango.Textf(
			"%d:%02d", int(rem.Hours()), int(rem.Minutes())%60))
		if out == nil {
			return nil
		}
		return out.OnClick(click.Left(func() {
			batt.Output(showBattPct)
		}))
	}
	batt.Output(showBattPct)

	// vol := volume.DefaultMixer().Output(func(v volume.Volume) bar.Output {
	// 	if v.Mute {
	// 		return outputs.
	// 			Pango(pango.Icon("fa-volume-mute"), spacer, "MUT").
	// 			Color(colors.Scheme("degraded"))
	// 	}
	// 	iconName := "off"
	// 	pct := v.Pct()
	// 	if pct > 66 {
	// 		iconName = "up"
	// 	} else if pct > 33 {
	// 		iconName = "down"
	// 	}
	// 	return outputs.Pango(
	// 		pango.Icon("fa-volume-"+iconName),
	// 		spacer,
	// 		pango.Textf("%2d%%", pct),
	// 	)
	// })

	loadAvg := sysinfo.New().Output(func(s sysinfo.Info) bar.Output {
		out := outputs.Textf("%0.2f %0.2f", s.Loads[0], s.Loads[2])
		// Load averages are unusually high for a few minutes after boot.
		if s.Uptime < 10*time.Minute {
			// so don't add colours until 10 minutes after system start.
			return out
		}
		switch {
		case s.Loads[0] > 128, s.Loads[2] > 64:
			out.Urgent(true)
		case s.Loads[0] > 64, s.Loads[2] > 32:
			out.Color(colors.Scheme("bad"))
		case s.Loads[0] > 32, s.Loads[2] > 16:
			out.Color(colors.Scheme("degraded"))
		}
		out.OnClick(startTaskManager)
		return out
	})

	freeMem := meminfo.New().Output(func(m meminfo.Info) bar.Output {
		out := outputs.Pango(pango.Icon("material-memory"), format.IBytesize(m.Available()))
		freeGigs := m.Available().Gigabytes()
		switch {
		case freeGigs < 0.5:
			out.Urgent(true)
		case freeGigs < 1:
			out.Color(colors.Scheme("bad"))
		case freeGigs < 2:
			out.Color(colors.Scheme("degraded"))
		case freeGigs > 12:
			out.Color(colors.Scheme("good"))
		}
		out.OnClick(startTaskManager)
		return out
	})

	temp := cputemp.New().
		RefreshInterval(2 * time.Second).
		Output(func(temp unit.Temperature) bar.Output {
			out := outputs.Pango(
				pango.Icon("mdi-fan"), spacer,
				pango.Textf("%2d℃", int(temp.Celsius())),
			)
			switch {
			case temp.Celsius() > 90:
				out.Urgent(true)
			case temp.Celsius() > 70:
				out.Color(colors.Scheme("bad"))
			case temp.Celsius() > 60:
				out.Color(colors.Scheme("degraded"))
			}
			return out
		})

	sub := netlink.Any()
	iface := sub.Get().Name
	sub.Unsubscribe()
	net := netspeed.New(iface).
		RefreshInterval(2 * time.Second).
		Output(func(s netspeed.Speeds) bar.Output {
			return outputs.Pango(
				pango.Icon("fa-upload"), spacer, pango.Textf("%7s", format.Byterate(s.Tx)),
				pango.Text(" ").Small(),
				pango.Icon("fa-download"), spacer, pango.Textf("%7s", format.Byterate(s.Rx)),
			)
		})

	mediaPlayer := builtins.NewMediaPlayer("spotify")

	grp, c := collapsing.Group(net, temp, freeMem, loadAvg)
	c.ButtonFunc(func(c collapsing.Controller) (start, end bar.Output) {
		if c.Expanded() {
			return outputs.Text(">").OnClick(click.Left(c.Collapse)),
				outputs.Text("<").OnClick(click.Left(c.Collapse))
		}
		icon := pango.Icon("fa-cogs").Color(colors.Scheme("dim-icon")).Small()

		text := pango.Textf("%s%%", "25")

		return outputs.Pango(icon, spacer, text).OnClick(click.Left(c.Expand)), nil
	})
	// vpn := vpn.DefaultInterface().Output(func(s vpn.State) bar.Output {
	// 	if s.Connected() {
	// 		return outputs.Pango(pango.Icon("fa-shield-alt")).Color(colors.Scheme("dim-icon"))
	// 	}
	// 	if s.Disconnected() {
	// 		return nil
	// 	}
	// 	return outputs.Text("...")
	// })

	vpn, _ := builtins.NewVPN()

	wlan, _ := builtins.NewWlan()

	//pango.Icon("fa-shield-alt")

	// ghNotify := github.New("%%GITHUB_CLIENT_ID%%", "%%GITHUB_CLIENT_SECRET%%").
	// 	Output(func(n github.Notifications) bar.Output {
	// 		if n.Total() == 0 {
	// 			return nil
	// 		}
	// 		out := outputs.Group(
	// 			pango.Icon("fab-github").
	// 				Concat(spacer).
	// 				ConcatTextf("%d", n.Total()))
	// 		mentions := n["mention"] + n["team_mention"]
	// 		if mentions > 0 {
	// 			out.Append(spacer)
	// 			out.Append(outputs.Pango(
	// 				pango.Icon("mdi-bell").
	// 					ConcatTextf("%d", mentions)).
	// 				Urgent(true))
	// 		}
	// 		return out.Glue().OnClick(
	// 			click.RunLeft("xdg-open", "https://github.com/notifications"))
	// 	})

	panic(barista.Run(
		mediaPlayer,
		vpn,
		wlan,
		grp,
		//ghNotify,
		batt,
		localtime,
	))
}

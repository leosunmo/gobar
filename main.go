package main

import (
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/leosunmo/barista"
	"github.com/leosunmo/barista/bar"
	"github.com/leosunmo/barista/base/click"
	"github.com/leosunmo/barista/base/watchers/netlink"
	"github.com/leosunmo/barista/colors"
	"github.com/leosunmo/barista/format"
	"github.com/leosunmo/barista/group/collapsing"
	"github.com/leosunmo/barista/logging"
	"github.com/leosunmo/barista/modules/battery"
	"github.com/leosunmo/barista/modules/clock"
	"github.com/leosunmo/barista/modules/cputemp"
	"github.com/leosunmo/barista/modules/meminfo"
	"github.com/leosunmo/barista/modules/netspeed"
	"github.com/leosunmo/barista/modules/sysinfo"
	"github.com/leosunmo/barista/modules/volume"
	"github.com/leosunmo/barista/modules/volume/pulseaudio"
	"github.com/leosunmo/barista/outputs"
	"github.com/leosunmo/barista/pango"
	"github.com/leosunmo/barista/pango/icons/symbols"

	"github.com/leosunmo/gobar/internal/builtins"
	"github.com/leosunmo/gobar/internal/cpu"
	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/martinlindhe/unit"
)

var spacer = pango.Text("  ").XXSmall()

var startTaskManager = click.RunLeft("gnome-system-monitor")

func home(path string) string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(usr.HomeDir, path)
}

func main() {
	logging.SetOutput(os.Stderr)
	err := symbols.LoadFile(home("/.config/gobar/fonts/MaterialSymbolsOutlined.codepoints"))
	if err != nil {
		logging.Log("failed to load font: %s", err.Error())
	}

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
				pango.Icon("symbol-event").Color(colors.Scheme("dim-icon")),
				spacer,
				now.Format("Mon Jan 2 "),
				pango.Icon("symbol-schedule").Color(colors.Scheme("dim-icon")),
				spacer,
				now.Format("15:04:05"),
				spacer,
			).OnClick(click.RunLeft("gsimplecal"))
		})

	vol := volume.New(pulseaudio.DefaultSink()).Output(func(v volume.Volume) bar.Output {
		iconName := "volume"
		switch {
		case v.Mute:
			iconName += "-off"
		case v.Pct() == 0:
			iconName += "-off"
		case v.Pct() <= 40:
			iconName += "-down"
		case v.Pct() > 40:
			iconName += "-up"
		}
		logging.Log("volume icon: %s", iconName)
		return outputs.Pango(
			pango.Icon("symbol-"+iconName).Color(colors.Scheme("dim-icon")),
			spacer,
			pango.Textf("%d%%", v.Pct()),
		).OnClick(
			func(e bar.Event) {
				switch e.Button {
				case bar.ButtonLeft:
					v.SetMuted(!v.Mute)
				case bar.ScrollDown:
					v.SetVolume(v.Vol - int64(float64(v.Max-v.Min)*float64(5)/100))
				case bar.ScrollUp:
					v.SetVolume(v.Vol + v.Min + int64(float64(v.Max-v.Min)*float64(5)/100))
				}
			},
		)
	})

	buildBattOutput := func(i battery.Info, disp *pango.Node) *bar.Segment {
		if i.Status == battery.Disconnected || i.Status == battery.Unknown {
			return nil
		}
		iconName := "battery"
		remainingPct := i.RemainingPct()
		if i.Status == battery.Charging {
			switch {
			case remainingPct > 80:
				iconName += "-charging-90"
			case remainingPct > 60:
				iconName += "-charging-80"
			case remainingPct > 50:
				iconName += "-charging-60"
			case remainingPct > 30:
				iconName += "-charging-50"
			case remainingPct > 20:
				iconName += "-charging-30"
			case remainingPct > 10:
				iconName += "-charging-20"
			default:
				iconName += "-charging-20"
			}
		} else {
			switch {
			case remainingPct > 95:
				iconName += "-full"
			case remainingPct > 80:
				iconName += "-6-bar"
			case remainingPct > 50:
				iconName += "-5-bar"
			case remainingPct > 40:
				iconName += "-4-bar"
			case remainingPct > 30:
				iconName += "-3-bar"
			case remainingPct > 20:
				iconName += "-2-bar"
			case remainingPct > 10:
				iconName += "-1-bar"
			case remainingPct > 5:
				iconName += "-0-bar"
			default:
				iconName += "-alert"
			}
		}
		out := outputs.Pango(pango.Icon("symbol-"+iconName).Color(colors.Scheme("dim-icon")), disp)
		switch {
		case remainingPct <= 5:
			out.Urgent(true)
		case remainingPct <= 15:
			out.Color(colors.Scheme("bad"))
		case remainingPct <= 25:
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
		out := outputs.Pango(pango.Icon("symbol-memory"), format.IBytesize(m.Available()))
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
				pango.Icon("symbol-mode-fan"), spacer,
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
				pango.Icon("symbol-keyboard-arrow-down"), spacer, pango.Textf("%7s", format.Byterate(s.Tx)),
				pango.Text(" ").Small(),
				pango.Icon("symbol-keyboard-arrow-up"), spacer, pango.Textf("%7s", format.Byterate(s.Rx)),
			)
		})

	grp, c := collapsing.Group(net, temp, freeMem, loadAvg)
	c.ButtonFunc(func(c collapsing.Controller) (start, end bar.Output) {
		if c.Expanded() {
			return outputs.Text(">").OnClick(click.Left(c.Collapse)),
				outputs.Text("<").OnClick(click.Left(c.Collapse))
		}
		return start, nil
	})

	cp := cpu.New(1 * time.Second).Output(func(stat cpu.CPUStat) bar.Output {
		icon := pango.Icon("symbol-settings").Color(colors.Scheme("dim-icon")).Small()
		return outputs.Pango(icon, spacer, pango.Textf("%4.1f%%", 100-stat.Idle)).OnClick(click.Left(c.Expand))
	}).Every(2 * time.Second)

	mediaPlayer := builtins.NewMediaPlayer("")

	vpn, _ := builtins.NewVPN()

	wlan, _ := builtins.NewWlan()

	panic(barista.Run(
		mediaPlayer,
		vpn,
		wlan,
		cp,
		grp,
		batt,
		vol,
		localtime,
	))
}

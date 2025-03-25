package builtins

import (
	"fmt"
	"time"

	"github.com/leosunmo/barista/bar"
	"github.com/leosunmo/barista/colors"
	"github.com/leosunmo/barista/logging"
	"github.com/leosunmo/barista/modules/media"
	"github.com/leosunmo/barista/outputs"
	"github.com/leosunmo/barista/pango"
	"github.com/leosunmo/gobar/internal/utils"
	"golang.org/x/time/rate"
)

var spacer = pango.Text(" ").XXSmall()

var excludedPlayers = []string{"chromium.*", "playerctld", "firefox", "opera", "vivaldi"}

// var excludedPlayers []string

func NewMediaPlayer(player string) bar.Module {
	playIcon := pango.Icon("symbol-play-arrow").Color(colors.Scheme("dim-icon"))
	pauseIcon := pango.Icon("symbol-pause").Color(colors.Scheme("dim-icon"))
	return NewMediaPlayerWithIcons(player, playIcon, pauseIcon)
}

var seekLimiter = rate.NewLimiter(rate.Every(50*time.Millisecond), 1)

func NewMediaPlayerWithIcons(player string, playIcon, pauseIcon *pango.Node) bar.Module {
	var icon *pango.Node
	mediaFormatter := func(m media.Info) bar.Output {
		if m.PlaybackStatus == media.Disconnected {
			logging.Log("Media player disconnected")
			return nil
		}
		artist := utils.Truncate(m.Artist, 20)
		title := utils.Truncate(m.Title, 40-len([]rune(artist)))
		if len(title) < 20 {
			artist = utils.Truncate(m.Artist, 40-len(title))
		}

		artistSong := pango.Textf("%s - %s", artist, title).Small()

		icon = makeMediaIcon(m, playIcon, pauseIcon)

		return outputs.Pango(icon, spacer, artistSong).OnClick(
			func(e bar.Event) {
				switch e.Button {
				case bar.ButtonLeft:
					m.PlayPause()
				case bar.ButtonRight:
					m.Next()
				case bar.ButtonMiddle:
					m.Previous()
				case bar.ScrollUp, bar.ScrollRight:
					if seekLimiter.Allow() {
						m.Seek(time.Second)
					}
				case bar.ScrollDown, bar.ScrollLeft:
					if seekLimiter.Allow() {
						m.Seek(-time.Second)
					}
				}
			},
		)
	}

	if player != "" {
		mod := media.New(player).RepeatingOutput(mediaFormatter)
		return mod
	}
	mod := media.Auto(excludedPlayers...).RepeatingOutput(mediaFormatter)
	return mod
}

func formatMediaTime(d time.Duration) string {
	h, m, s := utils.Hms(d)
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func makeMediaIcon(m media.Info, playIcon, pauseIcon *pango.Node) *pango.Node {
	var icon *pango.Node

	if m.PlaybackStatus == media.Playing {
		icon = playIcon
	} else {
		icon = pauseIcon
	}
	if m.PlaybackStatus == media.Playing {
		icon = pango.New(playIcon)
	} else {
		icon = pango.New(pauseIcon)
	}
	icon.Append(spacer, pango.Textf("%s/%s",
		formatMediaTime(m.Position()),
		formatMediaTime(m.Length)).Small())
	return icon.Color(colors.Scheme("dim-icon")).Small()
}

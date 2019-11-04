package builtins

import (
	"fmt"
	"time"

	"barista.run/bar"
	"barista.run/colors"
	"barista.run/modules/media"
	"barista.run/outputs"
	"barista.run/pango"
	"github.com/leosunmo/gobar/internal/utils"
	"golang.org/x/time/rate"
)

func NewMediaPlayer(player string) bar.Module {
	playIcon := pango.Icon("material-play-arrow").Color(colors.Scheme("dim-icon"))
	pauseIcon := pango.Icon("material-pause").Color(colors.Scheme("dim-icon"))
	return NewMediaPlayerWithIcons(player, playIcon, pauseIcon)
}

func NewMediaPlayerWithIcons(player string, playIcon, pauseIcon *pango.Node) bar.Module {

	var spacer = pango.Text(" ").XXSmall()
	var icon *pango.Node
	mediaFormatter := func(m media.Info) bar.Output {
		if m.PlaybackStatus == media.Stopped || m.PlaybackStatus == media.Disconnected {
			return nil
		}
		artist := utils.Truncate(m.Artist, 20)
		title := utils.Truncate(m.Title, 40-len([]rune(artist)))
		if len(title) < 20 {
			artist = utils.Truncate(m.Artist, 40-len(title))
		}

		artistSong := pango.Textf("%s - %s", artist, title).Small()

		// Custom behaviour for Spotify
		if m.PlayerName == "spotify" {
			if m.PlaybackStatus == media.Playing {
				icon = playIcon
			} else {
				icon = pauseIcon
			}
			return outputs.Pango(icon, spacer, artistSong).OnClick(
				func(e bar.Event) {
					if m.PlayerName == "spotify" {
						switch e.Button {
						case bar.ButtonLeft:
							m.PlayPause()
						case bar.ButtonRight:
							m.Next()
						case bar.ButtonMiddle:
							m.Previous()
						}
					} else {
						switch e.Button {
						case bar.ButtonLeft:
							m.PlayPause()
						case bar.ScrollDown, bar.ScrollRight:
							if rate.NewLimiter(rate.Every(50*time.Millisecond), 1).Allow() {
								m.Seek(time.Second)
							}
						case bar.ButtonBack:
							m.Previous()
						case bar.ScrollUp, bar.ScrollLeft:
							if rate.NewLimiter(rate.Every(50*time.Millisecond), 1).Allow() {
								m.Seek(-time.Second)
							}
						case bar.ButtonForward:
							m.Next()
						}
					}
				},
			)
		}
		if m.PlaybackStatus == media.Playing {
			icon.Append(
				spacer, pango.Textf("%s/%s",
					formatMediaTime(m.Position()),
					formatMediaTime(m.Length)),
			)
		}
		return outputs.Pango(icon, spacer, artistSong)
	}

	if player != "" {
		mod := media.New(player).Output(mediaFormatter)
		return mod
	}
	mod := media.Auto().Output(mediaFormatter)
	return mod

}

func formatMediaTime(d time.Duration) string {
	h, m, s := utils.Hms(d)
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

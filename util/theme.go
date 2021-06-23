package util

import (
	"image/color"

	"golang.org/x/image/font"
)

type Theme struct {
	Face font.Face

	Title            color.Color
	Background       color.Color
	Empty            color.Color
	Text             color.Color
	Highlight        color.Color
	HighlightSlider  color.Color
	BackgroundSlider color.Color
	NextHighlight    color.Color
	Rail             color.Color
	Infobar          color.Color
	ButtonUp         color.Color
	ButtonOver       color.Color
	ButtonDown       color.Color
	VolumeBg         color.Color
	VolumeBgOver     color.Color
}

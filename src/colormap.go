package main

import (
	"io"
	"sort"

	"github.com/fatih/color"
)

var colorMap = map[string]color.Attribute{
	"FGBLACK":      color.FgBlack,
	"FGRED":        color.FgRed,
	"FGGREEN":      color.FgGreen,
	"FGYELLOW":     color.FgYellow,
	"FGBLUE":       color.FgBlue,
	"FGMAGENTA":    color.FgMagenta,
	"FGCYAN":       color.FgCyan,
	"FGWHITE":      color.FgWhite,
	"FGHIBLACK":    color.FgHiBlack,
	"FGHIRED":      color.FgHiRed,
	"FGHIGREEN":    color.FgHiGreen,
	"FGHIYELLOW":   color.FgHiYellow,
	"FGHIBLUE":     color.FgHiBlue,
	"FGHIMAGENTA":  color.FgHiMagenta,
	"FGHICYAN":     color.FgHiCyan,
	"FGHIWHITE":    color.FgHiWhite,
	"BGBLACK":      color.BgBlack,
	"BGRED":        color.BgRed,
	"BGGREEN":      color.BgGreen,
	"BGYELLOW":     color.BgYellow,
	"BGBLUE":       color.BgBlue,
	"BGMAGENTA":    color.BgMagenta,
	"BGCYAN":       color.BgCyan,
	"BGWHITE":      color.BgWhite,
	"BGHIBLACK":    color.BgHiBlack,
	"BGHIRED":      color.BgHiRed,
	"BGHIGREEN":    color.BgHiGreen,
	"BGHIYELLOW":   color.BgHiYellow,
	"BGHIBLUE":     color.BgHiBlue,
	"BGHIMAGENTA":  color.BgHiMagenta,
	"BGHICYAN":     color.BgHiCyan,
	"BGHIWHITE":    color.BgHiWhite,
	"REVERSEVIDEO": color.ReverseVideo,
}

func displayColorList(w io.Writer) {
	keys := make([]string, 0)
	for k := range colorMap {
		keys = append(keys, k)
	}
	sizeKeys := len(keys)
	sort.Strings(keys)
	flexPrintln(w, "Supported color attributes:")
	for i, attrib := range keys {
		flexPrintf(w, "%s", attrib)
		if i < sizeKeys-1 {
			flexPrintf(w, "%s", ", ")
		}
	}
	flexPrintln(w, "")
}

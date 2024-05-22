package log

import (
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"
)

const (
	black = iota + 30
	red
	green
	yellow
	blue
	magenta
	cyan
	white
)

const (
	brightBlack = iota + 90
	brightRed
	brightGreen
	brightYellow
	brightBlue
	brightMagenta
	brightCyan
	brightWhite
)

const (
	reset              = 0
	increasedIntensity = 1
	decreasedIntensity = 2
	normalIntensity    = 22
)

func NewColorWriter(w io.Writer) zerolog.ConsoleWriter {
	o := zerolog.ConsoleWriter{
		Out:        w,
		TimeFormat: time.Kitchen,
	}

	o.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}

	o.FormatFieldName = func(i interface{}) string {
		return colorize(fmt.Sprintf("%s=", i), decreasedIntensity, normalIntensity, o.NoColor)
	}

	o.FormatErrFieldName = func(i interface{}) string {
		return colorAndIntensify(fmt.Sprintf("%s=", i), brightRed, decreasedIntensity, reset, o.NoColor)
	}

	o.FormatErrFieldValue = func(i interface{}) string {
		return colorize(fmt.Sprintf("%s", i), brightRed, reset, o.NoColor)
	}

	return o
}

func colorize(i interface{}, c, r int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", i)
	}

	return fmt.Sprintf("\033[%dm%v\033[%dm", c, i, r)
}

func colorAndIntensify(i interface{}, c, t, r int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", i)
	}

	return fmt.Sprintf("\033[%d;%dm%v\033[%dm", c, t, i, r)
}

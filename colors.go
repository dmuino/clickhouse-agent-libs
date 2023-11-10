package chagent

import "fmt"

// ansi colors
const (
	ColorsReset     = "\x1b[0m"
	ColorsFgRed     = "\x1b[31m"
	ColorsFgGreen   = "\x1b[32m"
	ColorsFgYellow  = "\x1b[33m"
	ColorsFgBlue    = "\x1b[34m"
	ColorsFgMagenta = "\x1b[35m"
	ColorsFgCyan    = "\x1b[36m"
	ColorsFgWhite   = "\x1b[37m"

	ColorsCheck = "\u2713"
	ColorsCross = "\u2717"
)

type Colorizer struct {
	enabled bool
}

func NewColorizer(enable bool) *Colorizer {
	return &Colorizer{enable}
}

func (c *Colorizer) Colorize(msg string, color string, dfltColor string) string {
	if c.enabled {
		if color != dfltColor {
			return fmt.Sprintf("%s%s%s", color, msg, dfltColor)
		} else {
			return fmt.Sprintf("%s%s", color, msg)
		}
	}
	return msg
}

func (c *Colorizer) Red(msg string, dfltColor string) string {
	return c.Colorize(msg, ColorsFgRed, dfltColor)
}

func (c *Colorizer) Green(msg string, dfltColor string) string {
	return c.Colorize(msg, ColorsFgGreen, dfltColor)
}

func (c *Colorizer) Yellow(msg string, dfltColor string) string {
	return c.Colorize(msg, ColorsFgYellow, dfltColor)
}

func (c *Colorizer) Blue(msg string, dfltColor string) string {
	return c.Colorize(msg, ColorsFgBlue, dfltColor)
}

func (c *Colorizer) Magenta(msg string, dfltColor string) string {
	return c.Colorize(msg, ColorsFgMagenta, dfltColor)
}

func (c *Colorizer) Cyan(msg string, dfltColor string) string {
	return c.Colorize(msg, ColorsFgCyan, dfltColor)
}

func (c *Colorizer) White(msg string, dfltColor string) string {
	return c.Colorize(msg, ColorsFgWhite, dfltColor)
}

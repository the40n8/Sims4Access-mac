package pairing

import (
	"strings"

	"rsc.io/qr"
)

// TerminalQR renders text as a QR code with Unicode half blocks (two
// modules per character row), dark on light, with the quiet zone scanners
// need. Works in any terminal with a monospace font.
func TerminalQR(text string) (string, error) {
	code, err := qr.Encode(text, qr.M)
	if err != nil {
		return "", err
	}
	const quiet = 2
	n := code.Size
	dark := func(x, y int) bool {
		return x >= 0 && y >= 0 && x < n && y < n && code.Black(x, y)
	}
	var b strings.Builder
	for y := -quiet; y < n+quiet; y += 2 {
		for x := -quiet; x < n+quiet; x++ {
			top, bottom := dark(x, y), dark(x, y+1)
			switch {
			case top && bottom:
				b.WriteRune(' ')
			case top:
				b.WriteRune('▄')
			case bottom:
				b.WriteRune('▀')
			default:
				b.WriteRune('█')
			}
		}
		b.WriteByte('\n')
	}
	return b.String(), nil
}

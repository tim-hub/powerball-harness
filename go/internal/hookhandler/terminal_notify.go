package hookhandler

// terminal_notify.go
// Shared helpers for building the CC 2.1.141+ hook JSON output `terminalSequence` field.
// Opt-in via the HARNESS_TERMINAL_NOTIFY env variable.
//
// Details: .claude/rules/hooks-2.1.139-plus.md
// Shell reference implementation: scripts/lib/terminal-notify.sh

import (
	"os"
	"strings"
)

// terminalNotifyMode is the resolved interpretation of the HARNESS_TERMINAL_NOTIFY env.
type terminalNotifyMode int

const (
	notifyOff     terminalNotifyMode = iota
	notifyBell                       // BEL character only
	notifyTitle                      // OSC 0: set window title
	notifyOSC9                       // OSC 9: iTerm2/ConEmu notification
	notifyDesktop                    // OSC 777: GNOME Terminal desktop notification
)

// resolveTerminalNotifyMode parses HARNESS_TERMINAL_NOTIFY. Unknown values return notifyOff.
func resolveTerminalNotifyMode() terminalNotifyMode {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("HARNESS_TERMINAL_NOTIFY"))) {
	case "", "0":
		return notifyOff
	case "1", "bell":
		return notifyBell
	case "title":
		return notifyTitle
	case "osc9":
		return notifyOSC9
	case "notify":
		return notifyDesktop
	default:
		// Unknown values are silently ignored to avoid breaking unknown environments.
		return notifyOff
	}
}

// sanitizeTerminalText removes control characters (0x00–0x1F, 0x7F) from title/body.
// Only printable characters are allowed to prevent terminal corruption and secret leakage.
func sanitizeTerminalText(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r < 0x20 || r == 0x7F {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// BuildTerminalSequence constructs the raw OSC escape sequence for terminalSequence.
//
// If title is empty, bell mode still emits BEL; all other modes return empty.
// Returns empty when HARNESS_TERMINAL_NOTIFY is unset (opt-in preserved).
//
// The return value is raw bytes; JSON callers let json.Marshal encode it.
func BuildTerminalSequence(title, body string) string {
	mode := resolveTerminalNotifyMode()
	if mode == notifyOff {
		return ""
	}

	cleanTitle := sanitizeTerminalText(title)
	cleanBody := sanitizeTerminalText(body)

	// Bell mode does not need a title; all other modes require one.
	if mode != notifyBell && cleanTitle == "" {
		return ""
	}

	const (
		esc = "\x1b"
		bel = "\x07"
	)

	switch mode {
	case notifyBell:
		return bel
	case notifyTitle:
		return esc + "]0;" + cleanTitle + bel
	case notifyOSC9:
		return esc + "]9;" + cleanTitle + bel
	case notifyDesktop:
		// OSC 777;notify;<title>;<body><BEL>
		if cleanBody != "" {
			return esc + "]777;notify;" + cleanTitle + ";" + cleanBody + bel
		}
		return esc + "]777;notify;" + cleanTitle + bel
	}
	return ""
}

// AugmentWithTerminalSequence adds a terminalSequence field to a hook response map.
// Does nothing when HARNESS_TERMINAL_NOTIFY is unset or title is empty (non-bell mode).
func AugmentWithTerminalSequence(resp map[string]interface{}, title, body string) {
	if resp == nil {
		return
	}
	seq := BuildTerminalSequence(title, body)
	if seq != "" {
		resp["terminalSequence"] = seq
	}
}

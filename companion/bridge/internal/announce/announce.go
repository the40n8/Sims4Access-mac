// Package announce decides which game events the Mac voice says by itself,
// with the same choices as the phone's settings, and words them (English,
// like the game and the mod's own speech).
package announce

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	v1 "github.com/the40n8/simsight-companion/companion/bridge/gen/ts4access/v1"
)

// Default: what the mod does not already say. It reads game windows and
// notifications itself, and announces Sim actions (its own setting).
func Default() *v1.MacAnnounce {
	return &v1.MacAnnounce{Groups: []string{"mood", "need"}, Social: "labels"}
}

const file = "mac-announce.json"

// Load reads the saved setting from dir, or the default.
func Load(dir string) *v1.MacAnnounce {
	b, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return Default()
	}
	var a struct {
		Groups []string `json:"groups"`
		Social string   `json:"social"`
	}
	if json.Unmarshal(b, &a) != nil {
		return Default()
	}
	return &v1.MacAnnounce{Groups: a.Groups, Social: a.Social}
}

// Save keeps the setting in dir.
func Save(dir string, a *v1.MacAnnounce) error {
	b, err := json.MarshalIndent(map[string]any{"groups": a.GetGroups(), "social": a.GetSocial()}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	return os.WriteFile(filepath.Join(dir, file), b, 0o600)
}

// group mirrors the phone's groupOf (companion/web/src/lib/prefs.svelte.ts).
func group(k v1.EventKind) string {
	switch k {
	case v1.EventKind_EVENT_KIND_MOOD:
		return "mood"
	case v1.EventKind_EVENT_KIND_MOODLET_ADDED, v1.EventKind_EVENT_KIND_MOODLET_REMOVED:
		return "moodlet"
	case v1.EventKind_EVENT_KIND_NEED:
		return "need"
	case v1.EventKind_EVENT_KIND_ACTION_STARTED, v1.EventKind_EVENT_KIND_ACTION_QUEUED:
		return "action"
	case v1.EventKind_EVENT_KIND_ACTIVE_SIM, v1.EventKind_EVENT_KIND_AWAY_STARTED, v1.EventKind_EVENT_KIND_AWAY_ENDED,
		v1.EventKind_EVENT_KIND_SPEED, v1.EventKind_EVENT_KIND_FUNDS:
		return "game"
	case v1.EventKind_EVENT_KIND_NOTIFICATION, v1.EventKind_EVENT_KIND_DIALOG, v1.EventKind_EVENT_KIND_PROBLEM,
		v1.EventKind_EVENT_KIND_VISITOR, v1.EventKind_EVENT_KIND_BILLS:
		return "messages"
	case v1.EventKind_EVENT_KIND_SOCIAL, v1.EventKind_EVENT_KIND_RELATIONSHIP:
		return "social"
	}
	return ""
}

func abs(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}

// Wants reports whether the Mac voice should say e under setting a
// (mirrors the phone's socialShown for conversations).
func Wants(a *v1.MacAnnounce, e *v1.Event) bool {
	// A blocked Sim is only shown by a bubble in the game: always said.
	if e.GetKind() == v1.EventKind_EVENT_KIND_ROUTE_FAIL {
		return true
	}
	g := group(e.GetKind())
	if g == "social" {
		switch a.GetSocial() {
		case "all":
			return true
		case "normal":
			return e.GetKind() == v1.EventKind_EVENT_KIND_RELATIONSHIP || abs(e.GetFriendshipDelta()) >= 2 || abs(e.GetRomanceDelta()) >= 2
		case "labels":
			return e.GetKind() == v1.EventKind_EVENT_KIND_RELATIONSHIP
		}
		return false
	}
	return g != "" && slices.Contains(a.GetGroups(), g)
}

func pretty(enum, prefix string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimPrefix(enum, prefix), "_", " "))
}

func signed(n int32) string {
	if n > 0 {
		return "+" + strconv.Itoa(int(n))
	}
	return strconv.Itoa(int(n))
}

var problems = map[string]string{"FIRE": "Fire! %s", "BROKEN": "Broken: %s", "PUDDLE": "A puddle on the floor",
	"SPOILED": "Spoiled food: %s", "FOOD": "Food ready: %s"}

// Why a Sim could not get there, from the game's TransitionFailureReasons.
func routeReason(reason, object string) string {
	switch reason {
	case "BLOCKING_OBJECT":
		if object != "" {
			return object + " is in the way"
		}
		return "something is in the way"
	case "RESERVATION":
		return "someone else is using it"
	case "BUILD_BUY":
		return "walls or building block the way"
	case "GOAL_ON_SLOPE":
		return "the spot is on a slope"
	case "INSUFFICIENT_HEAD_CLEARANCE":
		return "not enough room overhead"
	case "UNKNOWN", "":
		return "the way is blocked"
	}
	return "no way to get there: check doors and stairs"
}

// Text words an event for the Mac voice ("" for nothing to say).
func Text(e *v1.Event) string {
	sim := e.GetSimName()
	switch e.GetKind() {
	case v1.EventKind_EVENT_KIND_MOOD:
		return fmt.Sprintf("%s feels %s.", sim, strings.ToLower(e.GetDetail()))
	case v1.EventKind_EVENT_KIND_MOODLET_ADDED:
		return fmt.Sprintf("%s: %s.", sim, e.GetSubject())
	case v1.EventKind_EVENT_KIND_MOODLET_REMOVED:
		return fmt.Sprintf("%s: %s is over.", sim, e.GetSubject())
	case v1.EventKind_EVENT_KIND_NEED:
		return fmt.Sprintf("%s: %s %s.", sim, e.GetSubject(), pretty(e.GetDetail(), "NEED_LEVEL_"))
	case v1.EventKind_EVENT_KIND_ACTION_STARTED:
		return fmt.Sprintf("%s: %s.", sim, e.GetSubject())
	case v1.EventKind_EVENT_KIND_ACTIVE_SIM:
		return fmt.Sprintf("Now playing %s.", sim)
	case v1.EventKind_EVENT_KIND_SPEED:
		return pretty(e.GetSubject(), "SPEED_") + " speed."
	case v1.EventKind_EVENT_KIND_FUNDS:
		return fmt.Sprintf("Funds: %s simoleons.", e.GetSubject())
	case v1.EventKind_EVENT_KIND_AWAY_STARTED:
		return fmt.Sprintf("%s left for %s.", sim, e.GetSubject())
	case v1.EventKind_EVENT_KIND_AWAY_ENDED:
		return fmt.Sprintf("%s is back from %s.", sim, e.GetSubject())
	case v1.EventKind_EVENT_KIND_NOTIFICATION:
		if e.GetDetail() != "" {
			return e.GetSubject() + ": " + e.GetDetail()
		}
		return e.GetSubject()
	case v1.EventKind_EVENT_KIND_DIALOG:
		return "The game asks: " + e.GetSubject()
	case v1.EventKind_EVENT_KIND_PROBLEM:
		if f, ok := problems[e.GetSubject()]; ok {
			if strings.Contains(f, "%s") {
				return fmt.Sprintf(f, e.GetDetail()) + "."
			}
			return f + "."
		}
		return "Needs attention: " + e.GetDetail() + "."
	case v1.EventKind_EVENT_KIND_VISITOR:
		return e.GetSubject() + " arrived."
	case v1.EventKind_EVENT_KIND_BILLS:
		return "Bills are due: " + e.GetSubject() + " simoleons."
	case v1.EventKind_EVENT_KIND_SOCIAL:
		var parts []string
		if e.GetDetail() != "" {
			parts = append(parts, e.GetDetail())
		}
		if d := e.GetFriendshipDelta(); d != 0 {
			parts = append(parts, "friendship "+signed(d))
		}
		if d := e.GetRomanceDelta(); d != 0 {
			parts = append(parts, "romance "+signed(d))
		}
		return fmt.Sprintf("%s and %s: %s.", sim, e.GetSubject(), strings.Join(parts, ", "))
	case v1.EventKind_EVENT_KIND_ROUTE_FAIL:
		return fmt.Sprintf("%s can't %s: %s.", sim, strings.ToLower(strings.TrimRight(e.GetSubject(), ".…")), routeReason(e.GetDetail(), e.GetObjectName()))
	case v1.EventKind_EVENT_KIND_RELATIONSHIP:
		return fmt.Sprintf("%s and %s: now %s.", sim, e.GetSubject(), e.GetDetail())
	}
	return ""
}

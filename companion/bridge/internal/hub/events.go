package hub

import (
	"strconv"

	v1 "github.com/the40n8/Sims4Access-mac/companion/bridge/gen/ts4access/v1"
)

// Diff turns the change between two game states into events: mood changes,
// moodlets gained or lost, needs dropping to low or critical, actions
// starting, active Sim, speed and funds. No events while either state is
// outside a game (loading a household must not replay its whole state).
func Diff(old, cur *v1.GameState) []*v1.Event {
	if old == nil || cur == nil || !old.GetInGame() || !cur.GetInGame() {
		return nil
	}
	var out []*v1.Event
	add := func(kind v1.EventKind, sim *v1.Sim, subject, detail string) {
		e := &v1.Event{Kind: kind, Subject: subject, Detail: detail}
		if sim != nil {
			e.SimId, e.SimName = sim.GetId(), sim.GetFirstName()
		}
		out = append(out, e)
	}

	if os, cs := old.GetClock().GetSpeed(), cur.GetClock().GetSpeed(); os != cs && cs != v1.Speed_SPEED_UNSPECIFIED {
		add(v1.EventKind_EVENT_KIND_SPEED, nil, cs.String(), "")
	}
	if delta := cur.GetFunds() - old.GetFunds(); delta != 0 {
		add(v1.EventKind_EVENT_KIND_FUNDS, nil, strconv.FormatInt(cur.GetFunds(), 10), strconv.FormatInt(delta, 10))
	}

	diffDialogs(old, cur, add)

	before := map[string]*v1.Sim{}
	for _, s := range old.GetSims() {
		before[s.GetId()] = s
	}
	for _, s := range cur.GetSims() {
		prev, ok := before[s.GetId()]
		if !ok {
			continue
		}
		if s.GetActive() && !prev.GetActive() {
			add(v1.EventKind_EVENT_KIND_ACTIVE_SIM, s, "", "")
		}
		if was, now := prev.GetAwayAt(), s.GetAwayAt(); was != now {
			if now != "" {
				add(v1.EventKind_EVENT_KIND_AWAY_STARTED, s, now, s.GetUntil())
			} else {
				add(v1.EventKind_EVENT_KIND_AWAY_ENDED, s, was, "")
			}
		}
		if !s.GetHere() || !prev.GetHere() {
			continue
		}
		if mk := s.GetMood().GetKey(); mk != "" && mk != prev.GetMood().GetKey() {
			add(v1.EventKind_EVENT_KIND_MOOD, s, mk, s.GetMood().GetName())
		}
		diffMoodlets(prev, s, add)
		diffNeeds(prev, s, add)
		if a := running(s); a != nil {
			if p := running(prev); p == nil || p.GetId() != a.GetId() {
				add(v1.EventKind_EVENT_KIND_ACTION_STARTED, s, a.GetName(), "")
			}
		}
	}
	return out
}

type addFunc func(v1.EventKind, *v1.Sim, string, string)

func moodletID(m *v1.Moodlet) string {
	if m.GetKey() != "" {
		return m.GetKey()
	}
	return m.GetName()
}

func diffMoodlets(prev, cur *v1.Sim, add addFunc) {
	had := map[string]bool{}
	for _, m := range prev.GetMoodlets() {
		had[moodletID(m)] = true
	}
	has := map[string]bool{}
	for _, m := range cur.GetMoodlets() {
		has[moodletID(m)] = true
		if !had[moodletID(m)] {
			add(v1.EventKind_EVENT_KIND_MOODLET_ADDED, cur, m.GetName(), m.GetDescription())
		}
	}
	for _, m := range prev.GetMoodlets() {
		if !has[moodletID(m)] {
			add(v1.EventKind_EVENT_KIND_MOODLET_REMOVED, cur, m.GetName(), "")
		}
	}
}

// diffNeeds reports a need only when it gets worse and reaches low or
// critical, so a need hovering around a threshold does not repeat.
func diffNeeds(prev, cur *v1.Sim, add addFunc) {
	was := map[string]v1.NeedLevel{}
	for _, n := range prev.GetNeeds() {
		was[n.GetKey()] = n.GetLevel()
	}
	for _, n := range cur.GetNeeds() {
		lvl := n.GetLevel()
		if lvl >= v1.NeedLevel_NEED_LEVEL_LOW && lvl > was[n.GetKey()] && was[n.GetKey()] != v1.NeedLevel_NEED_LEVEL_UNSPECIFIED {
			add(v1.EventKind_EVENT_KIND_NEED, cur, n.GetKey(), lvl.String())
		}
	}
}

func running(s *v1.Sim) *v1.Action {
	for _, a := range s.GetActions() {
		if a.GetState() == v1.ActionState_ACTION_STATE_RUNNING {
			return a
		}
	}
	return nil
}

// diffDialogs announces a game window once its words are known (they can
// arrive a moment after the window itself).
func diffDialogs(old, cur *v1.GameState, add func(v1.EventKind, *v1.Sim, string, string)) {
	words := func(d *v1.Dialog) string {
		if d.GetTitle() != "" {
			return d.GetTitle()
		}
		return d.GetText()
	}
	known := map[string]bool{}
	for _, d := range old.GetDialogs() {
		known[d.GetId()] = words(d) != ""
	}
	for _, d := range cur.GetDialogs() {
		if w := words(d); w != "" && !known[d.GetId()] {
			add(v1.EventKind_EVENT_KIND_DIALOG, nil, w, d.GetText())
		}
	}
}

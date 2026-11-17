package hub

import (
	"testing"

	v1 "github.com/the40n8/Sims4Access-mac/companion/bridge/gen/ts4access/v1"
	"google.golang.org/protobuf/proto"
)

func base() *v1.GameState {
	return &v1.GameState{
		InGame: true,
		Funds:  1000,
		Clock:  &v1.Clock{Speed: v1.Speed_SPEED_NORMAL},
		Sims: []*v1.Sim{{
			Id: "1", FirstName: "Bella", Active: true, Here: true,
			Mood:     &v1.Mood{Key: "happy"},
			Needs:    []*v1.Need{{Key: "hunger", Percent: 70, Level: v1.NeedLevel_NEED_LEVEL_GOOD}},
			Moodlets: []*v1.Moodlet{{Key: "Buff_WellRested", Name: "Well Rested Buff"}},
			Actions:  []*v1.Action{{Id: "10", Name: "Read", State: v1.ActionState_ACTION_STATE_RUNNING}},
		}},
	}
}

func kinds(events []*v1.Event) []v1.EventKind {
	var out []v1.EventKind
	for _, e := range events {
		out = append(out, e.GetKind())
	}
	return out
}

func TestDiffNoEventsOutsideGame(t *testing.T) {
	loading := &v1.GameState{InGame: false}
	if got := Diff(loading, base()); len(got) != 0 {
		t.Fatalf("loading -> game produced %v", kinds(got))
	}
	if got := Diff(nil, base()); len(got) != 0 {
		t.Fatalf("first state produced %v", kinds(got))
	}
}

func TestDiffUnchanged(t *testing.T) {
	if got := Diff(base(), base()); len(got) != 0 {
		t.Fatalf("unchanged state produced %v", kinds(got))
	}
}

func TestDiffLocalizedMoodletNameIsNotANewMoodlet(t *testing.T) {
	cur := base()
	cur.Sims[0].Moodlets[0].Name = "Well Rested"
	if got := Diff(base(), cur); len(got) != 0 {
		t.Fatalf("renamed moodlet produced %v", kinds(got))
	}
}

func TestDiffChanges(t *testing.T) {
	cur := proto.Clone(base()).(*v1.GameState)
	s := cur.Sims[0]
	s.Mood = &v1.Mood{Key: "tense", Name: "Tense"}
	s.Needs[0] = &v1.Need{Key: "hunger", Percent: 20, Level: v1.NeedLevel_NEED_LEVEL_LOW}
	s.Moodlets = []*v1.Moodlet{{Key: "Buff_Hungry", Name: "Hungry"}}
	s.Actions = []*v1.Action{{Id: "11", Name: "Eat", State: v1.ActionState_ACTION_STATE_RUNNING}}
	cur.Clock.Speed = v1.Speed_SPEED_PAUSED
	cur.Funds = 900

	got := kinds(Diff(base(), cur))
	want := []v1.EventKind{
		v1.EventKind_EVENT_KIND_SPEED,
		v1.EventKind_EVENT_KIND_FUNDS,
		v1.EventKind_EVENT_KIND_MOOD,
		v1.EventKind_EVENT_KIND_MOODLET_ADDED,
		v1.EventKind_EVENT_KIND_MOODLET_REMOVED,
		v1.EventKind_EVENT_KIND_NEED,
		v1.EventKind_EVENT_KIND_ACTION_STARTED,
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("event %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestDiffWorkAndBack(t *testing.T) {
	home := base()
	work := base()
	work.Sims[0].Here, work.Sims[0].AwayAt, work.Sims[0].Until = false, "work", "5:00 PM"
	if got := kinds(Diff(home, work)); len(got) != 1 || got[0] != v1.EventKind_EVENT_KIND_AWAY_STARTED {
		t.Fatalf("leaving for work: got %v", got)
	}
	if got := kinds(Diff(work, home)); len(got) != 1 || got[0] != v1.EventKind_EVENT_KIND_AWAY_ENDED {
		t.Fatalf("back from work: got %v", got)
	}
}

func TestDiffNeedRecoveringIsQuiet(t *testing.T) {
	old := base()
	old.Sims[0].Needs[0].Level = v1.NeedLevel_NEED_LEVEL_CRITICAL
	cur := base()
	cur.Sims[0].Needs[0].Level = v1.NeedLevel_NEED_LEVEL_LOW
	if got := Diff(old, cur); len(got) != 0 {
		t.Fatalf("recovering need produced %v", kinds(got))
	}
}

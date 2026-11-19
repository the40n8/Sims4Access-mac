package hub

import (
	"testing"

	v1 "github.com/the40n8/simsight-companion/companion/bridge/gen/ts4access/v1"
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

func TestDiffDialogs(t *testing.T) {
	base := func(ds ...*v1.Dialog) *v1.GameState { return &v1.GameState{InGame: true, Dialogs: ds} }
	pending := &v1.Dialog{Id: "7"}
	named := &v1.Dialog{Id: "7", Title: "Cook what?"}
	if ev := Diff(base(), base(pending)); len(ev) != 0 {
		t.Fatalf("no words yet, want no event, got %v", ev)
	}
	ev := Diff(base(pending), base(named))
	if len(ev) != 1 || ev[0].GetKind() != v1.EventKind_EVENT_KIND_DIALOG || ev[0].GetSubject() != "Cook what?" {
		t.Fatalf("want one dialog event, got %v", ev)
	}
	if ev := Diff(base(named), base(named)); len(ev) != 0 {
		t.Fatalf("same window again, want no event, got %v", ev)
	}
}

func TestDiffProblems(t *testing.T) {
	st := func(ps ...*v1.Problem) *v1.GameState { return &v1.GameState{InGame: true, Problems: ps} }
	puddle := &v1.Problem{Kind: "PUDDLE", TargetId: "5", Name: "Puddle"}
	ev := Diff(st(), st(puddle))
	if len(ev) != 1 || ev[0].GetKind() != v1.EventKind_EVENT_KIND_PROBLEM || ev[0].GetSubject() != "PUDDLE" {
		t.Fatalf("want one problem event, got %v", ev)
	}
	if ev := Diff(st(puddle), st(puddle)); len(ev) != 0 {
		t.Fatalf("same puddle, want nothing, got %v", ev)
	}
}

func TestDiffVisitorsAndBills(t *testing.T) {
	before := &v1.GameState{InGame: true}
	after := &v1.GameState{InGame: true, BillsOwed: 340, Visitors: []*v1.Visitor{{SimId: "9", Name: "Nancy"}}}
	ev := Diff(before, after)
	kinds := map[v1.EventKind]string{}
	for _, e := range ev {
		kinds[e.GetKind()] = e.GetSubject()
	}
	if kinds[v1.EventKind_EVENT_KIND_VISITOR] != "Nancy" || kinds[v1.EventKind_EVENT_KIND_BILLS] != "340" {
		t.Fatalf("want visitor and bills events, got %v", ev)
	}
	if ev := Diff(after, after); len(ev) != 0 {
		t.Fatalf("nothing new, got %v", ev)
	}
}

func TestDiffSocial(t *testing.T) {
	sim := func(f int32, labels ...string) *v1.Sim {
		return &v1.Sim{Id: "1", FirstName: "Bella", Here: true,
			Actions:       []*v1.Action{{Id: "a", Name: "Tell Joke", State: v1.ActionState_ACTION_STATE_RUNNING}},
			Relationships: []*v1.Relationship{{SimId: "9", Name: "Nancy", Friendship: f, Labels: labels}}}
	}
	st := func(s *v1.Sim) *v1.GameState { return &v1.GameState{InGame: true, Sims: []*v1.Sim{s}} }
	var social, label *v1.Event
	for _, e := range Diff(st(sim(10)), st(sim(13, "Friend"))) {
		switch e.GetKind() {
		case v1.EventKind_EVENT_KIND_SOCIAL:
			social = e
		case v1.EventKind_EVENT_KIND_RELATIONSHIP:
			label = e
		}
	}
	if social == nil || social.GetFriendshipDelta() != 3 || social.GetDetail() != "Tell Joke" || social.GetSubject() != "Nancy" {
		t.Fatalf("want social +3 during Tell Joke, got %v", social)
	}
	if label == nil || label.GetDetail() != "Friend" {
		t.Fatalf("want new label Friend, got %v", label)
	}
}

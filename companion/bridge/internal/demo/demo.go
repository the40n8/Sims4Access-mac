// Package demo is a fake game for developing the phone app without The
// Sims 4: a two-Sim household whose needs drift, with commands that work.
package demo

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	v1 "github.com/the40n8/Sims4Access-mac/companion/bridge/gen/ts4access/v1"
	"github.com/the40n8/Sims4Access-mac/companion/bridge/internal/hub"
	"google.golang.org/protobuf/proto"
)

type game struct {
	hub     *hub.Hub
	cmds    chan *v1.ToGame
	state   *v1.GameState
	minutes int
	nextID  int
}

var activities = []string{"Cook Dinner", "Watch TV", "Chat", "Read a Book", "Take Shower", "Play Video Games", "Nap"}

// Run drives the fake household until ctx ends.
func Run(ctx context.Context, h *hub.Hub) {
	g := &game{hub: h, cmds: make(chan *v1.ToGame, 16), minutes: 8 * 60}
	g.state = &v1.GameState{
		InGame:        true,
		Funds:         20450,
		HouseholdName: "Goth",
		Clock:         &v1.Clock{Speed: v1.Speed_SPEED_NORMAL},
		Narration:     &v1.Narration{MacVoice: true, SimActions: true},
		Sims: []*v1.Sim{
			g.sim("1001", "Bella", "Goth", true),
			g.sim("1002", "Mortimer", "Goth", false),
		},
	}
	h.SetSender(func(c *v1.ToGame) error { g.cmds <- c; return nil })
	h.SetConnected(true)
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	g.publish()
	for {
		select {
		case <-ctx.Done():
			return
		case c := <-g.cmds:
			res := g.call(c)
			res.Id = c.GetId()
			h.Resolve(res)
			g.publish()
		case <-tick.C:
			g.step()
			g.publish()
		}
	}
}

func (g *game) sim(id, first, last string, active bool) *v1.Sim {
	s := &v1.Sim{Id: id, FirstName: first, LastName: last, Active: active, Here: true,
		Mood: &v1.Mood{Key: "happy", Name: "Happy", Intensity: 1}}
	for _, k := range []string{"hunger", "bladder", "energy", "hygiene", "fun", "social"} {
		s.Needs = append(s.Needs, &v1.Need{Key: k, Percent: int32(50 + rand.IntN(50))})
	}
	s.Moodlets = []*v1.Moodlet{{Name: "Well Rested", Description: "Slept like a log.", MoodKey: "energized", Weight: 1, MinutesLeft: 240}}
	g.addAction(s, activities[rand.IntN(len(activities))])
	return s
}

func (g *game) addAction(s *v1.Sim, name string) {
	g.nextID++
	state := v1.ActionState_ACTION_STATE_RUNNING
	if len(s.Actions) > 0 {
		state = v1.ActionState_ACTION_STATE_QUEUED
	}
	s.Actions = append(s.Actions, &v1.Action{Id: fmt.Sprint(g.nextID), Name: name, State: state, Cancelable: true})
}

func level(p int32) v1.NeedLevel {
	switch {
	case p < 15:
		return v1.NeedLevel_NEED_LEVEL_CRITICAL
	case p < 30:
		return v1.NeedLevel_NEED_LEVEL_LOW
	case p < 60:
		return v1.NeedLevel_NEED_LEVEL_OKAY
	default:
		return v1.NeedLevel_NEED_LEVEL_GOOD
	}
}

func (g *game) step() {
	st := g.state
	if st.Clock.Speed == v1.Speed_SPEED_PAUSED {
		return
	}
	g.minutes += map[v1.Speed]int{v1.Speed_SPEED_NORMAL: 1, v1.Speed_SPEED_FAST: 3, v1.Speed_SPEED_ULTRA: 8}[st.Clock.Speed]
	for _, s := range st.Sims {
		for _, n := range s.Needs {
			n.Percent = max(0, min(100, n.Percent+int32(rand.IntN(5))-3))
			n.Level = level(n.Percent)
		}
		if rand.IntN(12) == 0 {
			name := activities[rand.IntN(len(activities))]
			if len(s.Actions) > 0 {
				s.Actions = s.Actions[1:]
				if len(s.Actions) > 0 {
					s.Actions[0].State = v1.ActionState_ACTION_STATE_RUNNING
				}
			}
			g.addAction(s, name)
		}
	}
	if rand.IntN(20) == 0 {
		s := st.Sims[rand.IntN(len(st.Sims))]
		moods := []struct{ key, name string }{{"happy", "Happy"}, {"energized", "Energized"}, {"playful", "Playful"}, {"bored", "Bored"}, {"tense", "Tense"}}
		m := moods[rand.IntN(len(moods))]
		s.Mood = &v1.Mood{Key: m.key, Name: m.name, Intensity: 1}
	}
}

func (g *game) publish() {
	m := g.minutes % (24 * 60)
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	g.state.Clock.Day = days[(g.minutes/(24*60))%7]
	g.state.Clock.Time = fmt.Sprintf("%02d:%02d", m/60, m%60)
	g.state.Seq++
	g.hub.SetState(proto.Clone(g.state).(*v1.GameState))
}

func (g *game) find(id string) *v1.Sim {
	for _, s := range g.state.Sims {
		if s.Id == id {
			return s
		}
	}
	return nil
}

var demoTargets = []*v1.Target{
	{Id: "o1", Name: "Fridge", Group: "APPLIANCES", Distance: 3},
	{Id: "o2", Name: "Sofa", Group: "COMFORT", Distance: 5},
	{Id: "o3", Name: "Shower", Group: "PLUMBING", Distance: 8},
	{Id: "o4", Name: "Computer", Group: "ELECTRONICS", Distance: 9},
}

var demoMenus = map[string][]*v1.Interaction{
	"self": {{Id: "1", Name: "Check Mood", Enabled: true}, {Id: "2", Name: "Call Someone", Category: "Phone", Enabled: true}},
	"sim":  {{Id: "1", Name: "Chat", Category: "Friendly", Enabled: true}, {Id: "2", Name: "Tell Joke", Category: "Funny", Enabled: true}, {Id: "3", Name: "Ask to Be Boyfriend", Category: "Romance", Reason: "Relationship too low"}},
	"o1":   {{Id: "1", Name: "Quick Meal", Category: "Grab", Enabled: true}, {Id: "2", Name: "Cook", Enabled: true}},
	"o2":   {{Id: "1", Name: "Sit", Enabled: true}, {Id: "2", Name: "Nap", Enabled: true}},
	"o3":   {{Id: "1", Name: "Take Shower", Enabled: true}},
	"o4":   {{Id: "1", Name: "Play Games", Enabled: true}, {Id: "2", Name: "Browse Web", Enabled: true}},
}

func item(t *v1.Target, side string) *v1.RoomItem { return &v1.RoomItem{Target: t, Side: side} }

// demoHouse: a small two-floor house around the demo objects.
func demoHouse() *v1.HouseMap {
	return &v1.HouseMap{Floors: []int32{0, 1}, Rooms: []*v1.Room{
		{Id: "r1", Kind: "KITCHEN", Floor: 0, Width: 4, Depth: 5, Area: 20, Part: "NW", Here: true, Sims: []string{"Bella"},
			Links: []*v1.RoomLink{{RoomId: "r2", Door: false, Side: "E"}, {RoomId: "", Door: true, Side: "N"}},
			Items: []*v1.RoomItem{item(demoTargets[0], "N"), item(&v1.Target{Id: "o5", Name: "Counter", Group: "SURFACE", Kind: "Counters"}, "W")}},
		{Id: "r2", Kind: "LIVING", Floor: 0, Width: 6, Depth: 5, Area: 30, Part: "NE", Distance: 5, Bearing: "E",
			Links: []*v1.RoomLink{{RoomId: "r1", Side: "W"}, {RoomId: "r3", Door: true, Side: "S"}, {RoomId: "", Door: true, Side: "E"}},
			Items: []*v1.RoomItem{item(demoTargets[1], "C"), item(&v1.Target{Id: "o6", Name: "TV", Group: "ELECTRONICS", Kind: "Televisions"}, "N")}},
		{Id: "r3", Kind: "BATHROOM", Floor: 0, Width: 2, Depth: 3, Area: 6, Part: "S", Distance: 8, Bearing: "SE",
			Links: []*v1.RoomLink{{RoomId: "r2", Door: true, Side: "N"}},
			Items: []*v1.RoomItem{item(demoTargets[2], "E")}},
		{Id: "r4", Kind: "BEDROOM", Number: 0, Floor: 1, Width: 4, Depth: 4, Area: 16, Part: "C", Distance: 9, Bearing: "N",
			Items: []*v1.RoomItem{item(demoTargets[3], "S"), item(&v1.Target{Id: "o7", Name: "Double Bed", Group: "COMFORT", Kind: "Beds"}, "N")}},
	}}
}

func (g *game) call(c *v1.ToGame) *v1.CommandResult {
	switch cmd := c.Command.(type) {
	case *v1.ToGame_GetHouseMap:
		return &v1.CommandResult{Outcome: &v1.CommandOutcome{Ok: true}, HouseMap: demoHouse()}
	case *v1.ToGame_ListTargets:
		var targets []*v1.Target
		for _, s := range g.state.Sims {
			targets = append(targets, &v1.Target{Id: s.Id, Name: s.FirstName + " " + s.LastName, IsSim: true, IsSelf: s.Active, Group: "SIMS", Distance: 2})
		}
		return &v1.CommandResult{Outcome: &v1.CommandOutcome{Ok: true}, Targets: append(targets, demoTargets...)}
	case *v1.ToGame_GetInteractions:
		key := cmd.GetInteractions.GetTargetId()
		if s := g.find(key); s != nil {
			key = "sim"
			if s.Active {
				key = "self"
			}
		}
		menu, ok := demoMenus[key]
		if !ok {
			return &v1.CommandResult{Outcome: &v1.CommandOutcome{Reason: "target_unavailable"}}
		}
		return &v1.CommandResult{Outcome: &v1.CommandOutcome{Ok: true}, Interactions: menu}
	case *v1.ToGame_RunInteraction:
		for _, s := range g.state.Sims {
			if s.Active {
				g.addAction(s, "Interaction "+cmd.RunInteraction.GetInteractionId())
			}
		}
		return &v1.CommandResult{Outcome: &v1.CommandOutcome{Ok: true}}
	}
	return &v1.CommandResult{Outcome: g.apply(c)}
}

func (g *game) apply(c *v1.ToGame) *v1.CommandOutcome {
	ok := &v1.CommandOutcome{Ok: true}
	switch cmd := c.Command.(type) {
	case *v1.ToGame_SetSpeed:
		g.state.Clock.Speed = cmd.SetSpeed.GetSpeed()
		return ok
	case *v1.ToGame_SelectSim:
		target := g.find(cmd.SelectSim.GetSimId())
		if target == nil {
			return &v1.CommandOutcome{Reason: "sim_unavailable"}
		}
		for _, s := range g.state.Sims {
			s.Active = s == target
		}
		return ok
	case *v1.ToGame_CancelAction:
		s := g.find(cmd.CancelAction.GetSimId())
		if s == nil {
			return &v1.CommandOutcome{Reason: "sim_unavailable"}
		}
		for i, a := range s.Actions {
			if a.Id == cmd.CancelAction.GetActionId() {
				s.Actions = append(s.Actions[:i], s.Actions[i+1:]...)
				if len(s.Actions) > 0 {
					s.Actions[0].State = v1.ActionState_ACTION_STATE_RUNNING
				}
				return ok
			}
		}
		return &v1.CommandOutcome{Reason: "action_finished"}
	case *v1.ToGame_SetNarration:
		g.state.Narration = cmd.SetNarration.GetNarration()
		return ok
	}
	return &v1.CommandOutcome{Reason: "unknown_command"}
}

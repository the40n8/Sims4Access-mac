package announce

import (
	"testing"

	v1 "github.com/the40n8/Sims4Access-mac/companion/bridge/gen/ts4access/v1"
)

func TestWantsAndText(t *testing.T) {
	a := Default()
	need := &v1.Event{Kind: v1.EventKind_EVENT_KIND_NEED, SimName: "Bella", Subject: "hunger", Detail: "NEED_LEVEL_CRITICAL"}
	if !Wants(a, need) || Text(need) != "Bella: hunger critical." {
		t.Fatalf("need: wants=%v text=%q", Wants(a, need), Text(need))
	}
	dialog := &v1.Event{Kind: v1.EventKind_EVENT_KIND_DIALOG, Subject: "Cook"}
	if Wants(a, dialog) {
		t.Fatal("game windows are off by default (the mod reads them)")
	}
	chat := &v1.Event{Kind: v1.EventKind_EVENT_KIND_SOCIAL, SimName: "Bella", Subject: "Nancy", Detail: "Tell Joke", FriendshipDelta: 3}
	if Wants(a, chat) {
		t.Fatal("labels level: no social moves")
	}
	a.Social = "normal"
	if !Wants(a, chat) || Text(chat) != "Bella and Nancy: Tell Joke, friendship +3." {
		t.Fatalf("social: %q", Text(chat))
	}
	if Wants(a, &v1.Event{Kind: v1.EventKind_EVENT_KIND_SOCIAL, FriendshipDelta: 1}) {
		t.Fatal("normal ignores drifts under 2")
	}
}

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	if err := Save(dir, &v1.MacAnnounce{Groups: []string{"messages"}, Social: "all"}); err != nil {
		t.Fatal(err)
	}
	got := Load(dir)
	if len(got.GetGroups()) != 1 || got.GetGroups()[0] != "messages" || got.GetSocial() != "all" {
		t.Fatalf("got %v", got)
	}
}

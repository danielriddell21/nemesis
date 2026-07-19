package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/danielriddell21/nemesis/internal/gui"
)

func TestMsgJSONRoundTrip(t *testing.T) {
	orig := gui.Msg{
		Type: "state",
		State: &gui.StateMsg{
			Tick: 12, PlayerX: 1.5, PlayerY: 2.5, AlienState: "hunt",
			Path: [][2]int{{3, 4}, {3, 5}}, Menace: 0.7, Done: 1, Total: 3,
		},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(orig); err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Error("encoding is not line-delimited")
	}
	var got gui.Msg
	if err := json.NewDecoder(&buf).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Fatalf("round trip changed message:\n%+v\n%+v", got, orig)
	}
}

func TestHubBroadcastsToOthersAndStoresHello(t *testing.T) {
	h := newHub("self")
	a := make(chan gui.Msg, 4)
	b := make(chan gui.Msg, 4)
	ida := h.addParticipant(a, nil)
	h.addParticipant(b, nil)

	h.handle(ida, gui.Msg{Type: "hello", Seed: 7})

	select {
	case m := <-b:
		if m.Seed != 7 {
			t.Fatalf("b received seed %d, want 7", m.Seed)
		}
	default:
		t.Fatal("other participant did not receive the hello")
	}
	select {
	case <-a:
		t.Fatal("sender should not receive its own message")
	default:
	}
	if h.last.Seed != 7 {
		t.Fatalf("hub.last seed = %d, want 7 stored", h.last.Seed)
	}
}

func TestHubSendsLastHelloToNewParticipant(t *testing.T) {
	// A visualiser that arrives after the run began still needs the level.
	h := newHub("self")
	ida := h.addParticipant(make(chan gui.Msg, 4), nil)
	h.handle(ida, gui.Msg{Type: "hello", Seed: 3})

	late := make(chan gui.Msg, 4)
	h.addParticipant(late, nil)
	select {
	case m := <-late:
		if m.Seed != 3 {
			t.Fatalf("late participant got seed %d, want 3", m.Seed)
		}
	default:
		t.Fatal("late participant did not receive the last hello")
	}
}

func TestHubDropClosesChannel(t *testing.T) {
	h := newHub("self")
	a := make(chan gui.Msg, 4)
	id := h.addParticipant(a, nil)
	h.handle(id, gui.Msg{Type: eofType})
	if _, ok := <-a; ok {
		t.Fatal("participant channel not closed after drop")
	}
}

func TestChildLinkBridgesStdio(t *testing.T) {
	inR, inW := io.Pipe()
	var out bytes.Buffer
	link := childLink(inR, &out)

	go func() {
		enc := json.NewEncoder(inW)
		_ = enc.Encode(gui.Msg{Type: "hello", Seed: 42})
		_ = inW.Close()
	}()

	select {
	case m := <-link.In:
		if m.Type != "hello" || m.Seed != 42 {
			t.Fatalf("got %+v", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no message arrived from stdin")
	}
	// Stdin closing must close In so the window can terminate.
	select {
	case _, ok := <-link.In:
		if ok {
			t.Fatal("expected closed In after EOF")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("In not closed after stdin EOF")
	}

	link.Out <- gui.Msg{Type: "events"}
	deadline := time.Now().Add(2 * time.Second)
	for out.Len() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !strings.Contains(out.String(), `"events"`) {
		t.Fatalf("stdout writer produced %q", out.String())
	}
}

func TestExecuteVersion(t *testing.T) {
	cmd := newRootCmd("1.2.3")
	if cmd.Version != "1.2.3" {
		t.Fatalf("version = %q", cmd.Version)
	}
	if cmd.Flags().Lookup("child") == nil || !cmd.Flags().Lookup("child").Hidden {
		t.Fatal("--child flag should exist and be hidden")
	}
}

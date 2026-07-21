package cli

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/danielriddell21/crucible/hub"

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

func TestRouteClassifiesMessages(t *testing.T) {
	// The opening hello is shared state a late visualiser needs; per-tick
	// batches only fan out; anything else is ignored. The hub mechanics behind
	// each route are covered by crucible/hub's own tests.
	tests := []struct {
		typ  string
		want hub.Route
	}{
		{"hello", hub.RouteState},
		{"state", hub.RouteBroadcast},
		{"events", hub.RouteBroadcast},
		{"", hub.RouteNone},
		{"unknown", hub.RouteNone},
	}
	for _, tt := range tests {
		if got := route(gui.Msg{Type: tt.typ}); got != tt.want {
			t.Errorf("route(%q) = %v, want %v", tt.typ, got, tt.want)
		}
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

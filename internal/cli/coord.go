package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/danielriddell21/nemesis/internal/gui"
)

const eofType = "_eof"

type hub struct {
	self   string
	inbox  chan srcMsg
	done   chan struct{}
	mu     sync.Mutex
	parts  map[int]*participant
	nextID int
	last   gui.Msg
}

type participant struct {
	out chan gui.Msg
	cmd *exec.Cmd
}

type srcMsg struct {
	id int
	m  gui.Msg
}

func newHub(self string) *hub {
	return &hub{self: self, inbox: make(chan srcMsg, 128), done: make(chan struct{}), parts: map[int]*participant{}}
}

func (h *hub) addParticipant(out chan gui.Msg, cmd *exec.Cmd) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	id := h.nextID
	h.nextID++
	h.parts[id] = &participant{out: out, cmd: cmd}
	if h.last.Type == "hello" {
		trySend(out, h.last)
	}
	return id
}

func (h *hub) run() {
	// Select on done rather than ranging over inbox: child reader goroutines may
	// still send after teardown, so inbox is never closed.
	for {
		select {
		case sm := <-h.inbox:
			h.handle(sm.id, sm.m)
		case <-h.done:
			return
		}
	}
}

func (h *hub) handle(src int, m gui.Msg) {
	switch m.Type {
	case "hello":
		h.mu.Lock()
		h.last = m
		h.mu.Unlock()
		h.broadcastExcept(src, m)
	case "state", "events":
		h.broadcastExcept(src, m)
	case eofType:
		h.drop(src)
	}
}

func (h *hub) broadcastExcept(src int, m gui.Msg) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for id, p := range h.parts {
		if id != src {
			trySend(p.out, m)
		}
	}
}

func (h *hub) drop(id int) {
	h.mu.Lock()
	p := h.parts[id]
	delete(h.parts, id)
	if p != nil {
		close(p.out)
	}
	h.mu.Unlock()
	if p != nil && p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
}

func (h *hub) spawnChild(index int) {
	cmd := exec.CommandContext(context.Background(), h.self, fmt.Sprintf("--child=%d", index))
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "spawn visualiser window: stdin pipe: %v\n", err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "spawn visualiser window: stdout pipe: %v\n", err)
		return
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "spawn visualiser window: start: %v\n", err)
		return
	}

	out := make(chan gui.Msg, 64)
	id := h.addParticipant(out, cmd)

	go func() { // hub -> child stdin
		enc := json.NewEncoder(stdin)
		for m := range out {
			if enc.Encode(m) != nil {
				break
			}
		}
		_ = stdin.Close()
	}()
	go func() { // child stdout -> hub, then signal removal and reap
		dec := json.NewDecoder(stdout)
		for {
			var m gui.Msg
			if dec.Decode(&m) != nil {
				break
			}
			h.inbox <- srcMsg{id: id, m: m}
		}
		h.inbox <- srcMsg{id: id, m: gui.Msg{Type: eofType}}
		_ = cmd.Wait()
	}()
}

func (h *hub) shutdown() {
	h.mu.Lock()
	cmds := make([]*exec.Cmd, 0, len(h.parts))
	for _, p := range h.parts {
		if p.cmd != nil {
			cmds = append(cmds, p.cmd)
		}
	}
	h.mu.Unlock()
	for _, c := range cmds {
		if c.Process != nil {
			_ = c.Process.Kill()
		}
	}
	close(h.done)
}

func trySend(ch chan gui.Msg, m gui.Msg) {
	select {
	case ch <- m:
	default:
	}
}

func runLeader(cfg gui.Config) error {
	if !cfg.Visualiser {
		cfg.Role = gui.RoleGame
		if err := gui.Run(cfg); err != nil {
			return fmt.Errorf("run gui: %w", err)
		}
		return nil
	}
	return lead(cfg, os.Args[0], gui.Run)
}

func lead(cfg gui.Config, self string, runWindow func(gui.Config) error) error {
	h := newHub(self)
	leaderIn := make(chan gui.Msg, 64)
	leaderOut := make(chan gui.Msg, 64)
	h.addParticipant(leaderIn, nil) // the game window is id 0
	go h.run()
	go func() {
		for m := range leaderOut {
			h.inbox <- srcMsg{id: 0, m: m}
		}
	}()
	h.spawnChild(1) // the visualiser

	cfg.Role = gui.RoleGame
	cfg.Link = &gui.Link{In: leaderIn, Out: leaderOut}
	err := runWindow(cfg)
	close(leaderOut)
	h.shutdown()
	if err != nil {
		return fmt.Errorf("run game: %w", err)
	}
	return nil
}

func runChild(cfg gui.Config) error {
	cfg.Role = gui.RoleVisualiser
	cfg.Link = childLink(os.Stdin, os.Stdout)
	if err := gui.Run(cfg); err != nil {
		return fmt.Errorf("run visualiser: %w", err)
	}
	return nil
}

func childLink(r io.Reader, w io.Writer) *gui.Link {
	in := make(chan gui.Msg, 64)
	out := make(chan gui.Msg, 64)
	go func() { // leader stdin -> in
		dec := json.NewDecoder(r)
		for {
			var m gui.Msg
			if dec.Decode(&m) != nil {
				break
			}
			in <- m
		}
		close(in)
	}()
	go func() { // out -> leader via stdout
		enc := json.NewEncoder(w)
		for m := range out {
			if enc.Encode(m) != nil {
				break
			}
		}
	}()
	return &gui.Link{In: in, Out: out}
}

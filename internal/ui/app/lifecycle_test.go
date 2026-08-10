package app

import (
	"testing"
)

type stubSessions struct {
	resets int
}

func (s *stubSessions) EnsureSession() (int64, error) { return 1, nil }
func (s *stubSessions) ResetSession() error           { s.resets++; return nil }

func TestShutdownCmdFinalizesSession(t *testing.T) {
	sess := &stubSessions{}
	m := Model{Dispatcher: &Dispatcher{Sessions: sess}}

	msg := m.ShutdownCmd()()
	if _, ok := msg.(ShutdownMsg); !ok {
		t.Fatalf("expected ShutdownMsg, got %T", msg)
	}
	if sess.resets != 1 {
		t.Errorf("ResetSession calls = %d, want 1", sess.resets)
	}
}

func TestShutdownCmdNilDispatcher(t *testing.T) {
	m := Model{}
	msg := m.ShutdownCmd()()
	if _, ok := msg.(ShutdownMsg); !ok {
		t.Fatalf("expected ShutdownMsg, got %T", msg)
	}
}

package main

import "testing"

func TestDispatchNoArgs(t *testing.T) {
	if code := dispatch(nil); code != 1 {
		t.Errorf("dispatch(nil) = %d, want 1", code)
	}
}

func TestDispatchUnknownCommand(t *testing.T) {
	if code := dispatch([]string{"bogus"}); code != 1 {
		t.Errorf("dispatch(bogus) = %d, want 1", code)
	}
}

func TestDispatchVersion(t *testing.T) {
	if code := dispatch([]string{"version"}); code != 0 {
		t.Errorf("dispatch(version) = %d, want 0", code)
	}
}

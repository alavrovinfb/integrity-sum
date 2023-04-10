package process

import "testing"

func TestGetPID(t *testing.T) {
	// Skipping the case where the process is found
	// because it is not possible to test it in a cross-platform way,
	// and mocking the GetPID function won't give us much.

	// The process is not found
	pid, err := GetPID("someUnknownProcess")
	if err != ErrProcessNotFound {
		t.Errorf("expected error %v, got %v", ErrProcessNotFound, err)
	}
	if pid != 0 {
		t.Errorf("PID should be zero")
	}
}

package main

import (
	"testing"
)


func TestPingHost(t *testing.T) {
	t.Run("run ip test", func(t *testing.T) {
		t.Parallel()
		_, err := pingHost("127.0.0.1", 2)
		if err != nil {
			t.Errorf("failed to ping v4: %v", err)
		}
	})

	t.Run("run ip test", func(t *testing.T) {
		t.Parallel()
		_, err := pingHost("::1", 2)
		if err != nil {
			t.Errorf("failed to ping v6: %v", err)
		}
	})
}


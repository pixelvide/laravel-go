package root

import (
	"testing"
)

func TestSetInfo(t *testing.T) {
	// Save original values
	origUse := rootCmd.Use
	origShort := rootCmd.Short
	origLong := rootCmd.Long
	defer func() {
		// Restore original values
		rootCmd.Use = origUse
		rootCmd.Short = origShort
		rootCmd.Long = origLong
	}()

	use := "test-app"
	short := "Test Short"
	long := "Test Long Description"

	SetInfo(use, short, long)

	if rootCmd.Use != use {
		t.Errorf("Expected Use to be %s, got %s", use, rootCmd.Use)
	}
	if rootCmd.Short != short {
		t.Errorf("Expected Short to be %s, got %s", short, rootCmd.Short)
	}
	if rootCmd.Long != long {
		t.Errorf("Expected Long to be %s, got %s", long, rootCmd.Long)
	}
}

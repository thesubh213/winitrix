//go:build windows

package cmd

import (
	"testing"

	"golang.org/x/sys/windows/registry"
)

// newTestKey creates a throwaway HKCU key so PATH logic can be exercised
// without ever touching the real user Environment key.
func newTestKey(t *testing.T) registry.Key {
	t.Helper()

	path := `Software\WinitrixTests\` + t.Name()
	key, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.QUERY_VALUE|registry.SET_VALUE|registry.WRITE)
	if err != nil {
		t.Fatalf("create test key: %v", err)
	}
	t.Cleanup(func() {
		if err := registry.DeleteKey(registry.CURRENT_USER, path); err != nil {
			t.Logf("cleanup: delete test key: %v", err)
		}
		_ = key.Close()
	})
	return key
}

func TestEnsurePathInKeyAppendsAndPreservesType(t *testing.T) {
	key := newTestKey(t)

	original := `C:\Windows;C:\Tools`
	if err := key.SetExpandStringValue("Path", original); err != nil {
		t.Fatalf("seed registry value: %v", err)
	}

	updated, err := ensurePathInKey(key, `C:\Users\test\Winitrix\bin`)
	if err != nil {
		t.Fatalf("ensurePathInKey: %v", err)
	}
	if !updated {
		t.Fatal("expected PATH to be updated")
	}

	value, valueType, err := key.GetStringValue("Path")
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	want := original + `;C:\Users\test\Winitrix\bin`
	if value != want {
		t.Fatalf("PATH = %q, want %q", value, want)
	}
	if valueType != registry.EXPAND_SZ {
		t.Fatalf("value type = %d, want REG_EXPAND_SZ (%d)", valueType, registry.EXPAND_SZ)
	}

	// A second run is a no-op.
	updated, err = ensurePathInKey(key, `C:\Users\test\Winitrix\bin`)
	if err != nil {
		t.Fatalf("second ensurePathInKey: %v", err)
	}
	if updated {
		t.Fatal("expected no update on the second run")
	}
}

func TestEnsurePathInKeyRefusesUnexpectedType(t *testing.T) {
	key := newTestKey(t)

	// A non-string Path value (e.g. corrupted or hand-edited) must never be
	// overwritten: the old code treated this read failure as an empty PATH and
	// replaced the whole value.
	if err := key.SetDWordValue("Path", 1); err != nil {
		t.Fatalf("seed registry value: %v", err)
	}

	if _, err := ensurePathInKey(key, `C:\bin`); err == nil {
		t.Fatal("expected an error for a non-string PATH value")
	}

	// The original value must be untouched.
	dword, _, err := key.GetIntegerValue("Path")
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if dword != 1 {
		t.Fatalf("PATH value was modified: %d", dword)
	}
	if _, _, err := key.GetStringValue("Path"); err == nil {
		t.Fatal("value type changed to string")
	}
}

func TestEnsurePathInKeyCreatesMissingValue(t *testing.T) {
	key := newTestKey(t)

	updated, err := ensurePathInKey(key, `C:\bin`)
	if err != nil {
		t.Fatalf("ensurePathInKey: %v", err)
	}
	if !updated {
		t.Fatal("expected PATH to be created")
	}

	value, _, err := key.GetStringValue("Path")
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if value != `C:\bin` {
		t.Fatalf("PATH = %q, want %q", value, `C:\bin`)
	}
}

func TestPathContains(t *testing.T) {
	list := `C:\Windows;C:\Program Files\Git\cmd;`
	if !pathContains(list, `c:\program files\git\cmd`) {
		t.Fatal("expected case-insensitive match")
	}
	if !pathContains(list, `C:\Windows\`) {
		t.Fatal("expected cleaned-path match")
	}
	if pathContains(list, `C:\Other`) {
		t.Fatal("unexpected match")
	}
}

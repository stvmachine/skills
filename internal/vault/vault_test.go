package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagerPaths(t *testing.T) {
	m := New()
	if !strings.HasSuffix(m.EnvDir, ".stevmachine-skills") {
		t.Fatalf("unexpected env dir: %s", m.EnvDir)
	}

	if m.keysFile("") != filepath.Join(m.EnvDir, ".env.keys") {
		t.Error("default keys file wrong")
	}
	if m.keysFile("prod") != filepath.Join(m.EnvDir, ".env.prod.keys") {
		t.Error("env keys file wrong")
	}
	if m.envFile("prod") != filepath.Join(m.EnvDir, ".env.prod") {
		t.Error("env file wrong")
	}
}

func TestIsInitialized(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{EnvDir: dir}
	if m.IsInitialized("") {
		t.Error("should not be initialized")
	}
	_ = os.WriteFile(filepath.Join(dir, ".env.keys"), []byte("x"), 0o600)
	if !m.IsInitialized("") {
		t.Error("should be initialized")
	}
}

func TestInitVault(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{EnvDir: dir}
	// dotenvx won't be available, so InitVault will fail at encrypt step
	_ = m.InitVault("")
	if _, err := os.Stat(filepath.Join(dir, ".env")); os.IsNotExist(err) {
		t.Error(".env should be created even if encrypt fails")
	}
	info, _ := os.Stat(dir)
	if info.Mode().Perm()&0o777 != 0o700 {
		t.Logf("dir perms are %o", info.Mode().Perm())
	}
}

func TestSetAndList(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{EnvDir: dir}

	if err := m.Set("KEY", "value", ""); err != nil {
		// dotenvx may not be available in CI
		if !strings.Contains(err.Error(), "dotenvx") {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Skip("dotenvx not available")
	}

	val, err := m.Get("KEY", "")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if val != "value" {
		t.Errorf("expected KEY=value, got: %s", val)
	}

	// Update existing key
	if err := m.Set("KEY", "new", ""); err != nil {
		t.Fatalf("Set update error: %v", err)
	}
	val, err = m.Get("KEY", "")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if val != "new" {
		t.Errorf("expected KEY=new, got: %s", val)
	}
}

func TestListVars(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{EnvDir: dir}
	_ = os.WriteFile(filepath.Join(dir, ".env"), []byte("A=1\nB=2\n# comment\n\n"), 0o600)

	vars, err := m.ListVars("")
	if err != nil {
		t.Fatalf("ListVars error: %v", err)
	}
	if vars["A"] != "1" || vars["B"] != "2" {
		t.Errorf("unexpected vars: %v", vars)
	}
	if _, ok := vars["# comment"]; ok {
		t.Error("comments should be ignored")
	}
}

func TestListMasked(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{EnvDir: dir}
	_ = os.WriteFile(filepath.Join(dir, ".env"), []byte("SHORT=ab\nLONG=abcdefghijklmnop\n"), 0o600)

	masked := m.ListMasked("")
	if masked["SHORT"] != "****" {
		t.Errorf("short mask wrong: %s", masked["SHORT"])
	}
	if masked["LONG"] != "abcd****mnop" {
		t.Errorf("long mask wrong: %s", masked["LONG"])
	}
}

func TestDestroyPlaintext(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{EnvDir: dir}
	f := filepath.Join(dir, ".env")
	_ = os.WriteFile(f, []byte("secret"), 0o600)
	m.DestroyPlaintext("")
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Error(".env should be removed")
	}
}

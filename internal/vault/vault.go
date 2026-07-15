package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Manager struct {
	EnvDir string
}

func New() *Manager {
	home, _ := os.UserHomeDir()
	return &Manager{
		EnvDir: filepath.Join(home, ".stevmachine-skills"),
	}
}

func (m *Manager) keysFile(env string) string {
	if env != "" {
		return filepath.Join(m.EnvDir, fmt.Sprintf(".env.%s.keys", env))
	}
	return filepath.Join(m.EnvDir, ".env.keys")
}

func (m *Manager) envFile(env string) string {
	if env != "" {
		return filepath.Join(m.EnvDir, fmt.Sprintf(".env.%s", env))
	}
	return filepath.Join(m.EnvDir, ".env")
}

func (m *Manager) IsInitialized(env string) bool {
	_, err := os.Stat(m.keysFile(env))
	return err == nil
}

func (m *Manager) InitVault(env string) error {
	_ = os.MkdirAll(m.EnvDir, 0o700)
	f, err := os.OpenFile(m.envFile(env), os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_ = f.Close()
	return m.Encrypt(env)
}

func findDotenvx() string {
	if p, err := exec.LookPath("dotenvx"); err == nil {
		return p
	}
	home, _ := os.UserHomeDir()
	candidates := []string{
		"/usr/local/bin/dotenvx",
		"/opt/homebrew/bin/dotenvx",
		filepath.Join(home, ".local/bin/dotenvx"),
		filepath.Join(home, "node_modules/.bin/dotenvx"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func runDotenvx(dir string, args ...string) ([]byte, error) {
	bin := findDotenvx()
	if bin == "" {
		return nil, fmt.Errorf("dotenvx not found")
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func (m *Manager) Encrypt(env string) error {
	envFile := m.envFile(env)
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return fmt.Errorf("no .env file to encrypt")
	}
	out, err := runDotenvx(m.EnvDir, "encrypt", "-f", envFile)
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(out))
	}
	return nil
}

func (m *Manager) Decrypt(env string) error {
	envFile := m.envFile(env)
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return fmt.Errorf("no .env file to decrypt")
	}
	out, err := runDotenvx(m.EnvDir, "decrypt", "-f", envFile)
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(out))
	}
	return nil
}

func (m *Manager) Set(key, value, env string) error {
	envFile := m.envFile(env)
	out, err := runDotenvx(m.EnvDir, "set", key, value, "-f", envFile)
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(out))
	}
	return nil
}

func (m *Manager) Get(key, env string) (string, error) {
	out, err := runDotenvx(m.EnvDir, "get", key, "-f", m.envFile(env))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (m *Manager) ListVars(env string) (map[string]string, error) {
	out, err := runDotenvx(m.EnvDir, "get", "-f", m.envFile(env))
	if err != nil {
		return nil, err
	}
	var raw map[string]string
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for k, v := range raw {
		if k == "DOTENV_PUBLIC_KEY" {
			continue
		}
		result[k] = v
	}
	return result, nil
}

func (m *Manager) ListMasked(env string) map[string]string {
	vars, err := m.ListVars(env)
	if err != nil {
		return nil
	}
	for k, v := range vars {
		if len(v) <= 8 {
			vars[k] = "****"
		} else {
			vars[k] = v[:4] + "****" + v[len(v)-4:]
		}
	}
	return vars
}

func (m *Manager) Rotate(env string) error {
	envFile := m.envFile(env)
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return fmt.Errorf("no .env file to rotate")
	}
	out, err := runDotenvx(m.EnvDir, "rotate", "-f", envFile)
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(out))
	}
	return nil
}

func (m *Manager) DestroyPlaintext(env string) {
	envFile := m.envFile(env)
	info, err := os.Stat(envFile)
	if err != nil {
		return
	}
	zeros := make([]byte, info.Size())
	_ = os.WriteFile(envFile, zeros, 0o600)
	_ = os.Remove(envFile)
}

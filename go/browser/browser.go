package browser

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type BrowserRef struct {
	Browser *rod.Browser
	ID      string
}

type Manager struct {
	browsers []*BrowserRef
	nextID   int
}

func NewManager() *Manager {
	return &Manager{}
}

type Config struct {
	Headless bool
	Stealth  bool
	Proxy    string
}

func DefaultConfig() Config {
	return Config{Headless: true, Stealth: true}
}

func (m *Manager) Launch(cfg Config) (*BrowserRef, error) {
	u := launcher.New().Headless(cfg.Headless)
	if cfg.Stealth {
		u.Delete("enable-automation")
	}
	controlURL, err := u.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch: %w", err)
	}
	b := rod.New().ControlURL(controlURL)
	if err := b.Connect(); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	ref := &BrowserRef{
		Browser: b,
		ID:      fmt.Sprintf("b%d", m.nextID),
	}
	m.nextID++
	m.browsers = append(m.browsers, ref)
	return ref, nil
}

func (m *Manager) Find(id string) *BrowserRef {
	for _, ref := range m.browsers {
		if ref.ID == id {
			return ref
		}
	}
	return nil
}

func (m *Manager) CloseAll() {
	for _, ref := range m.browsers {
		_ = ref.Browser.Close()
	}
	m.browsers = nil
}
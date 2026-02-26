package alert

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"cryptobar/internal/config"
	"cryptobar/internal/price"
)

const cooldownDuration = 5 * time.Minute

type Manager struct {
	store     *price.Store
	mu        sync.Mutex
	cooldowns map[string]time.Time // pair -> last alert time
}

func NewManager(store *price.Store) *Manager {
	m := &Manager{
		store:     store,
		cooldowns: make(map[string]time.Time),
	}
	store.OnUpdate(m.check)
	return m
}

func (m *Manager) check(pair string, data price.PriceData) {
	alerts := config.GetAlerts()

	for _, rule := range alerts {
		if rule.Pair != pair || !rule.Enabled {
			continue
		}

		m.mu.Lock()
		lastAlert, hasCooldown := m.cooldowns[pair]
		m.mu.Unlock()

		if hasCooldown && time.Since(lastAlert) < cooldownDuration {
			continue
		}

		var triggered bool
		var direction string

		if rule.Above > 0 && data.Price >= rule.Above {
			triggered = true
			direction = "above"
		}
		if rule.Below > 0 && data.Price <= rule.Below {
			triggered = true
			direction = "below"
		}

		if triggered {
			m.mu.Lock()
			m.cooldowns[pair] = time.Now()
			m.mu.Unlock()

			symbol := pair
			title := fmt.Sprintf("CryptoBar Price Alert")
			msg := fmt.Sprintf("%s is %s target: $%s", symbol, direction, data.FormatPrice(-1))
			go sendNotification(title, msg)
			log.Printf("[Alert] %s: %s", title, msg)
		}
	}
}

func sendNotification(title, message string) {
	script := fmt.Sprintf(`display notification "%s" with title "%s" sound name "Glass"`, message, title)
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		log.Printf("[Alert] notification error: %v", err)
	}
}

package nervo

import (
	"log"
	"time"
)

func watchManagerHealth(m *Manager, treatment func()) {
	t := time.NewTicker(time.Second)

	for {
		<-t.C
		if err := m.pingWithTimeout(time.Second * 20); err != nil {
			log.Println("detected unhealthy manager, applying treatment...")
			treatment()
		}
	}
}

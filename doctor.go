package nervo

import (
	"fmt"
	"log"
	"os/exec"
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

func resetUsb() (output string, err error) {
	cmd := exec.Command(
		"sh",
		"-c",
		fmt.Sprintf(`
			echo 0 > /sys/devices/platform/soc/3f980000.usb/buspower
			sleep 1
			echo 1 > /sys/devices/platform/soc/3f980000.usb/buspower
			sleep 1
		`),
	)

	out, execErr := cmd.CombinedOutput()
	output = string(out)
	err = execErr

	return
}

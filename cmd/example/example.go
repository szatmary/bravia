package main

import (
	"fmt"
	"os"
	"time"

	"github.com/szatmary/bravia"
)

// To enable this feature on your device see:
// See https://pro-bravia.sony.net/develop/integrate/ssip/overview/index.html

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage %s [ipaddress:port]\n", os.Args[0])
		os.Exit(2)
	}

	// Connect to a bravia devices
	tv := bravia.NewBravia(os.Args[1])
	if tv == nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to a device at %s\n", os.Args[1])
		os.Exit(1)
	}
	defer tv.Close()

	// Create channel to receive device notifications
	notify := tv.MakeNotifyChan()
	go func() {
		for {
			if notify == nil {
				return
			}

			select {
			case msg := <-notify:
				switch msg.Command {
				case bravia.CONNECTED:
					val, err := msg.Bool()
					fmt.Printf("Connected to device? %v (err: %v)", val, err)
				case bravia.POWR:
					val, err := msg.Bool()
					fmt.Printf("Power on? %v, (err: %v)", val, err)
				case bravia.AMUT:
					val, err := msg.Bool()
					fmt.Printf("Audio Muted? %v, (err: %v)", val, err)
				case bravia.PMUT:
					val, err := msg.Bool()
					fmt.Printf("Picture Muted? %v, (err: %v)", val, err)
				case bravia.VOLU:
					val, err := msg.Int()
					fmt.Printf("Current Volume: %d, (err: %v)", val, err)
				case bravia.INPT:
					intutType, inputNum, err := msg.Ints()
					name, ok := bravia.InputName[intutType]
					if !ok {
						name = "Unknown"
					}
					fmt.Printf("Current Input: %s%d, (err: %v)", name, inputNum, err)
				}
			}
		}
	}()

	macaddr, err := tv.GetMacAddress("eth0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	fmt.Printf("Device MAC address: %s\n", macaddr)
	tv.SetPowerStatus(true)
	time.Sleep(2 * time.Second)
	tv.SetAudioVolume(0)
	tv.SetInput(bravia.HDMI, 1)
	time.Sleep(2 * time.Second)
	// Ir messages emulate remote button presses
	// Many IR messaged don't have feebback
	tv.SendIrMessage(bravia.IrHelp)
	time.Sleep(1 * time.Second)
	tv.SendIrMessage(bravia.IrDown)
	time.Sleep(1 * time.Second)
	tv.SendIrMessage(bravia.IrDown)
	time.Sleep(1 * time.Second)
	tv.SendIrMessage(bravia.IrUp)
	time.Sleep(1 * time.Second)
	tv.SendIrMessage(bravia.IrUp)
	time.Sleep(1 * time.Second)
	tv.SetPowerStatus(false)
	os.Exit(0)

}

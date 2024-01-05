package utility

import (
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func CheckMicrocontrollerHealth(client mqtt.Client, responseReceived *bool) {
	timeout := time.After(10 * time.Second)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			if !*responseReceived {
				log.Fatalf("Cannot connect to Microcontroller")
			}
			return
		case <-tick:
			if *responseReceived {
				log.Println("Microcontroller is online")
				return
			}
		}
	}
}

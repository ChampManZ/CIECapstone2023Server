package utility

import (
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func CheckMicrocontrollerHealth(client mqtt.Client, responseReceived *bool) {
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(1 * time.Second) //fix mem leak lol
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			if !*responseReceived {
				log.Fatalf("Cannot connect to Microcontroller")
			}
			return
		case <-ticker.C:
			if *responseReceived {
				log.Println("Microcontroller is online")
				return
			}
		}
	}
}

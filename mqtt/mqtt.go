package mqttx

import (
	conx "capstone/server/controller"
	"capstone/server/entity"
	"capstone/server/utility"
	"capstone/server/utility/config"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func NewMqttx(conf *config.Config) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", conf.MQTT_server, conf.MQTT_port))
	opts.SetClientID("go-mqtt-client")
	opts.SetUsername(conf.MQTT_username)
	opts.SetPassword(conf.MQTT_password)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return client
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected to MQTT Broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

func OnSignal(mc conx.Controller) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		messageString := string(msg.Payload())
		log.Printf("Received message: %s from topic: %s\n", messageString, msg.Topic())
		if messageString == "1" {
			mc.IncrementGlobalCounter()
			log.Printf("Counter: %d", mc.GlobalCounter)
			var previous, current, next *entity.Student = nil, nil, nil
			var prevPayload, currPayload, nextPayload *entity.IndividualPayload = nil, nil, nil
			if prevStudent, ok := mc.GetStudentByCounter(mc.GlobalCounter - 1); ok {
				previous = prevStudent
				prevPayload = &entity.IndividualPayload{
					Type: "student name",
					Data: entity.StudentPayload{
						OrderOfReading: mc.GlobalCounter - 1,
						Name:           previous.FirstName + " " + previous.LastName,
						Reading:        previous.Notes,
						Certificate:    previous.Certificate,
					},
				}
				log.Printf("Previous: %+v", previous)
			}

			if currentStudent, ok := mc.GetStudentByCounter(mc.GlobalCounter); ok {
				current = currentStudent
				currPayload = &entity.IndividualPayload{
					Type: "student name",
					Data: entity.StudentPayload{
						OrderOfReading: mc.GlobalCounter,
						Name:           current.FirstName + " " + current.LastName,
						Reading:        current.Notes,
						Certificate:    current.Certificate,
					},
				}
				log.Printf("Current: %+v", current)
			}

			if nextStudent, ok := mc.GetStudentByCounter(mc.GlobalCounter + 1); ok {
				next = nextStudent
				nextPayload = &entity.IndividualPayload{
					Type: "student name",
					Data: entity.StudentPayload{
						OrderOfReading: mc.GlobalCounter + 1,
						Name:           next.FirstName + " " + next.LastName,
						Reading:        next.Notes,
						Certificate:    next.Certificate,
					},
				}
				log.Printf("Next: %+v", next)
			}

			payload := entity.AnnouncePayload{
				Previous: prevPayload,
				Current:  currPayload,
				Next:     nextPayload,
			}

			jsonData, err := json.Marshal(payload)
			if err != nil {
				log.Fatal(err)
			}

			client.Publish("announce", 0, false, jsonData)

			//TODO: replace
			err = utility.WriteStringToFile("failsave.txt", strconv.Itoa(mc.GlobalCounter))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func OnHealthcheck(mc conx.Controller) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		messageString := string(msg.Payload())
		if messageString == "ok" {
			mc.MicrocontrollerAlive = true
		}
	}
}

func RegisterCallBacks(c mqtt.Client, mc conx.Controller) {
	token := c.Subscribe("signal", 0, OnSignal(mc))
	token.Wait()
	token = c.Subscribe("healthcheck", 0, OnHealthcheck(mc))
	token.Wait()
}

package mqttx

import (
	conx "capstone/server/controller"
	"capstone/server/entity"
	"capstone/server/utility/config"
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func NewMqttx(conf *config.Config) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", conf.MQTT_server, conf.MQTT_port))
	opts.SetClientID("go-mqtt-clientx")
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

func OnSignal(mc *conx.Controller) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		messageString := string(msg.Payload())
		log.Printf("Received message: %s from topic: %s\n", messageString, msg.Topic())
		if messageString == "1" {
			mc.IncrementGlobalCounter()
			log.Printf("Counter: %d", mc.GlobalCounter)
			payload := entity.AnnounceMQTTPayload{}
			var data entity.IndividualPayload
			index := mc.GlobalCounter + 3
			if index >= 0 && index < len(mc.Script) {
				data = mc.Script[index]
			}

			if data.Type == "student name" {
				payload = entity.AnnounceMQTTPayload{
					Revert:        false,
					CurrentNumber: data.Data.(entity.StudentPayload).Order,
					MaxNumber:     data.Data.(entity.StudentPayload).FacultyMax,
					Session:       data.Data.(entity.StudentPayload).Session,
					Faculty:       data.Data.(entity.StudentPayload).Faculty,
					Block: entity.IndividualPayload{
						Type: "student name",
						Data: data.Data,
					},
				}
			} else if data.Type == "script" {
				payload = entity.AnnounceMQTTPayload{
					Revert:  true,
					Session: data.Data.(entity.AnnouncerPayload).Session,
					Faculty: data.Data.(entity.AnnouncerPayload).Faculty,
					Block: entity.IndividualPayload{
						Type: "script",
						Data: data.Data,
					},
				}
			}

			payload.Mode = mc.Mode

			jsonData, err := json.Marshal(payload)
			if err != nil {
				log.Fatal(err)
			}

			client.Publish("announce", 2, false, jsonData)

			//TODO: fail save
		} else if messageString == "2" {
			mc.DecrementGlobalCounter()
			log.Printf("Counter: %d", mc.GlobalCounter)
			payload := entity.AnnounceMQTTPayload{}
			var data entity.IndividualPayload
			index := mc.GlobalCounter + 3
			if index >= 0 && index < len(mc.Script) {
				data = mc.Script[index]
			}

			if data.Type == "student name" {
				payload = entity.AnnounceMQTTPayload{
					Revert:  false,
					Session: data.Data.(entity.StudentPayload).Session,
					Faculty: data.Data.(entity.StudentPayload).Faculty,
					Block: entity.IndividualPayload{
						Type: "student name",
						Data: data.Data,
					},
				}
			} else if data.Type == "script" {
				payload = entity.AnnounceMQTTPayload{
					Revert:  true,
					Session: data.Data.(entity.AnnouncerPayload).Session,
					Faculty: data.Data.(entity.AnnouncerPayload).Faculty,
					Block: entity.IndividualPayload{
						Type: "script",
						Data: data.Data,
					},
				}
			}

			payload.Mode = mc.Mode

			jsonData, err := json.Marshal(payload)
			if err != nil {
				log.Fatal(err)
			}

			client.Publish("announce", 2, false, jsonData)
		} else if messageString == "3" {
			mc.IncrementGlobalCounter()
			mc.IncrementGlobalCounter()
			log.Printf("Counter: %d", mc.GlobalCounter)
			payload := entity.AnnounceMQTTPayload{}
			var data entity.IndividualPayload
			index := mc.GlobalCounter + 3
			if index >= 0 && index < len(mc.Script) {
				data = mc.Script[index]
			}

			if data.Type == "student name" {
				payload = entity.AnnounceMQTTPayload{
					Revert:  false,
					Session: data.Data.(entity.StudentPayload).Session,
					Faculty: data.Data.(entity.StudentPayload).Faculty,
					Block: entity.IndividualPayload{
						Type: "student name",
						Data: data.Data,
					},
				}
			} else if data.Type == "script" {
				payload = entity.AnnounceMQTTPayload{
					Revert:  true,
					Session: data.Data.(entity.AnnouncerPayload).Session,
					Faculty: data.Data.(entity.AnnouncerPayload).Faculty,
					Block: entity.IndividualPayload{
						Type: "script",
						Data: data.Data,
					},
				}
			}

			payload.Mode = mc.Mode

			jsonData, err := json.Marshal(payload)
			if err != nil {
				log.Fatal(err)
			}

			client.Publish("announce", 2, false, jsonData)
		}
	}
}

func OnHealthcheck(mc *conx.Controller) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		messageString := string(msg.Payload())
		log.Println(messageString)
		if messageString == "ok" {
			mc.MicrocontrollerAlive = true
		}
	}
}

func RegisterCallBacks(c mqtt.Client, mc *conx.Controller) {
	token := c.Subscribe("signal", 0, OnSignal(mc))
	token.Wait()
	token = c.Subscribe("healthcheck", 0, OnHealthcheck(mc))
	token.Wait()
}

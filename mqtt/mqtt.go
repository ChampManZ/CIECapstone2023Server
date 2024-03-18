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
		//Normal case (Forward)
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
					Action:  "increase",
					Session: data.Data.(entity.StudentPayload).Session,
					Faculty: data.Data.(entity.StudentPayload).Faculty,
				}
			} else if data.Type == "script" {
				payload = entity.AnnounceMQTTPayload{
					Action:  "increase",
					Session: data.Data.(entity.AnnouncerPayload).Session,
					Faculty: data.Data.(entity.AnnouncerPayload).Faculty,
				}
			}

			for _, v := range mc.Script[mc.GlobalCounter:] {
				if v.Type == "student name" {
					payload.CurrentNumber = v.Data.(entity.StudentPayload).Order
					payload.MaxNumber = v.Data.(entity.StudentPayload).FacultyMax
					break
				}
				log.Printf("Script Data: %v", payload.Block.Data)
			}

			payload.Mode = mc.Mode

			log.Printf("Payload: %v", payload)

			jsonData, err := json.Marshal(payload)
			if err != nil {
				log.Fatal(err)
			}

			response := mc.PrepareDashboardMQTT()
			jsonData2, err := json.Marshal(response)
			if err != nil {
				log.Fatal(err)
			}

			client.Publish("announce", 2, false, jsonData)
			client.Publish("dashboard", 2, false, jsonData2)

			if mc.Script[mc.GlobalCounter].Type == "script" && !mc.Paused {
				mc.TogglePause()
				log.Println("MQTT: pause publishing.")
			} else if mc.Script[mc.GlobalCounter].Type == "student name" && mc.Paused {
				mc.TogglePause()
				log.Println("MQTT: resume publishing.")
			}

			//TODO: fail save
			//Reverse Case
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
					Action:        "decrease",
					CurrentNumber: data.Data.(entity.StudentPayload).Order,
					MaxNumber:     data.Data.(entity.StudentPayload).FacultyMax,
					Session:       data.Data.(entity.StudentPayload).Session,
					Faculty:       data.Data.(entity.StudentPayload).Faculty,
				}
			} else if data.Type == "script" {
				payload = entity.AnnounceMQTTPayload{
					Action:  "decrease",
					Session: data.Data.(entity.AnnouncerPayload).Session,
					Faculty: data.Data.(entity.AnnouncerPayload).Faculty,
				}

			}
			for _, v := range mc.Script[mc.GlobalCounter:] {
				if v.Type == "student name" {
					payload.CurrentNumber = v.Data.(entity.StudentPayload).Order
					payload.MaxNumber = v.Data.(entity.StudentPayload).FacultyMax
					break
				}
			}

			payload.Mode = mc.Mode

			jsonData, err := json.Marshal(payload)
			if err != nil {
				log.Fatal(err)
			}

			response := mc.PrepareDashboardMQTT()
			jsonData2, err := json.Marshal(response)
			if err != nil {
				log.Fatal(err)
			}

			if mc.Script[mc.GlobalCounter].Type == "script" && !mc.Paused {
				mc.TogglePause()
				log.Println("MQTT: pause publishing.")
			} else if mc.Script[mc.GlobalCounter].Type == "student name" && mc.Paused {
				mc.TogglePause()
				log.Println("MQTT: resume publishing.")
			}

			client.Publish("announce", 2, false, jsonData)
			client.Publish("dashboard", 2, false, jsonData2)
			//Skip Case
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
					Action:        "increase",
					CurrentNumber: data.Data.(entity.StudentPayload).Order,
					MaxNumber:     data.Data.(entity.StudentPayload).FacultyMax,
					Session:       data.Data.(entity.StudentPayload).Session,
					Faculty:       data.Data.(entity.StudentPayload).Faculty,
				}
			} else if data.Type == "script" {
				payload = entity.AnnounceMQTTPayload{
					Action:  "increase",
					Session: data.Data.(entity.AnnouncerPayload).Session,
					Faculty: data.Data.(entity.AnnouncerPayload).Faculty,
				}

			}
			for _, v := range mc.Script[mc.GlobalCounter:] {
				if v.Type == "student name" {
					payload.CurrentNumber = v.Data.(entity.StudentPayload).Order
					payload.MaxNumber = v.Data.(entity.StudentPayload).FacultyMax
					break
				}
			}

			payload.Mode = mc.Mode

			jsonData, err := json.Marshal(payload)
			if err != nil {
				log.Fatal(err)
			}

			response := mc.PrepareDashboardMQTT()
			jsonData2, err := json.Marshal(response)
			if err != nil {
				log.Fatal(err)
			}

			client.Publish("announce", 2, false, jsonData)
			client.Publish("dashboard", 2, false, jsonData2)
		} else if messageString == "4" {
			log.Printf("Counter: %d", mc.GlobalCounter)
			payload := entity.AnnounceMQTTPayload{
				Action: "reset",
			}

			jsonData, err := json.Marshal(payload)
			if err != nil {
				log.Fatal(err)
			}

			if mc.Script[mc.GlobalCounter].Type == "script" && !mc.Paused {
				mc.TogglePause()
				log.Println("MQTT: pause publishing.")
			} else if mc.Script[mc.GlobalCounter].Type == "student name" && mc.Paused {
				mc.TogglePause()
				log.Println("MQTT: resume publishing.")
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

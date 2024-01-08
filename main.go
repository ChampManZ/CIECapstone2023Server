package main

import (
	"capstone/server/entity"
	"capstone/server/handlers"
	"capstone/server/utility"
	"capstone/server/utility/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/labstack/echo/v4"
)

var Counter int
var MicrocontrollerAlive bool = false
var StudentList map[int]entity.Student

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected to MQTT Broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

var onSignal mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	messageString := string(msg.Payload())
	if messageString == "1" {
		Counter++
		var previous, current, next *entity.Student = nil, nil, nil
		if prevStudent, ok := StudentList[Counter-1]; ok {
			previous = &prevStudent
		}

		if currentStudent, ok := StudentList[Counter]; ok {
			current = &currentStudent
		}

		if nextStudent, ok := StudentList[Counter+1]; ok {
			next = &nextStudent
		}

		payload := entity.Payload{
			Previous: previous,
			Current:  current,
			Next:     next,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Fatal(err)
		}

		client.Publish("announce", 0, false, jsonData)

		file, err := os.Open("failsave.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		file.WriteString(strconv.Itoa(Counter))
	}
}

var onHealthcheck mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	messageString := string(msg.Payload())
	if messageString == "ok" {
		MicrocontrollerAlive = true
	}
}

func main() {
	config.Setup()
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", config.GlobalConfig.MQTT_server, config.GlobalConfig.MQTT_port))
	opts.SetClientID("go-mqtt-client")
	opts.SetUsername(config.GlobalConfig.MQTT_username)
	opts.SetPassword(config.GlobalConfig.MQTT_password)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	token := client.Subscribe("signal", 0, onSignal)
	token.Wait()
	token = client.Subscribe("healthcheck", 0, onHealthcheck)
	token.Wait()

	client.Publish("healthcheck", 0, false, "CHK")
	utility.CheckMicrocontrollerHealth(client, &MicrocontrollerAlive)

	if _, err := os.Stat("failsave.txt"); os.IsNotExist(err) {
		//get backup
		err = utility.DownloadFile(config.GlobalConfig.Download_URL, "failsave.txt")
		if err != nil {
			log.Printf("Error downloading file: %s", err)

			//recover logic
			file, err := os.Create("failsave.txt")
			if err != nil {
				log.Fatal(err)
			}
			file.WriteString(strconv.Itoa(Counter))
			file.Close()
		}
		Counter = 0
	} else {
		file, err := os.Open("failsave.txt")
		if err != nil {
			log.Fatal(err)
		}
		byteValue, _ := io.ReadAll(file)
		Counter, err = strconv.Atoi(string(byteValue))
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}

	//init echo
	e := echo.New()
	handlers.RegisterRoutes(e)
	// Start server
	go func() {
		if err := e.Start(":8443"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("Shutting down the server")
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	client.Disconnect(250)
}

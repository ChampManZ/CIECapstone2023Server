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

var MainController = handlers.NewController()

// TODO: move all these to mqtt/mqtt.go
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
	log.Printf("Received message: %s from topic: %s\n", messageString, msg.Topic())
	if messageString == "1" {
		MainController.GlobalCounter++
		log.Printf("Counter: %d", MainController.GlobalCounter)
		var previous, current, next *entity.Student = nil, nil, nil
		var prevPayload, currPayload, nextPayload *entity.IndividualPayload = nil, nil, nil
		if prevStudent, ok := MainController.StudentList[MainController.GlobalCounter-1]; ok {
			previous = &prevStudent
			prevPayload = &entity.IndividualPayload{
				Type: "student name",
				Data: entity.StudentPayload{
					OrderOfReading: MainController.GlobalCounter - 1,
					Name:           previous.FirstName + " " + previous.LastName,
					Reading:        previous.Certificate,
					Note:           previous.Notes,
					Certificate:    previous.Certificate,
				},
			}
			log.Printf("Previous: %+v", previous)
		}

		if currentStudent, ok := MainController.StudentList[MainController.GlobalCounter]; ok {
			current = &currentStudent
			currPayload = &entity.IndividualPayload{
				Type: "student name",
				Data: entity.StudentPayload{
					OrderOfReading: MainController.GlobalCounter,
					Name:           current.FirstName + " " + current.LastName,
					Reading:        current.Certificate,
					Note:           current.Notes,
					Certificate:    current.Certificate,
				},
			}
			log.Printf("Current: %+v", current)
		}

		if nextStudent, ok := MainController.StudentList[MainController.GlobalCounter+1]; ok {
			next = &nextStudent
			nextPayload = &entity.IndividualPayload{
				Type: "student name",
				Data: entity.StudentPayload{
					OrderOfReading: MainController.GlobalCounter + 1,
					Name:           next.FirstName + " " + next.LastName,
					Reading:        next.Certificate,
					Note:           next.Notes,
					Certificate:    next.Certificate,
				},
			}
			log.Printf("Next: %+v", next)
		}

		payload := entity.Payload{
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
		err = utility.WriteStringToFile("failsave.txt", strconv.Itoa(MainController.GlobalCounter))
		if err != nil {
			log.Fatal(err)
		}
	}
}

var onHealthcheck mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	messageString := string(msg.Payload())
	if messageString == "ok" {
		MainController.MicrocontrollerAlive = true
	}
}

func main() {
	config.Setup()
	//TODO: move all these to mqtt/mqtt.go
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
	utility.CheckMicrocontrollerHealth(client, &MainController.MicrocontrollerAlive)

	//TO DO: Replace with MySQL
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
			file.WriteString(strconv.Itoa(MainController.GlobalCounter))
			file.Close()
		}
		MainController.GlobalCounter = 0
	} else {
		file, err := os.Open("failsave.txt")
		if err != nil {
			log.Fatal(err)
		}
		byteValue, _ := io.ReadAll(file)
		MainController.GlobalCounter, err = strconv.Atoi(string(byteValue))
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}

	MySQL_DNS := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s", config.GlobalConfig.MySQL_username,
		config.GlobalConfig.MySQL_password, config.GlobalConfig.MySQL_host, config.GlobalConfig.MySQL_port, config.GlobalConfig.MySQL_dbname)
	var err error
	MainController.MySQLConn, err = utility.NewMySQLConn(MySQL_DNS)
	if err != nil {
		log.Fatal(err)
	}

	//TO DO: replace
	err = utility.ReadCSVIntoMap("result.csv", &MainController.StudentList)

	if err != nil {
		log.Fatal(err)
	}

	//init echo
	e := echo.New()
	MainController.RegisterRoutes(e)
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

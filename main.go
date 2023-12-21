package main

import (
	"capstone/server/handlers"
	"capstone/server/utility/config"
	"context"
	"fmt"
	"io/ioutil"
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

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	config.Setup()
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", config.GlobalConfig.MQTT_server, config.GlobalConfig.MQTT_port))
	opts.SetClientID("go-mqtt-client")
	opts.SetUsername("")
	opts.SetPassword("")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if _, err := os.Stat("failsave.txt"); os.IsNotExist(err) {
		Counter = 0
		file, err := os.Create("failsave.txt")
		if err != nil {
			log.Fatal(err)
		}
		file.WriteString(strconv.Itoa(Counter))
		file.Close()
	} else {
		file, err := os.Open("failsave.txt")
		if err != nil {
			log.Fatal(err)
		}
		byteValue, _ := ioutil.ReadAll(file)
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

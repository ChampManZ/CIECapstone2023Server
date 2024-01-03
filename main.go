package main

import (
	"capstone/server/handlers"
	"capstone/server/utility"
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

var onSignal mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	messageString := string(msg.Payload())
	switch messageString {
	case "1":
		Counter += 1
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

	//get backup
	utility.DownloadFile(config.GlobalConfig.Download_URL, "failsave_backup.txt")

	if _, err := os.Stat("failsave.txt"); os.IsNotExist(err) {
		Counter = 0
		file, err := os.Create("failsave.txt")
		if err != nil {
			log.Fatal(err)
		}
		file.WriteString(strconv.Itoa(Counter))
		file.Close()
	} else {
		//checksum
		checksumLocal, err := utility.CalculateChecksum("failsave.txt")
		if err != nil {
			log.Printf("Error checksum failed: %v", err)
		}

		checksumDownloaded, err := utility.CalculateChecksum("failsave_backup.txt")
		if err != nil {
			log.Printf("Error checksum failed: %v", err)
		}

		if checksumLocal == checksumDownloaded {
			fmt.Println("Checksum matched")
		}

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

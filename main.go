package main

import (
	conx "capstone/server/controller"
	"capstone/server/handlers"
	mqttx "capstone/server/mqtt"
	"capstone/server/utility"
	"capstone/server/utility/config"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config.Setup()
	var MainController = conx.NewController()

	client := mqttx.NewMqttx(config.GlobalConfig)
	client.Publish("healthcheck", 0, false, "CHK")
	utility.CheckMicrocontrollerHealth(client, &MainController.MicrocontrollerAlive)

	MySQL_DNS := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.GlobalConfig.MySQL_username,
		config.GlobalConfig.MySQL_password, config.GlobalConfig.MySQL_host, config.GlobalConfig.MySQL_port, config.GlobalConfig.MySQL_dbname)

	var err error
	MainController.MySQLConn, err = utility.NewMySQLConn(MySQL_DNS)
	if err != nil {
		log.Fatal(err)
	}

	MainController.GlobalCounter = MainController.MySQLConn.QueryCounter()
	MainController.StudentList, err = MainController.MySQLConn.QueryStudentsToMap()
	if err != nil {
		log.Fatal(err)
	}

	//init echo
	hl := handlers.NewHandlers(MainController)
	hl.RegisterRoutes(hl.Echo)

	// Start server
	go func() {
		if err := hl.Echo.Start(":8443"); err != nil && err != http.ErrServerClosed {
			hl.Echo.Logger.Fatal("Shutting down the server")
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := hl.Echo.Shutdown(ctx); err != nil {
		hl.Echo.Logger.Fatal(err)
	}

	client.Disconnect(250)
}

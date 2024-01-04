package main

import (
	"capstone/server/handlers"
	"capstone/server/utility/config"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type Student struct {
	StudentID   string
	FirstName   string
	LastName    string
	Certificate string
}

type StudentNote struct {
	StudentID   string
	FirstName   string
	LastName    string
	Certificate string
	Note        string
}

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

func readStudentInfo(filePath string) (map[string]Student, error) {
	studentData := make(map[string]Student)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		student := Student{
			StudentID:   record[0],
			FirstName:   record[1],
			LastName:    record[2],
			Certificate: record[3],
		}

		studentData[student.StudentID] = student
	}

	return studentData, nil
}

func queryStudentNote(studentID string) (string, error) {
	db, err := sqlx.Connect("mysql", "root:Sammax20011558_@tcp(127.0.0.1:3306)/ciecapstone2023")
	if err != nil {
		return "", err
	}
	defer db.Close()

	var note string
	err = db.Get(&note, "SELECT notes FROM student_notes WHERE studentID = ?", studentID)
	if err != nil {
		return "", err
	}

	return note, nil
}

func updateStudentNote(studentID, newNote string) error {
	db, err := sqlx.Connect("mysql", "root:Sammax20011558_@tcp(127.0.0.1:3306)/ciecapstone2023")
	if err != nil {
		return err
	}
	defer db.Close()

	query := "UPDATE student_notes SET notes = ? WHERE studentID = ?"

	_, err = db.Exec(query, newNote, studentID)
	if err != nil {
		return err
	}

	fmt.Printf("Note updated successfully for studentID: %s\n", studentID)
	return nil
}

func updateStudentNoteHandler(c echo.Context) error {
	studentID := c.Param("studentID")
	newNote := c.Param("newNote")

	err := updateStudentNote(studentID, newNote)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Note updated successfully"})
}

func getStudentNotesHandler(c echo.Context) error {
	studentNoteList, err := getStudentNotes()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Return the joined data as JSON response
	return c.JSON(http.StatusOK, studentNoteList)
}

func getStudentNotes() ([]StudentNote, error) {
	csvFilePath := `D:\Thanapat Work\CIE 4th Year\Capstone Project\Server\CIECapstone2023Server\student_list\student_list.csv`

	csvData, err := readStudentInfo(csvFilePath)
	if err != nil {
		log.Fatal(err)
	}

	var studentNoteList []StudentNote

	for studentID, student := range csvData {
		// Query student note
		note, err := queryStudentNote(studentID)
		if err != nil {
			log.Printf("Error querying student note: %v\n", err)
			continue
		}

		joinedData := StudentNote{
			StudentID:   student.StudentID,
			FirstName:   student.FirstName,
			LastName:    student.LastName,
			Certificate: student.Certificate,
			Note:        note,
		}

		// Append to the slice
		studentNoteList = append(studentNoteList, joinedData)
	}

	// Print or do further processing with the joined data
	for _, data := range studentNoteList {
		fmt.Printf("StudentID: %s, FirstName: %s, LastName: %s, Certificate: %s, Note: %s\n",
			data.StudentID, data.FirstName, data.LastName, data.Certificate, data.Note)
	}

	return studentNoteList, nil
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

	// Performance Test
	startTime := time.Now()

	//init echo
	e := echo.New()
	handlers.RegisterRoutes(e)

	// Register routes
	e.PUT("/updateStudentNote/:studentID/:newNote", updateStudentNoteHandler)
	e.GET("/getStudentNotes", getStudentNotesHandler)

	// Start server
	go func() {
		if err := e.Start(":8443"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("Shutting down the server")
		}
	}()

	elapsed := time.Since(startTime)
	fmt.Printf("Time taken for the join operation: %s\n", elapsed)

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

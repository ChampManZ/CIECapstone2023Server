package main

import (
	"capstone/server/handlers"
	// "capstone/server/utility/config"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
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

func queryStudentNoteAndSaveToCSV(csvFilePath, outputFilePath string) error {
	csvData, err := readStudentInfo(csvFilePath)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close() // Ensure the file is closed after writing

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{"StudentID", "FirstName", "LastName", "Certificate", "Note"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for studentID, student := range csvData {
		note, err := queryStudentNote(studentID)
		if err != nil {
			log.Printf("Error querying student note: %v\n", err)
			continue
		}

		studentNote := StudentNote{
			StudentID:   student.StudentID,
			FirstName:   student.FirstName,
			LastName:    student.LastName,
			Certificate: student.Certificate,
			Note:        note,
		}

		// Write studentNote to CSV
		record := []string{
			studentNote.StudentID,
			studentNote.FirstName,
			studentNote.LastName,
			studentNote.Certificate,
			studentNote.Note,
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func saveStudentNoteToCSV(studentNote StudentNote, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{
		studentNote.StudentID,
		studentNote.FirstName,
		studentNote.LastName,
		studentNote.Certificate,
		studentNote.Note,
	}

	if err := writer.Write(record); err != nil {
		return err
	}

	return nil
}

func queryStudentNote(studentID string) (string, error) {
	db, err := sqlx.Connect("mysql", "root:Sammax20011558_@tcp(127.0.0.1:3306)/ciecapstone2023")
	if err != nil {
		return "", err
	}
	defer db.Close()

	var note string
	err = db.Get(&note, "SELECT notes FROM student_notes_big WHERE studentID = ?", studentID)
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
	csvFilePath := `/Users/champthanapat/Documents/CIE Capstone/CIECapstone2023Server/student_list/student_list_big.csv`

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

func readCSV(filePath string) (map[int][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	resultMap := make(map[int][]string)
	//skip header
	for _, record := range records[1:] {
		id, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, err
		}
		resultMap[id] = record[1:]
	}
	return resultMap, nil
}

func joinCSVs(file1, file2 string) (map[int]StudentNote, error) {
	data1, err := readCSV(file1)
	if err != nil {
		return nil, err
	}

	data2, err := readCSV(file2)
	if err != nil {
		return nil, err
	}

	result := make(map[int]StudentNote)
	for id, values := range data1 {
		if notes, ok := data2[id]; ok {

			idStr := strconv.Itoa(id)

			result[id] = StudentNote{
				StudentID:   idStr,
				FirstName:   values[0],
				LastName:    values[1],
				Certificate: values[2],
				Note:        strings.Join(notes, ", "),
			}
		}
	}
	return result, nil
}

func saveToCSV(students map[int]StudentNote, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Headers, not necessary
	header := []string{"StudentID", "FirstName", "LastName", "Certificate", "Notes"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Sort ID, also not necessary
	var ids []int
	for id := range students {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		student := students[id]
		record := []string{
			student.StudentID,
			student.FirstName,
			student.LastName,
			student.Certificate,
			student.Note,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// config.Setup()
	// opts := mqtt.NewClientOptions()
	// opts.AddBroker(fmt.Sprintf("tcp://%s:%d", config.GlobalConfig.MQTT_server, config.GlobalConfig.MQTT_port))
	// opts.SetClientID("go-mqtt-client")
	// opts.SetUsername("")
	// opts.SetPassword("")
	// opts.SetDefaultPublishHandler(messagePubHandler)
	// opts.OnConnect = connectHandler
	// opts.OnConnectionLost = connectLostHandler
	// client := mqtt.NewClient(opts)
	// if token := client.Connect(); token.Wait() && token.Error() != nil {
	// 	panic(token.Error())
	// }

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
	var mStart runtime.MemStats
	runtime.ReadMemStats(&mStart)
	csvFilePath := `/Users/champthanapat/Documents/CIE Capstone/CIECapstone2023Server/student_list/student_list_big.csv`
	// outputFilePath := `/Users/champthanapat/Documents/CIE Capstone/CIECapstone2023Server/student_list/student_notes_big.csv`

	// if err := queryStudentNoteAndSaveToCSV(csvFilePath, outputFilePath); err != nil {
	// 	log.Fatal(err)
	// }

	studentsTest, err := joinCSVs(csvFilePath, `/Users/champthanapat/Documents/CIE Capstone/CIECapstone2023Server/student_list/student_notes_big.csv`)
	if err != nil {
		fmt.Println(err)
		return
	}

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
	var mEnd runtime.MemStats
	runtime.ReadMemStats(&mEnd)

	// Calculate memory usage
	memoryUsed := mEnd.Alloc - mStart.Alloc // bytes

	//mem used
	fmt.Printf("\nResource usages: \n")
	fmt.Printf("Time taken: %s\n", elapsed)
	fmt.Printf("Memory used: %.2f MB\n", float64(memoryUsed)/1024/1024)

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	saveToCSV(studentsTest, "result.csv")
	// client.Disconnect(250)
}

package utility

import (
	"capstone/server/entity"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

func ReadCSV(filePath string) (map[int][]string, error) {
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

func JoinCSVs(file1, file2 string) (map[int]entity.Student, error) {
	data1, err := ReadCSV(file1)
	if err != nil {
		return nil, err
	}

	data2, err := ReadCSV(file2)
	if err != nil {
		return nil, err
	}

	result := make(map[int]entity.Student)
	for id, values := range data1 {
		if notes, ok := data2[id]; ok {
			result[id] = entity.Student{
				StudentID:   id,
				FirstName:   values[0],
				LastName:    values[1],
				Certificate: values[2],
				Reading:     strings.Join(notes, ", "),
			}
		}
	}
	return result, nil
}

func SaveToCSV(students map[int]entity.Student, filePath string) error {
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
			strconv.Itoa(student.StudentID),
			student.FirstName,
			student.LastName,
			student.Certificate,
			student.Reading,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// update field with safety check
func UpdateStudentField(students map[int]entity.Student, studentID int, field, newValue string) error {
	// Check if the studentID exists in the map
	student, exists := students[studentID]
	if !exists {
		return fmt.Errorf("student does not exist %d", studentID)
	}

	switch field {
	case "FirstName":
		student.FirstName = newValue
	case "LastName":
		student.LastName = newValue
	case "Certificate":
		student.Certificate = newValue
	case "Notes":
		student.Reading = newValue
	default:
		return fmt.Errorf("invalid field: %s", field)
	}

	students[studentID] = student
	return nil
}

func ReadCSVIntoMap(filePath string, studentsMap *map[int]entity.Student) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records := make(map[int]entity.Student)

	// drop header
	_, err = reader.Read()
	if err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		StudentID, err := strconv.Atoi(record[0])
		if err != nil {
			return fmt.Errorf("failed to parse ID for %v: %v", record, err)
		}

		records[StudentID] = entity.Student{
			StudentID:   StudentID,
			FirstName:   record[1],
			LastName:    record[2],
			Certificate: record[3],
			Reading:     record[4],
		}
	}

	*studentsMap = records

	return nil
}

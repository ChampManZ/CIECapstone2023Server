package utility

import (
	"capstone/server/entity"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"unicode"
	"unicode/utf8"
)

func CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func WriteStringToFile(filePath string, content string) error {
	file, err := os.OpenFile("failsave.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	file.Sync()
	return nil
}

func DownloadFile(url string, filepath string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Non 200 status code: %d", resp.StatusCode)
		return err
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// workaround bcz reg is a piece of shit
func convertAndUnmarshal(data []byte) ([]entity.StudentData, error) {
	var intermediate interface{}
	err := json.Unmarshal(data, &intermediate)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(intermediate).Kind() == reflect.Slice {
		slice, ok := intermediate.([]interface{})
		if !ok {
			return nil, fmt.Errorf("type assertion to slice of interface{} failed")
		}
		var modifiedSlice []map[string]interface{}

		for _, item := range slice {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("type assertion to map[string]interface{} failed")
			}
			modifiedMap := make(map[string]interface{})
			for key, value := range itemMap {
				switch v := value.(type) {
				case float64:
					modifiedMap[key] = fmt.Sprintf("%v", v)
				case bool:
					modifiedMap[key] = fmt.Sprintf("%t", v)
				default:
					modifiedMap[key] = fmt.Sprintf("%s", v)
				}
			}
			modifiedSlice = append(modifiedSlice, modifiedMap)
		}

		modifiedData, err := json.Marshal(modifiedSlice)
		if err != nil {
			return nil, err
		}

		var students []entity.StudentData
		err = json.Unmarshal(modifiedData, &students)
		if err != nil {
			return nil, err
		}

		return students, nil
	}

	return nil, fmt.Errorf("expected JSON array at the top level")
}

func FetchRegistraData(url string) ([]entity.StudentData, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data []entity.StudentData
	//err = json.Unmarshal(body, &data)
	data, err = convertAndUnmarshal(body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func IsFirstCharNotEnglish(s string) bool {
	if s == "" {
		return false
	}

	r, _ := utf8.DecodeRuneInString(s)
	return !unicode.IsLetter(r) || (r < 'A' || r > 'z' || (r > 'Z' && r < 'a'))
}

func IsNotInList(str string, list []string) bool {
	for _, s := range list {
		if s == str {
			return false
		}
	}
	return true
}

func FindStudentByOrder(students map[int]entity.Student, order int) int {
	for key, student := range students {
		if student.OrderOfReceive == order {
			return key
		}
	}
	return -1 // Not found
}

func AnnouncerAlreadyAdded(announcers []entity.AnnouncerGroupByFaculty, announcerID int) bool {
	for _, announcer := range announcers {
		if announcer.AnnouncerID == announcerID {
			return true
		}
	}
	return false
}

package utility

import (
	"capstone/server/entity"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

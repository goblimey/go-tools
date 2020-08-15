package testsupport

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
)

// Helper methods for tests.

// MakeUUID creates a UUID.  See https://yourbasic.org/golang/generate-uuid-guid/.
//
func MakeUUID() string {
	// This produces something like "9e0825f2-e557-28df-93b7-a01c789f36a8".
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

// CreateWorkingDirectory create a working directory and makes it the current
// directory.
//
func CreateWorkingDirectory() (string, error) {
	directoryName := "/tmp/" + MakeUUID()
	err := os.Mkdir(directoryName, os.ModePerm)
	if err != nil {
		return "", err
	}
	err = os.Chdir(directoryName)
	if err != nil {
		return "", err
	}
	return directoryName, nil
}

// RemoveWorkingDirectory removes the working directory and any files in it.
//
func RemoveWorkingDirectory(directoryName string) error {
	err := os.RemoveAll(directoryName)
	if err != nil {
		return err
	}
	return nil
}

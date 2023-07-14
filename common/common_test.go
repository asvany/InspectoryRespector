package common

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"testing"
)

func Test_InitEnv(t *testing.T) {
	InitEnv("../")

}

func Test_InitDISPLAY(t *testing.T) {
	InitEnv("../")

	InitDISPLAY()

}


func calculateMD5(filePath string,t *testing.T) (string) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		t.Error(err)
	}

	return fmt.Sprintf("%x", hash.Sum(result))
}

func Test_Copy(t *testing.T){
	InitEnv("../")

	temp_filename := "/tmp/common_test.go.bak"
	defer func() {
		os.Remove(temp_filename)
	}()


	err := Copy("common_test.go", temp_filename)
	if err != nil {
		t.Error(err)
	}

	if calculateMD5("common_test.go",t) != calculateMD5(temp_filename,t) {
		t.Error("MD5 not match")
	}

}

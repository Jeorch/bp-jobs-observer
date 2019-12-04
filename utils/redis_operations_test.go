package utils

import (
	"fmt"
	"github.com/PharbersDeveloper/bp-jobs-observer/tests"
	"testing"
)

func TestAddKey2Redis(t *testing.T) {

	tests.SetEnv()

	key := "HELLO"
	count, err := AddKey2Redis(key)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("count", count)
}

func TestCheckKeyExistInRedis(t *testing.T) {

	tests.SetEnv()

	key := "HELLO"
	exist, err := CheckKeyExistInRedis(key)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("exist=", exist)

}

package system

import (
	"fmt"
	"log"
)

func ErrLog(funcName string, args ...string) {
	log.Println(fmt.Sprintf("%s: %s", funcName, args))
}

func FatalLog(funcName string, args ...string) {
	log.Fatal(fmt.Sprintf("%s: %s", funcName, args))
}

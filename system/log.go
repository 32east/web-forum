package system

import (
	"fmt"
	"log"
)

func ErrLog(funcName string, args ...string) error {
	fmtPrint := fmt.Sprintf("%s: %s", funcName, args)
	log.Println(fmtPrint)
	return fmt.Errorf(fmtPrint)
}

func FatalLog(funcName string, args ...string) {
	log.Fatal(fmt.Sprintf("%s: %s", funcName, args))
}

package system

import (
	"os"
	"strings"
)

func RegisterEnvironment() {
	var file, err = os.ReadFile(".env")

	if err != nil {
		panic(err)
	}

	var envRead = string(file)

	for _, str := range strings.Fields(envRead) {
		var indexOfSeparator = strings.Index(str, "=")
		var key = str[:indexOfSeparator]
		var value = str[indexOfSeparator+1:]
		var envErr = os.Setenv(key, value)

		if envErr != nil {
			panic(envErr)
		}
	}
}

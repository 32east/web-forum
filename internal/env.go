package internal

import (
	"os"
	"strings"
)

func RegisterEnvironment() {
	file, err := os.ReadFile(".env")

	if err != nil {
		panic(err)
	}

	envRead := string(file)

	for _, str := range strings.Fields(envRead) {
		indexOfSeparator := strings.Index(str, "=")
		key := str[:indexOfSeparator]
		value := str[indexOfSeparator+1:]

		err := os.Setenv(key, value)

		if err != nil {
			panic(err)
		}
	}
}

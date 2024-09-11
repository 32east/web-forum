package internal

import (
	"strings"
	"unicode/utf8"
)

func Utf8Length(str string) int {
	convertToByte := []byte(str)
	utf8count := 0

	for len(convertToByte) > 0 {
		_, size := utf8.DecodeRune(convertToByte)
		utf8count += 1

		convertToByte = convertToByte[size:]
	}

	return utf8count
}

func FormatString(str string) string {
	str = strings.Replace(str, "\r", "", -1)

	for {
		findTabSpaces := strings.Index(str, "\n\n\n")

		if findTabSpaces == -1 {
			break
		}

		str = strings.Replace(str, "\n\n\n", "\n\n", -1)
	}

	return str
}

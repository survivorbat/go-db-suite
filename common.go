package dbsuite

import "strings"

const maxDBNameLength = 64

func toSafeDBName(in string) string {
	output := make([]rune, 0, len(in))

	for _, char := range strings.ToLower(in) {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			output = append(output, char)
		}
	}

	if len(output) <= maxDBNameLength {
		return string(output)
	}

	return string(output[len(output)-maxDBNameLength:])
}

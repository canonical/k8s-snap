package utils

import (
	"regexp"
)

// YamlCommentLines adds "# " at the beginning of each line.
func YamlCommentLines(content []byte) []byte {
	re := regexp.MustCompile("(?m)^")
	out := re.ReplaceAll(content, []byte("# "))
	return out
}

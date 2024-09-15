package iofs

import (
	"bytes"
	"regexp"
	"strings"
)

var (
	splitMarkRe  = regexp.MustCompile("(?m)^---")
	rmCommentsRe = regexp.MustCompile("(?m)^#.*$")
)

func splitYAML(data []byte) []string {
	// Sanitize.
	data = bytes.TrimSpace(data)
	data = rmCommentsRe.ReplaceAll(data, []byte(""))

	// Split (YAML can declare multiple files in the same file using `---`).
	dataSplit := splitMarkRe.Split(string(data), -1)

	// Remove empty splits.
	nonEmptyData := []string{}
	for _, d := range dataSplit {
		d = strings.TrimSpace(d)
		if d != "" {
			nonEmptyData = append(nonEmptyData, d)
		}
	}

	return nonEmptyData
}

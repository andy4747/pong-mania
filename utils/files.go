package utils

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func FileExtension(filename string) string {
	var extension string = ""
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		extension = parts[len(parts)-1]
	}
	return extension
}

func NewProfileFileName(extenstion string) string {
	uid := uuid.NewString()
	filename := fmt.Sprintf("%s.%s", uid, extenstion)
	return filename
}

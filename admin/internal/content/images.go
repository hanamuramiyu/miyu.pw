package content

import (
	"strings"
)

func ValidateImagePath(workID, path string) error {
	if path == "" {
		return invalid("path", "is required")
	}
	if strings.ContainsAny(path, "\\?#") {
		return invalid("path", "must not contain \\, ? or #")
	}
	if strings.Contains(path, "..") {
		return invalid("path", "must not contain ..")
	}
	const prefix = "/media/"
	if !strings.HasPrefix(path, prefix) {
		return invalid("path", "must start with /media/")
	}
	if !strings.HasSuffix(path, ".webp") {
		return invalid("path", "must be a .webp file")
	}

	rest := strings.TrimPrefix(path, prefix)
	if rest == "" || strings.HasPrefix(rest, "/") {
		return invalid("path", "must name a file")
	}

	if workID == "" {
		return nil
	}

	workPrefix := workID + "/"
	if !strings.HasPrefix(rest, workPrefix) {
		return invalid("path", "must belong to this work")
	}
	name := strings.TrimPrefix(rest, workPrefix)
	if name == "" || strings.Contains(name, "/") {
		return invalid("path", "must be a file in the work media directory")
	}
	return nil
}

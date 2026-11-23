package content

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	dateLayout     = "2006-01-02"
	fullURLMessage = "expected a full http(s) URL with a host"
)

var (
	slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	hexPattern  = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
)

func validateWork(w *Work) error {
	if w == nil {
		return invalid("work", "is required")
	}

	w.Title = strings.TrimSpace(w.Title)
	w.Summary = strings.TrimSpace(w.Summary)

	if w.Tools == nil {
		w.Tools = []string{}
	}
	if w.Links == nil {
		w.Links = []Link{}
	}
	if w.Colors == nil {
		w.Colors = []Color{}
	}
	if w.Fonts == nil {
		w.Fonts = []Font{}
	}
	if w.Images == nil {
		w.Images = []Image{}
	}

	if !slugPattern.MatchString(w.Slug) {
		return invalid("slug", "must contain lowercase letters, digits and single hyphens")
	}
	if w.Title == "" {
		return invalid("title", "is required")
	}
	if w.Summary == "" {
		return invalid("summary", "is required")
	}
	if err := validateDate("date", w.Date); err != nil {
		return err
	}
	if err := validateDate("updated", w.Updated); err != nil {
		return err
	}

	for i, link := range w.Links {
		if err := validateHref(link.Href); err != nil {
			return invalid(fmt.Sprintf("links[%d].href", i), err.Error())
		}
	}
	for i, color := range w.Colors {
		if !hexPattern.MatchString(color.Hex) {
			return invalid(fmt.Sprintf("colors[%d].hex", i), "must be a 3- or 6-digit hex colour starting with #")
		}
	}
	for i, image := range w.Images {
		if err := ValidateImagePath("", image.Path); err != nil {
			var validation *ValidationError
			if errors.As(err, &validation) {
				return invalid(fmt.Sprintf("images[%d].path", i), validation.Message)
			}
			return invalid(fmt.Sprintf("images[%d].path", i), "is invalid")
		}
	}

	return nil
}

func validateDate(field, value string) error {
	parsed, err := time.Parse(dateLayout, value)
	if err != nil || parsed.Format(dateLayout) != value {
		return invalid(field, "must be a real date in YYYY-MM-DD format")
	}
	return nil
}

func validateHref(value string) error {
	if strings.ContainsAny(value, " \t\r\n\v\f") {
		return errors.New(fullURLMessage)
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return errors.New(fullURLMessage)
	}
	scheme := strings.ToLower(parsed.Scheme)
	if (scheme != "http" && scheme != "https") || parsed.Host == "" || parsed.Hostname() == "" {
		return errors.New(fullURLMessage)
	}
	return nil
}

package content

import (
	"errors"
	"testing"
)

func validWork() *Work {
	return &Work{
		Slug:     "example-work",
		Title:    "Example work",
		Summary:  "A short summary",
		BodyMD:   "Body",
		Kind:     "Website",
		Audience: "For clients",
		Role:     "Development",
		Date:     "2025-01-02",
		Updated:  "2025-02-03",
		Tools:    []string{"Go"},
		Links:    []Link{{Label: "Site", Href: "https://example.com"}},
		Colors:   []Color{{Name: "Ink", Hex: "#141211"}},
		Fonts:    []Font{{Name: "Manrope", Role: "Body"}},
		Images:   []Image{{Path: "/media/a.webp", Alt: "A", Width: 800, Height: 600}},
	}
}

func TestValidateWork(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Work)
		field  string
	}{
		{name: "valid", mutate: func(*Work) {}},
		{name: "empty slug", mutate: func(w *Work) { w.Slug = "" }, field: "slug"},
		{name: "uppercase slug", mutate: func(w *Work) { w.Slug = "Example" }, field: "slug"},
		{name: "space in slug", mutate: func(w *Work) { w.Slug = "example work" }, field: "slug"},
		{name: "double hyphen", mutate: func(w *Work) { w.Slug = "example--work" }, field: "slug"},
		{name: "leading hyphen", mutate: func(w *Work) { w.Slug = "-example" }, field: "slug"},
		{name: "trailing hyphen", mutate: func(w *Work) { w.Slug = "example-" }, field: "slug"},
		{name: "underscore", mutate: func(w *Work) { w.Slug = "example_work" }, field: "slug"},
		{name: "empty title", mutate: func(w *Work) { w.Title = "" }, field: "title"},
		{name: "blank title", mutate: func(w *Work) { w.Title = "   " }, field: "title"},
		{name: "empty summary", mutate: func(w *Work) { w.Summary = "" }, field: "summary"},
		{name: "blank summary", mutate: func(w *Work) { w.Summary = " \t\n " }, field: "summary"},
		{name: "impossible date", mutate: func(w *Work) { w.Date = "2025-02-30" }, field: "date"},
		{name: "short date", mutate: func(w *Work) { w.Date = "2025-1-2" }, field: "date"},
		{name: "date with time", mutate: func(w *Work) { w.Date = "2025-01-02T00:00:00Z" }, field: "date"},
		{name: "empty date", mutate: func(w *Work) { w.Date = "" }, field: "date"},
		{name: "bad updated", mutate: func(w *Work) { w.Updated = "2025-13-01" }, field: "updated"},
		{name: "short hex", mutate: func(w *Work) { w.Colors[0].Hex = "#12" }, field: "colors[0].hex"},
		{name: "long hex", mutate: func(w *Work) { w.Colors[0].Hex = "#1234567" }, field: "colors[0].hex"},
		{name: "hex without hash", mutate: func(w *Work) { w.Colors[0].Hex = "123456" }, field: "colors[0].hex"},
		{name: "hex with invalid digit", mutate: func(w *Work) { w.Colors[0].Hex = "#12345g" }, field: "colors[0].hex"},
		{name: "relative href", mutate: func(w *Work) { w.Links[0].Href = "/work" }, field: "links[0].href"},
		{name: "ftp href", mutate: func(w *Work) { w.Links[0].Href = "ftp://example.com" }, field: "links[0].href"},
		{name: "href without host", mutate: func(w *Work) { w.Links[0].Href = "https://" }, field: "links[0].href"},
		{name: "href with empty path host", mutate: func(w *Work) { w.Links[0].Href = "https:///x" }, field: "links[0].href"},
		{name: "empty http host", mutate: func(w *Work) { w.Links[0].Href = "http://" }, field: "links[0].href"},
		{name: "href with space", mutate: func(w *Work) { w.Links[0].Href = "https://exa mple.com" }, field: "links[0].href"},
		{name: "empty href", mutate: func(w *Work) { w.Links[0].Href = "" }, field: "links[0].href"},
		{name: "image traversal", mutate: func(w *Work) { w.Images[0].Path = "/media/../secret.webp" }, field: "images[0].path"},
		{name: "image without media prefix", mutate: func(w *Work) { w.Images[0].Path = "media/a.webp" }, field: "images[0].path"},
		{name: "image without webp extension", mutate: func(w *Work) { w.Images[0].Path = "/media/a.png" }, field: "images[0].path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			work := validWork()
			tt.mutate(work)

			err := validateWork(work)
			if tt.field == "" {
				if err != nil {
					t.Fatalf("validateWork() error = %v, want nil", err)
				}
				return
			}

			if !errors.Is(err, ErrValidation) {
				t.Fatalf("validateWork() error = %v, want ErrValidation", err)
			}
			var validation *ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("validateWork() error = %T, want *ValidationError", err)
			}
			if validation.Field != tt.field {
				t.Errorf("field = %q, want %q", validation.Field, tt.field)
			}
		})
	}
}

func TestValidateWorkAcceptsHexVariants(t *testing.T) {
	for _, value := range []string{"#abc", "#ABC", "#a1b2c3", "#A1B2C3"} {
		work := validWork()
		work.Colors[0].Hex = value
		if err := validateWork(work); err != nil {
			t.Errorf("hex %q rejected: %v", value, err)
		}
	}
}

func TestValidateWorkAcceptsFullHTTPURLs(t *testing.T) {
	for _, value := range []string{
		"https://example.com",
		"https://example.com/path?query=1#fragment",
		"http://example.com:8080/path",
		"HTTPS://EXAMPLE.COM",
	} {
		work := validWork()
		work.Links[0].Href = value
		if err := validateWork(work); err != nil {
			t.Errorf("href %q rejected: %v", value, err)
		}
	}
}

func TestValidateWorkRejectsIncompleteHTTPURLs(t *testing.T) {
	for _, value := range []string{
		"https://",
		"http://",
		"https:///x",
		"https://exa mple.com",
	} {
		work := validWork()
		work.Links[0].Href = value
		err := validateWork(work)
		var validation *ValidationError
		if !errors.As(err, &validation) {
			t.Fatalf("href %q error = %v, want *ValidationError", value, err)
		}
		if validation.Field != "links[0].href" {
			t.Errorf("field = %q, want links[0].href", validation.Field)
		}
		if validation.Message != fullURLMessage {
			t.Errorf("message = %q, want %q", validation.Message, fullURLMessage)
		}
	}
}

func TestValidateWorkTrimsTitleAndSummary(t *testing.T) {
	work := validWork()
	work.Title = "  Example work  "
	work.Summary = "\t A short summary \n"

	if err := validateWork(work); err != nil {
		t.Fatalf("validateWork() error = %v", err)
	}
	if work.Title != "Example work" {
		t.Errorf("title = %q, want trimmed value", work.Title)
	}
	if work.Summary != "A short summary" {
		t.Errorf("summary = %q, want trimmed value", work.Summary)
	}
}

func TestValidateWorkNil(t *testing.T) {
	if err := validateWork(nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("validateWork(nil) error = %v, want ErrValidation", err)
	}
}

func TestValidateImagePath(t *testing.T) {
	const workID = "0190e0e0-0000-7000-8000-000000000000"

	tests := []struct {
		name   string
		workID string
		path   string
		valid  bool
	}{
		{name: "work media file", workID: workID, path: "/media/" + workID + "/0123456789abcdef-640.webp", valid: true},
		{name: "generic media file", path: "/media/a.webp", valid: true},
		{name: "foreign work", workID: workID, path: "/media/0190e0e0-0000-7000-8000-000000000001/a.webp"},
		{name: "nested file", workID: workID, path: "/media/" + workID + "/nested/a.webp"},
		{name: "traversal", workID: workID, path: "/media/" + workID + "/../a.webp"},
		{name: "traversal in name", path: "/media/../a.webp"},
		{name: "backslash", path: "/media/a\\b.webp"},
		{name: "query", path: "/media/a.webp?x=1"},
		{name: "no prefix", path: "/static/a.webp"},
		{name: "empty", path: ""},
		{name: "not webp", path: "/media/a.png"},
		{name: "only directory", workID: workID, path: "/media/" + workID + "/"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateImagePath(test.workID, test.path)
			if test.valid {
				if err != nil {
					t.Fatalf("ValidateImagePath(%q) error = %v, want nil", test.path, err)
				}
				return
			}

			if !errors.Is(err, ErrValidation) {
				t.Fatalf("ValidateImagePath(%q) error = %v, want ErrValidation", test.path, err)
			}
			var validation *ValidationError
			if !errors.As(err, &validation) || validation.Field != "path" {
				t.Fatalf("error = %#v, want field path", err)
			}
		})
	}
}

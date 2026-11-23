package publish

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hanamuramiyu/miyu.pw/admin/internal/content"
)

type fakeStore struct {
	summaries     []content.WorkSummary
	works         map[string]*content.Work
	listErr       error
	getErr        error
	includeDrafts bool
}

func (f *fakeStore) List(_ context.Context, includeDrafts bool) ([]content.WorkSummary, error) {
	f.includeDrafts = includeDrafts
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.summaries, nil
}

func (f *fakeStore) Get(_ context.Context, id string) (*content.Work, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	work, ok := f.works[id]
	if !ok {
		return nil, content.ErrNotFound
	}
	return work, nil
}

func publishedWork() *content.Work {
	client := "Acme"
	return &content.Work{
		ID:        "work-1",
		Slug:      "monban",
		Title:     "monban",
		Summary:   "Modern Minecraft access control",
		BodyMD:    "## Why\n\nText.\n",
		Kind:      "Plugin for Minecraft servers",
		Audience:  "For myself",
		Role:      "Development and support",
		Client:    &client,
		Date:      "2025-11-07",
		Updated:   "2026-10-22",
		Published: true,
		Featured:  true,
		SortOrder: 1,
		Tools:     []string{"Java", "Gradle"},
		Links: []content.Link{
			{Label: "Repository", Href: "https://github.com/octocat/monban"},
		},
		Colors: []content.Color{{Name: "Background", Hex: "#141211"}},
		Fonts:  []content.Font{{Name: "Comfortaa", Role: "Headings"}},
	}
}

func summary(work *content.Work) content.WorkSummary {
	return content.WorkSummary{ID: work.ID, Slug: work.Slug, Published: work.Published}
}

func TestExportWritesPublishedWorksOnly(t *testing.T) {
	work := publishedWork()
	draft := &content.Work{ID: "work-2", Slug: "draft", Title: "Draft", Published: false}
	store := &fakeStore{
		summaries: []content.WorkSummary{summary(work), summary(draft)},
		works:     map[string]*content.Work{work.ID: work, draft.ID: draft},
	}

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "works", "monban"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "works", "monban", "old.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "works", "stale"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Export(context.Background(), store, dir); err != nil {
		t.Fatalf("Export: %v", err)
	}
	if store.includeDrafts {
		t.Error("List was called with includeDrafts=true, want false")
	}

	data, err := os.ReadFile(filepath.Join(dir, "works", "monban", "index.md"))
	if err != nil {
		t.Fatalf("read exported work: %v", err)
	}

	want := `---
title: "monban"
summary: "Modern Minecraft access control"
date: "2025-11-07"
updated: "2026-10-22"
tools: ["Java", "Gradle"]
kind: "Plugin for Minecraft servers"
audience: "For myself"
featured: true
order: 1
published: true
role: "Development and support"
client: "Acme"
links:
  - label: "Repository"
    href: "https://github.com/octocat/monban"
colors:
  - name: "Background"
    hex: "#141211"
fonts:
  - name: "Comfortaa"
    role: "Headings"
---

## Why

Text.
`
	if string(data) != want {
		t.Errorf("exported work =\n%s\nwant:\n%s", data, want)
	}

	if _, err := os.Stat(filepath.Join(dir, "works", "draft")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("draft directory status = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "works", "stale")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("stale directory status = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "works", "monban", "old.md")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("old file status = %v, want not exist", err)
	}
}

func TestExportOmitsEmptyOptionalFields(t *testing.T) {
	work := &content.Work{
		ID:        "work-1",
		Slug:      "minimal",
		Title:     "Minimal",
		Summary:   "Only the required fields",
		BodyMD:    "Body.\n",
		Date:      "2026-01-02",
		Published: true,
		Tools:     []string{"Go"},
	}
	store := &fakeStore{summaries: []content.WorkSummary{summary(work)}, works: map[string]*content.Work{work.ID: work}}

	dir := t.TempDir()
	if err := Export(context.Background(), store, dir); err != nil {
		t.Fatalf("Export: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "works", "minimal", "index.md"))
	if err != nil {
		t.Fatalf("read exported work: %v", err)
	}

	want := `---
title: "Minimal"
summary: "Only the required fields"
date: "2026-01-02"
tools: ["Go"]
featured: false
order: 0
published: true
---

Body.
`
	if string(data) != want {
		t.Errorf("exported work =\n%s\nwant:\n%s", data, want)
	}

	for _, key := range []string{"updated:", "kind:", "audience:", "role:", "client:", "links:", "colors:", "fonts:", "images:"} {
		if strings.Contains(string(data), key) {
			t.Errorf("exported work contains %q, want it omitted", key)
		}
	}
}

func TestExportKeepsPreviousStagingOnInvalidWork(t *testing.T) {
	work := publishedWork()
	work.Tools = nil
	store := &fakeStore{summaries: []content.WorkSummary{summary(work)}, works: map[string]*content.Work{work.ID: work}}

	dir := t.TempDir()
	previous := filepath.Join(dir, "works", "previous", "index.md")
	if err := os.MkdirAll(filepath.Dir(previous), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(previous, []byte("previous"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Export(context.Background(), store, dir)
	var invalid *InvalidWorkError
	if !errors.As(err, &invalid) {
		t.Fatalf("Export error = %v, want InvalidWorkError", err)
	}
	if !strings.Contains(err.Error(), "at least one tool") {
		t.Errorf("error = %q, want a tool hint", err)
	}

	data, readErr := os.ReadFile(previous)
	if readErr != nil {
		t.Fatalf("previous staging was removed: %v", readErr)
	}
	if string(data) != "previous" {
		t.Errorf("previous staging = %q, want previous", data)
	}
}

func TestExportWritesImagesInOrder(t *testing.T) {
	work := publishedWork()
	work.Images = []content.Image{
		{
			Path:    "/media/work-1/aaa-1200.webp",
			Alt:     "Permission editor",
			Caption: "Role tree",
			Width:   1200,
			Height:  630,
			IsCover: true,
		},
		{
			Path:   "/media/work-1/bbb-800.webp",
			Alt:    "Command output",
			Width:  800,
			Height: 450,
		},
	}
	store := &fakeStore{summaries: []content.WorkSummary{summary(work)}, works: map[string]*content.Work{work.ID: work}}

	dir := t.TempDir()
	if err := Export(context.Background(), store, dir); err != nil {
		t.Fatalf("Export: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "works", "monban", "index.md"))
	if err != nil {
		t.Fatalf("read exported work: %v", err)
	}

	want := `images:
  - src: "/media/work-1/aaa-1200.webp"
    alt: "Permission editor"
    caption: "Role tree"
    width: 1200
    height: 630
    cover: true
  - src: "/media/work-1/bbb-800.webp"
    alt: "Command output"
    width: 800
    height: 450
---`
	if !strings.Contains(string(data), want) {
		t.Errorf("exported work =\n%s\nwant it to contain:\n%s", data, want)
	}
	if strings.Count(string(data), "cover: true") != 1 {
		t.Errorf("exported work has %d cover markers, want 1", strings.Count(string(data), "cover: true"))
	}
}

func TestExportRejectsInvalidWorks(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*content.Work)
		want   string
	}{
		{
			name:   "invalid slug",
			mutate: func(w *content.Work) { w.Slug = "../escape" },
			want:   "slug",
		},
		{
			name:   "empty body",
			mutate: func(w *content.Work) { w.BodyMD = " \n\t" },
			want:   "body is empty",
		},
		{
			name:   "long title",
			mutate: func(w *content.Work) { w.Title = strings.Repeat("a", 121) },
			want:   "title is longer",
		},
		{
			name:   "too many tools",
			mutate: func(w *content.Work) { w.Tools = make([]string, 13) },
			want:   "12 tools",
		},
		{
			name:   "empty tool",
			mutate: func(w *content.Work) { w.Tools = []string{"Java", " "} },
			want:   "tools[1]",
		},
		{
			name:   "negative order",
			mutate: func(w *content.Work) { w.SortOrder = -1 },
			want:   "order",
		},
		{
			name:   "invalid link",
			mutate: func(w *content.Work) { w.Links = []content.Link{{Label: "Bad", Href: "javascript:alert(1)"}} },
			want:   "http(s)",
		},
		{
			name:   "invalid colour",
			mutate: func(w *content.Work) { w.Colors = []content.Color{{Name: "Bad", Hex: "red"}} },
			want:   "hex",
		},
		{
			name:   "empty font role",
			mutate: func(w *content.Work) { w.Fonts = []content.Font{{Name: "Comfortaa"}} },
			want:   "fonts[0].role",
		},
		{
			name: "image path outside media",
			mutate: func(w *content.Work) {
				w.Images = []content.Image{{Path: "images/cover.webp", Alt: "Cover", Width: 1200, Height: 630}}
			},
			want: "images[0].src",
		},
		{
			name: "image path traversal",
			mutate: func(w *content.Work) {
				w.Images = []content.Image{{Path: "/media/../cover.webp", Alt: "Cover", Width: 1200, Height: 630}}
			},
			want: "..",
		},
		{
			name: "empty image alt",
			mutate: func(w *content.Work) {
				w.Images = []content.Image{{Path: "/media/work-1/cover.webp", Alt: " ", Width: 1200, Height: 630}}
			},
			want: "images[0].alt",
		},
		{
			name: "non-positive image size",
			mutate: func(w *content.Work) {
				w.Images = []content.Image{{Path: "/media/work-1/cover.webp", Alt: "Cover", Width: 0, Height: 630}}
			},
			want: "positive",
		},
		{
			name: "two covers",
			mutate: func(w *content.Work) {
				w.Images = []content.Image{
					{Path: "/media/work-1/a.webp", Alt: "A", Width: 1200, Height: 630, IsCover: true},
					{Path: "/media/work-1/b.webp", Alt: "B", Width: 800, Height: 450, IsCover: true},
				}
			},
			want: "only one image can be the cover",
		},
		{
			name: "long image path",
			mutate: func(w *content.Work) {
				w.Images = []content.Image{{Path: "/media/work-1/" + strings.Repeat("a", 300) + ".webp", Alt: "A", Width: 1200, Height: 630}}
			},
			want: "images[0].src is longer",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			work := publishedWork()
			test.mutate(work)
			store := &fakeStore{summaries: []content.WorkSummary{summary(work)}, works: map[string]*content.Work{work.ID: work}}

			err := Export(context.Background(), store, t.TempDir())
			var invalid *InvalidWorkError
			if !errors.As(err, &invalid) {
				t.Fatalf("Export error = %v, want InvalidWorkError", err)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Errorf("error = %q, want %q", err, test.want)
			}
		})
	}
}

func TestExportRejectsInvalidBodyLink(t *testing.T) {
	work := publishedWork()
	work.Slug = "123"
	work.BodyMD = "[link text](https://)\n"
	store := &fakeStore{summaries: []content.WorkSummary{summary(work)}, works: map[string]*content.Work{work.ID: work}}

	err := Export(context.Background(), store, t.TempDir())
	var invalid *InvalidWorkError
	if !errors.As(err, &invalid) {
		t.Fatalf("Export error = %v, want InvalidWorkError", err)
	}
	if invalid.Slug != "123" {
		t.Errorf("slug = %q, want 123", invalid.Slug)
	}
	want := `work "123": link "https://" is not a valid URL`
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
}

func TestExportRejectsInvalidBodyTargets(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty target", body: "[x]()\n"},
		{name: "empty image target", body: "![x]()\n"},
		{name: "bare anchor", body: "[x](#)\n"},
		{name: "empty host", body: "[x](https://)\n"},
		{name: "empty http host", body: "[x](http://)\n"},
		{name: "path without host", body: "[x](https:///path)\n"},
		{name: "space in url", body: "[x](https://exa mple.com)\n"},
		{name: "relative link", body: "[x](relative/path.md)\n"},
		{name: "parent image path", body: "![x](../image.webp)\n"},
		{name: "relative image with parent", body: "![x](path/../image.webp)\n"},
		{name: "image invalid host", body: "![x](https://)\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			work := publishedWork()
			work.BodyMD = test.body
			store := &fakeStore{summaries: []content.WorkSummary{summary(work)}, works: map[string]*content.Work{work.ID: work}}

			err := Export(context.Background(), store, t.TempDir())
			var invalid *InvalidWorkError
			if !errors.As(err, &invalid) {
				t.Fatalf("Export error = %v, want InvalidWorkError", err)
			}
			if invalid.Slug != work.Slug {
				t.Errorf("slug = %q, want %q", invalid.Slug, work.Slug)
			}
			if !strings.Contains(err.Error(), "is not a valid URL") {
				t.Errorf("error = %q, want an invalid URL hint", err)
			}
		})
	}
}

func TestExportAcceptsValidBodyTargets(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "absolute link", body: "[x](https://example.com/path)\n"},
		{name: "absolute link with title", body: `[x](https://example.com "Example")` + "\n"},
		{name: "internal link", body: "[x](/works/other/)\n"},
		{name: "anchor link", body: "[x](#section)\n"},
		{name: "relative image", body: "![x](path/to/image.webp)\n"},
		{name: "absolute image", body: "![x](/media/image.webp)\n"},
		{name: "image inside link", body: "[![x](path/to/image.webp)](/works/other/)\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			work := publishedWork()
			work.BodyMD = test.body
			store := &fakeStore{summaries: []content.WorkSummary{summary(work)}, works: map[string]*content.Work{work.ID: work}}

			if err := Export(context.Background(), store, t.TempDir()); err != nil {
				t.Fatalf("Export: %v", err)
			}
		})
	}
}

func TestExportRejectsDuplicateSlugs(t *testing.T) {
	work := publishedWork()
	other := publishedWork()
	other.ID = "work-2"
	store := &fakeStore{
		summaries: []content.WorkSummary{summary(work), summary(other)},
		works:     map[string]*content.Work{work.ID: work, other.ID: other},
	}

	err := Export(context.Background(), store, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "duplicate slug") {
		t.Fatalf("Export error = %v, want duplicate slug", err)
	}
}

func TestExportSkipsUnpublishedEvenIfListed(t *testing.T) {
	work := publishedWork()
	work.Published = false
	store := &fakeStore{summaries: []content.WorkSummary{summary(work)}, works: map[string]*content.Work{work.ID: work}}

	dir := t.TempDir()
	if err := Export(context.Background(), store, dir); err != nil {
		t.Fatalf("Export: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(dir, "works"))
	if err != nil {
		t.Fatalf("read works directory: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("works directory has %d entries, want 0", len(entries))
	}
}

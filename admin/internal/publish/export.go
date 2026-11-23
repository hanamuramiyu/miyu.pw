package publish

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/hanamuramiyu/miyu.pw/admin/internal/content"
)

const (
	dateLayout     = "2006-01-02"
	fullURLMessage = "expected a full http(s) URL with a host"
)

var (
	slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	hexPattern  = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
)

type Store interface {
	List(ctx context.Context, includeDrafts bool) ([]content.WorkSummary, error)
	Get(ctx context.Context, id string) (*content.Work, error)
}

type InvalidWorkError struct {
	Slug   string
	Reason string
}

func (e *InvalidWorkError) Error() string {
	if e.Slug == "" {
		return e.Reason
	}
	return fmt.Sprintf("work %q: %s", e.Slug, e.Reason)
}

func Export(ctx context.Context, store Store, dir string) error {
	if store == nil {
		return fmt.Errorf("publish: store is not configured")
	}
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("publish: export directory is not configured")
	}

	summaries, err := store.List(ctx, false)
	if err != nil {
		return fmt.Errorf("list published works: %w", err)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create export directory: %w", err)
	}

	staging, err := os.MkdirTemp(dir, ".works-")
	if err != nil {
		return fmt.Errorf("create temporary works directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()

	if err := os.Chmod(staging, 0o755); err != nil {
		return fmt.Errorf("prepare temporary works directory: %w", err)
	}

	seen := make(map[string]bool, len(summaries))
	for _, summary := range summaries {
		work, err := store.Get(ctx, summary.ID)
		if err != nil {
			return fmt.Errorf("load work %q: %w", summary.Slug, err)
		}
		if work == nil {
			return fmt.Errorf("load work %q: empty result", summary.Slug)
		}
		if !work.Published {
			continue
		}
		if seen[work.Slug] {
			return fmt.Errorf("load work %q: duplicate slug", work.Slug)
		}
		seen[work.Slug] = true

		if err := writeWork(staging, work); err != nil {
			return err
		}
	}

	target := filepath.Join(dir, "works")
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("replace works directory: %w", err)
	}
	if err := os.Rename(staging, target); err != nil {
		return fmt.Errorf("replace works directory: %w", err)
	}
	return nil
}

func writeWork(root string, work *content.Work) error {
	if !slugPattern.MatchString(work.Slug) {
		return invalidWork(work.Slug, "slug must contain lowercase letters, digits and single hyphens")
	}

	frontmatter, err := buildFrontmatter(work)
	if err != nil {
		return err
	}

	if err := validateBodyLinks(work.Slug, work.BodyMD); err != nil {
		return err
	}

	workDir := filepath.Join(root, work.Slug)
	if err := os.RemoveAll(workDir); err != nil {
		return fmt.Errorf("prepare work %q: %w", work.Slug, err)
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("prepare work %q: %w", work.Slug, err)
	}

	content := frontmatter + work.BodyMD
	if err := os.WriteFile(filepath.Join(workDir, "index.md"), []byte(content), 0o644); err != nil {
		return fmt.Errorf("write work %q: %w", work.Slug, err)
	}
	return nil
}

func buildFrontmatter(work *content.Work) (string, error) {
	title := strings.TrimSpace(work.Title)
	if title == "" {
		return "", invalidWork(work.Slug, "title is required")
	}
	if err := checkLength(work.Slug, "title", title, 120); err != nil {
		return "", err
	}

	summary := strings.TrimSpace(work.Summary)
	if summary == "" {
		return "", invalidWork(work.Slug, "summary is required")
	}
	if err := checkLength(work.Slug, "summary", summary, 200); err != nil {
		return "", err
	}

	if err := checkDate(work.Slug, "date", work.Date); err != nil {
		return "", err
	}
	updated := strings.TrimSpace(work.Updated)
	if updated != "" {
		if err := checkDate(work.Slug, "updated", updated); err != nil {
			return "", err
		}
	}

	if len(work.Tools) == 0 {
		return "", invalidWork(work.Slug, "at least one tool is required")
	}
	if len(work.Tools) > 12 {
		return "", invalidWork(work.Slug, "no more than 12 tools are allowed")
	}
	for i, tool := range work.Tools {
		name := strings.TrimSpace(tool)
		if name == "" {
			return "", invalidWork(work.Slug, fmt.Sprintf("tools[%d] is empty", i))
		}
		if err := checkLength(work.Slug, fmt.Sprintf("tools[%d]", i), name, 60); err != nil {
			return "", err
		}
	}

	if work.SortOrder < 0 {
		return "", invalidWork(work.Slug, "order must not be negative")
	}
	if strings.TrimSpace(work.BodyMD) == "" {
		return "", invalidWork(work.Slug, "the Markdown body is empty")
	}

	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %s\n", yamlString(title))
	fmt.Fprintf(&b, "summary: %s\n", yamlString(summary))
	fmt.Fprintf(&b, "date: %s\n", yamlString(work.Date))
	if updated != "" {
		fmt.Fprintf(&b, "updated: %s\n", yamlString(updated))
	}
	b.WriteString("tools: [")
	for i, tool := range work.Tools {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(yamlString(strings.TrimSpace(tool)))
	}
	b.WriteString("]\n")

	if kind := strings.TrimSpace(work.Kind); kind != "" {
		if err := checkLength(work.Slug, "kind", kind, 80); err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "kind: %s\n", yamlString(kind))
	}
	if audience := strings.TrimSpace(work.Audience); audience != "" {
		if err := checkLength(work.Slug, "audience", audience, 120); err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "audience: %s\n", yamlString(audience))
	}

	fmt.Fprintf(&b, "featured: %t\n", work.Featured)
	fmt.Fprintf(&b, "order: %d\n", work.SortOrder)
	b.WriteString("published: true\n")

	if role := strings.TrimSpace(work.Role); role != "" {
		if err := checkLength(work.Slug, "role", role, 120); err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "role: %s\n", yamlString(role))
	}
	if work.Client != nil {
		client := strings.TrimSpace(*work.Client)
		if client != "" {
			if err := checkLength(work.Slug, "client", client, 120); err != nil {
				return "", err
			}
			fmt.Fprintf(&b, "client: %s\n", yamlString(client))
		}
	}

	if len(work.Links) > 0 {
		if len(work.Links) > 8 {
			return "", invalidWork(work.Slug, "no more than 8 links are allowed")
		}
		b.WriteString("links:\n")
		for i, link := range work.Links {
			label := strings.TrimSpace(link.Label)
			href := strings.TrimSpace(link.Href)
			if label == "" {
				return "", invalidWork(work.Slug, fmt.Sprintf("links[%d].label is required", i))
			}
			if err := checkLength(work.Slug, fmt.Sprintf("links[%d].label", i), label, 60); err != nil {
				return "", err
			}
			if err := checkHref(href); err != nil {
				return "", invalidWork(work.Slug, fmt.Sprintf("links[%d].href %s", i, err))
			}
			fmt.Fprintf(&b, "  - label: %s\n", yamlString(label))
			fmt.Fprintf(&b, "    href: %s\n", yamlString(href))
		}
	}

	if len(work.Colors) > 0 {
		if len(work.Colors) > 12 {
			return "", invalidWork(work.Slug, "no more than 12 colours are allowed")
		}
		b.WriteString("colors:\n")
		for i, color := range work.Colors {
			name := strings.TrimSpace(color.Name)
			hex := strings.TrimSpace(color.Hex)
			if name == "" {
				return "", invalidWork(work.Slug, fmt.Sprintf("colors[%d].name is required", i))
			}
			if err := checkLength(work.Slug, fmt.Sprintf("colors[%d].name", i), name, 60); err != nil {
				return "", err
			}
			if !hexPattern.MatchString(hex) {
				return "", invalidWork(work.Slug, fmt.Sprintf("colors[%d].hex must be a 3- or 6-digit hex colour", i))
			}
			fmt.Fprintf(&b, "  - name: %s\n", yamlString(name))
			fmt.Fprintf(&b, "    hex: %s\n", yamlString(hex))
		}
	}

	if len(work.Fonts) > 0 {
		if len(work.Fonts) > 8 {
			return "", invalidWork(work.Slug, "no more than 8 fonts are allowed")
		}
		b.WriteString("fonts:\n")
		for i, font := range work.Fonts {
			name := strings.TrimSpace(font.Name)
			role := strings.TrimSpace(font.Role)
			if name == "" {
				return "", invalidWork(work.Slug, fmt.Sprintf("fonts[%d].name is required", i))
			}
			if role == "" {
				return "", invalidWork(work.Slug, fmt.Sprintf("fonts[%d].role is required", i))
			}
			if err := checkLength(work.Slug, fmt.Sprintf("fonts[%d].name", i), name, 80); err != nil {
				return "", err
			}
			if err := checkLength(work.Slug, fmt.Sprintf("fonts[%d].role", i), role, 80); err != nil {
				return "", err
			}
			fmt.Fprintf(&b, "  - name: %s\n", yamlString(name))
			fmt.Fprintf(&b, "    role: %s\n", yamlString(role))
		}
	}

	if len(work.Images) > 0 {
		b.WriteString("images:\n")
		coverSeen := false
		for i, image := range work.Images {
			src := strings.TrimSpace(image.Path)
			alt := strings.TrimSpace(image.Alt)
			if !strings.HasPrefix(src, "/media/") {
				return "", invalidWork(work.Slug, fmt.Sprintf("images[%d].src must start with /media/", i))
			}
			if strings.Contains(src, "..") {
				return "", invalidWork(work.Slug, fmt.Sprintf("images[%d].src must not contain ..", i))
			}
			if err := checkLength(work.Slug, fmt.Sprintf("images[%d].src", i), src, 300); err != nil {
				return "", err
			}
			if alt == "" {
				return "", invalidWork(work.Slug, fmt.Sprintf("images[%d].alt is required", i))
			}
			if err := checkLength(work.Slug, fmt.Sprintf("images[%d].alt", i), alt, 300); err != nil {
				return "", err
			}
			caption := strings.TrimSpace(image.Caption)
			if caption != "" {
				if err := checkLength(work.Slug, fmt.Sprintf("images[%d].caption", i), caption, 200); err != nil {
					return "", err
				}
			}
			if image.Width <= 0 || image.Height <= 0 {
				return "", invalidWork(work.Slug, fmt.Sprintf("images[%d].width and height must be positive", i))
			}
			fmt.Fprintf(&b, "  - src: %s\n", yamlString(src))
			fmt.Fprintf(&b, "    alt: %s\n", yamlString(alt))
			if caption != "" {
				fmt.Fprintf(&b, "    caption: %s\n", yamlString(caption))
			}
			fmt.Fprintf(&b, "    width: %d\n", image.Width)
			fmt.Fprintf(&b, "    height: %d\n", image.Height)
			if image.IsCover {
				if coverSeen {
					return "", invalidWork(work.Slug, "only one image can be the cover")
				}
				coverSeen = true
				b.WriteString("    cover: true\n")
			}
		}
	}

	b.WriteString("---\n\n")
	return b.String(), nil
}

func invalidWork(slug, reason string) error {
	return &InvalidWorkError{Slug: slug, Reason: reason}
}

func checkLength(slug, field, value string, max int) error {
	if len(utf16.Encode([]rune(value))) > max {
		return invalidWork(slug, fmt.Sprintf("%s is longer than %d characters", field, max))
	}
	return nil
}

func checkDate(slug, field, value string) error {
	parsed, err := time.Parse(dateLayout, value)
	if err != nil || parsed.Format(dateLayout) != value {
		return invalidWork(slug, fmt.Sprintf("%s must be a real date in YYYY-MM-DD format", field))
	}
	return nil
}

func checkHref(value string) error {
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

type markdownTarget struct {
	target string
	image  bool
}

func validateBodyLinks(slug, body string) error {
	for _, found := range markdownTargets(body) {
		target, ok := parseMarkdownTarget(found.target)
		if !ok || !validBodyTarget(target, found.image) {
			return invalidWork(slug, fmt.Sprintf("link %q is not a valid URL", target))
		}
	}
	return nil
}

func markdownTargets(body string) []markdownTarget {
	var targets []markdownTarget

	offset := 0
	for {
		index := strings.Index(body[offset:], "](")
		if index < 0 {
			break
		}
		index += offset
		offset = index + 2

		if index > 0 && body[index-1] == '\\' {
			continue
		}
		open, ok := matchingOpenBracket(body, index)
		if !ok {
			continue
		}
		end := strings.Index(body[index+2:], ")")
		if end < 0 {
			break
		}
		raw := body[index+2 : index+2+end]
		image := open > 0 && body[open-1] == '!'
		targets = append(targets, markdownTarget{target: raw, image: image})
		offset = index + 2 + end + 1
	}

	return targets
}

func matchingOpenBracket(body string, close int) (int, bool) {
	depth := 0
	for i := close - 1; i >= 0; i-- {
		switch body[i] {
		case ']':
			depth++
		case '[':
			if depth == 0 {
				return i, true
			}
			depth--
		}
	}
	return 0, false
}

func parseMarkdownTarget(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}

	if strings.HasPrefix(raw, "<") {
		end := strings.Index(raw, ">")
		if end < 0 {
			return raw, false
		}
		target := raw[1:end]
		rest := strings.TrimSpace(raw[end+1:])
		if rest != "" && !isMarkdownTitle(rest) {
			return raw, false
		}
		return target, true
	}

	if fields := strings.Fields(raw); len(fields) > 1 {
		rest := strings.TrimSpace(raw[len(fields[0]):])
		if !isMarkdownTitle(rest) {
			return raw, false
		}
		return fields[0], true
	}

	return raw, true
}

func isMarkdownTitle(value string) bool {
	if len(value) < 2 {
		return false
	}
	last := value[len(value)-1]
	switch value[0] {
	case '"':
		return last == '"'
	case '\'':
		return last == '\''
	case '(':
		return last == ')'
	}
	return false
}

func validBodyTarget(target string, image bool) bool {
	if strings.ContainsAny(target, " \t\r\n\v\f") {
		return false
	}
	if target == "" || target == "#" {
		return false
	}
	if strings.HasPrefix(target, "#") {
		return true
	}
	lower := strings.ToLower(target)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return checkHref(target) == nil
	}
	if strings.HasPrefix(target, "/") {
		return true
	}
	parsed, err := url.Parse(target)
	if err != nil {
		return false
	}
	if parsed.Scheme != "" {
		return false
	}
	if !image {
		return false
	}
	return !strings.Contains(target, "..")
}

func yamlString(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return `""`
	}
	return string(encoded)
}

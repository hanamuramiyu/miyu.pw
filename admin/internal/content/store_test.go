package content

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testRepository(t *testing.T) *Repository {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse DATABASE_URL: %v", err)
	}

	schema := fmt.Sprintf("content_test_%d", time.Now().UnixNano())
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}

	t.Cleanup(func() {
		dropCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(dropCtx, "DROP SCHEMA IF EXISTS "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
		pool.Close()
	})

	if _, err := pool.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	for _, name := range []string{"0001_init.sql", "0003_updated_at.sql"} {
		migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if _, err := pool.Exec(ctx, string(migration)); err != nil {
			t.Fatalf("apply %s: %v", name, err)
		}
	}

	return NewRepository(pool)
}

func sampleWork(slug string, published bool) *Work {
	return &Work{
		Slug:      slug,
		Title:     "Title " + slug,
		Summary:   "Summary " + slug,
		BodyMD:    "Body",
		Kind:      "Website",
		Audience:  "For clients",
		Role:      "Development",
		Date:      "2025-01-02",
		Updated:   "2025-02-03",
		Published: published,
		Featured:  true,
		SortOrder: 3,
		Tools:     []string{"Go", "PostgreSQL"},
		Links:     []Link{{Label: "Site", Href: "https://example.com"}},
		Colors:    []Color{{Name: "Ink", Hex: "#141211"}},
		Fonts:     []Font{{Name: "Manrope", Role: "Body"}},
		Images: []Image{{
			Path: "/media/a.webp", Alt: "A", Caption: "Caption",
			Width: 800, Height: 600, IsCover: true,
		}},
	}
}

func TestCreateAndGet(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()

	client := "Acme"
	want := sampleWork("first-work", true)
	want.Client = &client

	if err := repo.Create(ctx, want); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if want.ID == "" {
		t.Fatal("Create() did not fill the id")
	}
	if want.CreatedAt.IsZero() || want.UpdatedAt.IsZero() {
		t.Fatal("Create() did not fill the timestamps")
	}

	got, err := repo.Get(ctx, want.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.ID != want.ID || got.Slug != want.Slug || got.Title != want.Title {
		t.Errorf("identity fields = %+v, want %+v", got, want)
	}
	if got.Date != "2025-01-02" || got.Updated != "2025-02-03" {
		t.Errorf("dates = %q/%q, want 2025-01-02/2025-02-03", got.Date, got.Updated)
	}
	if got.Client == nil || *got.Client != client {
		t.Errorf("client = %v, want %q", got.Client, client)
	}
	if !got.Published || !got.Featured || got.SortOrder != 3 {
		t.Errorf("flags = published:%v featured:%v sort:%d", got.Published, got.Featured, got.SortOrder)
	}
	if !reflect.DeepEqual(got.Tools, want.Tools) {
		t.Errorf("tools = %v, want %v", got.Tools, want.Tools)
	}
	if !reflect.DeepEqual(got.Links, want.Links) {
		t.Errorf("links = %v, want %v", got.Links, want.Links)
	}
	if !reflect.DeepEqual(got.Colors, want.Colors) {
		t.Errorf("colors = %v, want %v", got.Colors, want.Colors)
	}
	if !reflect.DeepEqual(got.Fonts, want.Fonts) {
		t.Errorf("fonts = %v, want %v", got.Fonts, want.Fonts)
	}
	if !reflect.DeepEqual(got.Images, want.Images) {
		t.Errorf("images = %v, want %v", got.Images, want.Images)
	}
}

func TestListFiltersDrafts(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()

	if err := repo.Create(ctx, sampleWork("draft-work", false)); err != nil {
		t.Fatalf("Create(draft) error = %v", err)
	}
	if err := repo.Create(ctx, sampleWork("published-work", true)); err != nil {
		t.Fatalf("Create(published) error = %v", err)
	}

	published, err := repo.List(ctx, false)
	if err != nil {
		t.Fatalf("List(false) error = %v", err)
	}
	if len(published) != 1 || published[0].Slug != "published-work" {
		t.Fatalf("List(false) = %+v, want only published-work", published)
	}

	all, err := repo.List(ctx, true)
	if err != nil {
		t.Fatalf("List(true) error = %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("List(true) returned %d works, want 2", len(all))
	}
}

func TestListEmptyReturnsEmptySlice(t *testing.T) {
	repo := testRepository(t)

	works, err := repo.List(context.Background(), true)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if works == nil {
		t.Fatal("List() returned nil, want an empty slice")
	}
	if len(works) != 0 {
		t.Fatalf("List() returned %d works, want 0", len(works))
	}
}

func TestGetMissingWork(t *testing.T) {
	repo := testRepository(t)

	_, err := repo.Get(context.Background(), "00000000-0000-0000-0000-0000000000aa")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestUpdateMissingWork(t *testing.T) {
	repo := testRepository(t)

	work := sampleWork("missing-work", false)
	work.ID = "00000000-0000-0000-0000-0000000000bb"
	if err := repo.Update(context.Background(), work); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update() error = %v, want ErrNotFound", err)
	}
}

func TestInvalidID(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()

	checks := []struct {
		name string
		call func() error
	}{
		{name: "get", call: func() error { _, err := repo.Get(ctx, "not-a-uuid"); return err }},
		{name: "delete", call: func() error { return repo.Delete(ctx, "not-a-uuid") }},
		{name: "publish", call: func() error { return repo.Publish(ctx, "not-a-uuid", true) }},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			err := check.call()
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("error = %v, want ErrValidation", err)
			}
			var validation *ValidationError
			if !errors.As(err, &validation) || validation.Field != "id" {
				t.Fatalf("error = %#v, want field id", err)
			}
		})
	}
}

func TestUpdateReplacesChildren(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()

	work := sampleWork("update-work", false)
	work.Tools = []string{"Go", "PostgreSQL"}
	work.Links = []Link{
		{Label: "Site", Href: "https://example.com"},
		{Label: "Docs", Href: "https://docs.example.com"},
	}
	work.Colors = []Color{
		{Name: "Ink", Hex: "#141211"},
		{Name: "Paper", Hex: "#ebe4dc"},
	}
	work.Fonts = []Font{
		{Name: "Manrope", Role: "Body"},
		{Name: "Comfortaa", Role: "Display"},
	}
	work.Images = []Image{
		{Path: "/media/a.webp", Alt: "A", Width: 800, Height: 600},
		{Path: "/media/b.webp", Alt: "B", Width: 800, Height: 600},
	}
	if err := repo.Create(ctx, work); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	work.Tools = []string{"Rust"}
	work.Links = []Link{{Label: "Repository", Href: "https://github.com/example"}}
	work.Colors = []Color{{Name: "Paper", Hex: "#fff"}}
	work.Fonts = nil
	work.Images = []Image{
		{Path: "/media/c.webp", Alt: "C", Width: 1200, Height: 630, IsCover: true},
		{Path: "/media/d.webp", Alt: "D"},
	}
	if err := repo.Update(ctx, work); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.Get(ctx, work.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !reflect.DeepEqual(got.Tools, work.Tools) {
		t.Errorf("tools = %v, want %v", got.Tools, work.Tools)
	}
	if !reflect.DeepEqual(got.Links, work.Links) {
		t.Errorf("links = %v, want %v", got.Links, work.Links)
	}
	if !reflect.DeepEqual(got.Colors, work.Colors) {
		t.Errorf("colors = %v, want %v", got.Colors, work.Colors)
	}
	if len(got.Fonts) != 0 {
		t.Errorf("fonts = %v, want none", got.Fonts)
	}
	if !reflect.DeepEqual(got.Images, work.Images) {
		t.Errorf("images = %v, want %v", got.Images, work.Images)
	}
}

func TestDeleteAndPublish(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()

	work := sampleWork("delete-work", false)
	if err := repo.Create(ctx, work); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Publish(ctx, work.ID, true); err != nil {
		t.Fatalf("Publish(true) error = %v", err)
	}
	got, err := repo.Get(ctx, work.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !got.Published {
		t.Error("work is not published after Publish(true)")
	}

	if err := repo.Publish(ctx, work.ID, false); err != nil {
		t.Fatalf("Publish(false) error = %v", err)
	}
	got, err = repo.Get(ctx, work.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Published {
		t.Error("work is published after Publish(false)")
	}

	if err := repo.Delete(ctx, work.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repo.Get(ctx, work.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() after Delete error = %v, want ErrNotFound", err)
	}
	if err := repo.Delete(ctx, work.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second Delete() error = %v, want ErrNotFound", err)
	}
	if err := repo.Publish(ctx, work.ID, true); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Publish() on deleted work error = %v, want ErrNotFound", err)
	}
}

func TestCreateRejectsDuplicateSlug(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()

	if err := repo.Create(ctx, sampleWork("same-slug", false)); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	err := repo.Create(ctx, sampleWork("same-slug", false))
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("second Create() error = %v, want ErrValidation", err)
	}
	var validation *ValidationError
	if !errors.As(err, &validation) || validation.Field != "slug" {
		t.Fatalf("error = %#v, want field slug", err)
	}
}

func TestCreateRejectsInvalidWorkBeforeSQL(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()

	work := sampleWork("invalid-work", false)
	work.Summary = "  "
	err := repo.Create(ctx, work)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("Create() error = %v, want ErrValidation", err)
	}
}

package content

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const selectWorkSQL = `
	SELECT id::text, slug, title, summary, body_md, kind, audience, role, client,
	       to_char(date, 'YYYY-MM-DD'), to_char(updated, 'YYYY-MM-DD'),
	       published, featured, sort_order, created_at, updated_at
	FROM works
	WHERE id = $1`

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context, includeDrafts bool) ([]WorkSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, slug, title, summary, kind,
		       to_char(date, 'YYYY-MM-DD'), to_char(updated, 'YYYY-MM-DD'),
		       published, featured, sort_order, created_at, updated_at
		FROM works
		WHERE $1 OR published
		ORDER BY sort_order, date DESC, slug`, includeDrafts)
	if err != nil {
		return nil, fmt.Errorf("select works: %w", err)
	}
	defer rows.Close()

	summaries := make([]WorkSummary, 0)
	for rows.Next() {
		var summary WorkSummary
		if err := rows.Scan(
			&summary.ID, &summary.Slug, &summary.Title, &summary.Summary, &summary.Kind,
			&summary.Date, &summary.Updated, &summary.Published, &summary.Featured,
			&summary.SortOrder, &summary.CreatedAt, &summary.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan work summary: %w", err)
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate works: %w", err)
	}

	return summaries, nil
}

func (r *Repository) Get(ctx context.Context, id string) (*Work, error) {
	idParam, err := uuidParam(id)
	if err != nil {
		return nil, err
	}

	work := &Work{}
	var client *string
	err = r.pool.QueryRow(ctx, selectWorkSQL, idParam).Scan(
		&work.ID, &work.Slug, &work.Title, &work.Summary, &work.BodyMD,
		&work.Kind, &work.Audience, &work.Role, &client,
		&work.Date, &work.Updated, &work.Published, &work.Featured, &work.SortOrder,
		&work.CreatedAt, &work.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select work: %w", err)
	}
	work.Client = client

	if err := loadChildren(ctx, r.pool, idParam, work); err != nil {
		return nil, err
	}

	return work, nil
}

func (r *Repository) Create(ctx context.Context, w *Work) error {
	if err := validateWork(w); err != nil {
		return err
	}
	idParam, err := optionalUUID(w.ID)
	if err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create work: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id pgtype.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO works (
			id, slug, title, summary, body_md, kind, audience, role, client,
			date, updated, published, featured, sort_order
		) VALUES (
			COALESCE($1, gen_random_uuid()), $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14
		)
		RETURNING id, created_at, updated_at`,
		idParam, w.Slug, w.Title, w.Summary, w.BodyMD, w.Kind, w.Audience, w.Role,
		w.Client, parseDate(w.Date), parseDate(w.Updated), w.Published, w.Featured,
		w.SortOrder,
	).Scan(&id, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return mapWriteError(err)
	}
	w.ID = id.String()

	if err := replaceChildren(ctx, tx, id, w); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit create work: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, w *Work) error {
	if err := validateWork(w); err != nil {
		return err
	}
	idParam, err := uuidParam(w.ID)
	if err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin update work: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx, `
		UPDATE works SET
			slug = $2, title = $3, summary = $4, body_md = $5, kind = $6,
			audience = $7, role = $8, client = $9, date = $10, updated = $11,
			published = $12, featured = $13, sort_order = $14
		WHERE id = $1
		RETURNING created_at, updated_at`,
		idParam, w.Slug, w.Title, w.Summary, w.BodyMD, w.Kind, w.Audience, w.Role,
		w.Client, parseDate(w.Date), parseDate(w.Updated), w.Published, w.Featured,
		w.SortOrder,
	).Scan(&w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return mapWriteError(err)
	}

	if err := replaceChildren(ctx, tx, idParam, w); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit update work: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	idParam, err := uuidParam(id)
	if err != nil {
		return err
	}

	tag, err := r.pool.Exec(ctx, `DELETE FROM works WHERE id = $1`, idParam)
	if err != nil {
		return fmt.Errorf("delete work: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Publish(ctx context.Context, id string, published bool) error {
	idParam, err := uuidParam(id)
	if err != nil {
		return err
	}

	tag, err := r.pool.Exec(ctx, `UPDATE works SET published = $2 WHERE id = $1`, idParam, published)
	if err != nil {
		return fmt.Errorf("publish work: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type queryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func loadChildren(ctx context.Context, q queryer, id pgtype.UUID, w *Work) error {
	w.Tools = []string{}
	w.Links = []Link{}
	w.Colors = []Color{}
	w.Fonts = []Font{}
	w.Images = []Image{}

	if err := loadTools(ctx, q, id, w); err != nil {
		return err
	}
	if err := loadLinks(ctx, q, id, w); err != nil {
		return err
	}
	if err := loadColors(ctx, q, id, w); err != nil {
		return err
	}
	if err := loadFonts(ctx, q, id, w); err != nil {
		return err
	}
	return loadImages(ctx, q, id, w)
}

func loadTools(ctx context.Context, q queryer, id pgtype.UUID, w *Work) error {
	rows, err := q.Query(ctx, `SELECT name FROM work_tools WHERE work_id = $1 ORDER BY position`, id)
	if err != nil {
		return fmt.Errorf("select work tools: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("scan work tool: %w", err)
		}
		w.Tools = append(w.Tools, name)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate work tools: %w", err)
	}
	return nil
}

func loadLinks(ctx context.Context, q queryer, id pgtype.UUID, w *Work) error {
	rows, err := q.Query(ctx, `SELECT label, href FROM work_links WHERE work_id = $1 ORDER BY position`, id)
	if err != nil {
		return fmt.Errorf("select work links: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.Label, &link.Href); err != nil {
			return fmt.Errorf("scan work link: %w", err)
		}
		w.Links = append(w.Links, link)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate work links: %w", err)
	}
	return nil
}

func loadColors(ctx context.Context, q queryer, id pgtype.UUID, w *Work) error {
	rows, err := q.Query(ctx, `SELECT name, hex FROM work_colors WHERE work_id = $1 ORDER BY position`, id)
	if err != nil {
		return fmt.Errorf("select work colors: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var color Color
		if err := rows.Scan(&color.Name, &color.Hex); err != nil {
			return fmt.Errorf("scan work color: %w", err)
		}
		w.Colors = append(w.Colors, color)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate work colors: %w", err)
	}
	return nil
}

func loadFonts(ctx context.Context, q queryer, id pgtype.UUID, w *Work) error {
	rows, err := q.Query(ctx, `SELECT name, role FROM work_fonts WHERE work_id = $1 ORDER BY position`, id)
	if err != nil {
		return fmt.Errorf("select work fonts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var font Font
		if err := rows.Scan(&font.Name, &font.Role); err != nil {
			return fmt.Errorf("scan work font: %w", err)
		}
		w.Fonts = append(w.Fonts, font)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate work fonts: %w", err)
	}
	return nil
}

func loadImages(ctx context.Context, q queryer, id pgtype.UUID, w *Work) error {
	rows, err := q.Query(ctx, `
		SELECT path, alt, caption, width, height, is_cover
		FROM work_images
		WHERE work_id = $1
		ORDER BY position`, id)
	if err != nil {
		return fmt.Errorf("select work images: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var image Image
		var caption *string
		var width, height *int
		if err := rows.Scan(&image.Path, &image.Alt, &caption, &width, &height, &image.IsCover); err != nil {
			return fmt.Errorf("scan work image: %w", err)
		}
		if caption != nil {
			image.Caption = *caption
		}
		if width != nil {
			image.Width = *width
		}
		if height != nil {
			image.Height = *height
		}
		w.Images = append(w.Images, image)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate work images: %w", err)
	}
	return nil
}

func replaceChildren(ctx context.Context, tx pgx.Tx, id pgtype.UUID, w *Work) error {
	if _, err := tx.Exec(ctx, `DELETE FROM work_tools WHERE work_id = $1`, id); err != nil {
		return fmt.Errorf("delete work tools: %w", err)
	}
	for i, name := range w.Tools {
		if _, err := tx.Exec(ctx,
			`INSERT INTO work_tools (work_id, name, position) VALUES ($1, $2, $3)`,
			id, name, i+1,
		); err != nil {
			return mapWriteError(err)
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM work_links WHERE work_id = $1`, id); err != nil {
		return fmt.Errorf("delete work links: %w", err)
	}
	for i, link := range w.Links {
		if _, err := tx.Exec(ctx,
			`INSERT INTO work_links (work_id, label, href, position) VALUES ($1, $2, $3, $4)`,
			id, link.Label, link.Href, i+1,
		); err != nil {
			return mapWriteError(err)
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM work_colors WHERE work_id = $1`, id); err != nil {
		return fmt.Errorf("delete work colors: %w", err)
	}
	for i, color := range w.Colors {
		if _, err := tx.Exec(ctx,
			`INSERT INTO work_colors (work_id, name, hex, position) VALUES ($1, $2, $3, $4)`,
			id, color.Name, color.Hex, i+1,
		); err != nil {
			return mapWriteError(err)
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM work_fonts WHERE work_id = $1`, id); err != nil {
		return fmt.Errorf("delete work fonts: %w", err)
	}
	for i, font := range w.Fonts {
		if _, err := tx.Exec(ctx,
			`INSERT INTO work_fonts (work_id, name, role, position) VALUES ($1, $2, $3, $4)`,
			id, font.Name, font.Role, i+1,
		); err != nil {
			return mapWriteError(err)
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM work_images WHERE work_id = $1`, id); err != nil {
		return fmt.Errorf("delete work images: %w", err)
	}
	for i, image := range w.Images {
		if _, err := tx.Exec(ctx, `
			INSERT INTO work_images (work_id, path, alt, caption, width, height, position, is_cover)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			id, image.Path, image.Alt, nullableString(image.Caption),
			nullableInt(image.Width), nullableInt(image.Height), i+1, image.IsCover,
		); err != nil {
			return mapWriteError(err)
		}
	}

	return nil
}

func uuidParam(id string) (pgtype.UUID, error) {
	var parsed pgtype.UUID
	if err := parsed.Scan(id); err != nil {
		return pgtype.UUID{}, invalid("id", "must be a valid uuid")
	}
	return parsed, nil
}

func optionalUUID(id string) (pgtype.UUID, error) {
	if id == "" {
		return pgtype.UUID{}, nil
	}
	return uuidParam(id)
}

func parseDate(value string) time.Time {
	parsed, _ := time.Parse(dateLayout, value)
	return parsed
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == "23505" && pgErr.ConstraintName == "works_slug_key":
			return invalid("slug", "is already in use")
		case pgErr.Code == "23503":
			return ErrNotFound
		}
	}
	return err
}

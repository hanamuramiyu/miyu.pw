package content

import "time"

type Work struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	BodyMD    string    `json:"bodyMd"`
	Kind      string    `json:"kind"`
	Audience  string    `json:"audience"`
	Role      string    `json:"role"`
	Client    *string   `json:"client"`
	Date      string    `json:"date"`
	Updated   string    `json:"updated"`
	Published bool      `json:"published"`
	Featured  bool      `json:"featured"`
	SortOrder int       `json:"sortOrder"`
	Tools     []string  `json:"tools"`
	Links     []Link    `json:"links"`
	Colors    []Color   `json:"colors"`
	Fonts     []Font    `json:"fonts"`
	Images    []Image   `json:"images"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type WorkSummary struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Kind      string    `json:"kind"`
	Date      string    `json:"date"`
	Updated   string    `json:"updated"`
	Published bool      `json:"published"`
	Featured  bool      `json:"featured"`
	SortOrder int       `json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Link struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

type Color struct {
	Name string `json:"name"`
	Hex  string `json:"hex"`
}

type Font struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type Image struct {
	Path    string `json:"path"`
	Alt     string `json:"alt"`
	Caption string `json:"caption"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	IsCover bool   `json:"isCover"`
}

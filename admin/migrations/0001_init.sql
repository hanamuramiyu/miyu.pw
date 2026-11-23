
CREATE TABLE works (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    title       text NOT NULL,
    summary     text NOT NULL CHECK (btrim(summary) <> ''),
    body_md     text NOT NULL DEFAULT '',
    kind        text NOT NULL DEFAULT '',
    audience    text NOT NULL DEFAULT '',
    role        text NOT NULL DEFAULT '',
    client      text,
    date        date NOT NULL,
    updated     date NOT NULL,
    published   boolean NOT NULL DEFAULT false,
    featured    boolean NOT NULL DEFAULT false,
    sort_order  integer NOT NULL DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX works_published_idx ON works (published);
CREATE INDEX works_featured_sort_idx ON works (featured, sort_order);

CREATE TABLE work_tools (
    work_id   uuid NOT NULL REFERENCES works (id) ON DELETE CASCADE,
    name      text NOT NULL,
    position  integer NOT NULL,
    PRIMARY KEY (work_id, position)
);

CREATE TABLE work_links (
    work_id   uuid NOT NULL REFERENCES works (id) ON DELETE CASCADE,
    label     text NOT NULL,
    href      text NOT NULL,
    position  integer NOT NULL,
    PRIMARY KEY (work_id, position)
);

CREATE TABLE work_colors (
    work_id   uuid NOT NULL REFERENCES works (id) ON DELETE CASCADE,
    name      text NOT NULL,
    hex       text NOT NULL,
    position  integer NOT NULL,
    PRIMARY KEY (work_id, position)
);

CREATE TABLE work_fonts (
    work_id   uuid NOT NULL REFERENCES works (id) ON DELETE CASCADE,
    name      text NOT NULL,
    role      text NOT NULL,
    position  integer NOT NULL,
    PRIMARY KEY (work_id, position)
);

CREATE TABLE work_images (
    work_id   uuid NOT NULL REFERENCES works (id) ON DELETE CASCADE,
    path      text NOT NULL,
    alt       text NOT NULL DEFAULT '',
    caption   text,
    width     integer,
    height    integer,
    position  integer NOT NULL,
    is_cover  boolean NOT NULL DEFAULT false,
    PRIMARY KEY (work_id, position)
);

CREATE TABLE sessions (
    id          text PRIMARY KEY,
    created_at  timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL
);

CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);

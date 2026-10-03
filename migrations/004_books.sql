CREATE TABLE IF NOT EXISTS books (
    id         SERIAL       PRIMARY KEY,
    isbn       VARCHAR(17)  NOT NULL,
    title      VARCHAR(200) NOT NULL,
    author     VARCHAR(150) NOT NULL,
    publisher  VARCHAR(150) NOT NULL,
    year       INTEGER      NOT NULL,
    stock      INTEGER      NOT NULL DEFAULT 0,
    category   VARCHAR(50)  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS books_isbn_key     ON books (isbn);
CREATE INDEX        IF NOT EXISTS books_title_idx    ON books (title);
CREATE INDEX        IF NOT EXISTS books_category_idx ON books (category);

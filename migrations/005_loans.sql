CREATE TABLE IF NOT EXISTS loans (
    id          BIGSERIAL   PRIMARY KEY,
    book_id     INTEGER     NOT NULL REFERENCES books(id),
    user_id     INTEGER     NOT NULL REFERENCES users(id),
    loaned_by   INTEGER     NOT NULL REFERENCES users(id),
    loan_date   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_date    TIMESTAMPTZ NOT NULL,
    return_date TIMESTAMPTZ,
    status      VARCHAR(10) NOT NULL DEFAULT 'borrowed'
                CHECK (status IN ('borrowed', 'returned')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS loans_user_id_idx ON loans (user_id);
CREATE INDEX IF NOT EXISTS loans_book_id_idx ON loans (book_id);
CREATE INDEX IF NOT EXISTS loans_status_idx  ON loans (status);

-- Cursor pagination buku: urut berdasarkan title ASC, id ASC
-- (title unik secara praktis, tapi id sebagai tiebreaker)
-- Kondisi cursor: (title, id) > ($1, $2) karena urutan ASC
CREATE INDEX IF NOT EXISTS books_title_id_asc_idx
    ON books (title ASC, id ASC);

CREATE TABLE IF NOT EXISTS releases (
    id BIGSERIAL PRIMARY KEY,
    name TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    image VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS filehashes (
    id BIGSERIAL PRIMARY KEY,
    full_file_name TEXT NOT NULL,
    algorithm CHAR(16) NOT NULL,
    hash_sum VARCHAR NOT NULL,
    name_pod TEXT,
    release_id BIGINT,
    FOREIGN KEY (release_id) REFERENCES releases (id) ON DELETE CASCADE
);

CREATE INDEX filehashes_release_name ON releases (name);
CREATE INDEX filehashes_release_id ON filehashes (releases_id);

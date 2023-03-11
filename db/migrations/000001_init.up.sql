CREATE TABLE IF NOT EXISTS releases (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    release_type VARCHAR(32) NOT NULL,
    image VARCHAR(256) NOT NULL
);

CREATE TABLE IF NOT EXISTS filehashes (
    id BIGSERIAL PRIMARY KEY,
    full_file_name VARCHAR(256) NOT NULL,
    algorithm CHAR(16) NOT NULL,
    hash_sum VARCHAR(256) NOT NULL,
    name_pod VARCHAR(256) NOT NULL,
    release_id BIGINT NOT NULL,
    FOREIGN KEY (release_id) REFERENCES releases (id) ON DELETE CASCADE
);

CREATE INDEX filehashes_release_name ON releases (name);
CREATE INDEX filehashes_release_id ON filehashes (release_id);

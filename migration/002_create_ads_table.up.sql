CREATE TABLE ads (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    image_url TEXT NOT NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price > 0),
    author_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ads_author_id ON ads(author_id);
CREATE INDEX idx_ads_created_at ON ads(created_at);
CREATE INDEX idx_ads_price ON ads(price);
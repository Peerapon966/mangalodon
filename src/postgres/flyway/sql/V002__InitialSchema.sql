BEGIN;
  CREATE TABLE IF NOT EXISTS site (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    base_url VARCHAR(255) NOT NULL,
    path JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS scraper (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    site_id INT NOT NULL REFERENCES site(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  );

  DO $$
  BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'manga_status') THEN
      CREATE TYPE manga_status AS ENUM ('ongoing', 'completed');
    END IF;
  END
  $$;

  CREATE TABLE IF NOT EXISTS manga (
    id SERIAL PRIMARY KEY,
    title VARCHAR(512) NOT NULL UNIQUE,
    manga_id VARCHAR(512) NOT NULL UNIQUE,
    first_chapter NUMERIC(5,1) NOT NULL,
    last_chapter NUMERIC(5,1) NOT NULL,
    current_chapter NUMERIC(5,1) NOT NULL,
    status manga_status DEFAULT 'ongoing',
    chapter_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS source (
    id SERIAL PRIMARY KEY,
    manga_id INT NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    scraper_id INT NOT NULL REFERENCES scraper(id) ON DELETE CASCADE,
    url VARCHAR(512) NOT NULL,
    priority INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (manga_id, scraper_id, priority)
  );

  DO $$
  BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
      CREATE TYPE job_status AS ENUM ('unknown', 'working', 'completed', 'error');
    END IF;
  END
  $$;

  CREATE TABLE IF NOT EXISTS job (
    id SERIAL PRIMARY KEY,
    manga_id INT NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    scraper_id INT REFERENCES scraper(id) ON DELETE SET NULL,
    status job_status DEFAULT 'unknown',
    chapter NUMERIC(5,1) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (manga_id, chapter)
  );

  CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = CURRENT_TIMESTAMP;
        RETURN NEW;
  END;
  $$ language 'plpgsql';

  CREATE OR REPLACE TRIGGER update_site_updated_at
  BEFORE UPDATE ON site
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

  CREATE OR REPLACE TRIGGER update_scraper_updated_at
  BEFORE UPDATE ON scraper
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

  CREATE OR REPLACE TRIGGER update_manga_updated_at
  BEFORE UPDATE ON manga
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

  CREATE OR REPLACE TRIGGER update_source_updated_at
  BEFORE UPDATE ON source
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

  CREATE OR REPLACE TRIGGER update_job_updated_at
  BEFORE UPDATE ON job
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();
COMMIT;
import { Hono } from 'hono';
import { sql, sqlRead } from '../db.js';

export const mangasRoute = new Hono();

function slugify(text: string) {
  return text
    .toString()
    .toLowerCase()
    .trim()
    .replace(/\s+/g, '-')
    .replace(/[^\w\-]+/g, '')
    .replace(/\-\-+/g, '-');
}

mangasRoute.get('/', async (c) => {
  try {
    const page = parseInt(c.req.query('page') || '1', 10);
    const limit = 20;
    const offset = (page - 1) * limit;

    const mangas = await sqlRead`
      SELECT 
        m.*,
        (
          SELECT COALESCE(json_agg(j.*), '[]'::json)
          FROM job j
          WHERE j.manga_id = m.id AND j.status IN ('working', 'error')
        ) as active_jobs
      FROM manga m
      ORDER BY 
        (m.current_chapter < m.last_chapter) DESC,
        m.chapter_updated_at DESC NULLS LAST,
        m.title ASC
      LIMIT ${limit} OFFSET ${offset}
    `;

    return c.json(mangas);
  } catch (error) {
    console.error('Error fetching mangas:', error);
    return c.json({ error: 'Failed to fetch mangas' }, 500);
  }
});

mangasRoute.get('/:manga_id', async (c) => {
  const mangaId = c.req.param('manga_id');
  try {
    const [manga] = await sqlRead`
      SELECT 
        m.*,
        (
          SELECT COALESCE(json_agg(s.*), '[]'::json)
          FROM source s
          WHERE s.manga_id = m.id
        ) as sources
      FROM manga m
      WHERE m.manga_id = ${mangaId}
    `;

    if (!manga) {
      return c.json({ error: 'Manga not found' }, 404);
    }

    return c.json(manga);
  } catch (error) {
    console.error('Error fetching manga:', error);
    return c.json({ error: 'Failed to fetch manga' }, 500);
  }
});

mangasRoute.post('/', async (c) => {
  try {
    const body = await c.req.json();
    const { title, firstChapter, sources } = body;

    if (!title || firstChapter == null || !sources || !sources.length) {
      return c.json({ error: 'Missing required fields' }, 400);
    }

    const mangaId = slugify(title);

    const result = await sql.begin(async (tx) => {
      const [existing] = await tx`SELECT id FROM manga WHERE manga_id = ${mangaId}`;
      if (existing) {
        throw new Error('CONFLICT');
      }

      const [manga] = await tx`
        INSERT INTO manga (title, manga_id, first_chapter, last_chapter, current_chapter)
        VALUES (${title}, ${mangaId}, ${firstChapter}, ${firstChapter}, ${firstChapter})
        RETURNING *
      `;

      const insertedSources = [];
      for (const src of sources) {
        const [scraper] = await tx`SELECT id, url_format FROM scraper WHERE id = ${src.scraperId}`;
        if (!scraper) throw new Error(`Scraper ${src.scraperId} not found`);

        const formattedUrl = scraper.url_format.replace('${id}', src.sourceMangaId);

        const [source] = await tx`
          INSERT INTO source (manga_id, scraper_id, url, priority)
          VALUES (${manga.id}, ${scraper.id}, ${formattedUrl}, ${src.priority})
          RETURNING *
        `;
        insertedSources.push(source);
      }

      return { ...manga, sources: insertedSources };
    });

    return c.json(result, 201);
  } catch (error: any) {
    if (error.message === 'CONFLICT') {
      return c.json({ error: 'Manga already exists' }, 409);
    }
    console.error('Error creating manga:', error);
    return c.json({ error: 'Failed to create manga' }, 500);
  }
});

mangasRoute.delete('/:manga_id', async (c) => {
  const mangaId = c.req.param('manga_id');
  try {
    await sql`DELETE FROM manga WHERE manga_id = ${mangaId}`;
    return c.json({ success: true });
  } catch (error) {
    console.error('Error deleting manga:', error);
    return c.json({ error: 'Failed to delete manga' }, 500);
  }
});

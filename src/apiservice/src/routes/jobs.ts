import { Hono } from 'hono';
import { sql } from '../db.js';
import { broadcastJobStatus } from '../ws.js';

export const jobsRoute = new Hono();

jobsRoute.post('/webhook', async (c) => {
  try {
    const body = await c.req.json();
    const { mangaId, chapter, status, scraperId } = body;

    if (!mangaId || chapter == null || !status) {
      return c.json({ error: 'Missing required fields' }, 400);
    }

    const [manga] = await sql`SELECT id, last_chapter FROM manga WHERE manga_id = ${mangaId}`;
    if (!manga) {
      return c.json({ error: 'Manga not found' }, 404);
    }

    await sql`
      INSERT INTO job (manga_id, chapter, status, scraper_id)
      VALUES (${manga.id}, ${chapter}, ${status}, ${scraperId || null})
      ON CONFLICT (manga_id, chapter) 
      DO UPDATE SET status = EXCLUDED.status, scraper_id = EXCLUDED.scraper_id
    `;

    if (status === 'completed') {
      const numChapter = Number(chapter);
      if (numChapter > Number(manga.last_chapter)) {
        await sql`
          UPDATE manga 
          SET last_chapter = ${numChapter}
          WHERE id = ${manga.id}
        `;
      }
    }

    broadcastJobStatus(mangaId, chapter, status);

    return c.json({ success: true });
  } catch (error) {
    console.error('Error processing webhook:', error);
    return c.json({ error: 'Failed to process webhook' }, 500);
  }
});

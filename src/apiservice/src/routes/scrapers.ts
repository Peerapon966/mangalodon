import { Hono } from 'hono';
import { sql, sqlRead } from '../db.js';
import type { Scraper } from '../types.js';

export const scrapersRoute = new Hono();

scrapersRoute.get('/', async (c) => {
  try {
    const scrapers = await sqlRead<Scraper[]>`
      SELECT * FROM scraper ORDER BY id ASC
    `;
    return c.json(scrapers);
  } catch (error) {
    console.error('Error fetching scrapers:', error);
    return c.json({ error: 'Failed to fetch scrapers' }, 500);
  }
});

scrapersRoute.post('/', async (c) => {
  try {
    const body = await c.req.json();
    const { name, url_format, note } = body;

    if (!name || !url_format) {
      return c.json({ error: 'Missing required fields' }, 400);
    }

    const [scraper] = await sql<Scraper[]>`
      INSERT INTO scraper (name, url_format, note)
      VALUES (${name}, ${url_format}, ${note || null})
      RETURNING *
    `;

    return c.json(scraper, 201);
  } catch (error) {
    console.error('Error creating scraper:', error);
    return c.json({ error: 'Failed to create scraper' }, 500);
  }
});

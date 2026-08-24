import { Hono } from 'hono';
import { sql } from '../db.js';
import * as fs from 'fs/promises';
import * as path from 'path';

export const chaptersRoute = new Hono();

chaptersRoute.get('/:manga_id/chapters/:chapter', async (c) => {
  const mangaId = c.req.param('manga_id');
  const chapter = c.req.param('chapter');
  
  try {
    const dirPath = path.join('/assets', mangaId, chapter);
    const files = await fs.readdir(dirPath);
    
    const webpFiles = files
      .filter(file => file.endsWith('.webp'))
      .sort((a, b) => {
        const numA = parseInt(a.replace('.webp', ''), 10);
        const numB = parseInt(b.replace('.webp', ''), 10);
        return numA - numB;
      });

    const pages = webpFiles.map(file => `/assets/${mangaId}/${chapter}/${file}`);
    return c.json({ pages });
  } catch (error: any) {
    if (error.code === 'ENOENT') {
      return c.json({ error: 'Chapter not found or downloaded yet' }, 404);
    }
    console.error('Error reading chapter directory:', error);
    return c.json({ error: 'Internal server error' }, 500);
  }
});

chaptersRoute.put('/:manga_id/read', async (c) => {
  const mangaId = c.req.param('manga_id');
  
  try {
    const body = await c.req.json();
    const { chapter } = body;
    
    if (chapter == null) {
      return c.json({ error: 'Missing chapter number' }, 400);
    }

    const [updated] = await sql`
      UPDATE manga 
      SET current_chapter = ${chapter} 
      WHERE manga_id = ${mangaId}
      RETURNING id
    `;

    if (!updated) {
      return c.json({ error: 'Manga not found' }, 404);
    }

    return c.json({ success: true });
  } catch (error) {
    console.error('Error updating read progress:', error);
    return c.json({ error: 'Failed to update reading progress' }, 500);
  }
});

chaptersRoute.delete('/:manga_id/chapters/:chapter', async (c) => {
  const mangaId = c.req.param('manga_id');
  const chapter = c.req.param('chapter');
  
  try {
    const dirPath = path.join('/assets', mangaId, chapter);
    await fs.rm(dirPath, { recursive: true, force: true });
    
    await sql`
      DELETE FROM job 
      USING manga 
      WHERE job.manga_id = manga.id 
        AND manga.manga_id = ${mangaId} 
        AND job.chapter = ${chapter}
    `;

    return c.json({ success: true, message: 'Chapter deleted' });
  } catch (error) {
    console.error('Error deleting chapter:', error);
    return c.json({ error: 'Failed to delete chapter' }, 500);
  }
});

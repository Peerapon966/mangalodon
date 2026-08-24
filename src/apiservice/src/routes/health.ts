import { Hono } from 'hono';
import { sql } from '../db.js';

export const healthRoute = new Hono();

healthRoute.get('/', async (c) => {
  try {
    await sql`SELECT 1 as connected`;
    return c.json({ status: 'ok', database: 'connected' });
  } catch (error) {
    console.error('Database connection error:', error);
    return c.json({ status: 'error', database: 'disconnected' }, 500);
  }
});

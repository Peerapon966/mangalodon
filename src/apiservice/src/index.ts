import { serve } from '@hono/node-server';
import { Hono } from 'hono';
import { logger } from 'hono/logger';
import { cors } from 'hono/cors';

import { healthRoute } from './routes/health.js';
import { scrapersRoute } from './routes/scrapers.js';
import { mangasRoute } from './routes/mangas.js';
import { chaptersRoute } from './routes/chapters.js';
import { scrapeRoute } from './routes/scrape.js';
import { jobsRoute } from './routes/jobs.js';
import { initWebSocketServer } from './ws.js';

const app = new Hono();

app.use('*', async (c, next) => {
  if (c.req.path === '/api/health') {
    return next();
  }
  // Call the Hono logger middleware for all other routes
  return logger()(c, next);
});

app.use('*', cors());

app.route('/api/health', healthRoute);
app.route('/api/v1/scrapers', scrapersRoute);
app.route('/api/v1/mangas', mangasRoute);
app.route('/api/v1/mangas', chaptersRoute); // Mount chapters under mangas
app.route('/api/v1/scrape', scrapeRoute);
app.route('/api/v1/jobs', jobsRoute);

app.get('/', (c) => {
  return c.text('Mangalodon API Service is running!');
});

const port = 3000;
console.log(`Server is running on port ${port}`);

const server = serve({
  fetch: app.fetch,
  port
});

initWebSocketServer(server as any);

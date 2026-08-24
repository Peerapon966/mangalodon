import { Hono } from 'hono';
import amqplib from 'amqplib';

export const scrapeRoute = new Hono();

const RABBITMQ_URL = process.env.RABBITMQ_URL || 'amqp://localhost';
const QUEUE_NAME = process.env.RABBITMQ_QUEUE || 'scrape_queue';

let channel: amqplib.Channel | null = null;

async function initRabbitMQ() {
  try {
    const connection = await amqplib.connect(RABBITMQ_URL);
    channel = await connection.createChannel();
    await channel.assertQueue(QUEUE_NAME, { durable: true });
    console.log(`Connected to RabbitMQ, queue: ${QUEUE_NAME}`);
  } catch (error) {
    console.error('Failed to connect to RabbitMQ:', error);
  }
}

initRabbitMQ();

scrapeRoute.post('/', async (c) => {
  try {
    const body = await c.req.json();
    const { mangaId, chapter } = body;

    if (!mangaId || chapter == null) {
      return c.json({ error: 'Missing mangaId or chapter' }, 400);
    }

    if (!channel) {
      await initRabbitMQ();
      if (!channel) {
        return c.json({ error: 'Message queue unavailable' }, 503);
      }
    }

    const message = JSON.stringify({ mangaId, chapter });
    channel.sendToQueue(QUEUE_NAME, Buffer.from(message), { persistent: true });

    return c.json({ success: true, message: 'Scrape job queued' }, 202);
  } catch (error) {
    console.error('Error queuing job:', error);
    return c.json({ error: 'Failed to queue scrape job' }, 500);
  }
});

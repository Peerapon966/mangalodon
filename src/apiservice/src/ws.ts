import { WebSocketServer, WebSocket } from 'ws';
import type { Server } from 'node:http';

let wss: WebSocketServer | null = null;
const clients = new Set<WebSocket>();

export function initWebSocketServer(server: Server) {
  wss = new WebSocketServer({ server, path: '/ws' });

  wss.on('connection', (ws) => {
    clients.add(ws);
    
    ws.on('close', () => {
      clients.delete(ws);
    });

    ws.on('error', (err) => {
      console.error('WebSocket error:', err);
    });
  });
}

export function broadcastJobStatus(mangaId: string, chapter: number | string, status: string) {
  if (!wss) return;

  const payload = JSON.stringify({
    event: 'JOB_STATUS_UPDATED',
    data: {
      mangaId,
      chapter: Number(chapter),
      status
    }
  });

  for (const client of clients) {
    if (client.readyState === WebSocket.OPEN) {
      client.send(payload);
    }
  }
}

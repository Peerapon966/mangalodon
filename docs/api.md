# Mangalodon API Specification

## 1. Manga Management

### GET `/api/v1/mangas`
- **Description**: List mangas for the library view. 
  - **Sorting**: Mangas with unread chapters appear first (using a computed field like `current_chapter < last_chapter`), followed by a secondary sort of `chapter_updated_at DESC` (mangas that recently had a new chapter added), and finally alphabetically by `title`.
- **Query Params**: `?page=1` (20 items per page pagination)
- **Internal Logic**:
  1. Query `manga` table.
  2. `LEFT JOIN` on `job` table to fetch scraping statuses, but **only filter for `status IN ('working', 'error')`**. (Completed jobs are ignored by the frontend and eventually pruned by your cronjob).
  3. Sort by unread logic, return 20 items.

### GET `/api/v1/mangas/<manga_id>`
- **Description**: Get detailed manga information, including all available chapters from `first_chapter` to `last_chapter`.
- **Response**:
```json
{
  "id": 1,
  "title": "One Punch Man",
  "mangaId": "one-punch-man",
  "firstChapter": 1,
  "lastChapter": 224,
  "currentChapter": 223,
  "status": "ongoing",
  "sources": [
    {
      "scraperId": 1,
      "sourceMangaId": "one_punch",
      "priority": 1,
      "url": "..."
    }
  ]
}
```
- **Internal Logic**:
  1. Query `manga` table by `manga_id`.
  2. Return metadata and compute available chapters list.

### POST `/api/v1/mangas`
- **Description**: Register a new manga.
- **Request Body**:
```json
{
  "title": "One Punch Man",
  "firstChapter": 1,
  "sources": [
    {
      "scraperId": 1,
      "sourceMangaId": "one-punch-man",
      "priority": 1
    }
  ]
}
```
- **Internal Logic**:
  1. Receive `title`, `firstChapter`, and `sources` array.
  2. Generate slug (`manga_id`). Query DB. If exists, abort and return `409 Conflict`.
  3. Insert `manga` row.
  4. Loop through `sources`. Query `scraper` table to get `url_format` (e.g., `https://example.com/manga/${id}`).
  5. Replace the single `${...}` token in `url_format` with the provided `sourceMangaId` string.
  6. Save the final constructed URL into the `source` table.

### PUT `/api/v1/mangas/<manga_id>`
- **Description**: Update an existing manga's metadata, sources, or priority list.

### DELETE `/api/v1/mangas/<manga_id>`
- **Description**: Delete a manga completely.

---

## 2. Reading & Chapters

### GET `/api/v1/mangas/<manga_id>/chapters/<chapter>`
- **Description**: Fetch reading data for a specific chapter. Returns an array of available WEBP image URLs that the frontend can render.
- **Response**:
```json
{
  "pages": [
    "/assets/one-punch-man/137/1.webp",
    "/assets/one-punch-man/137/2.webp"
  ]
}
```
- **Internal Logic**:
  1. Read the local filesystem directory: `/assets/<manga_id>/<chapter>/`.
  2. Find all `.webp` files.
  3. Sort them numerically (`1.webp`, `2.webp`).
  4. Return array of constructed static URLs.

### PUT `/api/v1/mangas/<manga_id>/read`
- **Description**: Update the user's reading progress.
- **Request Body**:
```json
{
  "chapter": 137
}
```
- **Response**: `200 OK`
- **Internal Logic**:
  1. Execute `UPDATE manga SET current_chapter = :chapter WHERE manga_id = :manga_id`.

### DELETE `/api/v1/mangas/<manga_id>/chapters/<chapter>`
- **Description**: Delete a specific chapter (e.g., if scraping failed or pages are broken).

---

## 3. Scraping & Jobs

### POST `/api/v1/scrape`
- **Description**: Force-scrape a specific chapter immediately.
- **Request Body**:
```json
{
  "mangaId": "one-punch-man",
  "chapter": 137
}
```
- **Internal Logic**:
  1. Insert a job into the queue for immediate scraping.

### POST `/api/v1/jobs/webhook`
- **Description**: **Internal Webhook.** Go scraper calls this to update NestJS API when a job status changes.
- **Request Body**:
```json
{
  "mangaId": "one-punch-man",
  "chapter": 137,
  "status": "completed",
  "scraperId": 1
}
```
- **Internal Logic**:
  1. Update `job` table status.
  2. If `status == completed`, check if `chapter > manga.last_chapter`. If so, `UPDATE manga SET last_chapter = :chapter`.
  3. Broadcast WebSocket event to frontend to refresh UI.

### WebSocket `ws://<domain>/ws`
- **Description**: Frontend connects here to receive real-time updates when a scrape job changes status.
- **Event Payload**:
```json
{
  "event": "JOB_STATUS_UPDATED",
  "data": {
    "mangaId": "one-punch-man",
    "chapter": 137,
    "status": "completed"
  }
}
```

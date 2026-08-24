export type MangaStatus = 'ongoing' | 'completed';
export type JobStatus = 'unknown' | 'working' | 'completed' | 'error';

export interface Manga {
  id: number;
  title: string;
  manga_id: string;
  first_chapter: string | number;
  last_chapter: string | number;
  current_chapter: string | number;
  status: MangaStatus;
  chapter_updated_at: Date;
  created_at: Date;
  updated_at: Date;
}

export interface Scraper {
  id: number;
  name: string;
  note: string | null;
  url_format: string;
  created_at: Date;
  updated_at: Date;
}

export interface Source {
  id: number;
  manga_id: number;
  scraper_id: number;
  url: string;
  priority: number;
  created_at: Date;
  updated_at: Date;
}

export interface Job {
  id: number;
  manga_id: number;
  scraper_id: number | null;
  status: JobStatus;
  chapter: string | number;
  created_at: Date;
  updated_at: Date;
}

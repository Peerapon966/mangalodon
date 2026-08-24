import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';

export default function MangaDetail() {
  const { mangaId } = useParams();
  const [manga, setManga] = useState(null);

  useEffect(() => {
    fetch(`/api/v1/mangas/\${mangaId}`)
      .then(res => res.json())
      .then(data => setManga(data));
  }, [mangaId]);

  if (!manga) return <div className="loading">Loading...</div>;

  const chapters = [];
  for (let i = Number(manga.last_chapter); i >= Number(manga.first_chapter); i--) {
    chapters.push(i);
  }

  return (
    <div className="manga-detail">
      <h2>{manga.title}</h2>
      <div className="chapter-list">
        {chapters.map(ch => (
          <Link 
            to={`/manga/\${manga.manga_id}/chapter/\${ch}`} 
            key={ch} 
            className={`chapter-item \${Number(manga.current_chapter) === ch ? 'current' : ''}`}
          >
            Chapter {ch}
          </Link>
        ))}
      </div>
    </div>
  );
}

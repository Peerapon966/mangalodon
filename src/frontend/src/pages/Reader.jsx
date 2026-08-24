import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

export default function Reader() {
  const { mangaId, chapter } = useParams();
  const [pages, setPages] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    fetch(`/api/v1/mangas/\${mangaId}/chapters/\${chapter}`)
      .then(res => res.json())
      .then(data => setPages(data.pages || []));

    fetch(`/api/v1/mangas/\${mangaId}/read`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ chapter: Number(chapter) })
    });
  }, [mangaId, chapter]);

  return (
    <div className="reader">
      <div className="reader-nav">
        <button onClick={() => navigate(-1)} className="btn-back">Back</button>
        <span>Chapter {chapter}</span>
      </div>
      <div className="pages-container">
        {pages.map((url, i) => (
          <img key={i} src={url} alt={`Page \${i+1}`} className="manga-page" />
        ))}
        {pages.length === 0 && <p className="empty-msg">No pages found for this chapter. It might still be scraping or failed.</p>}
      </div>
    </div>
  );
}

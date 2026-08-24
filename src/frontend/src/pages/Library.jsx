import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { io } from 'socket.io-client';

export default function Library() {
  const [mangas, setMangas] = useState([]);

  useEffect(() => {
    fetch('/api/v1/mangas?page=1')
      .then(res => res.json())
      .then(data => setMangas(data));

    const socket = io();
    socket.on('JOB_STATUS_UPDATED', (data) => {
      setMangas(prev => prev.map(m => {
        if (m.manga_id === data.mangaId) {
          return { ...m, job_status: data.status };
        }
        return m;
      }));
    });
    
    return () => socket.disconnect();
  }, []);

  return (
    <div className="library">
      {mangas.map(manga => (
        <Link to={`/manga/\${manga.manga_id}`} key={manga.id} className="manga-card">
          <div className="cover-wrapper">
            <img src={`/assets/\${manga.manga_id}/cover.webp`} alt={manga.title} onError={(e) => e.target.src = '/placeholder.png'} />
            {manga.job_status === 'working' && <div className="job-badge working">Scraping...</div>}
            {manga.job_status === 'error' && <div className="job-badge error">Failed</div>}
          </div>
          <div className="info">
            <h3>{manga.title}</h3>
            <p>
              {manga.current_chapter} / {manga.last_chapter} 
              {Number(manga.current_chapter) < Number(manga.last_chapter) && <span className="unread-dot" />}
            </p>
          </div>
        </Link>
      ))}
      {mangas.length === 0 && <p className="empty-msg">Your library is empty. Add a manga to start reading!</p>}
    </div>
  );
}

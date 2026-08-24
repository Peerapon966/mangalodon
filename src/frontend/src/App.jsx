import React from 'react';
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import Library from './pages/Library';
import MangaDetail from './pages/MangaDetail';
import Reader from './pages/Reader';

import AddManga from './pages/AddManga';

function App() {
  return (
    <BrowserRouter>
      <div className="app-container">
        <header className="navbar">
          <h1>Mangalodon</h1>
          <nav>
            <Link to="/" className="nav-link">Library</Link>
            <Link to="/add" className="nav-link">Add Manga</Link>
          </nav>
        </header>
        <main className="main-content">
          <Routes>
            <Route path="/" element={<Library />} />
            <Route path="/add" element={<AddManga />} />
            <Route path="/manga/:mangaId" element={<MangaDetail />} />
            <Route path="/manga/:mangaId/chapter/:chapter" element={<Reader />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  );
}

export default App;

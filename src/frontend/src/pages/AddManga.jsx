import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

const AddManga = () => {
  const [title, setTitle] = useState('');
  const [firstChapter, setFirstChapter] = useState('');
  const [sources, setSources] = useState([
    { scraperId: '', sourceMangaId: '', priority: 1 }
  ]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  const handleAddSource = () => {
    setSources([...sources, { scraperId: '', sourceMangaId: '', priority: sources.length + 1 }]);
  };

  const handleRemoveSource = (index) => {
    const newSources = [...sources];
    newSources.splice(index, 1);
    setSources(newSources);
  };

  const handleSourceChange = (index, field, value) => {
    const newSources = [...sources];
    newSources[index][field] = value;
    setSources(newSources);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const payload = {
        title,
        firstChapter: Number(firstChapter),
        sources: sources.map(s => ({
          scraperId: Number(s.scraperId),
          sourceMangaId: s.sourceMangaId,
          priority: Number(s.priority)
        }))
      };

      const res = await fetch('/api/v1/mangas', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.message || 'Failed to add manga');
      }

      navigate('/');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="add-manga-container">
      <h2>Add New Manga</h2>
      {error && <div className="error-message">{error}</div>}
      
      <form onSubmit={handleSubmit} className="add-manga-form">
        <div className="form-group">
          <label>Title</label>
          <input 
            type="text" 
            value={title} 
            onChange={(e) => setTitle(e.target.value)} 
            placeholder="e.g. One Punch Man"
            required 
          />
        </div>

        <div className="form-group">
          <label>First Chapter</label>
          <input 
            type="number" 
            step="0.1"
            value={firstChapter} 
            onChange={(e) => setFirstChapter(e.target.value)} 
            placeholder="e.g. 1"
            required 
          />
        </div>

        <h3>Sources</h3>
        {sources.map((source, index) => (
          <div key={index} className="source-block">
            <div className="source-header">
              <h4>Source {index + 1}</h4>
              {sources.length > 1 && (
                <button type="button" className="btn-remove" onClick={() => handleRemoveSource(index)}>
                  &times;
                </button>
              )}
            </div>
            
            <div className="source-fields">
              <div className="form-group">
                <label>Scraper ID</label>
                <input 
                  type="number" 
                  value={source.scraperId} 
                  onChange={(e) => handleSourceChange(index, 'scraperId', e.target.value)} 
                  placeholder="e.g. 1"
                  required 
                />
              </div>

              <div className="form-group">
                <label>Source Manga ID (Slug)</label>
                <input 
                  type="text" 
                  value={source.sourceMangaId} 
                  onChange={(e) => handleSourceChange(index, 'sourceMangaId', e.target.value)} 
                  placeholder="e.g. one-punch-man"
                  required 
                />
              </div>

              <div className="form-group">
                <label>Priority</label>
                <input 
                  type="number" 
                  value={source.priority} 
                  onChange={(e) => handleSourceChange(index, 'priority', e.target.value)} 
                  placeholder="1 (Highest)"
                  required 
                />
              </div>
            </div>
          </div>
        ))}

        <button type="button" className="btn-secondary" onClick={handleAddSource}>
          + Add Another Source
        </button>

        <div className="form-actions">
          <button type="button" className="btn-cancel" onClick={() => navigate('/')}>Cancel</button>
          <button type="submit" className="btn-primary" disabled={loading}>
            {loading ? 'Adding...' : 'Save Manga'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default AddManga;

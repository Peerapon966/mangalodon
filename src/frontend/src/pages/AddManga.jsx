import React, { useState, useEffect, useContext, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { FormContext } from '../App';
import { DndContext, closestCenter, KeyboardSensor, PointerSensor, TouchSensor, useSensor, useSensors, DragOverlay } from '@dnd-kit/core';
import { arrayMove, SortableContext, sortableKeyboardCoordinates, verticalListSortingStrategy, useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';

const generateId = () => Math.random().toString(36).substr(2, 9);

const SourceItem = React.forwardRef(({ source, index, sourcesLength, handleRemoveSource, handleSourceChange, availableScrapers, style, isDragging, dragListeners, dragAttributes, isOverlay }, ref) => {
  return (
    <div ref={ref} style={style} className={`source-block ${isDragging ? 'dragging-dnd' : ''} ${isOverlay ? 'overlay-dnd' : ''}`}>
      <div className="source-header">
        <div className="source-title-group">
          <span className="drag-handle" title="Drag to reorder" {...dragAttributes} {...dragListeners}>≡</span>
          <h4>Source {index + 1}</h4>
        </div>
        <button 
          type="button" 
          className="btn-remove" 
          onClick={() => handleRemoveSource(index)}
          style={{ visibility: sourcesLength > 1 ? 'visible' : 'hidden' }}
          title="Remove Source"
        >
          &times;
        </button>
      </div>
      
      <div className="form-row grid-2">
        <div className="form-group">
          <label>Scraper</label>
          <select
            value={source.scraperId}
            onChange={(e) => handleSourceChange(index, 'scraperId', e.target.value)}
            required
          >
            <option value="">Select a scraper...</option>
            {availableScrapers.map(scraper => (
              <option key={scraper.id} value={scraper.id}>
                {scraper.name}
              </option>
            ))}
          </select>
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
      </div>
    </div>
  );
});

function SortableSourceItem({ source, index, sourcesLength, handleRemoveSource, handleSourceChange, availableScrapers }) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: source.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.3 : 1, // Dim the original item while dragging
  };

  return (
    <SourceItem
      ref={setNodeRef}
      style={style}
      source={source}
      index={index}
      sourcesLength={sourcesLength}
      handleRemoveSource={handleRemoveSource}
      handleSourceChange={handleSourceChange}
      availableScrapers={availableScrapers}
      dragListeners={listeners}
      dragAttributes={attributes}
    />
  );
}

const AddManga = () => {
  const [title, setTitle] = useState('');
  const [firstChapter, setFirstChapter] = useState('');
  const [sources, setSources] = useState([
    { id: generateId(), scraperId: '', sourceMangaId: '' }
  ]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [availableScrapers, setAvailableScrapers] = useState([]);
  const [activeId, setActiveId] = useState(null);
  const navigate = useNavigate();
  const { setIsDirty, requestNavigate } = useContext(FormContext);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 5,
      }
    }),
    useSensor(TouchSensor, {
      activationConstraint: {
        delay: 150,
        tolerance: 5,
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const sourceIds = useMemo(() => sources.map(s => s.id), [sources]);

  useEffect(() => {
    const fetchScrapers = async () => {
      try {
        const res = await fetch('/api/v1/scrapers');
        if (res.ok) {
          const data = await res.json();
          setAvailableScrapers(data);
        }
      } catch (err) {
        console.error('Failed to fetch scrapers', err);
      }
    };
    fetchScrapers();
  }, []);

  const handleAddSource = () => {
    setSources([...sources, { id: generateId(), scraperId: '', sourceMangaId: '' }]);
    setIsDirty(true);
  };

  const handleRemoveSource = (index) => {
    const newSources = [...sources];
    newSources.splice(index, 1);
    setSources(newSources);
    setIsDirty(true);
  };

  const handleSourceChange = (index, field, value) => {
    const newSources = [...sources];
    newSources[index][field] = value;
    setSources(newSources);
    setIsDirty(true);
  };

  const handleDragStart = (event) => {
    setActiveId(event.active.id);
  };

  const handleDragCancel = () => {
    setActiveId(null);
  };

  const handleDragEnd = (event) => {
    const { active, over } = event;
    setActiveId(null);
    
    if (over && active.id !== over.id) {
      setSources((items) => {
        const oldIndex = items.findIndex(i => i.id === active.id);
        const newIndex = items.findIndex(i => i.id === over.id);
        
        const newArray = arrayMove(items, oldIndex, newIndex);
        setIsDirty(true);
        return newArray;
      });
    }
  };

  const activeSource = activeId ? sources.find(s => s.id === activeId) : null;
  const activeIndex = activeId ? sources.findIndex(s => s.id === activeId) : -1;

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const payload = {
        title,
        firstChapter: Number(firstChapter),
        sources: sources.map((s, idx) => ({
          scraperId: Number(s.scraperId),
          sourceMangaId: s.sourceMangaId,
          priority: idx + 1
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

      setIsDirty(false); // Successfully saved
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
        <div className="form-row grid-2">
          <div className="form-group">
            <label>Title</label>
            <input 
              type="text" 
              value={title} 
              onChange={(e) => { setTitle(e.target.value); setIsDirty(true); }} 
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
              onChange={(e) => { setFirstChapter(e.target.value); setIsDirty(true); }} 
              placeholder="e.g. 1"
              required 
            />
          </div>
        </div>

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', marginBottom: '0.5rem', marginTop: '1.5rem' }}>
          <h3 style={{ margin: 0 }}>Sources</h3>
          <span style={{ fontSize: '0.85rem', color: '#94a3b8' }}>*Drag to reorder priority (Top is highest)</span>
        </div>

        <DndContext 
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
          onDragCancel={handleDragCancel}
        >
          <SortableContext 
            items={sourceIds}
            strategy={verticalListSortingStrategy}
          >
            {sources.map((source, index) => (
              <SortableSourceItem
                key={source.id}
                source={source}
                index={index}
                sourcesLength={sources.length}
                handleRemoveSource={handleRemoveSource}
                handleSourceChange={handleSourceChange}
                availableScrapers={availableScrapers}
              />
            ))}
          </SortableContext>
          
          <DragOverlay>
            {activeId && activeSource ? (
              <SourceItem
                source={activeSource}
                index={activeIndex}
                sourcesLength={sources.length}
                handleRemoveSource={() => {}}
                handleSourceChange={() => {}}
                availableScrapers={availableScrapers}
                isOverlay
                isDragging
              />
            ) : null}
          </DragOverlay>
        </DndContext>

        <button type="button" className="btn-add-source" onClick={handleAddSource}>
          + Add Another Source
        </button>

        <div className="form-actions">
          <button type="button" className="btn-cancel" onClick={() => requestNavigate('/')}>Cancel</button>
          <button type="submit" className="btn-primary" disabled={loading}>
            {loading ? 'Adding...' : 'Save Manga'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default AddManga;

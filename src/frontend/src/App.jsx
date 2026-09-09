import React, { useState, createContext, useContext } from 'react';
import { BrowserRouter, Routes, Route, Link, useLocation, useNavigate } from 'react-router-dom';
import Library from './pages/Library';
import MangaDetail from './pages/MangaDetail';
import Reader from './pages/Reader';
import AddManga from './pages/AddManga';

export const FormContext = createContext();

function FabNavigation() {
  const location = useLocation();
  const { requestNavigate } = useContext(FormContext);
  
  const isAddPage = location.pathname === '/add';

  const handleToggle = () => {
    if (isAddPage) {
      requestNavigate('/');
    } else {
      requestNavigate('/add');
    }
  };

  return (
    <button className="fab-nav" onClick={handleToggle} title={isAddPage ? "Back to Library" : "Add Manga"}>
      {isAddPage ? (
        <svg className="fab-icon" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      ) : (
        <svg className="fab-icon" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
      )}
    </button>
  );
}

function NavigationWrapper({ children }) {
  const [isDirty, setIsDirty] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [pendingPath, setPendingPath] = useState(null);
  const navigate = useNavigate();

  const requestNavigate = (path) => {
    if (isDirty) {
      setPendingPath(path);
      setShowModal(true);
    } else {
      navigate(path);
    }
  };

  const confirmLeave = () => {
    setShowModal(false);
    setIsDirty(false);
    navigate(pendingPath);
  };

  const cancelLeave = () => {
    setShowModal(false);
    setPendingPath(null);
  };

  return (
    <FormContext.Provider value={{ isDirty, setIsDirty, requestNavigate }}>
      {children}
      {showModal && (
        <div className="modal-overlay">
          <div className="modal-content">
            <h3>Unsaved Changes</h3>
            <p>You have unsaved changes. Are you sure you want to leave? Your edits will be lost.</p>
            <div className="modal-actions">
              <button className="btn-secondary" onClick={cancelLeave}>Stay</button>
              <button className="btn-danger" onClick={confirmLeave}>Leave</button>
            </div>
          </div>
        </div>
      )}
    </FormContext.Provider>
  );
}

function App() {
  return (
    <BrowserRouter>
      <NavigationWrapper>
        <div className="app-container">
          <header className="navbar">
            <h1>Mangalodon</h1>
          </header>
          <main className="main-content">
            <Routes>
              <Route path="/" element={<Library />} />
              <Route path="/add" element={<AddManga />} />
              <Route path="/manga/:mangaId" element={<MangaDetail />} />
              <Route path="/manga/:mangaId/chapter/:chapter" element={<Reader />} />
            </Routes>
          </main>
          <FabNavigation />
        </div>
      </NavigationWrapper>
    </BrowserRouter>
  );
}

export default App;

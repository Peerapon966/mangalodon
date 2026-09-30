BEGIN;
  INSERT INTO site (id, name, base_url, path) VALUES 
    (1, 'NEKOPOST', 'https://www.nekopost.net', '{"recommend": "/manga", "series": "/manga/<sid>", "chapter": "/manga/<sid>/<cid>"}'),
    (2, 'NICEOPPAI', 'https://www.niceoppai.net', '{"recommend": "/", "series": "/<sid>", "chapter": "/<sid>/<cid>"}'),
    (3, 'MangaKakalot', 'https://www.mangakakalot.gg', '{"recommend": "/", "series": "/manga/<sid>", "chapter": "/manga/<sid>/<cid>"}');  

  INSERT INTO scraper (id, name, site_id, url_regex) VALUES 
    (1, 'NEKOPOST', 1, '(?:https:\/\/www\.)?nekopost\.net\/manga\/([^/]+)'),
    (2, 'NICEOPPAI', 2, '(?:https:\/\/www\.)?niceoppai\.net\/([^/]+)'),
    (3, 'MangaKakalot', 3, '(?:https:\/\/www\.)?mangakakalot\.gg\/manga\/([^/]+)');

  INSERT INTO manga (id, title, manga_id, first_chapter, last_chapter, current_chapter, status, poster) VALUES 
    (1, 'One Punch Man', 'one-punch-man', 130, 217, 217, 'ongoing', NULL);
  
  INSERT INTO source (id, manga_id, scraper_id, series_id, priority) VALUES 
    (1, 1, 2, 'One-Punch-Man', 210, 1),
    (2, 1, 3, 'one-punch-man', 217, 2);
COMMIT;
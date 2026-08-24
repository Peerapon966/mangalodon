const fs = require('fs');
let text = fs.readFileSync('index.css', 'utf8');
text = text.replace(/\x00/g, '');
if (text.charCodeAt(0) === 0xFEFF) {
  text = text.slice(1);
}
text = text.replace(/\uFEFF/g, '');
fs.writeFileSync('index.css', text, 'utf8');

const app = require('./app');
const dotenv = require('dotenv');
const http = require('http');
const fs = require('fs');

dotenv.config();

const PORT = process.env.PORT || 3000;

// const options = {
//   key: fs.readFileSync(process.env.CERT_KEY),
//   cert: fs.readFileSync(process.env.CERT_PEM),
// };

const server = http.createServer(app);

server.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`);
});

const http = require('http');

const port = 3333;
 
const server = http.createServer((req, res) => {
  const foo = process.env.FOO;
  console.log(`Sending response. FOO=${foo}`);
  res.statusCode = 200;
  res.setHeader('Content-Type', 'text/plain');
  res.end('Hello World');
});
 
server.listen(port, () => {
  console.log(`Server running at port ${port}`);
});

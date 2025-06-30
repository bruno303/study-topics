const express = require('express');
const apiRouter = require('./routes/api');

const app = express();

// Middleware
app.use(express.json());

// Routes
app.use('/api', apiRouter);

// 404 Handler
app.use((req, res) => {
  res.status(404).json({ message: 'Not Found' });
});

module.exports = app;

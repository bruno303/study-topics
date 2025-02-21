const express = require("express");
const { getMovies } = require("./repository");
const { PORT, VIDEO_DIR, HLS_DIR } = require("./constants");
const { isProcessed, processMovie, getAudioTracks } = require("./movie");

const app = express();

app.post("/process/:movie", async (req, res) => {
  const movieName = req.params.movie;
  const movies = getMovies(VIDEO_DIR);
  const movie = movies.find(m => m.name === movieName);

  if (!movie) {
      return res.status(404).json({ error: "Movie not found" });
  }

  try {
    console.log(`Starting FFmpeg processing...`);
    await processMovie(movie);
    res.status(200).json({ url: `/hls/${movie.name}/master.m3u8` });
  } catch (error) {
      res.status(500).json({ error: "Error processing movie" });
  }
});

// API to process & play a movie
app.get("/play/:movie", async (req, res) => {
  const movieName = req.params.movie;
  const movies = getMovies(VIDEO_DIR);
  const movie = movies.find(m => m.name === movieName);

  if (!movie) {
      return res.status(404).json({ error: "Movie not found" });
  }

  try {
      if (!isProcessed(movie)) {
          console.log(`Movie not processed, starting FFmpeg...`);
          await processMovie(movie);
      }
      res.json({ url: `/hls/${movie.name}/master.m3u8` });
  } catch (error) {
      res.status(500).json({ error: "Error processing movie" });
  }
});

// API to list all movies
app.get("/movies", async (req, res) => {
  const movies = getMovies(VIDEO_DIR).map(movie => ({
      name: movie.name,
      path: movie.path,
      hlsPath: movie.hlsPath,
      subtitles: movie.subtitles,
      audioTracks: [],
  }));

  for (const movie of movies) {
      movie.audioTracks = await getAudioTracks(movie.path);
  }

  res.json(movies);
});

// Serve HLS files
app.use("/hls", express.static(HLS_DIR));
app.use("/subtitles", express.static(VIDEO_DIR));
app.use(express.static("static"));

app.listen(PORT, () => {
    console.log(`Server running at http://localhost:${PORT}`);
});

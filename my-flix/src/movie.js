const fs = require("fs");
const path = require("path");
const { spawn } = require("child_process");
const { VIDEO_DIR } = require("./constants");

function processMovie(movie) {
  return new Promise((resolve, reject) => {
      console.log(`Processing movie: ${movie.name}...`);

      fs.mkdirSync(movie.hlsPath, { recursive: true });

      const outputPlaylist = path.join(movie.hlsPath, "master.m3u8");

      // const ffmpeg = spawn("ffmpeg", [
      //     "-i", movie.path,
      //     "-map", "0:v:0", // Video track
      //     "-map", "0:a?",  // All audio tracks
      //     "-map", "0:s?",  // All subtitle tracks
      //     "-codec:v", "libx264",
      //     "-codec:a", "aac",
      //     "-preset", "veryfast",
      //     "-hls_time", "10",
      //     "-hls_list_size", "0",
      //     "-hls_flags", "independent_segments",
      //     "-hls_playlist_type", "vod",
      //     "-f", "hls",
      //     outputPlaylist
      // ]);

      // Convert subtitles to .vtt format
      movie.subtitles.map((sub) => {
          const vttFile = path.join(VIDEO_DIR, sub.replace("/subtitles/", ""));
          const srtFile = path.join(VIDEO_DIR, sub.replace("/subtitles/", "").replace(".vtt", ".srt"));
          const ffmpegSub = spawn("ffmpeg", [
              "-i", srtFile, vttFile
          ]);

          ffmpegSub.stderr.on("data", (data) => console.log(`FFmpeg Subtitles: ${data}`));
          ffmpegSub.on("close", (code) => {
              if (code !== 0) console.error("Error converting subtitles");
          });
      });

      // ffmpeg.stderr.on("data", (data) => console.log(`FFmpeg: ${data}`));
      // ffmpeg.on("close", (code) => {
      //     if (code === 0) {
      //         console.log(`Finished processing: ${movie.name}`);
      //         resolve();
      //     } else {
      //         console.error(`Error processing ${movie.name}`);
      //         reject(new Error(`FFmpeg exited with code ${code}`));
      //     }
      // });
  });
}

function isProcessed(movie) {
  return fs.existsSync(path.join(movie.hlsPath, "master.m3u8"));
}

function getAudioTracks(filePath) {
  return new Promise((resolve) => {
      const ffmpeg = spawn("ffmpeg", ["-i", filePath]);

      let output = "";
      ffmpeg.stderr.on("data", (data) => (output += data.toString()));

      ffmpeg.on("close", () => {
          const audioTracks = [];
          const regex = /Stream #\d+:(\d+).*: Audio:.*\((.*?)\)/g;
          let match;
          while ((match = regex.exec(output)) !== null) {
              audioTracks.push({ index: match[1], language: match[2] });
          }
          resolve(audioTracks);
      });
  });
}

module.exports = { processMovie, isProcessed, getAudioTracks };

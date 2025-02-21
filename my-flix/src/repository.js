const fs = require("fs");
const path = require("path");
const { VIDEO_DIR, HLS_DIR } = require("./constants");
const { subscribe } = require("diagnostics_channel");

function getMovies(dir, fileList = []) {
  const files = fs.readdirSync(dir);
  
  files.forEach((file) => {
      const fullPath = path.join(dir, file);
      const stat = fs.statSync(fullPath);

      if (stat.isDirectory()) {
          getMovies(fullPath, fileList);
      } else if (file.endsWith(".mp4")) {
          // Find subtitles with the same name
          const baseName = file.replace(".mp4", "");
          const subtitles = fs
              .readdirSync(dir)
              .filter(sub => sub.startsWith(baseName) && sub.endsWith(".vtt"))
              .map(sub => `/subtitles/${path.relative(VIDEO_DIR, path.join(dir, sub))}`);

          fileList.push({
              name: path.basename(file, ".mp4"),
              path: fullPath,
              hlsPath: path.join(HLS_DIR, path.basename(file, ".mp4")),
              subtitles: subtitles,
          });
      }
  });

  return fileList;
}

module.exports = { getMovies };

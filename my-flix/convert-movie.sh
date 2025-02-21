#!/bin/bash

INPUT_FILE="/mnt/c/Users/Pichau/Desktop/Game Of Thrones - 1ª Temporada (2011) 720p Dual Áudio - Douglasvip/S01E01 - Winter Is Coming.mp4"

# ffmpeg -i "$INPUT_FILE" \
#    -map 0:v:0 -map 0:a:0 -map 0:a:1 \
#    -c:v libx264 -crf 23 -preset fast -c:a aac -b:a 128k \
#    -c:s webvtt -f hls -hls_list_size 0 -hls_time 10 \
#    -var_stream_map "v:0,a:0,a:1" \
#    -master_pl_name master.m3u8 -y "$(pwd)/hls/S01E01 - Winter Is Coming/output_%v.m3u8"



ffmpeg -i "$INPUT_FILE" \
   -map v:0 -map a:0 -map a:1 \
   -metadata:s:a:0 language=pt-BR -metadata:s:a:1 language=eng \
   -metadata:s:a:0 title="Português (Brasil)" -metadata:s:a:1 title="English" \
   -var_stream_map "v:0,a:0 a:1" \
   -f hls -hls_time 10 -hls_playlist_type vod \
   -hls_flags independent_segments \
   -hls_segment_filename "$(pwd)/hls/2/output_%v_%03d.ts" \
   -master_pl_name "master.m3u8" \
   "$(pwd)/hls/2/output_%v.m3u8"

# ffmpeg -i "$INPUT_FILE" \
#    -map 0:v:0 -c:v libx264 -crf 23 -preset fast -f hls -hls_time 10 -hls_list_size 0 -hls_segment_filename "$(pwd)/hls/3/output_v%03d.ts" "$(pwd)/hls/3/output_v.m3u8" \
#    -map 0:a:0 -c:a aac -b:a 128k -f hls -hls_time 10 -hls_list_size 0 -hls_segment_filename "$(pwd)/hls/3/output_a0_%03d.ts" "$(pwd)/hls/3/output_a0.m3u8" \
#    -map 0:a:1 -c:a aac -b:a 128k -f hls -hls_time 10 -hls_list_size 0 -hls_segment_filename "$(pwd)/hls/3/output_a1_%03d.ts" "$(pwd)/hls/3/output_a1.m3u8" \
#    -master_pl_name "$(pwd)/hls/3/master.m3u8"

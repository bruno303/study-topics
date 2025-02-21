#!/bin/bash

SRT_SUBITTLE=""
SRT_SUBTITLE_OUTPUT=""
VTT_FILE=""

## Convert subtitle to UTF-8
iconv -f ISO-8859-1 -t UTF-8 "$SRT_SUBITTLE" -o "$SRT_SUBTITLE_OUTPUT"

## Convert SRT to VTT
ffmpeg -sub_charenc UTF-8 -i "$SRT_SUBTITLE_OUTPUT" -c:s webvtt "$VTT_FILE"

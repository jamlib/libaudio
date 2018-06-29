package ffmpeg

import (
  "fmt"
  "bytes"
  "errors"
  "os/exec"
)

type ffmpeg struct {
  Bin string
}

type Metadata struct {
  Artist string
  Album string
  Disc string
  Track string
  Title string
  Date string
  Artwork string
}

// new ffmpeg wrapper where args can be added
func New() (*ffmpeg, error) {
  bin, err := exec.LookPath("ffmpeg")
  if err != nil {
    return &ffmpeg{}, errors.New("ffmpeg not found on system\n")
  }
  return &ffmpeg{ Bin: bin }, nil
}

// run ffmpeg (capture stdout & stderr)
func (f *ffmpeg) Exec(args ...string) (string, error) {
  exec := exec.Command(f.Bin, args...)

  var out bytes.Buffer
  var stderr bytes.Buffer
  exec.Stdout = &out
  exec.Stderr = &stderr

  err := exec.Run()
  if err != nil {
    return "", errors.New(fmt.Sprint(err) + ": " + stderr.String())
  }
  return out.String(), nil
}

// optimize image as embedded album art
func (f *ffmpeg) OptimizeAlbumArt(input, output string) (string, error) {
  return f.Exec([]string{ "-i", input, "-y", "-qscale:v", "2",
    "-vf", "scale=500:-1", output }...)
}

// convert lossless to mp3
func (f *ffmpeg) ToMp3(input, quality string, meta Metadata, output string) (string, error) {
  a := []string{ "-i", input }

  if len(meta.Artwork) > 0 {
    a = append(a, "-i", meta.Artwork)
  }

  // mp3 audio codec
  a = append(a, "-map", "0:a", "-c:a")
  if quality == "copy" {
    a = append(a, "copy")
  } else {
    a = append(a, "libmp3lame")

    // mp3 audio quality
    if quality == "320" {
      a = append(a, "-b:a", "320k")
    } else {
      a = append(a, "-qscale:a", "0")
    }
  }

  // id3v2 metadata
  a = append(a, "-id3v2_version", "4")
  a = append(a, "-metadata", "artist=" + meta.Artist)
  a = append(a, "-metadata", "album=" + meta.Album)
  a = append(a, "-metadata", "disc=" + meta.Disc)
  a = append(a, "-metadata", "track=" + meta.Track)
  a = append(a, "-metadata", "title=" + meta.Title)
  a = append(a, "-metadata", "date=" + meta.Date)

  // embedd album artwork
  if len(meta.Artwork) > 0 {
    a = append(a, "-map", "1:v", "-c:v", "copy", "-metadata:s:v",
      "title=Album cover", "-metadata:s:v", "comment=Cover (Front)")
  }

  a = append(a, "-y", output)
  return f.Exec(a...)
}

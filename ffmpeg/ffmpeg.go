package ffmpeg

import (
  "fmt"
  "bytes"
  "errors"
  "os/exec"
)

type ffmpeg struct {
  *exec.Cmd
}

type Metadata struct {
  Artist string
  Album string
  Title string
  Track string
  Date string
  Artwork string
}

// check that ffmpeg is installed on system
func Which() (string, error) {
  bin, err := exec.LookPath("ffmpeg")
  if err != nil {
    return "", errors.New("ffmpeg not found on system\n")
  }
  return bin, nil
}

// new ffmpeg wrapper where args can be added
func New(args ...string) *ffmpeg {
  bin, _ := Which()
  return &ffmpeg{exec.Command(bin, args...)}
}

// add additional arguments
func (f *ffmpeg) AddArgs(args ...string) {
  f.Args = append(f.Args, args...)
}

// run ffmpeg (capture stdout & stderr)
func (f *ffmpeg) Exec() (string, error) {
  var out bytes.Buffer
  var stderr bytes.Buffer
  f.Stdout = &out
  f.Stderr = &stderr

  err := f.Run()
  if err != nil {
    return "", errors.New(fmt.Sprint(err) + ": " + stderr.String())
  }
  return out.String(), nil
}

// optimize image as embedded album art
func OptimizeAlbumArt(input, output string) (string, error) {
  return New("-i", input, "-y", "-qscale:v", "2",
    "-vf", "scale=500:-1", output).Exec()
}

// convert lossless to mp3
func ToMp3(input, quality string, meta Metadata, output string) (string, error) {
  f := New("-i", input)

  if len(meta.Artwork) > 0 {
    f.AddArgs("-i", meta.Artwork)
  }

  // mp3 audio codec
  f.AddArgs("-map", "0:a", "-codec:a", "libmp3lame")

  // mp3 audio quality
  if quality == "320" {
    f.AddArgs("-b:a", "320k")
  } else {
    f.AddArgs("-qscale:a", "0")
  }

  // id3v2 metadata
  f.AddArgs("-id3v2_version", "4")
  f.AddArgs("-metadata", "artist=" + meta.Artist)
  f.AddArgs("-metadata", "album=" + meta.Album)
  f.AddArgs("-metadata", "title=" + meta.Title)
  f.AddArgs("-metadata", "track=" + meta.Track)
  f.AddArgs("-metadata", "date=" + meta.Date)

  // embedd album artwork
  if len(meta.Artwork) > 0 {
    f.AddArgs("-map", "1:v", "-c:v", "copy", "-metadata:s:v", "title=Album cover",
      "-metadata:s:v", "comment=Cover (Front)")
  }

  f.AddArgs("-y", output)
  return f.Exec()
}

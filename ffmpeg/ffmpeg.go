package ffmpeg

import (
  "os"
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

type Mp3Config struct {
  Input, Quality, Output string
  Meta Metadata
  Fix bool
}

// mp3 quality helper function
func (f *ffmpeg) mp3Quality(q string) []string {
  a := []string{}

  if q == "copy" {
    a = append(a, "copy")
  } else {
    a = append(a, "libmp3lame")

    // mp3 audio quality
    if q == "320" {
      a = append(a, "-b:a", "320k")
    } else {
      a = append(a, "-qscale:a", "0")
    }
  }

  return a
}

// convert to mp3
func (f *ffmpeg) ToMp3(c *Mp3Config) (string, error) {
  a := []string{ "-i" }

  // if track length displays outrageous number like 1035:36:51
  // copy w/o metadata, then add metadata fixes it
  fixOut := c.Output[:len(c.Output)-4] + "-fix.mp3"
  if c.Fix {
    b := []string{ "-i", c.Input, "-map_metadata", "-1", "-c:a" }
    b = append(b, f.mp3Quality(c.Quality)...)
    b = append(b, "-y", fixOut)

    s, err := f.Exec(b...)
    if err != nil {
      return s, err
    }

    // next input is fixed output
    a = append(a, fixOut)

    // do not need to convert again
    c.Quality = "copy"
  } else {
    a = append(a, c.Input)
  }

  if len(c.Meta.Artwork) > 0 {
    a = append(a, "-i", c.Meta.Artwork)
  }

  // mp3 audio codec
  a = append(a, "-map", "0:a", "-c:a")
  a = append(a, f.mp3Quality(c.Quality)...)

  // id3v2 metadata
  a = append(a, "-id3v2_version", "4")
  a = append(a, "-metadata", "artist=" + c.Meta.Artist)
  a = append(a, "-metadata", "album=" + c.Meta.Album)
  a = append(a, "-metadata", "disc=" + c.Meta.Disc)
  a = append(a, "-metadata", "track=" + c.Meta.Track)
  a = append(a, "-metadata", "title=" + c.Meta.Title)
  a = append(a, "-metadata", "date=" + c.Meta.Date)

  // embedd album artwork
  if len(c.Meta.Artwork) > 0 {
    a = append(a, "-map", "1:v", "-c:v", "copy", "-metadata:s:v",
      "title=Album cover", "-metadata:s:v", "comment=Cover (Front)")
  }

  a = append(a, "-y", c.Output)
  s, err := f.Exec(a...)

  if c.Fix && err == nil {
    err = os.Remove(fixOut)
  }

  return s, err
}

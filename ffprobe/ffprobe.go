package ffprobe

import (
  "bytes"
  "os/exec"
  "encoding/json"
)

type Ffprober interface {
  GetData(filePath string) (*Data, error)
  EmbeddedImage() (int, int, bool)
}

type ffprobe struct {
  Bin string
  Data *Data
}

type Data struct {
  Streams            []*Stream   `json:"streams"`
  Format             *Format     `json:"format"`
}

type Stream struct {
  // common
  Index              int         `json:"index"`
  CodecName          string      `json:"codec_name"`
  CodecLongName      string      `json:"codec_long_name"`
  CodecType          string      `json:"codec_type"`
  CodecTimeBase      string      `json:"codec_time_base"`
  CodecTagString     string      `json:"codec_tag_string"`
  CodecTag           string      `json:"codec_tag"`
  // audio
  Channels           int         `json:"channels"`
  ChannelLayout      string      `json:"channel_layout"`
  SampleFmt          string      `json:"sample_fmt"`
  SampleRate         string      `json:"sample_rate"`
  TimeBase           string      `json:"time_base"`
  StartPts           int         `json:"start_pts"`
  StartTime          string      `json:"start_time"`
  DurationTs         uint64      `json:"duration_ts"`
  Duration           string      `json:"duration"`
  // audio MP3
  BitRate            string      `json:"bit_rate"`
  // audio FLAC
  BitsPerRawSample   string      `json:"bits_per_raw_sample"`
  // image
  Width              int         `json:"width"`
  Height             int         `json:"height"`
  PixFmt             string      `json:"pix_fmt"`
}

type Format struct {
  Filename           string      `json:"filename"`
  NumStreams         int         `json:"nb_streams"`
  NumPrograms        int         `json:"nb_programs"`
  FormatName         string      `json:"format_name"`
  FormatLongName     string      `json:"format_long_name"`
  StartTime          float64     `json:"start_time,string"`
  Duration           float64     `json:"duration,string"`
  Size               string      `json:"size"`
  BitRate            string      `json:"bit_rate"`
  Score              int         `json:"probe_score"`
  Tags               *Tags       `json:"tags"`
}

type Tags struct {
  Album              string      `json:"album"`
  Artist             string      `json:"artist"`
  Comment            string      `json:"comment"`
  Date               string      `json:"date"`
  Genre              string      `json:"genre"`
  Disc               string      `json:"disc"`
  DiscTotal          string      `json:"disctotal"`
  Encoder            string      `json:"encoder"`
  Title              string      `json:"title"`
  Track              string      `json:"track"`
  TrackTotal         string      `json:"tracktotal"`
}

func New() (*ffprobe, error) {
  // check that ffprobe is installed on system
  bin, err := exec.LookPath("ffprobe")
  if err != nil {
    return &ffprobe{}, err
  }
  return &ffprobe{ Bin: bin }, nil
}

func (f *ffprobe) GetData(filePath string) (*Data, error) {
  data := &Data{}

  cmd := exec.Command(
    f.Bin, "-v", "quiet", "-print_format", "json",
    "-show_streams", "-show_format", filePath,
  )
  var out bytes.Buffer
  cmd.Stdout = &out

  err := cmd.Run()
  if err != nil {
    return data, err
  }

  err = json.Unmarshal(out.Bytes(), data)
  if err != nil {
    return data, err
  }

  f.Data = data
  return data, nil
}

// if embedded image, return width, height
func (f *ffprobe) EmbeddedImage() (int, int, bool) {
  if len(f.Data.Streams) > 1 && f.Data.Streams[1].Width > 0 &&
    f.Data.Streams[1].Height > 0 {
    return f.Data.Streams[1].Width, f.Data.Streams[1].Height, true
  }
  return 0, 0, false
}

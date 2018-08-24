package ffprobe

import (
  "io/ioutil"
  "encoding/json"
)

type MockFfprobe struct {
  Width int
  Embedded string
}

func (m *MockFfprobe) EmbeddedImage() (int, int, bool) {
  if len(m.Embedded) > 0 {
    return m.Width, 0, true
  }
  return 0, 0, false
}

func (m *MockFfprobe) GetData(filePath string) (*Data, error) {
  d := &Data{ Format: &Format{ Tags: &Tags{} } }

  raw, err := ioutil.ReadFile(filePath)
  if err != nil {
    return d, err
  }

  err = json.Unmarshal(raw, d.Format.Tags)
  if err != nil {
    return d, err
  }

  return d, nil
}

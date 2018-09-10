package fsutil

import (
  "os"
  "strings"
  "testing"
  "io/ioutil"
  "path/filepath"
)

type TestFile struct {
  Name, Contents string
}

func CreateTestFiles(t *testing.T, files []*TestFile) (string, []string) {
  td, err := ioutil.TempDir("", "")
  if err != nil {
    t.Fatal(err)
  }

  paths := []string{}
  for i := range files {
    if len(files[i].Name) == 0 {
      continue
    }

    pa := strings.Split(files[i].Name, "/")
    p := filepath.Join(td, filepath.Join(pa[:len(pa)-1]...))

    // create parent dirs
    if len(pa) > 1 {
      err := os.MkdirAll(p, 0777)
      if err != nil {
        t.Fatal(err)
      }
    }

    // create file
    if len(pa[len(pa)-1]) > 0 {
      fullpath := filepath.Join(p, pa[len(pa)-1])
      err := ioutil.WriteFile(fullpath, []byte(files[i].Contents), 0644)
      if err != nil {
        t.Fatal(err)
      }
    }

    // append to paths
    paths = append(paths, filepath.Join(td, files[i].Name))
  }

  return td, paths
}

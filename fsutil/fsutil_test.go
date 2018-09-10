package fsutil

import (
  "os"
  "errors"
  "strings"
  "testing"
  "io/ioutil"
  "path/filepath"
)

func testFilesFullPath(t *testing.T, f func(dir string, files []string)) {
  newFiles := []*TestFile{
    {"file1", "abcde"},
    {"file2.jpeg", "a"},
    {"dir1/file3.JPG", "acddfefsefd"},
    {"dir1/dir2/file4.png", "dfadfd"},
  }

  dir := CreateTestFiles(t, newFiles)
  defer os.RemoveAll(dir)

  files := []string{}
  for i := range newFiles {
    files = append(files, filepath.Join(dir, newFiles[i].Name))
  }

  f(dir, files)
}


func TestCopyFile(t *testing.T) {
  testFilesFullPath(t, func(dir string, files []string) {
    // destination dir
    td, err := ioutil.TempDir("", "")
    if err != nil {
      t.Fatal(err)
    }
    defer os.RemoveAll(td)

    tests := []struct {
      src, dest string
      error error
    }{
      { src: files[2], dest: filepath.Join(td, "file1"), error: nil },
      { src: files[3], dest: filepath.Join(td, "file3"), error: nil },
      { src: "audiocc-file-def-dne", dest: files[1],
        error: errors.New("audiocc-file-def-dne"),
      },
    }

    for i := range tests {
      e := CopyFile(tests[i].src, tests[i].dest)
      if e == nil && tests[i].error == nil {
        break
      }
      if e.Error() != tests[i].error.Error() {
        t.Errorf("Expected %#v, got %#v", tests[i].error.Error(), e.Error())
      }
    }
  })
}

func TestFilesByExtensionImages(t *testing.T) {
  files := []*TestFile{
    {"file1", ""},
    {"file2.jpeg", ""},
    {"dir1/file3.JPG", ""},
    {"dir1/dir2/file4.png", ""},
  }

  result := []string{
    "dir1/dir2/file4.png",
    "dir1/file3.JPG",
    "file2.jpeg",
  }

  dir := CreateTestFiles(t, files)
  defer os.RemoveAll(dir)

  paths := FilesImage(dir)
  if strings.Join(paths, "\n") != strings.Join(result, "\n") {
    t.Errorf("Expected %v, got %v", result, paths)
  }
}

func TestFilesByExtensionAudio(t *testing.T) {
  files := []*TestFile{
    {"not audio file", ""},
    {"file1.FLAC", ""},
    {"file2.m4a", ""},
    {"dir1/file3.mp3", ""},
    {"dir1/dir2/file4.mp4", ""},
    {"dir1/dir2/file5.SHN", ""},
    {"dir1/dir2/file6.WAV", ""},
  }

  result := []string{
    "dir1/dir2/file4.mp4",
    "dir1/dir2/file5.SHN",
    "dir1/dir2/file6.WAV",
    "dir1/file3.mp3",
    "file1.FLAC",
    "file2.m4a",
  }

  dir := CreateTestFiles(t, files)
  defer os.RemoveAll(dir)

  paths := FilesAudio(dir)
  if strings.Join(paths, "\n") != strings.Join(result, "\n") {
    t.Errorf("Expected %v, got %v", result, paths)
  }
}

func TestIsLarger(t *testing.T) {
  testFilesFullPath(t, func(dir string, files []string) {
    tests := []struct {
      src, dest string
      result bool
    }{
      { src: files[0], dest: files[1], result: true },
      { src: files[1], dest: files[0], result: false },
      { src: files[2], dest: files[3], result: true },
      { src: "audiocc-file-def-dne", dest: files[3], result: false },
    }

    for i := range tests {
      r := IsLarger(tests[i].src, tests[i].dest)
      if r != tests[i].result {
        t.Errorf("Expected %v, got %v", tests[i].result, r)
      }
    }
  })
}

func TestNthFileSize(t *testing.T) {
  tests := []struct {
    smallest bool
    result string
    other string
  }{
    { smallest: true, result: "file2.jpeg" },
    { smallest: false, result: "dir1/file3.JPG" },
    { other: "audiocc-file-def-dne", result: "" },
  }

  testFilesFullPath(t, func(dir string, files []string) {
    for i := range tests {
      // test file does not exist
      if len(tests[i].other) > 0 {
        files = []string{ tests[i].other }
      }
      r, err := NthFileSize(files, tests[i].smallest)

      // test errors by setting empty result
      if err != nil && tests[i].result == "" {
        continue
      }

      res := filepath.Join(dir, tests[i].result)
      if r != res {
        t.Errorf("Expected %v, got %v", res, r)
      }
    }
  })
}

func TestRenameFolder(t *testing.T) {
  testFiles := []*TestFile{
    {"dir1/file1", "abcde"},
    {"dir2/file2", "a"},
    {"dir3/file3", ""},
    {"dir4/file4", ""},
    {"dir6/file6", ""},
  }

  dir := CreateTestFiles(t, testFiles)
  defer os.RemoveAll(dir)

  tests := [][]string{
    {"dir2", "dir1", "dir1 (1)"},
    {"dir3", "dir1", "dir1 (2)"},
    {"dir4", "dir5", "dir5"},
    {"dir6", "path2/dir5", "path2/dir5"},
  }

  for i := 0; i < len(tests); i++ {
    r, err := RenameFolder(filepath.Join(dir, tests[i][0]), filepath.Join(dir, tests[i][1]))
    if err != nil {
      t.Errorf("Unexpected error %v", err.Error())
    }

    exp := filepath.Join(dir, tests[i][2])
    if r != exp {
      t.Errorf("Expected %v, got %v", exp, r)
    }
  }
}

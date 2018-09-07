package fsutil

import (
  "io"
  "os"
  "fmt"
  "sort"
  "strings"
  "path/filepath"
)

const PathSep = string(os.PathSeparator)
var ImageExts = []string{ "jpeg", "jpg", "png" }
var AudioExts = []string{ "flac", "m4a", "mp3", "mp4", "shn", "wav" }

func CopyFile(srcPath, destPath string) (err error) {
  srcFile, err := os.Open(srcPath)
  if err != nil {
    return
  }
  defer srcFile.Close()

  destFile, err := os.Create(destPath)
  if err != nil {
    return
  }
  defer destFile.Close()

  _, err = io.Copy(destFile, srcFile)
  if err != nil {
    return
  }

  err = destFile.Sync()
  return
}

func FilesAudio(dir string) []string {
  return FilesByExtension(dir, AudioExts)
}

func FilesImage(dir string) []string {
  return FilesByExtension(dir, ImageExts)
}

func FilesByExtension(dir string, exts []string) []string {
  files := []string{}

  // closure to pass to filepath.Walk
  walkFunc := func(p string, f os.FileInfo, err error) error {
    ext := filepath.Ext(p)
    if len(ext) == 0 {
      return nil
    }
    ext = strings.ToLower(ext[1:])

    x := sort.SearchStrings(exts, ext)
    if x < len(exts) && exts[x] == ext {
      // remove base directory
      p = p[len(dir):]
      // remove prefixed path separator
      if p[0] == os.PathSeparator {
        p = p[1:]
      }
      files = append(files, p)
    }

    return err
  }

  err := filepath.Walk(dir, walkFunc)
  if err != nil {
    return []string{}
  }
  // must sort: nested directories' files list first
  // char / sorts before A-Za-z0-9
  sort.Strings(files)

  return files
}

// true if destination does not exist or src has larger file size
func IsLarger(file, newFile string) bool {
  f, err := os.Open(file)
  defer f.Close()
  if err != nil {
    return false
  }

  f2, err := os.Open(newFile)
  defer f2.Close()

  if err == nil {
    fn, err := NthFileSize([]string{ file, newFile }, false)
    if err != nil || fn == newFile {
      return false
    }
  }

  return true
}

// index of smallest/largest file in slice of files
func NthFileSize(files []string, smallest bool) (string, error) {
  sizes := []int64{}

  found := -1
  for i := range files {
    in, err := os.Open(files[i])
    if err != nil {
      return "", err
    }
    defer in.Close()

    info, err := in.Stat()
    if err != nil {
      return "", err
    }

    sizes = append(sizes, info.Size())
    if found == -1 || (smallest && info.Size() < sizes[found]) ||
      (!smallest && info.Size() > sizes[found]) {
      found = i
    }
  }

  if found == -1 {
    return "", fmt.Errorf("File not found")
  }
  return files[found], nil
}

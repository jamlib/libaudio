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

// if dest folder already exists, merge audio file if Disc/Track not already present
// within. merge all images currently not present
func MergeFolder(src, dest string,
  indexFunc func(f string) (int, string)) (string, error) {

  // if folder already exists
  _, err := os.Stat(dest)
  if err == nil {
    // build dest audio file info maps
    destAudios := FilesAudio(dest)
    lookup := make(map[int]string, len(destAudios))
    for _, destFile := range destAudios {
      index, title := indexFunc(destFile)
      lookup[index] = title
    }

    // copy only src audio files that don't already exist
    copied := false
    for _, srcFile := range FilesAudio(src) {
      index, title := indexFunc(srcFile)
      if _, found := lookup[index]; !found {
        srcPath := filepath.Join(src, srcFile)

        // if not found, copy audio file
        _, f := filepath.Split(srcFile)
        err = CopyFile(srcPath, filepath.Join(dest, f))
        if err != nil {
          return dest, err
        }

        // add to lookup, ensure copied is true
        lookup[index] = title
        copied = true

        // remove source audio file
        err = os.Remove(srcPath)
        if err != nil {
          return dest, err
        }
      }
    }

    // copy all image files (if copied at least one audio file)
    if copied {
      for _, imgFile := range FilesImage(src) {
        imgPath := filepath.Join(src, imgFile)
        _, img := filepath.Split(imgFile)
        _ = CopyFile(imgPath, filepath.Join(dest, img))
      }
    }

    // if remaining audio files, rename to folder (x)
    if len(FilesAudio(src)) > 0 {
      return RenameFolder(src, dest)
    }

    // else delete folder
    err = os.RemoveAll(src)
    return dest, err
  }

  // folder doesn't exist
  return RenameFolder(src, dest)
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

// rename folder src to dest. if dest already exist, append (x)
// to folder name, iterate until folder not found
func RenameFolder(src, dest string) (string, error) {
  _, err := os.Stat(dest)
  if err == nil {
    // increment (x) until dir not found
    x := 0
    for {
      x += 1
      newDir := fmt.Sprintf("%v (%v)", dest, x)

      _, err := os.Stat(newDir)
      if err != nil {
        dest = newDir
        break
      }
    }
  }

  // create parent dir by trimming via filepath.Dir
  err = os.MkdirAll(filepath.Dir(dest), 0777)
  if err != nil {
    return dest, err
  }

  // rename to dest
  err = os.Rename(src, dest)
  return dest, err
}

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

// group sorted files by common directory
func BundleFiles(dir string, files []string, f func(bundle []int) error) error {
  var dirCur string
  bundle := []int{}

  // need final dir change
  files = append(files, "")

  for x := range files {
    d := filepath.Dir(filepath.Join(dir, files[x]))

    if dirCur == "" {
      dirCur = d
    }

    // if dir changes or last of all files
    if d != dirCur || x == len(files)-1 {
      err := f(bundle)
      if err != nil {
        return err
      }

      bundle = []int{}
      dirCur = d
    }

    bundle = append(bundle, x)
  }

  return nil
}

// copy a file from srcPath to destPath
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

// returns slice of all nested audio files
func FilesAudio(dir string) []string {
  return FilesByExtension(dir, AudioExts)
}

// returns slice of all nested image files
func FilesImage(dir string) []string {
  return FilesByExtension(dir, ImageExts)
}

// returns string slice of all nested files within path that match certain
// file extensions.
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

  // sort: nested files list first since char / sorts before A-Za-z0-9
  sort.Strings(files)

  return files
}

// returns true if dest does not exist or src has larger file size than dest
func IsLarger(src, dest string) bool {
  f, err := os.Open(src)
  defer f.Close()
  if err != nil {
    return false
  }

  f2, err := os.Open(dest)
  defer f2.Close()

  if err == nil {
    fn, err := NthFileSize([]string{ src, dest }, false)
    if err != nil || fn == dest {
      return false
    }
  }

  return true
}

// if dest folder exists, merge audio files that are not already present.
// indexFunc() should return a disc number and a track title which is stored
// in a lookup and used to determine if next audio file already exists.
func MergeFolder(src, dest string,
  indexFunc func(f string) (int, string)) (string, error) {

  // if dest folder does not exist, simply rename folder
  _, err := os.Stat(dest)
  if err != nil {
    return RenameFolder(src, dest)
  }

  // build dest audio file info maps
  destAudios := FilesAudio(dest)
  lookup := make(map[int]string, len(destAudios))
  for _, destFile := range destAudios {
    index, title := indexFunc(destFile)
    lookup[index] = title
  }

  // boolean to determine if any audio files were copied
  copied := false

  // copy only src audio files that don't already exist
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

  // if remaining audio files, rename to dest folder by including (x)
  if len(FilesAudio(src)) > 0 {
    return RenameFolder(src, dest)
  }

  // only junk remains in src, delete it
  err = os.RemoveAll(src)
  return dest, err
}

// return path of smallest/largest file by file size in slice of files
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

// rename folder src to dest. if dest already exists, append (x)
// to folder name, increment (x) until folder path not found
func RenameFolder(src, dest string) (string, error) {
  // check if dest not found
  _, err := os.Stat(dest)
  if err == nil {
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

  // create all parent directories
  err = os.MkdirAll(filepath.Dir(dest), 0777)
  if err != nil {
    return dest, err
  }

  // rename src to dest
  err = os.Rename(src, dest)
  return dest, err
}

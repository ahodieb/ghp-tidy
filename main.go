package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"iter"
	"os"
	"path"
	"strings"
	"time"
)

func main() {

	file := "/Volumes/Crucial X9/photos_31_08_2024/takeout-20240830T153532Z-001.tgz"
	a, err := NewPhotoArchive(file)
	if err != nil {
		panic(err)
	}

	dumpDir := fmt.Sprintf("dump-%s", time.Now().Format("20060102150405"))
	if err := os.Mkdir(dumpDir, 0755); err != nil {
		panic(err)
	}
	indx, err := os.Create(path.Join(dumpDir, "index.txt"))
	if err != nil {
		panic(err)
	}

	fmt.Println("Writing to: ", indx.Name())

	for e := range a.Entries() {
		//for e := range a.EntriesWithExt(".HEIC", ".jpg", ".mp4", ".json", ".jpeg", ".mov", ".png", ".mp") {
		_, _ = fmt.Fprintln(indx, e.String())
		if err := os.WriteFile(path.Join(dumpDir, e.Base()), e.Bytes, 0755); err != nil {
			panic(err)
		}
	}

	fmt.Println("Skipped:", len(a.skipped))
	for _, s := range a.skipped {
		fmt.Println(s)
	}

	fmt.Println("Failed:", len(a.failed))
	for _, s := range a.failed {
		fmt.Println(s)
	}
}

//dir := "/Volumes/Crucial X9/photos_31_08_2024"
//func lsDir(dir string) {
//	files, err := os.ReadDir(dir)
//	if err != nil {
//		panic(err)
//	}
//
//	for _, file := range files {
//		info, err := file.Info()
//		if err != nil {
//			panic(fmt.Errorf("failed to get info for %q, %w", file.Name(), err))
//		}
//
//		fmt.Println(file.Name(), info.Size())
//	}
//}

//
//func  ListFiles(path string) iter.Seq[string] {
//	return func(yield func(string) bool) {
//		for v := range s.m {
//			if !yield(v) {
//				return
//			}
//		}
//	}
//}

type Entry struct {
	Archive string
	Path    string
	Sha256  string
	Bytes   []byte
}

func (e *Entry) Base() string {
	return path.Base(e.Path)
}

func (e *Entry) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s", e.Sha256, e.Base(), e.Path, e.Archive)
}

type FailedEntry struct {
	Name  string
	Cause error
}

type PhotoArchive struct {
	Path string

	r       *tar.Reader
	skipped []string
	failed  []FailedEntry
}

func NewPhotoArchive(path string) (PhotoArchive, error) {
	f, err := os.Open(path)
	if err != nil {
		return PhotoArchive{Path: path, failed: []FailedEntry{{Cause: err}}}, err
	}

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return PhotoArchive{Path: path, failed: []FailedEntry{{Cause: err}}}, err
	}

	r := tar.NewReader(gzf)
	return PhotoArchive{Path: path, r: r}, nil
}

func (a *PhotoArchive) Entries() iter.Seq[Entry] {
	return func(yield func(entry Entry) bool) {
		for {
			header, err := a.r.Next()
			if err == io.EOF || header == nil {
				return
			}

			if header.Typeflag != tar.TypeReg {
				a.skipped = append(a.skipped, header.Name)
				continue
			}

			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, a.r); err != nil {
				a.failed = append(a.failed, FailedEntry{Name: header.Name, Cause: err})
				continue
			}
			data := buf.Bytes()
			hash := sha256.New()
			sum := base64.StdEncoding.EncodeToString(hash.Sum(data))

			e := Entry{Archive: a.Path, Path: header.Name, Sha256: sum, Bytes: data}
			if !yield(e) {
				return
			}
		}
	}
}

func (a *PhotoArchive) EntriesFiltered(f func(Entry) bool) iter.Seq[Entry] {
	return func(yield func(entry Entry) bool) {
		for e := range a.Entries() {
			if f(e) {
				if !yield(e) {
					return
				}
			} else {
				a.skipped = append(a.skipped, e.Path)
			}
		}
	}
}

func (a *PhotoArchive) EntriesWithExt(x ...string) iter.Seq[Entry] {
	ext := map[string]struct{}{}
	for _, e := range x {
		ext[strings.ToLower(e)] = struct{}{}
	}

	return a.EntriesFiltered(func(entry Entry) bool {
		_, match := ext[strings.ToLower(path.Ext(entry.Path))]
		return match
	})
}

func Filter[V any](f func(V) bool, s iter.Seq[V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for v := range s {
			if f(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

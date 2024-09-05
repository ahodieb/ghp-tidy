package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/urfave/cli/v2"
	"io"
	"iter"
	"log"
	"os"
	"path"
	"strings"
)

func main() {
	app := &cli.App{
		Name:  "gph-tidy",
		Usage: "tidy google photos takeout files",
		Commands: []*cli.Command{
			{
				Name:   "index",
				Usage:  "build an index of all the contents of the archives",
				Action: IndexCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output-file",
						Aliases: []string{"o"},
						Usage:   "Specify where to write the index",
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func IndexCmd(c *cli.Context) error {
	out := os.Stdout
	if outFile := c.String("output-file"); outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			return fmt.Errorf("could not create output file: %s, %w", outFile, err)
		}
		defer f.Close()
		out = f
	}

	var inputFiles []string

	input := c.Args()
	if input.Len() == 1 {
		singleInput := input.First()
		stat, err := os.Stat(singleInput)
		if err != nil {
			return fmt.Errorf("could not check input file: %s, %w", singleInput, err)
		}

		if stat.IsDir() {
			return fmt.Errorf("specifing a dir is not supported yet")
		}

		inputFiles = append(inputFiles, singleInput)
	} else {
		inputFiles = c.Args().Slice()
	}

	return index(inputFiles, out)
}

func index(input []string, out io.Writer) error {
	for _, f := range input {
		a, err := NewPhotoArchive(f)
		if err != nil {
			return fmt.Errorf("could open archive: %s,  %w", f, err)
		}

		for e := range a.Entries() {
			_, _ = fmt.Fprintln(out, e.String())
		}
	}

	return nil
}

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
	return fmt.Sprintf(`{"sha256":%q,"name":%q,"size":%d,"path":"%s/%s"}`, e.Sha256, e.Base(), len(e.Bytes), e.Archive, e.Path)
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
			hash.Write(data)
			b64Sum := base64.StdEncoding.EncodeToString(hash.Sum(nil))

			e := Entry{Archive: a.Path, Path: header.Name, Sha256: b64Sum, Bytes: data}
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

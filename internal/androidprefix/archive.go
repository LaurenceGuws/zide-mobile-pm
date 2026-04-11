package androidprefix

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type ArchiveStats struct {
	Files    int
	Dirs     int
	Symlinks int
	SHA256   string
	Size     int64
}

func WriteTarGz(sourceRoot string, outputPath string) (ArchiveStats, error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return ArchiveStats{}, err
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return ArchiveStats{}, err
	}
	defer file.Close()

	hash := sha256.New()
	counting := &countingWriter{writer: io.MultiWriter(file, hash)}
	gzipWriter := gzip.NewWriter(counting)
	gzipWriter.Name = filepath.Base(outputPath)
	gzipWriter.ModTime = time.Unix(0, 0).UTC()
	tarWriter := tar.NewWriter(gzipWriter)

	stats := ArchiveStats{}
	entries, err := collectEntries(sourceRoot)
	if err != nil {
		return ArchiveStats{}, err
	}
	for _, entry := range entries {
		if err := writeTarEntry(tarWriter, sourceRoot, entry, &stats); err != nil {
			return ArchiveStats{}, err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return ArchiveStats{}, err
	}
	if err := gzipWriter.Close(); err != nil {
		return ArchiveStats{}, err
	}
	if err := file.Sync(); err != nil {
		return ArchiveStats{}, err
	}
	stats.SHA256 = hex.EncodeToString(hash.Sum(nil))
	stats.Size = counting.count
	return stats, nil
}

func collectEntries(sourceRoot string) ([]string, error) {
	var entries []string
	err := filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == sourceRoot {
			return nil
		}
		relative, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		entries = append(entries, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(entries)
	return entries, nil
}

func writeTarEntry(writer *tar.Writer, sourceRoot string, relative string, stats *ArchiveStats) error {
	sourcePath := filepath.Join(sourceRoot, filepath.FromSlash(relative))
	info, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = relative
	header.ModTime = time.Unix(0, 0).UTC()
	header.AccessTime = time.Unix(0, 0).UTC()
	header.ChangeTime = time.Unix(0, 0).UTC()
	header.Uid = 0
	header.Gid = 0
	header.Uname = ""
	header.Gname = ""
	if info.Mode()&os.ModeSymlink != 0 {
		linkname, err := os.Readlink(sourcePath)
		if err != nil {
			return err
		}
		header.Linkname = filepath.ToSlash(linkname)
	}
	if info.IsDir() && header.Name[len(header.Name)-1] != '/' {
		header.Name += "/"
	}
	if err := writer.WriteHeader(header); err != nil {
		return err
	}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		stats.Symlinks++
	case info.IsDir():
		stats.Dirs++
	case info.Mode().IsRegular():
		file, err := os.Open(sourcePath)
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err := io.Copy(writer, file); err != nil {
			return err
		}
		stats.Files++
	default:
		return fmt.Errorf("unsupported staged file type: %s", relative)
	}
	return nil
}

type countingWriter struct {
	writer io.Writer
	count  int64
}

func (writer *countingWriter) Write(bytes []byte) (int, error) {
	n, err := writer.writer.Write(bytes)
	writer.count += int64(n)
	return n, err
}

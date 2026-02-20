package cmd

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const mewDir = ".mew"

type Snapshot struct {
	ID    string    `json:"id"`
	Title string    `json:"title"`
	Time  time.Time `json:"time"`
	File  string    `json:"file"`
}

type Index struct {
	Snaps []Snapshot `json:"snaps"`
}

func ensureDirs() error {
	return os.MkdirAll(filepath.Join(mewDir, "snaps"), 0o755)
}

func indexPath() string { return filepath.Join(mewDir, "index.json") }

func loadIndex() (Index, error) {
	var idx Index
	b, err := os.ReadFile(indexPath())
	if err != nil {
		if os.IsNotExist(err) {
			return idx, nil
		}
		return idx, err
	}
	if err := json.Unmarshal(b, &idx); err != nil {
		return idx, err
	}
	return idx, nil
}

func saveIndex(idx Index) error {
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath(), b, 0o644)
}

func doSnap(title string) (string, error) {
	if err := ensureDirs(); err != nil {
		return "", err
	}
	// determine running executable to avoid including it in the snapshot
	execPath, _ := os.Executable()
	execAbs, _ := filepath.Abs(execPath)

	tmpName := filepath.Join(mewDir, "tmp-"+time.Now().Format("20060102150405")+".tar.gz")
	f, err := os.Create(tmpName)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	mw := io.MultiWriter(f, h)
	gz := gzip.NewWriter(mw)
	tw := tar.NewWriter(gz)

	err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		// Skip .mew dir and its contents
		if d.IsDir() && d.Name() == strings.TrimPrefix(mewDir, string(os.PathSeparator)) {
			return filepath.SkipDir
		}
		if strings.HasPrefix(path, mewDir+string(os.PathSeparator)) || path == mewDir {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// skip the running executable (avoid trying to snapshot/restore it)
		if abs, err := filepath.Abs(path); err == nil && abs == execAbs {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = path
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		rf, err := os.Open(path)
		if err != nil {
			return err
		}
		if _, err := io.Copy(tw, rf); err != nil {
			rf.Close()
			return err
		}
		rf.Close()
		return nil
	})
	if err != nil {
		tw.Close()
		gz.Close()
		return "", err
	}
	tw.Close()
	gz.Close()
	f.Close()

	sum := hex.EncodeToString(h.Sum(nil))[:12]
	name := time.Now().Format("20060102T150405") + "-" + sum + ".tar.gz"
	dest := filepath.Join(mewDir, "snaps", name)
	if err := os.Rename(tmpName, dest); err != nil {
		return "", err
	}

	idx, err := loadIndex()
	if err != nil {
		return "", err
	}
	snap := Snapshot{ID: sum, Title: title, Time: time.Now(), File: dest}
	idx.Snaps = append([]Snapshot{snap}, idx.Snaps...)
	if err := saveIndex(idx); err != nil {
		return "", err
	}
	return sum, nil
}

func findSnapshot(key string) (*Snapshot, error) {
	idx, err := loadIndex()
	if err != nil {
		return nil, err
	}
	for _, s := range idx.Snaps {
		if s.ID == key || s.Title == key || strings.HasPrefix(s.ID, key) {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("snapshot not found: %s", key)
}

func doWind(key string) error {
	s, err := findSnapshot(key)
	if err != nil {
		return err
	}
	f, err := os.Open(s.File)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if strings.HasPrefix(hdr.Name, mewDir) {
			continue
		}
		target := hdr.Name
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		}
	}
	return nil
}

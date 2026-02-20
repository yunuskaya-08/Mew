package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommandCreatesIndex(t *testing.T) {
	td := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(td); err != nil {
		t.Fatal(err)
	}

	// run init
	initForce = false
	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatal(err)
	}
	idxPath := indexPath()
	if _, err := os.Stat(idxPath); err != nil {
		t.Fatalf("expected index at %s: %v", idxPath, err)
	}

	// write something then force init to overwrite
	if err := os.WriteFile(idxPath, []byte("garbage"), 0o644); err != nil {
		t.Fatal(err)
	}
	initForce = true
	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(idxPath)
	if err != nil {
		t.Fatal(err)
	}
	var idx Index
	if err := json.Unmarshal(b, &idx); err != nil {
		t.Fatalf("index not valid json after init --force: %v", err)
	}
	if len(idx.Snaps) != 0 {
		t.Fatalf("expected empty index after init --force, got %d snaps", len(idx.Snaps))
	}
}

func TestSnapWindCycle(t *testing.T) {
	td := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(td); err != nil {
		t.Fatal(err)
	}

	// create sample files
	if err := os.WriteFile("a.txt", []byte("alpha"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("sub", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("sub", "b.txt"), []byte("beta"), 0o644); err != nil {
		t.Fatal(err)
	}

	// create snapshot via command to exercise CLI layer
	if err := snapCmd.RunE(snapCmd, []string{"first"}); err != nil {
		t.Fatal(err)
	}

	idx, err := loadIndex()
	if err != nil {
		t.Fatal(err)
	}
	if len(idx.Snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(idx.Snaps))
	}
	s := idx.Snaps[0]
	if s.Title != "first" {
		t.Fatalf("snapshot title mismatch: %s", s.Title)
	}
	if len(s.ID) != 12 {
		t.Fatalf("unexpected id length: %s", s.ID)
	}

	// mutate files then restore
	if err := os.WriteFile("a.txt", []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll("sub"); err != nil {
		t.Fatal(err)
	}

	if err := windCmd.RunE(windCmd, []string{s.ID}); err != nil {
		t.Fatal(err)
	}

	// verify restore
	ba, err := os.ReadFile("a.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(ba) != "alpha" {
		t.Fatalf("a.txt content mismatch, got %q", string(ba))
	}
	bb, err := os.ReadFile(filepath.Join("sub", "b.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(bb) != "beta" {
		t.Fatalf("sub/b.txt content mismatch, got %q", string(bb))
	}
}

func TestListCommandOutput(t *testing.T) {
	td := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(td); err != nil {
		t.Fatal(err)
	}

	// no snaps -> prints 'no snapshots'
	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	if err := listCmd.RunE(listCmd, []string{}); err != nil {
		w.Close()
		os.Stdout = old
		t.Fatal(err)
	}
	w.Close()
	os.Stdout = old
	io.Copy(&buf, r)
	out := buf.String()
	if !strings.Contains(out, "no snapshots") {
		t.Fatalf("expected 'no snapshots', got %q", out)
	}

	// create a snap and ensure list prints it
	if err := os.WriteFile("x.txt", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := snapCmd.RunE(snapCmd, []string{"s1"}); err != nil {
		t.Fatal(err)
	}

	var buf2 bytes.Buffer
	r2, w2, _ := os.Pipe()
	old2 := os.Stdout
	os.Stdout = w2
	if err := listCmd.RunE(listCmd, []string{}); err != nil {
		w2.Close()
		os.Stdout = old2
		t.Fatal(err)
	}
	w2.Close()
	os.Stdout = old2
	io.Copy(&buf2, r2)
	out2 := buf2.String()
	idx, _ := loadIndex()
	if len(idx.Snaps) == 0 || !strings.Contains(out2, idx.Snaps[0].ID) {
		t.Fatalf("list output did not contain snapshot id; output=%q", out2)
	}
}

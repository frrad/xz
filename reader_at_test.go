package xz

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
)

const foxSentenceConst = "The quick brown fox jumps over the lazy dog.\n"

func TestReaderAtBlocks(t *testing.T) {
	f, fileSize := testOpenFile(t, "testfiles/fox.blocks.xz")
	testFilePart(t, f, fileSize, foxSentenceConst, 0, len(foxSentenceConst))
}

func BenchmarkBlocks(b *testing.B) {
	f, fileSize := testOpenFile(b,
		"/home/frederickrobinson/skyhub/index.csv.xz")

	conf := ReaderAtConfig{
		Len: fileSize,
	}
	r, err := conf.NewReaderAt(f)
	if err != nil {
		b.Fatalf("NewReader error %s", err)
	}

	bytesToRead := 100

	decompressedMaxLen := r.Size()
	random(b, bytesToRead, decompressedMaxLen, r)
}

func BenchmarkBlocksFromMemory(b *testing.B) {
	fileBytes, err := ioutil.ReadFile("/home/frederickrobinson/skyhub/index.csv.xz")
	fileReader := bytes.NewReader(fileBytes)

	conf := ReaderAtConfig{
		Len: int64(len(fileBytes)),
	}
	r, err := conf.NewReaderAt(fileReader)
	if err != nil {
		b.Fatalf("NewReader error %s", err)
	}

	bytesToRead := 100

	decompressedMaxLen := r.Size()
	random(b, bytesToRead, decompressedMaxLen, r)
}

func random(b *testing.B, bytesToRead int, decompressedMaxLen int64, r io.ReaderAt) {
	decompressedBytes := make([]byte, bytesToRead)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := rand.Int63n(decompressedMaxLen)

		_, err := r.ReadAt(decompressedBytes, int64(start))
		if err != nil && err != io.EOF {
			b.Fatalf("error while reading at: %v", err)
		}
	}
}

func TestReaderAtSimple(t *testing.T) {
	f, fileSize := testOpenFile(t, "testfiles/fox.xz")
	testFilePart(t, f, fileSize, foxSentenceConst, 0, 10)
}

func TestReaderAtMS(t *testing.T) {
	expect := foxSentenceConst + foxSentenceConst + foxSentenceConst + foxSentenceConst

	filePath := "testfiles/fox.blocks.xz"

	f, _ := testOpenFile(t, filePath)
	fData, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("Error reading file %s", err)
	}
	msBytes := testMultiStreams(fData)
	msB := bytes.NewReader(msBytes)

	start := len(foxSentenceConst)
	testFilePart(t, msB, int64(len(msBytes)), expect, start, len(expect)-start)
}

func testOpenFile(t testing.TB, filePath string) (*os.File, int64) {
	xz, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("os.Open(%q) error %s", filePath, err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("os.Stat(%q) error %s", filePath, err)
	}

	return xz, info.Size()
}

func testFilePart(t testing.TB, f io.ReaderAt, fileSize int64, expected string, start, size int) {
	conf := ReaderAtConfig{
		Len: fileSize,
	}
	r, err := conf.NewReaderAt(f)
	if err != nil {
		t.Fatalf("NewReader error %s", err)
	}

	decompressedBytes := make([]byte, size)
	n, err := r.ReadAt(decompressedBytes, int64(start))
	if err != nil {
		t.Fatalf("error while reading at: %v", err)
	}
	if n != len(decompressedBytes) {
		t.Fatalf("unexpectedly didn't read all")
	}

	subsetExpected := expected[start : start+size]
	if string(decompressedBytes) != subsetExpected {
		t.Fatalf("Unexpected decompression output. \"%s\" != \"%s\"",
			string(decompressedBytes), subsetExpected)
	}
}

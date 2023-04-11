package integritymonitor

import (
	"os"
	"testing"
)

func TestHashDir(t *testing.T) {
	rootPath := "/tmp/testroot/"
	testDir := "fancydir/"
	alg := "sha256"
	testDirPath := rootPath + testDir
	testFilePath := rootPath + testDir + "fancyfile"                                   // An empty file
	testFileHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" //A hash of an empty file

	// create test directory and file
	if err := os.MkdirAll(testDirPath, os.ModePerm); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if _, err := os.Create(testFilePath); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	defer os.RemoveAll(rootPath) // cleanup

	// call HashDir and verify result
	result := HashDir(rootPath, testDir, alg)
	if len(result) != 1 {
		t.Fatalf("HashDir returned unexpected number of results: %d", len(result))
	}
	if result[0].Hash != testFileHash {
		t.Fatalf("HashDir returned unexpected file hash: %s", result[0].Hash)
	}
}

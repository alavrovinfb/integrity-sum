package data

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type ChecksumsReader struct {
	r io.Reader
}

func NewFileStorage(reader io.Reader) *ChecksumsReader {
	return &ChecksumsReader{
		r: reader,
	}
}

func (fs *ChecksumsReader) Get() ([]*HashDataOutput, error) {
	fileScanner := bufio.NewScanner(fs.r)
	fileScanner.Split(bufio.ScanLines)
	var checkSums []*HashDataOutput

	for fileScanner.Scan() {
		sum, err := fs.parseRecord(fileScanner.Text())
		if err != nil {
			return nil, err
		}
		checkSums = append(checkSums, sum)
	}

	return checkSums, nil
}

func (fs *ChecksumsReader) parseRecord(rec string) (*HashDataOutput, error) {

	if len(rec) == 0 {
		return nil, fmt.Errorf("%s", "an empty hash record")
	}

	parts := strings.Split(rec, "  ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("incorrect hash record %s", rec)
	}

	return &HashDataOutput{
		Hash:         strings.TrimSpace(parts[0]),
		FullFileName: strings.TrimSpace(parts[1]),
	}, nil
}

package delta

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// GenerateBackwardPatch computes the diff to reconstruct the OLD text from the NEW text.
// Returns the patch text.
func GenerateBackwardPatch(oldText, newText string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(newText, oldText, true) // Notice: newText is text1, oldText is text2
	dmp.DiffCleanupSemantic(diffs)
	patches := dmp.PatchMake(newText, diffs)
	return dmp.PatchToText(patches)
}

// GenerateForwardPatch computes the diff to construct the NEW text from the OLD text.
func GenerateForwardPatch(oldText, newText string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldText, newText, true)
	dmp.DiffCleanupSemantic(diffs)
	patches := dmp.PatchMake(oldText, diffs)
	return dmp.PatchToText(patches)
}

// ApplyPatch applies a patch string to the base text.
func ApplyPatch(baseText, patchText string) (string, error) {
	dmp := diffmatchpatch.New()
	patches, err := dmp.PatchFromText(patchText)
	if err != nil {
		return "", err
	}
	result, _ := dmp.PatchApply(patches, baseText)
	// We ignore the boolean array of patch results for simplicity, assuming success
	return result, nil
}

// CompressGzip compresses a string into a gzip byte slice
func CompressGzip(data string) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write([]byte(data)); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecompressGzip decompresses a gzip byte slice back to a string
func DecompressGzip(data []byte) (string, error) {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer zr.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, zr); err != nil {
		return "", err
	}
	return buf.String(), nil
}

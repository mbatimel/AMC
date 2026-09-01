package main

import "testing"

func TestRequestBodyLimitIncludesBase64AndJSONOverhead(t *testing.T) {
	const maxFileSize = int64(10 * 1024 * 1024)
	encodedSize := ((maxFileSize + 2) / 3) * 4
	limit := int64(requestBodyLimit(maxFileSize))
	if limit <= encodedSize {
		t.Fatalf("limit=%d encoded=%d", limit, encodedSize)
	}
	if limit > 16*1024*1024 {
		t.Fatalf("limit=%d is unexpectedly large", limit)
	}
}

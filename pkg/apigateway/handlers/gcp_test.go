package handlers

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_uploadDataToGcsObject(t *testing.T) {
	data := bytes.NewReader([]byte("Hello world!"))
	bucket := "myproject001-367500-001"
	objectKey := "test/test_object_001.txt"
	err := uploadDataToGcsObject(data, bucket, objectKey)
	assert.Nil(t, err)
}

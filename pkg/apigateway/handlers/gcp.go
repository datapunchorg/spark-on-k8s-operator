package handlers

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"io"
	"time"
)

func uploadDataToGcsObject(data io.Reader, bucket, objectKey string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to run storage.NewClient: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	o := client.Bucket(bucket).Object(objectKey)

	// Optional: set a generation-match precondition to avoid potential race
	// conditions and data corruptions. The request to upload is aborted if the
	// objectKey's generation number does not match your precondition.
	// For an objectKey that does not yet exist, set the DoesNotExist precondition.
	// o = o.If(storage.Conditions{DoesNotExist: true})

	// If the live objectKey already exists in your bucket, set instead a
	// generation-match precondition using the live objectKey's generation number.
	// attrs, err := o.Attrs(ctx)
	// if err != nil {
	//      return fmt.Errorf("objectKey.Attrs: %v", err)
	// }
	// o = o.If(storage.Conditions{GenerationMatch: attrs.Generation})

	// Upload an objectKey with storage.Writer.
	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, data); err != nil {
		wc.Close()
		return fmt.Errorf("failed to copy data to %s/%s: %w", bucket, objectKey, err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close object %s/%s: %w", bucket, objectKey, err)
	}
	return nil
}

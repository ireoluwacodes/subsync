package storage

// GCS handles invoice PDF upload and signed URL generation.

type GCS struct{}

func NewGCS() *GCS { return &GCS{} }

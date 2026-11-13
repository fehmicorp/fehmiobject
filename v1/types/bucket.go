package types

import (
	"time"
)

type Bucket struct {
	Name      string    `bson:"name" json:"name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type Object struct {
	Bucket         string    `bson:"bucket" json:"bucket"`
	Key            string    `bson:"object_key" json:"object_key"`
	Size           int64     `bson:"size" json:"size"`
	ETag           string    `bson:"etag" json:"etag"`
	ChecksumSHA256 string    `bson:"checksum_sha256" json:"checksum_sha256"`
	ContentType    string    `bson:"content_type" json:"content_type"`
	StoragePath    string    `bson:"storage_path" json:"storage_path"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}

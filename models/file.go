package models

type File struct {
	ID         string `bson:"id" json:"id"`
	UserID     string `bson:"user_id" json:"user_id"`
	Filename   string `bson:"filename" json:"filename"`
	Size       int64  `bson:"size" json:"size"`
	Path       string `bson:"path" json:"path"`
	Hash       string `bson:"hash" json:"hash"`
	UploadedAt int64  `bson:"uploaded_at" json:"uploaded_at"`
}

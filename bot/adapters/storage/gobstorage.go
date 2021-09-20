package storage

import "encoding/gob"

type GobStorage interface {
	StorageAdapter
	SetOutput(*gob.Encoder)
}

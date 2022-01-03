package editor

import "io"

type UploadApplier interface {
	Upload(interface{}) (string, error)
	Apply(string) (io.ReadCloser, error)

	ApplyCode() string
}

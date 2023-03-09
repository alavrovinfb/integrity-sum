package integritymonitor

const (
	ErrTypeNewFile int = iota
	ErrTypeFileDeleted
	ErrTypeFileMismatch
)

type IntegrityError struct {
	Type int
	Path string
	Hash string
}

func (e *IntegrityError) Error() string {
	switch e.Type {
	case ErrTypeNewFile:
		return "new file found"
	case ErrTypeFileDeleted:
		return "file deleted"
	case ErrTypeFileMismatch:
		return "file content mismatch"
	}
	return "unknown integrity error"
}

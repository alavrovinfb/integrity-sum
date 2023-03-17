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
		return IntegrityMessageNewFileFound
	case ErrTypeFileDeleted:
		return IntegrityMessageFileDeleted
	case ErrTypeFileMismatch:
		return IntegrityMessageFileMismatch
	}
	return IntegrityMessageUnknownErr
}

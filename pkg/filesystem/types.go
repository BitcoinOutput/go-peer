package filesystem

type IFile interface {
	Read() ([]byte, error)
	Write([]byte) error
	IsExist() bool
}

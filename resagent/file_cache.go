package resagent

var _ Cache = (*FileCache)(nil)

type FileCache struct {
	dir string
}

// List implements Cache.
func (f *FileCache) List() []any {
	panic("unimplemented")
}

func NewFileCache(dir string) *FileCache {
	return &FileCache{dir: dir}
}

// Clear implements Cache.
func (f *FileCache) Clear() {
	panic("unimplemented")
}

// Delete implements Cache.
func (f *FileCache) Delete(rid string) {
	panic("unimplemented")
}

// Get implements Cache.
func (f *FileCache) Get(rid string) (any, bool) {
	panic("unimplemented")
}

// Set implements Cache.
func (f *FileCache) Set(rid string, data any) {
	panic("unimplemented")
}

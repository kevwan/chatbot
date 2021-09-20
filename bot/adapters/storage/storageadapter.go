package storage

type StorageAdapter interface {
	BuildIndex()
	Count() int
	Find(string) (map[string]int, bool)
	Search(string) []string
	Remove(string)
	Sync() error
	Update(string, map[string]int)
}

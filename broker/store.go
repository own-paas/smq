package broker

type Store interface {
	Raw() interface{}
	Txn() interface{}
	Open() error
	Put(key string, value interface{}) error
	Get(key string) (interface{}, error)
	All(prefix string) (interface{}, error)
	Del(key string) error
	Close() error
	Reset() error
}

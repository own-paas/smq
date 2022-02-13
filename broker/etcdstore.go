package broker

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

const (
	// DefaultTimeout default timeout
	DefaultTimeout = time.Second * 3
	// DefaultRequestTimeout default request timeout
	DefaultRequestTimeout = 10 * time.Second
	// DefaultSlowRequestTime default slow request time
	DefaultSlowRequestTime = time.Second * 1
)

// slowLogTxn wraps etcd transaction and log slow one.
type slowLogTxn struct {
	clientv3.Txn
	cancel context.CancelFunc
}

func newSlowLogTxn(client *clientv3.Client) clientv3.Txn {
	ctx, cancel := context.WithTimeout(client.Ctx(), DefaultRequestTimeout)
	return &slowLogTxn{
		Txn:    client.Txn(ctx),
		cancel: cancel,
	}
}

func (t *slowLogTxn) If(cs ...clientv3.Cmp) clientv3.Txn {
	return &slowLogTxn{
		Txn:    t.Txn.If(cs...),
		cancel: t.cancel,
	}
}

func (t *slowLogTxn) Then(ops ...clientv3.Op) clientv3.Txn {
	return &slowLogTxn{
		Txn:    t.Txn.Then(ops...),
		cancel: t.cancel,
	}
}

// Commit implements Txn Commit interface.
func (t *slowLogTxn) Commit() (*clientv3.TxnResponse, error) {
	start := time.Now()
	resp, err := t.Txn.Commit()
	t.cancel()

	cost := time.Now().Sub(start)
	if cost > DefaultSlowRequestTime {
		fmt.Sprintf("slow: txn runs too slow, resp=<%v> cost=<%s> errors:\n %+v",
			resp,
			cost,
			err)
	}

	return resp, err
}

type etcdStore struct {
	rootPrefix string
	rawClient  *clientv3.Client
}

func (self *etcdStore) Raw() interface{} {
	return self.rawClient
}

func (self *etcdStore) Txn() interface{} {
	return self.txn()
}

func (self *etcdStore) Open() (err error) {
	return
}

func (self *etcdStore) Put(key string, value interface{}) error {
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("value data type is not string")
	}

	_, err := self.txn().Then(clientv3.OpPut(key, v)).Commit()
	return err
}

func (self *etcdStore) Get(key string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(self.rawClient.Ctx(), DefaultRequestTimeout)
	defer cancel()

	resp, err := clientv3.NewKV(self.rawClient).Get(ctx, key)
	if nil != err {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	return resp.Kvs[0].Value, nil
}

func (self *etcdStore) All(prefix string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(self.rawClient.Ctx(), DefaultRequestTimeout)
	defer cancel()

	resp, err := clientv3.NewKV(self.rawClient).Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if nil != err {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	return resp.Kvs, nil
}

func (self *etcdStore) Del(key string) error {
	_, err := self.txn().Then(clientv3.OpDelete(key, clientv3.WithPrefix())).Commit()
	return err
}

func (self *etcdStore) Close() error {
	return self.rawClient.Close()
}

func (self *etcdStore) Reset() error {
	return self.Del(self.rootPrefix)
}

func (e *etcdStore) txn() clientv3.Txn {
	return newSlowLogTxn(e.rawClient)
}

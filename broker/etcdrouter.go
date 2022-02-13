package broker

import (
	"encoding/json"
	"fmt"
	genUUID "github.com/go-basic/uuid"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/model"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type etcdRouter struct {
	forwardDir      string
	forwardWatchDir string
	broker          *Broker
}

func (r *etcdRouter) Forward(nodeID string, data *model.Forward) error {
	key := fmt.Sprintf("%s/%s/%s", r.forwardDir, nodeID, genUUID.New())
	value, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.broker.store.Put(key, string(value))
}

func (r *etcdRouter) Receive() {
	etcdClient, _ := r.broker.store.Raw().(*clientv3.Client)
	watcher := clientv3.NewWatcher(etcdClient)
	defer watcher.Close()

	ctx := etcdClient.Ctx()
	for {
		rch := watcher.Watch(ctx, r.forwardWatchDir, clientv3.WithPrefix())
		for wresp := range rch {
			if wresp.Canceled {
				return
			}

			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.PUT {
					value := &model.Forward{}
					if len(ev.Kv.Value) > 0 {
						if err := json.Unmarshal(ev.Kv.Value, value); err != nil {
							global.LOGGER.Error("unmarshal forward data failed", zap.String("value", string(ev.Kv.Value)), zap.Error(err))
							continue
						}
					}

					if value != nil {
						cli, exist := r.broker.clients.Load(value.ClientID)
						if !exist {
							continue
						}

						c, ok := cli.(*client)
						if !ok {
							continue
						}

						if c.typ != LOCAL {
							continue
						}
						msg := &Message{
							client:  c,
							packet:  value.Packet,
							forward: true,
						}
						r.broker.SubmitWork(value.ClientID, msg)
					}

					global.LOGGER.Debug("delete forward key", zap.String("key", string(ev.Kv.Key)))
					r.broker.store.Del(string(ev.Kv.Key))
				}
			}
		}

		select {
		case <-ctx.Done():
			// server closed, return
			return
		default:
		}
	}
}

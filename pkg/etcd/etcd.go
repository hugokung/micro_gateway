package etcd

import (
	"context"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
)

type EtcdManger struct {
	//hosts       []string
	client *clientv3.Client
	//DialTimeout int
}

func NewEtcdManager(hosts []string, DialTimeout int) (*EtcdManger, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   hosts,
		DialTimeout: time.Duration(DialTimeout),
	})

	if err != nil {
		return nil, err
	}

	return &EtcdManger{
		client: cli,
	}, nil
}

func (e *EtcdManger) Close() {
	e.client.Close()
}

func (e *EtcdManger) GetServerListByPrefix(prefix string) ([]string, error) {
	resp, err := e.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	list := []string{}
	for _, kv := range resp.Kvs {
		list = append(list, string(kv.Value))
	}
	return list, nil
}

func (e *EtcdManger) WathServerListByPrefix(prefix string) (chan []string, chan error) {
	snapshots := make(chan []string)
	errCh := make(chan error)
	go func() {
		wCh := e.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
		for wResp := range wCh {
			for _, event := range wResp.Events {
				switch event.Type {
				case mvccpb.PUT, mvccpb.DELETE:
					snapshot, err := e.GetServerListByPrefix(prefix)
					if err != nil {
						errCh <- err
					}
					snapshots <- snapshot
				}
			}
		}
	}()
	return snapshots, errCh
}

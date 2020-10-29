/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package etcd

import (
	"context"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/value"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"testing"
)

func TestEtcdIndexer_Search(t *testing.T) {
	data := &client.PluginResponse{Revision: 1}
	c := &mockRegistry{}
	i := &Indexer{Client: c, Root: "/", Parser: value.BytesParser}

	// case: key does not contain prefix
	resp, err := i.Search(context.Background(), client.WithStrKey("a"))
	if err == nil || resp != nil {
		t.Fatalf("TestEtcdIndexer_Search failed")
	}

	// case: registry return err
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"))
	if err == nil || resp != nil {
		t.Fatalf("TestEtcdIndexer_Search failed")
	}

	// case: kvs is empty
	c.Response = data
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"))
	if err != nil || resp == nil || resp.Count != 0 {
		t.Fatalf("TestEtcdIndexer_Search failed")
	}

	// case: parse error
	data.Count = 2
	data.Kvs = []*mvccpb.KeyValue{{Key: []byte("/a/b"), Value: []byte("abc")}, {Key: []byte("/a/c"), Value: []byte("{}")}}
	old := i.Parser
	i.Parser = value.MapParser
	c.Response = data
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"))
	if err != nil || resp == nil || resp.Count != 2 || len(resp.Kvs) != 1 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	i.Parser = old

	// case: normal
	data.Count = 2
	data.Kvs = []*mvccpb.KeyValue{{Key: []byte("/a/b"), Value: []byte("abc")}, {Key: []byte("/a/c"), Value: []byte("{}")}}
	c.Response = data
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"))
	if err != nil || resp == nil || resp.Count != 2 || len(resp.Kvs) != 2 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"), client.WithKeyOnly())
	if err != nil || resp == nil || resp.Count != 2 || len(resp.Kvs) != 2 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"), client.WithCountOnly())
	if err != nil || resp == nil || resp.Count != 2 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
}

func TestCacheIndexer_Search(t *testing.T) {
	c := &mockCache{}
	cli := &mockRegistry{Response: &client.PluginResponse{
		Revision: 1,
		Kvs:      []*mvccpb.KeyValue{{Key: []byte("/a/b"), Value: []byte("abc")}},
		Count:    1,
	}}
	i := &CacheIndexer{
		Indexer: &Indexer{
			Root:   "/",
			Client: cli,
		},
		CacheIndexer: sd.NewCacheIndexer(c),
	}

	// case: key does not contain prefix
	resp, err := i.Search(context.Background(), client.WithStrKey("a"))
	if err == nil || resp != nil {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}

	// case: not match cache search remote
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"))
	if err != nil || resp == nil || resp.Count != 1 || string(resp.Kvs[0].Key) != "/a/b" {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}

	// no cache
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"), client.WithNoCache())
	if err != nil || resp == nil || resp.Count != 1 || string(resp.Kvs[0].Key) != "/a/b" {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}

	// rev > 0 or paging
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"), client.WithRev(1))
	if err != nil || resp == nil || resp.Count != 1 || string(resp.Kvs[0].Key) != "/a/b" {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"), client.WithOffset(0), client.WithLimit(1))
	if err != nil || resp == nil || resp.Count != 1 || string(resp.Kvs[0].Key) != "/a/b" {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}

	// not match cache
	c.Put("ka", &sd.KeyValue{Key: []byte("ka"), Value: []byte("va"), Version: 1, ModRevision: 1})
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"))
	if err != nil || resp == nil || resp.Count != 1 || string(resp.Kvs[0].Key) != "/a/b" {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"), client.WithPrefix())
	if err != nil || resp == nil || resp.Count != 1 || string(resp.Kvs[0].Key) != "/a/b" {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"), client.WithCountOnly())
	if err != nil || resp == nil || resp.Count != 1 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}

	// remote err
	oldResp := cli.Response
	cli.Response = nil
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"))
	if err == nil || resp != nil {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	cli.Response = oldResp

	// case: use cache index
	c.Put("/a", &sd.KeyValue{Key: []byte("/a"), Value: []byte("va"), Version: 1, ModRevision: 1})

	// exact match
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"))
	if err != nil || resp == nil || resp.Count != 1 || string(resp.Kvs[0].Value.([]byte)) != "va" {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"), client.WithCountOnly())
	if err != nil || resp == nil || resp.Count != 1 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}

	// prefix match
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"), client.WithPrefix())
	if err != nil || resp == nil || resp.Count != 1 || string(resp.Kvs[0].Value.([]byte)) != "va" {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	resp, err = i.Search(context.Background(), client.WithStrKey("/a"), client.WithPrefix(), client.WithCountOnly())
	if err != nil || resp == nil || resp.Count != 1 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}

	// mismatch cache but cacheonly
	resp, err = i.Search(context.Background(), client.WithStrKey("/a/b"), client.WithCacheOnly())
	if err != nil || resp == nil || resp.Count != 0 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
}

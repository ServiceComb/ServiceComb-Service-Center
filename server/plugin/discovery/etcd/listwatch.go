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
	"fmt"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
)

type ListWatchOptions struct {
	*discovery.Config
	Context context.Context
}

func (lo *ListWatchOptions) String() string {
	return fmt.Sprintf("{Config: %v, timeout: %s}", lo.Config, lo.Timeout)
}

type ListWatch interface {
	List(opts ListWatchOptions) (*registry.PluginResponse, error)
	// not support new multiple watchers
	Watch(opts ListWatchOptions) Watcher
	//
	DoWatch(ctx context.Context, f func(*registry.PluginResponse)) error
	Revision() int64
}

// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package notify

import (
	"github.com/apache/servicecomb-service-center/pkg/notify"
	"time"
)

const (
	AddJobTimeout          = 1 * time.Second
	SendTimeout            = 5 * time.Second
	HeartbeatTimeout       = 30 * time.Second
	InstanceEventQueueSize = 5000
)

var INSTANCE = notify.RegisterType("INSTANCE", InstanceEventQueueSize)
var notifyService *notify.NotifyService

func init() {
	notifyService = notify.NewNotifyService()
}

func NotifyCenter() *notify.NotifyService {
	return notifyService
}

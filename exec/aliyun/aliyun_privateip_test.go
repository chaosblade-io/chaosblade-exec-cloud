/*
 * Copyright 2025 The ChaosBlade Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package aliyun

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAliyunPrivateIpUnassign(t *testing.T) {
	result := unassignPrivateIpAddress(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "regionId", "networkInterfaceId", []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"})
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAliyunPrivateIpAssign(t *testing.T) {
	result := assignPrivateIpAddress(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "regionId", "networkInterfaceId", []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"})
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAliyunNetworkInterfaceAttributeDescribe(t *testing.T) {
	_, _err := describeNetworkInterfaceAttributeStatus(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "regionId", "networkInterfaceId")
	assert.NotNil(t, _err, "they should be equal")
}

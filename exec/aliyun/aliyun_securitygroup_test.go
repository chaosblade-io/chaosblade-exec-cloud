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

func TestAliyunSecurityGroupDelete(t *testing.T) {
	result := deleteSecurityGroup(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "regionId", "securityGroupId")
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAliyunSecurityGroupRemove(t *testing.T) {
	result := removeInstanceFromSecurityGroup(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "regionId", "securityGroupId", "networkInterfaceId", "instance1")
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAliyunSecurityGroupAdd(t *testing.T) {
	result := addInstanceToSecurityGroup(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "regionId", "securityGroupId", "networkInterfaceId", "instance1")
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAliyunDescribeSecurityGroup(t *testing.T) {
	_, _err := describeInstancesSecurityGroup(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "cn-hangzhou", "i-xx")
	assert.NotNil(t, _err, "they should be equal")
}

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

func TestAliyunPublicIpRelease(t *testing.T) {
	result := releasePublicIpAddress(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "regionId", "1.1.1.1", "instance1")
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAliyunPublicIpAssociate(t *testing.T) {
	result := allocatePublicIpAddress(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "regionId", "1.1.1.1", "instance1")
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAliyunPublicIpUnassociateEip(t *testing.T) {
	result := unassociateEipAddress(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "regionId", "allocationId", "instance1")
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAliyunPublicIpAssociateEip(t *testing.T) {
	result := associateEipAddress(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "regionId", "allocationId", "instance1")
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAliyunDescribeEipAddresses(t *testing.T) {
	_, _err := describeEipAddresses(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "regionId", "allocationId", "instance1")
	assert.NotNil(t, _err, "they should be equal")
}

func TestAliyunDescribeInstances(t *testing.T) {
	_, _err := describeInstances(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "cn-hangzhou", "i-xx")
	assert.NotNil(t, _err, "they should be equal")
}

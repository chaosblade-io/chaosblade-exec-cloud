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

package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAwsEcsStart(t *testing.T) {
	result := startAwsInstances(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "us-west-2", []string{"instance1", "instance2", "instance3"})
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAwsEcsStop(t *testing.T) {
	result := stopAwsInstances(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "us-west-2", []string{"instance1", "instance2", "instance3"})
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAwsEcsReboot(t *testing.T) {
	result := rebootAwsInstances(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "us-west-2", []string{"instance1", "instance2", "instance3"})
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAwsEcsDelete(t *testing.T) {
	result := deleteAwsInstances(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "us-west-2", []string{"instance1", "instance2", "instance3"})
	assert.Equal(t, int32(56002), result.Code, "they should be equal")
}

func TestAwsEcsDescribe(t *testing.T) {
	_, _err := describeInstancesStatus(context.WithValue(context.Background(), "uid", "123"), "accessKeyId", "accessKeySecret", "us-west-2", []string{"instance1", "instance2", "instance3"})
	assert.NotNil(t, _err, "they should be equal")
}

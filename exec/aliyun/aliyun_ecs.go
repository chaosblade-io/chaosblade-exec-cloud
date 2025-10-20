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
	"os"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/chaosblade-io/chaosblade-spec-go/channel"
	"github.com/chaosblade-io/chaosblade-spec-go/log"
	"github.com/chaosblade-io/chaosblade-spec-go/spec"
	"github.com/chaosblade-io/chaosblade-spec-go/util"

	"github.com/chaosblade-io/chaosblade-exec-cloud/exec"
	"github.com/chaosblade-io/chaosblade-exec-cloud/exec/category"
)

const EcsBin = "chaos_aliyun_ecs"

type EcsActionSpec struct {
	spec.BaseExpActionCommandSpec
}

func NewEcsActionSpec() spec.ExpActionCommandSpec {
	return &EcsActionSpec{
		spec.BaseExpActionCommandSpec{
			ActionFlags: []spec.ExpFlagSpec{
				&spec.ExpFlag{
					Name: "accessKeyId",
					Desc: "the accessKeyId of aliyun, if not provided, get from env ACCESS_KEY_ID",
				},
				&spec.ExpFlag{
					Name: "accessKeySecret",
					Desc: "the accessKeySecret of aliyun, if not provided, get from env ACCESS_KEY_SECRET",
				},
				&spec.ExpFlag{
					Name: "regionId",
					Desc: "the regionId of aliyun",
				},
				&spec.ExpFlag{
					Name: "type",
					Desc: "the operation of instances, support start, stop, reboot, etc",
				},
				&spec.ExpFlag{
					Name: "instances",
					Desc: "the instances list, split by comma",
				},
			},
			ActionExecutor: &EcsExecutor{},
			ActionExample: `
# stop instances which instance id is i-x,i-y
blade create aliyun ecs --accessKeyId xxx --accessKeySecret yyy --regionId cn-qingdao --type stop --instances i-x,i-y

# start instances which instance id is i-x,i-y
blade create aliyun ecs --accessKeyId xxx --accessKeySecret yyy --regionId cn-qingdao --type start --instances i-x,i-y

# reboot instances which instance id is i-x,i-y
blade create aliyun ecs --accessKeyId xxx --accessKeySecret yyy --regionId cn-qingdao --type reboot --instances i-x,i-y`,
			ActionPrograms:   []string{EcsBin},
			ActionCategories: []string{category.Cloud + "_" + category.Aliyun + "_" + category.Ecs},
		},
	}
}

func (*EcsActionSpec) Name() string {
	return "ecs"
}

func (*EcsActionSpec) Aliases() []string {
	return []string{}
}

func (*EcsActionSpec) ShortDesc() string {
	return "do some aliyun ecs Operations, like stop, start, reboot"
}

func (b *EcsActionSpec) LongDesc() string {
	if b.ActionLongDesc != "" {
		return b.ActionLongDesc
	}
	return "do some aliyun ecs Operations, like stop, start, reboot"
}

type EcsExecutor struct {
	channel spec.Channel
}

func (*EcsExecutor) Name() string {
	return "ecs"
}

var localChannel = channel.NewLocalChannel()

func (be *EcsExecutor) Exec(uid string, ctx context.Context, model *spec.ExpModel) *spec.Response {
	if be.channel == nil {
		util.Errorf(uid, util.GetRunFuncName(), spec.ChannelNil.Msg)
		return spec.ResponseFailWithFlags(spec.ChannelNil)
	}
	accessKeyId := model.ActionFlags["accessKeyId"]
	accessKeySecret := model.ActionFlags["accessKeySecret"]
	regionId := model.ActionFlags["regionId"]
	operationType := model.ActionFlags["type"]
	instances := model.ActionFlags["instances"]
	if accessKeyId == "" {
		val, ok := os.LookupEnv("ACCESS_KEY_ID")
		if !ok {
			log.Errorf(ctx, "could not get ACCESS_KEY_ID from env or parameter!")
			return spec.ResponseFailWithFlags(spec.ParameterLess, "accessKeyId")
		}
		accessKeyId = val
	}

	if accessKeySecret == "" {
		val, ok := os.LookupEnv("ACCESS_KEY_SECRET")
		if !ok {
			log.Errorf(ctx, "could not get ACCESS_KEY_SECRET from env or parameter!")
			return spec.ResponseFailWithFlags(spec.ParameterLess, "accessKeySecret")
		}
		accessKeySecret = val
	}

	if regionId == "" {
		log.Errorf(ctx, "regionId is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "regionId")
	}

	if operationType == "" {
		log.Errorf(ctx, "operationType is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "type")
	}

	if instances == "" {
		log.Errorf(ctx, "instances is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "instances")
	}
	instancesArray := strings.Split(instances, ",")
	instanceStatusMap, _err := describeInstancesStatus(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	if _err != nil {
		return spec.ResponseFailWithFlags(spec.ParameterRequestFailed, "describe instances status failed")
	}

	for _, instance := range instancesArray {
		if (instanceStatusMap[instance] == "Running" && operationType == "start") || (instanceStatusMap[instance] == "Stopped" && operationType == "stop") {
			return be.stop(ctx, operationType, accessKeyId, accessKeySecret, regionId, instancesArray)
		}
	}
	return be.start(ctx, operationType, accessKeyId, accessKeySecret, regionId, instancesArray)
}

func (be *EcsExecutor) start(ctx context.Context, operationType, accessKeyId, accessKeySecret, regionId string, instancesArray []string) *spec.Response {
	switch operationType {
	case "start":
		return startInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	case "stop":
		return stopInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	case "reboot":
		return rebootInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	// case "delete":
	//	return deleteInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support start, stop, reboot)")
	}
	select {}
}

func (be *EcsExecutor) stop(ctx context.Context, operationType, accessKeyId, accessKeySecret, regionId string, instancesArray []string) *spec.Response {
	switch operationType {
	case "start":
		return stopInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	case "stop":
		return startInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	// case "reboot":
	//	return rebootInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	// case "delete":
	//	return deleteInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support start, stop, reboot)")
	}
	ctx = context.WithValue(ctx, "bin", EcsBin)
	return exec.Destroy(ctx, be.channel, "aliyun es")
}

func (be *EcsExecutor) SetChannel(channel spec.Channel) {
	be.channel = channel
}

func CreateClient(accessKeyId *string, accessKeySecret *string, regionId string) (_result *ecs20140526.Client, _err error) {
	config := &openapi.Config{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}
	// 访问的域名
	config.Endpoint = tea.String("ecs." + regionId + ".aliyuncs.com")
	_result = &ecs20140526.Client{}
	_result, _err = ecs20140526.NewClient(config)
	return _result, _err
}

// start instances
func startInstances(ctx context.Context, accessKeyId, accessKeySecret, regionId string, instances []string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	startInstancesRequest := &ecs20140526.StartInstancesRequest{
		InstanceId: tea.StringSlice(instances),
		RegionId:   tea.String(regionId),
	}
	_, _err = client.StartInstances(startInstancesRequest)
	if _err != nil {
		log.Errorf(ctx, "start aliyun instances failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "start aliyun instances failed")
	}
	return spec.Success()
}

// stop instances
func stopInstances(ctx context.Context, accessKeyId, accessKeySecret, regionId string, instances []string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	stopInstancesRequest := &ecs20140526.StopInstancesRequest{
		InstanceId: tea.StringSlice(instances),
		RegionId:   tea.String(regionId),
	}
	_, _err = client.StopInstances(stopInstancesRequest)
	if _err != nil {
		log.Errorf(ctx, "stop aliyun instances failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "stop aliyun instances failed")
	}
	return spec.Success()
}

// reboot instances
func rebootInstances(ctx context.Context, accessKeyId, accessKeySecret, regionId string, instances []string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	rebootInstancesRequest := &ecs20140526.RebootInstancesRequest{
		InstanceId: tea.StringSlice(instances),
		RegionId:   tea.String(regionId),
	}
	_, _err = client.RebootInstances(rebootInstancesRequest)
	if _err != nil {
		log.Errorf(ctx, "reboot aliyun instances failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "restart aliyun instances failed")
	}
	return spec.Success()
}

// delete instances
func deleteInstances(ctx context.Context, accessKeyId, accessKeySecret, regionId string, instances []string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	deleteInstancesRequest := &ecs20140526.DeleteInstancesRequest{
		InstanceId: tea.StringSlice(instances),
		RegionId:   tea.String(regionId),
	}
	_, _err = client.DeleteInstances(deleteInstancesRequest)
	if _err != nil {
		log.Errorf(ctx, "delete aliyun instances failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "delete aliyun instances failed")
	}
	return spec.Success()
}

// describe instances status
func describeInstancesStatus(ctx context.Context, accessKeyId, accessKeySecret, regionId string, instances []string) (_result map[string]string, _err error) {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return _result, _err
	}
	describeInstanceStatusRequest := &ecs20140526.DescribeInstanceStatusRequest{
		InstanceId: tea.StringSlice(instances),
		RegionId:   tea.String(regionId),
	}
	response, _err := client.DescribeInstanceStatus(describeInstanceStatusRequest)
	if _err != nil {
		log.Errorf(ctx, "describe aliyun instances status failed, err: %s", _err.Error())
		return _result, _err
	}
	instanceStatusList := response.Body.InstanceStatuses.InstanceStatus
	statusMap := map[string]string{}
	for _, instanceStatus := range instanceStatusList {
		statusMap[*instanceStatus.InstanceId] = *instanceStatus.Status
	}
	_result = statusMap
	return _result, _err
}

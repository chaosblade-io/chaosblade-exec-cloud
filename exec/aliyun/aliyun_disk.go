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

	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/chaosblade-io/chaosblade-spec-go/log"
	"github.com/chaosblade-io/chaosblade-spec-go/spec"
	"github.com/chaosblade-io/chaosblade-spec-go/util"

	"github.com/chaosblade-io/chaosblade-exec-cloud/exec"
	"github.com/chaosblade-io/chaosblade-exec-cloud/exec/category"
)

const DiskBin = "chaos_aliyun_disk"

type DiskActionSpec struct {
	spec.BaseExpActionCommandSpec
}

func NewDiskActionSpec() spec.ExpActionCommandSpec {
	return &DiskActionSpec{
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
					Name: "type",
					Desc: "the operation of Disk, support delete etc",
				},
				&spec.ExpFlag{
					Name: "diskId",
					Desc: "the diskId",
				},
				&spec.ExpFlag{
					Name: "instanceId",
					Desc: "the instanceId",
				},
				&spec.ExpFlag{
					Name: "regionId",
					Desc: "the regionId of aliyun",
				},
			},
			ActionExecutor: &DiskExecutor{},
			ActionExample: `
# delete disk which disk id is i-x
blade create aliyun disk --accessKeyId xxx --accessKeySecret yyy --regionId cn-qingdao --type detach --instanceId i-x --diskId y`,
			ActionPrograms:   []string{DiskBin},
			ActionCategories: []string{category.Cloud + "_" + category.Aliyun + "_" + category.Disk},
		},
	}
}

func (*DiskActionSpec) Name() string {
	return "disk"
}

func (*DiskActionSpec) Aliases() []string {
	return []string{}
}

func (*DiskActionSpec) ShortDesc() string {
	return "do some aliyun diskId Operations, like detach"
}

func (b *DiskActionSpec) LongDesc() string {
	if b.ActionLongDesc != "" {
		return b.ActionLongDesc
	}
	return "do some aliyun diskId Operations, like detach"
}

type DiskExecutor struct {
	channel spec.Channel
}

func (*DiskExecutor) Name() string {
	return "disk"
}

func (be *DiskExecutor) Exec(uid string, ctx context.Context, model *spec.ExpModel) *spec.Response {
	if be.channel == nil {
		util.Errorf(uid, util.GetRunFuncName(), spec.ChannelNil.Msg)
		return spec.ResponseFailWithFlags(spec.ChannelNil)
	}
	accessKeyId := model.ActionFlags["accessKeyId"]
	accessKeySecret := model.ActionFlags["accessKeySecret"]
	operationType := model.ActionFlags["type"]
	diskId := model.ActionFlags["diskId"]
	instanceId := model.ActionFlags["instanceId"]
	regionId := model.ActionFlags["regionId"]

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

	if diskId == "" {
		log.Errorf(ctx, "diskId is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "diskId")
	}

	if instanceId == "" {
		log.Errorf(ctx, "instanceId is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "instanceId")
	}

	disksStatusMap, _err := describeDisksStatus(ctx, accessKeyId, accessKeySecret, regionId, instanceId)
	if _err != nil {
		return spec.ResponseFailWithFlags(spec.ParameterRequestFailed, "describe disks status failed")
	}

	if (disksStatusMap[diskId] != "In_use" && operationType == "detach") || (disksStatusMap[diskId] == "In_use" && operationType == "attach") {
		return be.stop(ctx, operationType, accessKeyId, accessKeySecret, regionId, diskId, instanceId)
	}
	return be.start(ctx, operationType, accessKeyId, accessKeySecret, regionId, diskId, instanceId)
}

func (be *DiskExecutor) start(ctx context.Context, operationType, accessKeyId, accessKeySecret, regionId, diskId, instanceId string) *spec.Response {
	switch operationType {
	case "detach":
		return detachDisk(ctx, accessKeyId, accessKeySecret, regionId, diskId, instanceId)
	case "attach":
		return attachDisk(ctx, accessKeyId, accessKeySecret, regionId, diskId, instanceId)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support detach)")
	}
	select {}
}

func (be *DiskExecutor) stop(ctx context.Context, operationType, accessKeyId, accessKeySecret, regionId, diskId, instanceId string) *spec.Response {
	switch operationType {
	case "detach":
		return attachDisk(ctx, accessKeyId, accessKeySecret, regionId, diskId, instanceId)
	case "attach":
		return detachDisk(ctx, accessKeyId, accessKeySecret, regionId, diskId, instanceId)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support detach)")
	}
	ctx = context.WithValue(ctx, "bin", DiskBin)
	return exec.Destroy(ctx, be.channel, "aliyun disk")
}

func (be *DiskExecutor) SetChannel(channel spec.Channel) {
	be.channel = channel
}

// detach disk
func detachDisk(ctx context.Context, accessKeyId, accessKeySecret, regionId, diskId, instanceId string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	detachDiskRequest := &ecs20140526.DetachDiskRequest{
		InstanceId:         tea.String(instanceId),
		DiskId:             tea.String(diskId),
		DeleteWithInstance: tea.Bool(false),
	}

	_, _err = client.DetachDisk(detachDiskRequest)
	if _err != nil {
		log.Errorf(ctx, "detach aliyun disk failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "detach aliyun disk failed")
	}
	return spec.Success()
}

// create disk
func attachDisk(ctx context.Context, accessKeyId, accessKeySecret, regionId, diskId, instanceId string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	attachDiskRequest := &ecs20140526.AttachDiskRequest{
		InstanceId: tea.String(instanceId),
		DiskId:     tea.String(diskId),
	}

	_, _err = client.AttachDisk(attachDiskRequest)
	if _err != nil {
		log.Errorf(ctx, "attach aliyun disk failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "attach aliyun disk failed")
	}
	return spec.Success()
}

// describe Disks status
func describeDisksStatus(ctx context.Context, accessKeyId, accessKeySecret, regionId, instanceId string) (_result map[string]string, _err error) {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return _result, _err
	}
	describeDisksRequest := &ecs20140526.DescribeDisksRequest{
		RegionId:   tea.String(regionId),
		InstanceId: tea.String(instanceId),
	}
	response, _err := client.DescribeDisks(describeDisksRequest)
	if _err != nil {
		log.Errorf(ctx, "describe aliyun Disk status failed, err: %s", _err.Error())
		return _result, _err
	}
	diskStatusList := response.Body.Disks.Disk
	statusMap := map[string]string{}
	for _, diskStatus := range diskStatusList {
		statusMap[*diskStatus.DiskId] = *diskStatus.Status
	}
	_result = statusMap
	return _result, _err
}

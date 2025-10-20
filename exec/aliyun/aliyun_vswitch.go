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

const VSwitchBin = "chaos_aliyun_vswitch"

type VSwitchActionSpec struct {
	spec.BaseExpActionCommandSpec
}

func NewVSwitchActionSpec() spec.ExpActionCommandSpec {
	return &VSwitchActionSpec{
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
					Desc: "the operation of VSwitch, support delete etc",
				},
				&spec.ExpFlag{
					Name: "vSwitchId",
					Desc: "the VSwitchId",
				},
			},
			ActionExecutor: &VSwitchExecutor{},
			ActionExample: `
# delete vSwitch which vSwitch id is i-x
blade create aliyun vSwitch --accessKeyId xxx --accessKeySecret yyy --type delete --vSwitchId i-x`,
			ActionPrograms:   []string{VSwitchBin},
			ActionCategories: []string{category.Cloud + "_" + category.Aliyun + "_" + category.VSwitch},
		},
	}
}

func (*VSwitchActionSpec) Name() string {
	return "vSwitch"
}

func (*VSwitchActionSpec) Aliases() []string {
	return []string{}
}

func (*VSwitchActionSpec) ShortDesc() string {
	return "do some aliyun vSwitchId Operations, like delete"
}

func (b *VSwitchActionSpec) LongDesc() string {
	if b.ActionLongDesc != "" {
		return b.ActionLongDesc
	}
	return "do some aliyun vSwitchId Operations, like delete"
}

type VSwitchExecutor struct {
	channel spec.Channel
}

func (*VSwitchExecutor) Name() string {
	return "vSwitch"
}

func (be *VSwitchExecutor) Exec(uid string, ctx context.Context, model *spec.ExpModel) *spec.Response {
	if be.channel == nil {
		util.Errorf(uid, util.GetRunFuncName(), spec.ChannelNil.Msg)
		return spec.ResponseFailWithFlags(spec.ChannelNil)
	}
	accessKeyId := model.ActionFlags["accessKeyId"]
	accessKeySecret := model.ActionFlags["accessKeySecret"]
	operationType := model.ActionFlags["type"]
	vSwitchId := model.ActionFlags["vSwitchId"]
	regionId := model.ActionFlags["regionId"]
	zoneId := model.ActionFlags["zoneId"]
	cidrBlock := model.ActionFlags["cidrBlock"]
	vpcId := model.ActionFlags["vpcId"]

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

	if vSwitchId == "" {
		log.Errorf(ctx, "vSwitchId is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "vSwitchId")
	}

	if zoneId == "" {
		log.Errorf(ctx, "zoneId is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "zoneId")
	}

	if cidrBlock == "" {
		log.Errorf(ctx, "cidrBlock is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "cidrBlock")
	}

	if vpcId == "" {
		log.Errorf(ctx, "vpcId is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "vpcId")
	}

	vSwitchesStatusMap, _err := describeVSwitchesStatus(ctx, accessKeyId, accessKeySecret, regionId)
	if _err != nil {
		log.Errorf(ctx, "describe VSwitches Status failed")
		return spec.ResponseFailWithFlags(spec.ParameterRequestFailed, "describe VSwitches Status failed")
	}

	if (vSwitchesStatusMap[vSwitchId] != "Available" && operationType == "delete") || (vSwitchesStatusMap[vSwitchId] == "Available" && operationType == "create") {
		return be.stop(ctx, operationType, accessKeyId, accessKeySecret, regionId, vSwitchId, zoneId, cidrBlock, vpcId)
	}
	return be.start(ctx, operationType, accessKeyId, accessKeySecret, regionId, vSwitchId, zoneId, cidrBlock, vpcId)
}

func (be *VSwitchExecutor) start(ctx context.Context, operationType, accessKeyId, accessKeySecret, regionId, vSwitchId, zoneId, cidrBlock, vpcId string) *spec.Response {
	switch operationType {
	case "delete":
		return deleteVSwitch(ctx, accessKeyId, accessKeySecret, regionId, vSwitchId)
	case "create":
		return createVSwitch(ctx, accessKeyId, accessKeySecret, regionId, zoneId, cidrBlock, vpcId)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support delete)")
	}
	select {}
}

func (be *VSwitchExecutor) stop(ctx context.Context, operationType, accessKeyId, accessKeySecret, regionId, vSwitchId, zoneId, cidrBlock, vpcId string) *spec.Response {
	switch operationType {
	case "delete":
		return createVSwitch(ctx, accessKeyId, accessKeySecret, regionId, zoneId, cidrBlock, vpcId)
	case "create":
		return deleteVSwitch(ctx, accessKeyId, accessKeySecret, regionId, vSwitchId)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support delete)")
	}
	ctx = context.WithValue(ctx, "bin", VSwitchBin)
	return exec.Destroy(ctx, be.channel, "aliyun vSwitch")
}

func (be *VSwitchExecutor) SetChannel(channel spec.Channel) {
	be.channel = channel
}

// delete vSwitch
func deleteVSwitch(ctx context.Context, accessKeyId, accessKeySecret, regionId, vSwitchId string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	deleteVSwitchRequest := &ecs20140526.DeleteVSwitchRequest{
		VSwitchId: tea.String(vSwitchId),
	}

	_, _err = client.DeleteVSwitch(deleteVSwitchRequest)
	if _err != nil {
		log.Errorf(ctx, "delete aliyun vSwitch failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "delete aliyun vSwitch failed")
	}
	return spec.Success()
}

// create vSwitch
func createVSwitch(ctx context.Context, accessKeyId, accessKeySecret, regionId, zoneId, cidrBlock, vpcId string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	createVSwitchRequest := &ecs20140526.CreateVSwitchRequest{
		ZoneId:    tea.String(zoneId),
		CidrBlock: tea.String(cidrBlock),
		VpcId:     tea.String(vpcId),
	}

	_, _err = client.CreateVSwitch(createVSwitchRequest)
	if _err != nil {
		log.Errorf(ctx, "create aliyun vSwitch failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun vSwitch failed")
	}
	return spec.Success()
}

// describe VSwitches status
func describeVSwitchesStatus(ctx context.Context, accessKeyId, accessKeySecret, regionId string) (_result map[string]string, _err error) {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return _result, _err
	}
	describeVSwitchesRequest := &ecs20140526.DescribeVSwitchesRequest{
		RegionId: tea.String(regionId),
	}
	response, _err := client.DescribeVSwitches(describeVSwitchesRequest)
	if _err != nil {
		log.Errorf(ctx, "describe aliyun VSwitch status failed, err: %s", _err.Error())
		return _result, _err
	}
	vSwitchStatusList := response.Body.VSwitches.VSwitch
	statusMap := map[string]string{}
	for _, vSwitchStatus := range vSwitchStatusList {
		statusMap[*vSwitchStatus.VSwitchId] = *vSwitchStatus.Status
	}
	_result = statusMap
	return _result, _err
}

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

const NetworkInterfaceBin = "chaos_aliyun_networkinterface"

type NetworkInterfaceActionSpec struct {
	spec.BaseExpActionCommandSpec
}

func NewNetworkInterfaceActionSpec() spec.ExpActionCommandSpec {
	return &NetworkInterfaceActionSpec{
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
					Name: "instanceId",
					Desc: "the ecs instanceId",
				},
				&spec.ExpFlag{
					Name: "networkInterfaceId",
					Desc: "the networkInterfaceId of aliyun",
				},
				&spec.ExpFlag{
					Name: "type",
					Desc: "the operation of NetworkInterface, support attach, detach etc",
				},
				&spec.ExpFlag{
					Name: "networkInterfaceId",
					Desc: "the NetworkInterfaceId",
				},
			},
			ActionExecutor: &NetworkInterfaceExecutor{},
			ActionExample: `
# attach networkInterface to instance i-x which networkInterface id is s-x
blade create aliyun networkInterface --accessKeyId xxx --accessKeySecret yyy --regionId cn-qingdao --type attach --networkInterfaceId s-x --instanceId i-x

# detach instance i-x from networkInterface which networkInterface id is s-x
blade create aliyun networkInterface --accessKeyId xxx --accessKeySecret yyy --regionId cn-qingdao --type detach --networkInterfaceId s-x --instanceId i-x`,
			ActionPrograms:   []string{NetworkInterfaceBin},
			ActionCategories: []string{category.Cloud + "_" + category.Aliyun + "_" + category.NetworkInterface},
		},
	}
}

func (*NetworkInterfaceActionSpec) Name() string {
	return "networkInterface"
}

func (*NetworkInterfaceActionSpec) Aliases() []string {
	return []string{}
}

func (*NetworkInterfaceActionSpec) ShortDesc() string {
	return "do some aliyun networkInterfaceId Operations, like detach, attach"
}

func (b *NetworkInterfaceActionSpec) LongDesc() string {
	if b.ActionLongDesc != "" {
		return b.ActionLongDesc
	}
	return "do some aliyun networkInterfaceId Operations, like detach, attach"
}

type NetworkInterfaceExecutor struct {
	channel spec.Channel
}

func (*NetworkInterfaceExecutor) Name() string {
	return "networkInterface"
}

func (be *NetworkInterfaceExecutor) Exec(uid string, ctx context.Context, model *spec.ExpModel) *spec.Response {
	if be.channel == nil {
		util.Errorf(uid, util.GetRunFuncName(), spec.ChannelNil.Msg)
		return spec.ResponseFailWithFlags(spec.ChannelNil)
	}
	accessKeyId := model.ActionFlags["accessKeyId"]
	accessKeySecret := model.ActionFlags["accessKeySecret"]
	regionId := model.ActionFlags["regionId"]
	operationType := model.ActionFlags["type"]
	networkInterfaceId := model.ActionFlags["networkInterfaceId"]
	instanceId := model.ActionFlags["instanceId"]

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

	if networkInterfaceId == "" {
		log.Errorf(ctx, "networkInterfaceId is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "networkInterfaceId")
	}

	if instanceId != "" && networkInterfaceId != "" {
		log.Errorf(ctx, "instanceId and networkInterfaceId can not exist both!")
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "instanceId and networkInterfaceId can not exist both")
	}

	networkInterfaceStatusMap, _err := describeNetworkInterfaceStatus(ctx, accessKeyId, accessKeySecret, regionId, networkInterfaceId, instanceId)
	if _err != nil {
		return spec.ResponseFailWithFlags(spec.ParameterRequestFailed, "describe networkInterface status failed")
	}
	if (networkInterfaceStatusMap[networkInterfaceId] != "InUse" && operationType == "detach") || (networkInterfaceStatusMap[networkInterfaceId] == "InUse" && operationType == "attach") {
		return be.stop(ctx, operationType, accessKeyId, accessKeySecret, regionId, networkInterfaceId, instanceId)
	}

	return be.start(ctx, operationType, accessKeyId, accessKeySecret, regionId, networkInterfaceId, instanceId)
}

func (be *NetworkInterfaceExecutor) start(ctx context.Context, operationType, accessKeyId, accessKeySecret, regionId, networkInterfaceId, instanceId string) *spec.Response {
	switch operationType {
	// case "delete":
	//	return deleteNetworkInterface(ctx, accessKeyId, accessKeySecret, regionId, networkInterfaceId)
	case "detach":
		return detachNetworkInterfaceFromInstance(ctx, accessKeyId, accessKeySecret, regionId, networkInterfaceId, instanceId)
	case "attach":
		return attachNetworkInterfaceFromInstance(ctx, accessKeyId, accessKeySecret, regionId, networkInterfaceId, instanceId)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support attach, detach)")
	}
	select {}
}

func (be *NetworkInterfaceExecutor) stop(ctx context.Context, operationType, accessKeyId, accessKeySecret, regionId, networkInterfaceId, instanceId string) *spec.Response {
	switch operationType {
	// case "delete":
	//	return deleteNetworkInterface(ctx, accessKeyId, accessKeySecret, regionId, networkInterfaceId)
	case "detach":
		return attachNetworkInterfaceFromInstance(ctx, accessKeyId, accessKeySecret, regionId, networkInterfaceId, instanceId)
	case "attach":
		return detachNetworkInterfaceFromInstance(ctx, accessKeyId, accessKeySecret, regionId, networkInterfaceId, instanceId)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support attach, detach)")
	}
	ctx = context.WithValue(ctx, "bin", NetworkInterfaceBin)
	return exec.Destroy(ctx, be.channel, "aliyun networkInterface")
}

func (be *NetworkInterfaceExecutor) SetChannel(channel spec.Channel) {
	be.channel = channel
}

// delete networkInterface
func deleteNetworkInterface(ctx context.Context, accessKeyId, accessKeySecret, regionId, networkInterfaceId string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	deleteNetworkInterfaceRequest := &ecs20140526.DeleteNetworkInterfaceRequest{
		RegionId:           tea.String(regionId),
		NetworkInterfaceId: tea.String(networkInterfaceId),
	}
	_, _err = client.DeleteNetworkInterface(deleteNetworkInterfaceRequest)

	if _err != nil {
		log.Errorf(ctx, "delete aliyun network interface failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "delete aliyun network interface failed")
	}
	return spec.Success()
}

// detach networkInterface from instance
func detachNetworkInterfaceFromInstance(ctx context.Context, accessKeyId, accessKeySecret, regionId, networkInterfaceId, instanceId string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}
	detachNetworkInterfaceRequest := &ecs20140526.DetachNetworkInterfaceRequest{
		RegionId:           tea.String(regionId),
		NetworkInterfaceId: tea.String(networkInterfaceId),
		InstanceId:         tea.String(instanceId),
	}
	_, _err = client.DetachNetworkInterface(detachNetworkInterfaceRequest)
	if _err != nil {
		log.Errorf(ctx, "detach aliyun network interface failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "detach aliyun network interface failed")
	}
	return spec.Success()
}

// attach networkInterface from instance
func attachNetworkInterfaceFromInstance(ctx context.Context, accessKeyId, accessKeySecret, regionId, networkInterfaceId, instanceId string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}
	attachNetworkInterfaceRequest := &ecs20140526.AttachNetworkInterfaceRequest{
		RegionId:           tea.String(regionId),
		NetworkInterfaceId: tea.String(networkInterfaceId),
		InstanceId:         tea.String(instanceId),
	}
	_, _err = client.AttachNetworkInterface(attachNetworkInterfaceRequest)
	if _err != nil {
		log.Errorf(ctx, "attach aliyun network interface failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "attach aliyun network interface failed")
	}
	return spec.Success()
}

// describe networkInterface status
func describeNetworkInterfaceStatus(ctx context.Context, accessKeyId, accessKeySecret, regionId, networkInterfaceId, instanceId string) (_result map[string]string, _err error) {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return _result, _err
	}
	describeNetworkInterfacesRequest := &ecs20140526.DescribeNetworkInterfacesRequest{
		InstanceId:         tea.String(instanceId),
		RegionId:           tea.String(regionId),
		NetworkInterfaceId: []*string{tea.String(networkInterfaceId)},
	}
	response, _err := client.DescribeNetworkInterfaces(describeNetworkInterfacesRequest)
	if _err != nil {
		log.Errorf(ctx, "describe aliyun networkInterface status failed, err: %s", _err.Error())
		return _result, _err
	}
	networkInterfaceStatusList := response.Body.NetworkInterfaceSets.NetworkInterfaceSet
	statusMap := map[string]string{}
	for _, networkInterfaceStatus := range networkInterfaceStatusList {
		statusMap[*networkInterfaceStatus.NetworkInterfaceId] = *networkInterfaceStatus.Status
	}
	_result = statusMap
	return _result, _err
}

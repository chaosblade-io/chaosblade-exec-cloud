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

const PublicIpBin = "chaos_aliyun_publicip"

type PublicIpActionSpec struct {
	spec.BaseExpActionCommandSpec
}

func NewPublicIpActionSpec() spec.ExpActionCommandSpec {
	return &PublicIpActionSpec{
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
					Desc: "the operation of PublicIp, support release, unassociate, etc",
				},
				&spec.ExpFlag{
					Name: "allocationId",
					Desc: "the allocationId",
				},
				&spec.ExpFlag{
					Name: "regionId",
					Desc: "the regionId of aliyun",
				},
				&spec.ExpFlag{
					Name: "publicIpAddress",
					Desc: "the PublicIpAddress",
				},
			},
			ActionExecutor: &PublicIpExecutor{},
			ActionExample: `
# release publicIp which publicIpAddress is 1.1.1.1
blade create aliyun publicIp --accessKeyId xxx --accessKeySecret yyy --type release --publicIpAddress 1.1.1.1

# unassociate publicIp from instance i-x which allocationId id is a-x
blade create aliyun publicIp --accessKeyId xxx --accessKeySecret yyy --type unassociate --instanceId i-x --allocationId a-x`,
			ActionPrograms:   []string{PublicIpBin},
			ActionCategories: []string{category.Cloud + "_" + category.Aliyun + "_" + category.PublicIp},
		},
	}
}

func (*PublicIpActionSpec) Name() string {
	return "publicIp"
}

func (*PublicIpActionSpec) Aliases() []string {
	return []string{}
}

func (*PublicIpActionSpec) ShortDesc() string {
	return "do some aliyun publicIp Operations, like release, unassociate"
}

func (b *PublicIpActionSpec) LongDesc() string {
	if b.ActionLongDesc != "" {
		return b.ActionLongDesc
	}
	return "do some aliyun publicIp Operations, like release, unassociate"
}

type PublicIpExecutor struct {
	channel spec.Channel
}

func (*PublicIpExecutor) Name() string {
	return "publicIp"
}

func (be *PublicIpExecutor) Exec(uid string, ctx context.Context, model *spec.ExpModel) *spec.Response {
	if be.channel == nil {
		util.Errorf(uid, util.GetRunFuncName(), spec.ChannelNil.Msg)
		return spec.ResponseFailWithFlags(spec.ChannelNil)
	}
	accessKeyId := model.ActionFlags["accessKeyId"]
	accessKeySecret := model.ActionFlags["accessKeySecret"]
	operationType := model.ActionFlags["type"]
	regionId := model.ActionFlags["regionId"]
	instanceId := model.ActionFlags["instanceId"]
	allocationId := model.ActionFlags["allocationId"]
	publicIpAddress := model.ActionFlags["publicIpAddress"]

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

	if operationType == "release" && publicIpAddress == "" {
		log.Errorf(ctx, "publicIpAddress is required when operationType is release!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "publicIpAddress")
	}

	if operationType == "unassociate" && allocationId == "" {
		log.Errorf(ctx, "allocationId is required when operationType is unassociate!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "allocationId")
	}

	if operationType == "unassociate" && instanceId == "" {
		log.Errorf(ctx, "instanceId is required when operationType is unassociate!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "instanceId")
	}

	if operationType == "release" || operationType == "associate" {
		ipStatusMap, _err := describeInstances(ctx, accessKeyId, accessKeySecret, regionId, instanceId)
		if _err != nil {
			return spec.ResponseFailWithFlags(spec.ParameterRequestFailed, "describe ip status failed")
		}
		isExist := false
		for i := 0; i < len(ipStatusMap[instanceId]); i++ {
			if ipStatusMap[instanceId][i] == publicIpAddress {
				isExist = true
			}
		}
		if (!isExist && operationType == "release") || (isExist && operationType == "associate") {
			return be.stop(ctx, operationType, accessKeyId, accessKeySecret, regionId, allocationId, instanceId, publicIpAddress)
		}
	}

	if operationType == "unassociateEip" || operationType == "associateEip" {
		eipStatusMap, _err := describeEipAddresses(ctx, accessKeyId, accessKeySecret, regionId, allocationId, publicIpAddress)
		if _err != nil {
			return spec.ResponseFailWithFlags(spec.ParameterRequestFailed, "describe eip status failed")
		}
		if (eipStatusMap[publicIpAddress] != "InUse" && operationType == "unassociateEip") || (eipStatusMap[publicIpAddress] == "InUse" && operationType == "associateEip") {
			return be.stop(ctx, operationType, accessKeyId, accessKeySecret, regionId, allocationId, instanceId, publicIpAddress)
		}
	}
	return be.start(ctx, operationType, accessKeyId, accessKeySecret, regionId, allocationId, instanceId, publicIpAddress)
}

func (be *PublicIpExecutor) start(ctx context.Context, operationType, accessKeyId, accessKeySecret, regionId, allocationId, instanceId, publicIpAddress string) *spec.Response {
	switch operationType {
	case "release":
		return releasePublicIpAddress(ctx, accessKeyId, accessKeySecret, regionId, publicIpAddress, instanceId)
	case "associate":
		return allocatePublicIpAddress(ctx, accessKeyId, accessKeySecret, regionId, publicIpAddress, instanceId)
	case "unassociateEip":
		return unassociateEipAddress(ctx, accessKeyId, accessKeySecret, regionId, allocationId, instanceId)
	case "associateEip":
		return associateEipAddress(ctx, accessKeyId, accessKeySecret, regionId, allocationId, instanceId)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support release, associate, unassociateEip, associateEip)")
	}
	select {}
}

func (be *PublicIpExecutor) stop(ctx context.Context, operationType, accessKeyId, accessKeySecret, regionId, allocationId, instanceId, publicIpAddress string) *spec.Response {
	switch operationType {
	case "release":
		return allocatePublicIpAddress(ctx, accessKeyId, accessKeySecret, regionId, publicIpAddress, instanceId)
	case "associate":
		return releasePublicIpAddress(ctx, accessKeyId, accessKeySecret, regionId, publicIpAddress, instanceId)
	case "unassociateEip":
		return associateEipAddress(ctx, accessKeyId, accessKeySecret, regionId, allocationId, instanceId)
	case "associateEip":
		return unassociateEipAddress(ctx, accessKeyId, accessKeySecret, regionId, allocationId, instanceId)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support release, associate, unassociateEip, associateEip)")
	}
	ctx = context.WithValue(ctx, "bin", PublicIpBin)
	return exec.Destroy(ctx, be.channel, "aliyun public Ip")
}

func (be *PublicIpExecutor) SetChannel(channel spec.Channel) {
	be.channel = channel
}

// release Public Ip
func releasePublicIpAddress(ctx context.Context, accessKeyId, accessKeySecret, regionId, publicIpAddress, instanceId string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	if instanceId != "" {
		releasePublicIpAddressRequest := &ecs20140526.ReleasePublicIpAddressRequest{
			PublicIpAddress: tea.String(publicIpAddress),
			InstanceId:      tea.String(instanceId),
		}
		_, _err = client.ReleasePublicIpAddress(releasePublicIpAddressRequest)
	} else {
		releasePublicIpAddressRequest := &ecs20140526.ReleasePublicIpAddressRequest{
			PublicIpAddress: tea.String(publicIpAddress),
		}
		_, _err = client.ReleasePublicIpAddress(releasePublicIpAddressRequest)
	}
	if _err != nil {
		log.Errorf(ctx, "release aliyun public Ip failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "release aliyun public Ip failed")
	}
	return spec.Success()
}

// allocate Public Ip
func allocatePublicIpAddress(ctx context.Context, accessKeyId, accessKeySecret, regionId, publicIpAddress, instanceId string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	if instanceId != "" {
		allocatePublicIpAddressRequest := &ecs20140526.AllocatePublicIpAddressRequest{
			IpAddress:  tea.String(publicIpAddress),
			InstanceId: tea.String(instanceId),
		}
		_, _err = client.AllocatePublicIpAddress(allocatePublicIpAddressRequest)
	}
	if _err != nil {
		log.Errorf(ctx, "allocate aliyun public Ip failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "allocate aliyun public Ip failed")
	}
	return spec.Success()
}

// unassociate Eip Address
func unassociateEipAddress(ctx context.Context, accessKeyId, accessKeySecret, regionId, allocationId, instanceId string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	if regionId != "" {
		unassociateEipAddressRequest := &ecs20140526.UnassociateEipAddressRequest{
			AllocationId: tea.String(allocationId),
			InstanceId:   tea.String(instanceId),
			RegionId:     tea.String(regionId),
		}
		_, _err = client.UnassociateEipAddress(unassociateEipAddressRequest)
	} else {
		unassociateEipAddressRequest := &ecs20140526.UnassociateEipAddressRequest{
			AllocationId: tea.String(allocationId),
			InstanceId:   tea.String(instanceId),
		}
		_, _err = client.UnassociateEipAddress(unassociateEipAddressRequest)
	}
	if _err != nil {
		log.Errorf(ctx, "unassociate aliyun Eip Address failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "unassociate aliyun Eip Address failed")
	}
	return spec.Success()
}

// associate Eip Address
func associateEipAddress(ctx context.Context, accessKeyId, accessKeySecret, regionId, allocationId, instanceId string) *spec.Response {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aliyun client failed")
	}

	if regionId != "" {
		associateEipAddressRequest := &ecs20140526.AssociateEipAddressRequest{
			AllocationId: tea.String(allocationId),
			InstanceId:   tea.String(instanceId),
			RegionId:     tea.String(regionId),
		}
		_, _err = client.AssociateEipAddress(associateEipAddressRequest)
	} else {
		associateEipAddressRequest := &ecs20140526.AssociateEipAddressRequest{
			AllocationId: tea.String(allocationId),
			InstanceId:   tea.String(instanceId),
		}
		_, _err = client.AssociateEipAddress(associateEipAddressRequest)
	}
	if _err != nil {
		log.Errorf(ctx, "associate aliyun Eip Address failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "associate aliyun Eip Address failed")
	}
	return spec.Success()
}

// describe eip addresses status
func describeEipAddresses(ctx context.Context, accessKeyId, accessKeySecret, regionId, allocationId, eipAddress string) (_result map[string]string, _err error) {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return _result, _err
	}
	describeEipAddressesRequest := &ecs20140526.DescribeEipAddressesRequest{
		AllocationId: tea.String(allocationId),
		RegionId:     tea.String(regionId),
		EipAddress:   tea.String(eipAddress),
	}
	response, _err := client.DescribeEipAddresses(describeEipAddressesRequest)
	if _err != nil {
		log.Errorf(ctx, "describe aliyun eip addresses status failed, err: %s", _err.Error())
		return _result, _err
	}
	eipAddressStatusList := response.Body.EipAddresses.EipAddress
	statusMap := map[string]string{}
	for _, eipAddressStatus := range eipAddressStatusList {
		statusMap[*eipAddressStatus.IpAddress] = *eipAddressStatus.Status
	}
	_result = statusMap
	return _result, _err
}

// describe instances status
func describeInstances(ctx context.Context, accessKeyId, accessKeySecret, regionId, instanceId string) (_result map[string][]string, _err error) {
	client, _err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret), regionId)
	if _err != nil {
		log.Errorf(ctx, "create aliyun client failed, err: %s", _err.Error())
		return _result, _err
	}
	describeInstancesRequest := &ecs20140526.DescribeInstancesRequest{
		InstanceIds: tea.String("[\"" + instanceId + "\"]"),
		RegionId:    tea.String(regionId),
	}
	response, _err := client.DescribeInstances(describeInstancesRequest)
	if _err != nil {
		log.Errorf(ctx, "describe aliyun instances status failed, err: %s", _err.Error())
		return _result, _err
	}
	instanceList := response.Body.Instances.Instance
	statusMap := map[string][]string{}
	for _, instanceStatus := range instanceList {
		var ipList []string
		for i := 0; i < len(instanceStatus.PublicIpAddress.IpAddress); i++ {
			ipList = append(ipList, *instanceStatus.PublicIpAddress.IpAddress[i])
		}
		statusMap[*instanceStatus.InstanceId] = ipList
	}
	_result = statusMap
	return _result, _err
}

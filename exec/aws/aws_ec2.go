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
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/chaosblade-io/chaosblade-spec-go/log"
	"github.com/chaosblade-io/chaosblade-spec-go/spec"
	"github.com/chaosblade-io/chaosblade-spec-go/util"

	"github.com/chaosblade-io/chaosblade-exec-cloud/exec"
	"github.com/chaosblade-io/chaosblade-exec-cloud/exec/category"
)

const Ec2Bin = "chaos_aws_ec2"

type EcsActionSpec struct {
	spec.BaseExpActionCommandSpec
}

func NewEcsActionSpec() spec.ExpActionCommandSpec {
	return &EcsActionSpec{
		spec.BaseExpActionCommandSpec{
			ActionFlags: []spec.ExpFlagSpec{
				&spec.ExpFlag{
					Name: "accessKeyId",
					Desc: "the accessKeyId of aws, if not provided, get from env ACCESS_KEY_ID",
				},
				&spec.ExpFlag{
					Name: "accessKeySecret",
					Desc: "the accessKeySecret of aws, if not provided, get from env ACCESS_KEY_SECRET",
				},
				&spec.ExpFlag{
					Name: "regionId",
					Desc: "the regionId of aws",
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
blade create aws ecs --accessKeyId xxx --accessKeySecret yyy --regionId us-west-2 --type stop --instances i-x,i-y

# start instances which instance id is i-x,i-y
blade create aws ecs --accessKeyId xxx --accessKeySecret yyy --regionId us-west-2 --type start --instances i-x,i-y

# reboot instances which instance id is i-x,i-y
blade create aws ecs --accessKeyId xxx --accessKeySecret yyy --regionId us-west-2 --type reboot --instances i-x,i-y`,
			ActionPrograms:   []string{Ec2Bin},
			ActionCategories: []string{category.Cloud + "_" + category.Aws + "_" + category.Ec2},
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
	return "do some aws ecs Operations, like stop, start, reboot"
}

func (b *EcsActionSpec) LongDesc() string {
	if b.ActionLongDesc != "" {
		return b.ActionLongDesc
	}
	return "do some aws ecs Operations, like stop, start, reboot"
}

type EcsExecutor struct {
	channel spec.Channel
}

func (*EcsExecutor) Name() string {
	return "ecs"
}

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
		return startAwsInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	case "stop":
		return stopAwsInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	case "reboot":
		return rebootAwsInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	// case "delete":
	//	return deleteAwsInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support start, stop, reboot)")
	}
	select {}
}

func (be *EcsExecutor) stop(ctx context.Context, operationType, accessKeyId, accessKeySecret, regionId string, instancesArray []string) *spec.Response {
	switch operationType {
	case "start":
		return stopAwsInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	case "stop":
		return startAwsInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	// case "reboot":
	//	return rebootAwsInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	// case "delete":
	//	return deleteAwsInstances(ctx, accessKeyId, accessKeySecret, regionId, instancesArray)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support start, stop, reboot)")
	}
	ctx = context.WithValue(ctx, "bin", Ec2Bin)
	return exec.Destroy(ctx, be.channel, "aws es")
}

func (be *EcsExecutor) SetChannel(channel spec.Channel) {
	be.channel = channel
}

func CreateConfig(accessKeyId, accessKeySecret, regionId string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(regionId),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessKeyId,
				SecretAccessKey: accessKeySecret,
				Source:          "chaosblade hard coded credentials",
			},
		}),
	)
	return cfg, err
}

// start instances
func startAwsInstances(ctx context.Context, accessKeyId, accessKeySecret, regionId string, instances []string) *spec.Response {
	cfg, _err := CreateConfig(accessKeyId, accessKeySecret, regionId)
	if _err != nil {
		log.Errorf(ctx, "create aws config failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aws config failed")
	}
	client := ec2.NewFromConfig(cfg)

	input := &ec2.StartInstancesInput{
		InstanceIds: instances,
	}

	_, _err = client.StartInstances(context.TODO(), input)
	if _err != nil {
		log.Errorf(ctx, "start aws instances failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "start aws instances failed")
	}
	return spec.Success()
}

// stop instances
func stopAwsInstances(ctx context.Context, accessKeyId, accessKeySecret, regionId string, instances []string) *spec.Response {
	cfg, _err := CreateConfig(accessKeyId, accessKeySecret, regionId)
	if _err != nil {
		log.Errorf(ctx, "create aws config failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aws config failed")
	}
	client := ec2.NewFromConfig(cfg)

	input := &ec2.StopInstancesInput{
		InstanceIds: instances,
	}
	_, _err = client.StopInstances(context.TODO(), input)
	if _err != nil {
		log.Errorf(ctx, "stop aws instances failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "stop aws instances failed")
	}
	return spec.Success()
}

// reboot instances
func rebootAwsInstances(ctx context.Context, accessKeyId, accessKeySecret, regionId string, instances []string) *spec.Response {
	cfg, _err := CreateConfig(accessKeyId, accessKeySecret, regionId)
	if _err != nil {
		log.Errorf(ctx, "create aws config failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aws config failed")
	}
	client := ec2.NewFromConfig(cfg)

	input := &ec2.RebootInstancesInput{
		InstanceIds: instances,
	}
	_, _err = client.RebootInstances(context.TODO(), input)
	if _err != nil {
		log.Errorf(ctx, "reboot aws instances failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "restart aws instances failed")
	}
	return spec.Success()
}

// delete instances
func deleteAwsInstances(ctx context.Context, accessKeyId, accessKeySecret, regionId string, instances []string) *spec.Response {
	cfg, _err := CreateConfig(accessKeyId, accessKeySecret, regionId)
	if _err != nil {
		log.Errorf(ctx, "create aws config failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create aws config failed")
	}
	client := ec2.NewFromConfig(cfg)

	input := &ec2.TerminateInstancesInput{
		InstanceIds: instances,
	}
	_, _err = client.TerminateInstances(context.TODO(), input)
	if _err != nil {
		log.Errorf(ctx, "delete aws instances failed, err: %s", _err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "delete aws instances failed")
	}
	return spec.Success()
}

// describe instances status
func describeInstancesStatus(ctx context.Context, accessKeyId, accessKeySecret, regionId string, instances []string) (_result map[string]string, _err error) {
	cfg, _err := CreateConfig(accessKeyId, accessKeySecret, regionId)
	if _err != nil {
		log.Errorf(ctx, "create aws config failed, err: %s", _err.Error())
		return _result, _err
	}
	client := ec2.NewFromConfig(cfg)

	// Create an input object with the instance IDs
	input := &ec2.DescribeInstanceStatusInput{
		InstanceIds: instances,
	}

	// Describe the instance status
	resp, _err := client.DescribeInstanceStatus(context.TODO(), input)
	if _err != nil {
		log.Errorf(ctx, "describe aws instances status failed, err: %s", _err.Error())
		return _result, _err
	}

	statusMap := map[string]string{}
	for _, status := range resp.InstanceStatuses {
		statusMap[*status.InstanceId] = string(status.InstanceState.Name)
	}

	_result = statusMap
	return _result, _err
}

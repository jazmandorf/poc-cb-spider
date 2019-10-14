// Copyright 2017 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

func init() {
	scopes := strings.Join([]string{
		compute.DevstorageFullControlScope,
		compute.ComputeScope,
	}, " ")
	fmt.Println("init :", scopes)
	//computeMain()
	// registerDemo("compute", scopes, computeMain)
}

type Config struct {
	Type         string `json:"type"`
	ProjectID    string `json:"project_id"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	ClientID     string `json:"client_id"`
	AuthURI      string `json:"auth_uri"`
	TokenURI     string `json:"token_uri"`
	AuthProvider string `json:"auth_provider_x509_cert_url"`
}
type InstanceInfo struct {
	zone         string
	region       string
	instnaceName string
}

type vmInstanceInfo struct {
	VMId string
}

func createInstance(service *compute.Service, conf Config, zone string, vmname string, diskname string) {

	projectID := conf.ProjectID

	prefix := "https://www.googleapis.com/compute/v1/projects/" + projectID
	imageURL := "https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/debian-7-wheezy-v20140606"
	zone = zone
	instanceName := vmname

	// Show the current images that are available.
	// res, err := service.Images.List(projectID).Do()
	// log.Printf("Got compute.Images.List, err: %#v, %v", res, err)

	instance := &compute.Instance{
		Name:        instanceName,
		Description: "compute sample instance",
		MachineType: prefix + "/zones/" + zone + "/machineTypes/n1-standard-1",
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    diskname,
					SourceImage: imageURL,
				},
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				AccessConfigs: []*compute.AccessConfig{
					{
						Type: "ONE_TO_ONE_NAT",
						Name: "External NAT",
					},
				},
				Network: prefix + "/global/networks/default",
			},
		},
		ServiceAccounts: []*compute.ServiceAccount{
			{
				Email: conf.ClientEmail,
				Scopes: []string{
					compute.DevstorageFullControlScope,
					compute.ComputeScope,
				},
			},
		},
	}

	op, err := service.Instances.Insert(projectID, zone, instance).Do()
	js, err := op.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Insert vm to marshal Json : ", string(js))
	log.Printf("Got compute.Operation, err: %#v, %v", op, err)
	etag := op.Header.Get("Etag")
	log.Printf("Etag=%v", etag)

	inst, err := service.Instances.Get(projectID, zone, instanceName).IfNoneMatch(etag).Do()
	log.Printf("Got compute.Instance, err: %#v, %v", inst, err)
	if googleapi.IsNotModified(err) {
		log.Printf("Instance not modified since insert.")
	} else {
		log.Printf("Instance modified since insert.")
	}

}
func getPublicIP(instance *compute.Instance) {
	fmt.Println("network Interface : ", instance.NetworkInterfaces[0].AccessConfigs[0].Name)
}
func getInstance(ctx context.Context, service *compute.Service, zone string, instanceName string, conf Config) *compute.Instance {
	/// ctx := context.Background()
	inst, err := service.Instances.Get(conf.ProjectID, zone, instanceName).Context(ctx).Do()
	//log.Printf("Got compute.Instance, err: %#v, %v", inst, err)
	if err != nil {
		log.Fatal(err)
	}
	js, err := inst.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("get Instance Marshal Json : ", string(js))
	fmt.Println("Instance status :", inst.Status)

	return inst
}

func stopVM(ctx context.Context, service *compute.Service, zone string, instanceName string, conf Config) (string, error) {
	// ctx := context.Background()

	inst, err := service.Instances.Stop(conf.ProjectID, zone, instanceName).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}
	js, err := inst.MarshalJSON()
	fmt.Println("Instance marshaljson :", string(js))
	log.Printf("Instances Stop, err: %#v, %v", inst, err)
	fmt.Println("Instance Status :", inst.Status)
	return inst.Status, err
}

func startVM(ctx context.Context, service *compute.Service, zone string, instanceName string, conf Config) (string, error) {

	getInst := getInstance(ctx, service, zone, instanceName, conf)

	if getInst.Status == "TERMINATED" {
		fmt.Println("터미네이터 이다.")
	}

	inst, err := service.Instances.Start(conf.ProjectID, zone, instanceName).Context(ctx).Do()
	js, err := inst.MarshalJSON()
	fmt.Println("Instance marshaljson :", string(js))
	log.Printf("StartVM, err: %#v, %v", inst, err)
	fmt.Println("Status :", inst.Status)
	fmt.Println("VM type : ", reflect.TypeOf(inst))
	return inst.Status, err
}

func deleteVM(ctx context.Context, service *compute.Service, zone string, instanceName string, conf Config) (string, error) {
	//ctx := context.Background()
	inst, err := service.Instances.Delete(conf.ProjectID, zone, instanceName).Context(ctx).Do()
	js, err := inst.MarshalJSON()
	fmt.Println("Instance marshaljson :", string(js))
	log.Printf("StartVM, err: %#v, %v", inst, err)
	fmt.Println("Status :", inst.Status)
	fmt.Println("VM type : ", reflect.TypeOf(inst))
	return inst.Status, err
}

func rebootVM(ctx context.Context, service *compute.Service, zone string, instanceName string, conf Config) (string, error) {
	//ctx := context.Background()
	st, err := stopVM(ctx, service, zone, instanceName, conf)
	if err != nil {
		log.Fatal(err)
	}

	return st, err
}

func ListPublicIP(ctx context.Context, service *compute.Service, conf Config, region string) (string, string) {
	list, err := service.Addresses.List(conf.ProjectID, region).Context(ctx).Do()
	listInfo, err := list.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("ListPublicIP Info : ", string(listInfo))
	fmt.Println("ListPublicIP[0] Name : ", list.Items[0].Name)
	fmt.Println("ListPublicIP[0] Address : ", list.Items[0].Address)
	//log.Printf("getGlovalAddressList, err: %#v, %v", list, err)
	name := list.Items[0].Name
	address := list.Items[0].Address
	return name, address
}
func getGlobalAddressList(ctx context.Context, service *compute.Service, config Config) {

	list, err := service.GlobalAddresses.List(config.ProjectID).Context(ctx).Do()
	log.Printf("getGlovalAddressList, err: %#v, %v", list, err)

}

func readFileConfig(filepath string) (Config, error) {

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	var config Config
	json.Unmarshal(data, &config)
	fmt.Println("readFileConfig Json : ", config.ClientEmail)

	return config, err

}

func connect(filePath string) *compute.Service {

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	s := string(data)
	d := []byte(s)
	conf, err := google.JWTConfigFromJSON(d, "https://www.googleapis.com/auth/compute")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connection Success : ", reflect.TypeOf(d))

	client := conf.Client(oauth2.NoContext)

	computeService, err := compute.New(client)

	return computeService

}

func CreatePublicIP(ctx context.Context, service *compute.Service, name string, region string, conf Config) {
	address := &compute.Address{
		Name: name,
	}
	info, err := service.Addresses.Insert(conf.ProjectID, region, address).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}
	infoJson, err := info.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("createPublicIP Info : ", string(infoJson))

}

func ListVM(ctx context.Context, service *compute.Service, zone string, conf Config) []byte {
	list, err := service.Instances.List(conf.ProjectID, zone).Do()
	if err != nil {
		log.Fatal(err)
	}

	listJson, err := list.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("List Vm : ", string(listJson))

	return listJson
}

func ListImage(ctx context.Context, service *compute.Service, conf Config) []byte {
	projectID := conf.ProjectID
	list, err := service.Images.List(projectID).Do()
	log.Printf("Got compute.Images.List, err: %#v, %v", list, err)
	req := service.Images.List(projectID)
	if err := req.Pages(ctx, func(page *compute.ImageList) error {
		for i, image := range page.Items {
			// TODO: Change code below to process each `image` resource:
			fmt.Printf("get ImagetList : %#v\n", image, i)
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	imageListJson, err := list.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("List Vm : ", string(imageListJson))

	return imageListJson
}

func main() {
	credentialFilePath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	config, _ := readFileConfig(credentialFilePath)
	//zone := "asia-northeast1-b"
	//instanceName := "cscmcloud"
	//diskname := "mzcsc21"
	region := "asia-northeast1"
	ctx := context.Background()

	client := connect(credentialFilePath)

	fmt.Println(reflect.TypeOf(client))
	fmt.Println("config Project ID : ", config.ProjectID)

	//createInstance(client, config, zone, instanceName, diskname)
	//instance := getInstance(ctx, client, zone, instanceName, config)
	//fmt.Println("output instance : ", instance)
	//getInstance(ctx, client, zone, instanceName, config)
	//stopVM(ctx, client, zone, instanceName, config)
	//startVM(ctx, client, zone, instanceName, config)
	//getGlobalAddressList(ctx, client, config)
	//getPublicIP(instance)
	CreatePublicIP(ctx, client, "staticip3", region, config)
	//name, address := ListPublicIP(ctx, client, config, region)
	//fmt.Println("output name : ", name)
	//fmt.Println("output address : ", address)
	//getVMlist := ListVM(ctx, client, zone, config)
	//fmt.Println("getVMList : ", string(getVMlist))
	getImagelist := ListImage(ctx, client, config)
	fmt.Println("getVMList : ", string(getImagelist))

}

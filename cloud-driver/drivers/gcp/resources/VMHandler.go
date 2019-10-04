// Proof of Concepts of CB-Spider.
// The CB-Spider is a sub-Framework of the Cloud-Barista Multi-Cloud Project.
// The CB-Spider Mission is to connect all the clouds with a single interface.
//
//      * Cloud-Barista: https://github.com/cloud-barista
//
// This is a Cloud Driver Example for PoC Test.
//
// by hyokyung.kim@innogrid.co.kr, 2019.07.

package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	compute "google.golang.org/api/compute/v1"

	"github.com/Azure/go-autorest/autorest/to"
	idrv "github.com/cloud-barista/poc-cb-spider/cloud-driver/interfaces"
	irs "github.com/cloud-barista/poc-cb-spider/cloud-driver/interfaces/resources"
)

type GCPVMHandler struct {
	Region idrv.RegionInfo
	Ctx    context.Context
	Client *compute.Service
	Credential idrv.CredentialInfo
}

func (vmHandler *GCPVMHandler) StartVM(vmReqInfo irs.VMReqInfo) (irs.VMInfo, error) {
	// Set VM Create Information
	// GCP 는 reqinfo에 ProjectID를 받아야 함.
	ctx := vmHandler.Ctx
	vmName := vmReqInfo.Name
	projectID := vmHandler.Credential.ProjectID
	prefix := "https://www.googleapis.com/compute/v1/projects/" + projectID
	imageURL := "https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/debian-7-wheezy-v20140606"
	zone := vmHandler.Region.Zone
	// instanceName := "cscmcloud"

	instance := &compute.Client.Instance{
		Name:        vmName,
		Description: "compute sample instance",
		MachineType: prefix + "/zones/" + zone + "/machineTypes/n1-standard-1",
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    "my-root-pd",
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
	
	op, err := vmHandler.Client.Instances.Insert(projectID, zone, instance).Do()
	
	
	// 이게 시작하는  api Start 내부 매개변수로 projectID, zone, InstanceID
	//vm, err := vmHandler.Client.Instances.Start(project string, zone string, instance string)
	vm, err := vmHandler.Client.Instances.Get(projectID,zone,vmName).Context(ctx).Do()
	if err != nil {
		panic(err)
	}

	vmInfo := mappingServerInfo(vm) // 따로 맵핑 정보 맞춰야 함.

	return vmInfo, nil
}
// stop이라고 보면 될듯
func (vmHandler *GCPVMHandler) SuspendVM(vmID string) {
	projectID := vmHandler.Credential.ProjectID
	zone := vmHandler.Region.Zone
	ctx := vmHandler.Ctx

	inst, err := service.Instances.Stop(projectID, zone, vmID).Context(ctx).Do()
	
	if err != nil {
		panic(err)
	}

	fmt.Println("instance stop status :", inst.Status)	
}

func (vmHandler *GCPVMHandler) ResumeVM(vmID string) {

	projectID := vmHandler.Credential.ProjectID
	zone := vmHandler.Region.Zone
	ctx := vmHandler.Ctx

	inst, err := service.Instances.Start(projectID, zone, vmID).Context(ctx).Do()
	
	if err != nil {
		panic(err)
	}

	fmt.Println("instance resume status :", inst.Status)	
	
}

func (vmHandler *GCPVMHandler) RebootVM(vmID string) {
	vmIdArr := strings.Split(vmID, ":")

	future, err := vmHandler.Client.Restart(vmHandler.Ctx, vmIdArr[0], vmIdArr[1])
	if err != nil {
		panic(err)
	}
	err = future.WaitForCompletionRef(vmHandler.Ctx, vmHandler.Client.Client)
	if err != nil {
		panic(err)
	}
}

func (vmHandler *GCPVMHandler) TerminateVM(vmID string) {
	projectID := vmHandler.Credential.ProjectID
	zone := vmHandler.Region.Zone
	ctx := vmHandler.Ctx

	inst, err := service.Instances.Stop(projectID, zone, vmID).Context(ctx).Do()
	
	if err != nil {
		panic(err)
	}

	fmt.Println("instance status :", inst.Status)	
}

func (vmHandler *GCPVMHandler) ListVMStatus() []*irs.VMStatusInfo {
	//serverList, err := vmHandler.Client.ListAll(vmHandler.Ctx)
	serverList, err := vmHandler.Client.List(vmHandler.Ctx, vmHandler.Region.ResourceGroup)
	if err != nil {
		panic(err)
	}

	var vmStatusList []*irs.VMStatusInfo
	for _, s := range serverList.Values() {
		if s.InstanceView != nil {
			/*if s.InstanceView != nil {
			    statusStr := getVmStatus(*s.InstanceView)
			    status := irs.VMStatus(statusStr)
			    status := vmHandler.GetVMStatus(*s.ID)
			    vmStatusInfo := irs.VMStatusInfo{
			        VmId:     *s.ID,
			        VmStatus: status,
			    }
			    vmStatusList = append(vmStatusList, &vmStatusInfo)
			}*/
			vmIdArr := strings.Split(*s.ID, "/")
			vmId := vmIdArr[4] + ":" + vmIdArr[8]
			status := vmHandler.GetVMStatus(vmId)
			vmStatusInfo := irs.VMStatusInfo{
				VmId:     *s.ID,
				VmStatus: status,
			}
			vmStatusList = append(vmStatusList, &vmStatusInfo)
		}
	}

	return vmStatusList
}

func (vmHandler *GCPVMHandler) GetVMStatus(vmID string) irs.VMStatus {
	vmIdArr := strings.Split(vmID, ":")
	instanceView, err := vmHandler.Client.InstanceView(vmHandler.Ctx, vmIdArr[0], vmIdArr[1])
	if err != nil {
		panic(err)
	}

	// Get powerState, provisioningState
	vmStatus := getVmStatus(instanceView)
	return irs.VMStatus(vmStatus)
}

func (vmHandler *GCPVMHandler) ListVM() []*irs.VMInfo {
	//serverList, err := vmHandler.Client.ListAll(vmHandler.Ctx)
	serverList, err := vmHandler.Client.List(vmHandler.Ctx, vmHandler.Region.ResourceGroup)
	if err != nil {
		panic(err)
	}

	var vmList []*irs.VMInfo
	for _, server := range serverList.Values() {
		vmInfo := mappingServerInfo(server)
		vmList = append(vmList, &vmInfo)
	}

	return vmList
}

func (vmHandler *GCPVMHandler) GetVM(vmID string) irs.VMInfo {
	vmIdArr := strings.Split(vmID, ":")
	vm, err := vmHandler.Client.Get(vmHandler.Ctx, vmIdArr[0], vmIdArr[1], compute.InstanceView)
	if err != nil {
		panic(err)
	}

	vmInfo := mappingServerInfo(vm)
	return vmInfo
}

func getVmStatus(instanceView compute.VirtualMachineInstanceView) string {
	var powerState, provisioningState string

	for _, stat := range *instanceView.Statuses {
		statArr := strings.Split(*stat.Code, "/")

		if statArr[0] == "PowerState" {
			powerState = statArr[1]
		} else if statArr[0] == "ProvisioningState" {
			provisioningState = statArr[1]
		}
	}

	// Set VM Status Info
	var vmState string
	if powerState != "" && provisioningState != "" {
		vmState = powerState + "(" + provisioningState + ")"
	} else if powerState != "" && provisioningState == "" {
		vmState = powerState
	} else if powerState == "" && provisioningState != "" {
		vmState = provisioningState
	} else {
		vmState = "-"
	}
	return vmState
}

func mappingServerInfo(server compute.VirtualMachine) irs.VMInfo {

	// Get Default VM Info
	vmInfo := irs.VMInfo{
		Name: *server.Name,
		Id:   *server.ID,
		Region: irs.RegionInfo{
			Region: *server.Location,
		},
		SpecID: string(server.VirtualMachineProperties.HardwareProfile.VMSize),
	}

	// Set VM Zone
	if server.Zones != nil {
		vmInfo.Region.Zone = (*server.Zones)[0]
	}

	// Set VM Image Info
	imageRef := server.VirtualMachineProperties.StorageProfile.ImageReference
	imageId := *imageRef.Publisher + ":" + *imageRef.Offer + ":" + *imageRef.Sku + ":" + *imageRef.Version
	vmInfo.ImageID = imageId

	// Set VNic Info
	niList := *server.NetworkProfile.NetworkInterfaces
	for _, ni := range niList {
		if ni.NetworkInterfaceReferenceProperties != nil {
			vmInfo.VNIC = *ni.ID
		}
	}

	// Set GuestUser Id/Pwd
	if server.VirtualMachineProperties.OsProfile.AdminUsername != nil {
		vmInfo.GuestUserID = *server.VirtualMachineProperties.OsProfile.AdminUsername
	}
	if server.VirtualMachineProperties.OsProfile.AdminPassword != nil {
		vmInfo.GuestUserID = *server.VirtualMachineProperties.OsProfile.AdminPassword
	}

	// Set BootDisk
	if server.VirtualMachineProperties.StorageProfile.OsDisk.Name != nil {
		vmInfo.GuestBootDisk = *server.VirtualMachineProperties.StorageProfile.OsDisk.Name
	}

	return vmInfo
}

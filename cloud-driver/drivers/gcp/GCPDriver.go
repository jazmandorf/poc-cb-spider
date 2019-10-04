// Proof of Concepts of CB-Spider.
// The CB-Spider is a sub-Framework of the Cloud-Barista Multi-Cloud Project.
// The CB-Spider Mission is to connect all the clouds with a single interface.
//
//      * Cloud-Barista: https://github.com/cloud-barista
//
// This is a Cloud Driver Example for PoC Test.
//
// by hyokyung.kim@innogrid.co.kr, 2019.07.

package gcp

import (
	"context"
	"log"
	"time"

	gcpcon "github.com/cloud-barista/poc-cb-spider/cloud-driver/drivers/gcp/connect"
	idrv "github.com/cloud-barista/poc-cb-spider/cloud-driver/interfaces"
	icon "github.com/cloud-barista/poc-cb-spider/cloud-driver/interfaces/connect"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	compute "google.golang.org/api/compute/v1"
)

type GCPDriver struct {
}

func (GCPDriver) GetDriverVersion() string {
	return "GCP DRIVER Version 1.0"
}

func (GCPDriver) GetDriverCapability() idrv.DriverCapabilityInfo {
	var drvCapabilityInfo idrv.DriverCapabilityInfo

	drvCapabilityInfo.ImageHandler = false
	drvCapabilityInfo.VNetworkHandler = false
	drvCapabilityInfo.SecurityHandler = false
	drvCapabilityInfo.KeyPairHandler = false
	drvCapabilityInfo.VNicHandler = false
	drvCapabilityInfo.PublicIPHandler = false
	drvCapabilityInfo.VMHandler = true

	return drvCapabilityInfo
}

func (driver *GCPDriver) ConnectCloud(connectionInfo idrv.ConnectionInfo) (icon.CloudConnection, error) {
	// 1. get info of credential and region for Test A Cloud from connectionInfo.
	// 2. create a client object(or service  object) of Test A Cloud with credential info.
	// 3. create CloudConnection Instance of "connect/TDA_CloudConnection".
	// 4. return CloudConnection Interface of TDA_CloudConnection.

	Ctx, VMClient, err := getVMClient(connectionInfo.CredentialInfo)
	if err != nil {
		return nil, err
	}

	iConn := gcpcon.GCPCloudConnection{
		Region:              connectionInfo.RegionInfo,
		Credential:          connectionInfo.CredentialInfo,
		Ctx:                 Ctx,
		VMClient:            VMClient,
		ImageClient:         VMClient,
		PublicIPClient:      VMClient,
		SecurityGroupClient: VMClient,
		VNetClient:          VMClient,
		VNicClient:          VMClient,
		SubnetClient:        VMClient,
	}
	return &iConn, nil
}

func getVMClient(credential idrv.CredentialInfo) (context.Context, *compute.Service, error) {

	// GCP 는  ClientSecret에
	// JSON byte읽은 걸 string으로 변환해서 credential.ClientSecret에 넣을꺼임
	data := []byte(credential.ClientSecret)
	authURL := "https://www.googleapis.com/auth/compute"
	conf, err := google.JWTConfigFromJSON(data, authURL)

	if err != nil {
		log.Fatal(err)
		return nil, nil, err
	}
	client := conf.Client(oauth2.NoContext)

	vmClient, err := compute.New(client)

	ctx, _ := context.WithTimeout(context.Background(), 600*time.Second)

	return ctx, &vmClient, nil
}

var TestDriver GCPDriver

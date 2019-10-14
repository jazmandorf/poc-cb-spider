package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	idrv "../../../interfaces"
	irs "../../../interfaces/resources"
	"github.com/davecgh/go-spew/spew"
	compute "google.golang.org/api/compute/v1"
)

type GCPPublicIPHandler struct {
	Region     idrv.RegionInfo
	Ctx        context.Context
	Client     *compute.Service
	Credential idrv.CredentialInfo
}

// @TODO: PublicIP 리소스 프로퍼티 정의 필요
type PublicIPInfo struct {
	Id                string
	Name              string
	Region            string
	CreationTimestamp string
	IpAddress         string
	NetworkTier       string // PREMIUM, STANDARD
	AddressType       string // External, INTERNAL, UNSPECIFIED_TYPE
	Status            string // IN_USE, RESERVED, RESERVING
}

func (publicIP *PublicIPInfo) setter(address network.PublicIPAddress) *PublicIPInfo {
	publicIP.Id = *address.ID
	publicIP.Name = *address.Name
	publicIP.Location = *address.Location
	publicIP.PublicIPAddressSku = fmt.Sprint(address.Sku.Name)
	publicIP.PublicIPAddressVersion = fmt.Sprint(address.PublicIPAddressVersion)
	publicIP.PublicIPAllocationMethod = fmt.Sprint(address.PublicIPAllocationMethod)
	publicIP.IPAddress = *address.IPAddress
	publicIP.IdleTimeoutInMinutes = *address.IdleTimeoutInMinutes

	return publicIP
}

func (publicIpHandler *GCPPublicIPHandler) CreatePublicIP(publicIPReqInfo irs.PublicIPReqInfo) (irs.PublicIPInfo, error) {

	// @TODO: PublicIP 생성 요청 파라미터 정의 필요
	type PublicIPReqInfo struct {
		PublicIPAddressSkuName       string
		PublicIPAddressVersion       string
		PublicIPAllocationMethod     string
		PublicIPIdleTimeoutInMinutes int32
	}
	reqInfo := PublicIPReqInfo{
		PublicIPAddressSkuName:       "Basic",
		PublicIPAddressVersion:       "IPv4",
		PublicIPAllocationMethod:     "Static",
		PublicIPIdleTimeoutInMinutes: 4,
	}

	publicIPArr := strings.Split(publicIPReqInfo.Id, ":")

	// Check PublicIP Exists
	publicIP, err := publicIpHandler.Client.Get(publicIpHandler.Ctx, publicIPArr[0], publicIPArr[1], "")
	if publicIP.ID != nil {
		errMsg := fmt.Sprintf("Public IP with name %s already exist", publicIPArr[1])
		createErr := errors.New(errMsg)
		return irs.PublicIPInfo{}, createErr
	}

	createOpts := network.PublicIPAddress{
		Sku: &network.PublicIPAddressSku{
			Name: network.PublicIPAddressSkuName(reqInfo.PublicIPAddressSkuName),
		},
		PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
			PublicIPAddressVersion:   network.IPVersion(reqInfo.PublicIPAddressVersion),
			PublicIPAllocationMethod: network.IPAllocationMethod(reqInfo.PublicIPAllocationMethod),
			IdleTimeoutInMinutes:     &reqInfo.PublicIPIdleTimeoutInMinutes,
		},
		Location: &publicIpHandler.Region.Region,
	}

	future, err := publicIpHandler.Client.CreateOrUpdate(publicIpHandler.Ctx, publicIPArr[0], publicIPArr[1], createOpts)
	if err != nil {
		return irs.PublicIPInfo{}, err
	}
	err = future.WaitForCompletionRef(publicIpHandler.Ctx, publicIpHandler.Client.Client)
	if err != nil {
		return irs.PublicIPInfo{}, err
	}

	// @TODO: 생성된 PublicIP 정보 리턴
	publicIPInfo, err := publicIpHandler.GetPublicIP(publicIPReqInfo.Id)
	if err != nil {
		return irs.PublicIPInfo{}, err
	}
	return publicIPInfo, nil
}

func (publicIpHandler *GCPPublicIPHandler) ListPublicIP() ([]*irs.PublicIPInfo, error) {
	//result, err := publicIpHandler.Client.ListAll(publicIpHandler.Ctx)
	result, err := publicIpHandler.Client.List(publicIpHandler.Ctx, publicIpHandler.Region.ResourceGroup)
	if err != nil {
		return nil, err
	}

	var publicIPList []*PublicIPInfo
	for _, publicIP := range result.Values() {
		publicIPInfo := new(PublicIPInfo).setter(publicIP)
		publicIPList = append(publicIPList, publicIPInfo)
	}

	spew.Dump(publicIPList)
	return nil, nil
}

func (publicIpHandler *GCPPublicIPHandler) GetPublicIP(publicIPID string) (irs.PublicIPInfo, error) {
	publicIPArr := strings.Split(publicIPID, ":")
	publicIP, err := publicIpHandler.Client.Get(publicIpHandler.Ctx, publicIPArr[0], publicIPArr[1], "")
	if err != nil {
		return irs.PublicIPInfo{}, err
	}

	publicIPInfo := new(PublicIPInfo).setter(publicIP)

	spew.Dump(publicIPInfo)
	return irs.PublicIPInfo{}, nil
}

func (publicIpHandler *GCPPublicIPHandler) DeletePublicIP(publicIPID string) (bool, error) {
	publicIPArr := strings.Split(publicIPID, ":")
	future, err := publicIpHandler.Client.Delete(publicIpHandler.Ctx, publicIPArr[0], publicIPArr[1])
	if err != nil {
		return false, err
	}
	err = future.WaitForCompletionRef(publicIpHandler.Ctx, publicIpHandler.Client.Client)
	if err != nil {
		return false, err
	}
	return true, nil
}

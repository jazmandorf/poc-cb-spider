package resources

import (
	"context"
	"fmt"

	idrv "../../../interfaces"
	irs "../../../interfaces/resources"
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
	Region            string // GCP
	CreationTimestamp string // GCP
	IpAddress         string // GCP
	NetworkTier       string // GCP : PREMIUM, STANDARD
	AddressType       string // GCP : External, INTERNAL, UNSPECIFIED_TYPE
	Status            string // GCP : IN_USE, RESERVED, RESERVING
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

	return publicIPInfo, nil
}

func (publicIpHandler *GCPPublicIPHandler) ListPublicIP() ([]*irs.PublicIPInfo, error) {

	return nil, nil
}

func (publicIpHandler *GCPPublicIPHandler) GetPublicIP(publicIPID string) (irs.PublicIPInfo, error) {

	return irs.PublicIPInfo{}, nil
}

func (publicIpHandler *GCPPublicIPHandler) DeletePublicIP(publicIPID string) (bool, error) {

	return true, nil
}

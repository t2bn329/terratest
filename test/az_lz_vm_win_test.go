package test

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"testing"

	//"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestWindowsVM(t *testing.T) {
	t.Parallel()

	foo := map[string]interface{}{
		"nullable_string":    nil,
		"nonnullable_string": "foo",
	}
	options := &terraform.Options{
		TerraformDir: "./fixtures/terraform-null",
		Vars:         map[string]interface{}{"foo": foo},
	}
	terraform.InitAndApply(t, options)

	yamlConfig, _ := ioutil.ReadFile("az-lz-test_data/vm-windows.yaml")
	expectedVMConfig := AzVM{}
	_ = yaml.Unmarshal([]byte(yamlConfig), &expectedVMConfig)
	fmt.Println("Searching for VM: ", expectedVMConfig.RsName)

	vmExists := false
	subscriptionID, resGroup, err := LocateResource(expectedVMConfig.RsType, expectedVMConfig.RsName)
	if err == nil {
		vmExists = true
	}

	// website::tag::4:: Verify key vault exists and ensure it matches the output.
	assert.True(t, vmExists, "VM does not exist")

	myVM, _ := azure.GetVirtualMachineE(expectedVMConfig.RsName, *resGroup.Name, subscriptionID)

	assert.Equal(t, expectedVMConfig.RsName, *myVM.Name)
	fmt.Println("vmSize: ", myVM.HardwareProfile.VMSize)

	// 1. Check the VM Size directly. This strategy gets one specific property of the VM per method.
	actualVMSize := azure.GetSizeOfVirtualMachine(t, expectedVMConfig.RsName, *resGroup.Name, subscriptionID)
	fmt.Println("method1 - VM Size: ", actualVMSize)
	assert.Contains(t, actualVMSize, expectedVMConfig.Parameters.VMSize)

	// 2. Check the VM size by reference. This strategy is beneficial when checking multiple properties
	// by using one VM reference. Optional parameters have to be checked first to avoid nil panics.
	actualVMSize = myVM.HardwareProfile.VMSize
	fmt.Println("method2 - VM Size: ", actualVMSize)
	assert.Contains(t, actualVMSize, expectedVMConfig.Parameters.VMSize)

	// 3. Check the VM size by instance. This strategy is beneficial when checking multiple properties
	// by using one VM instance and making calls against it with the added benefit of property check abstraction.
	vmInstance := azure.Instance{VirtualMachine: myVM}
	actualVMSize = vmInstance.GetVirtualMachineInstanceSize()
	fmt.Println("method3 - VM Size: ", actualVMSize)
	assert.Contains(t, actualVMSize, expectedVMConfig.Parameters.VMSize)

	//osType := myVM.StorageProfile.OsDisk.OsType
	assert.Contains(t, myVM.StorageProfile.OsDisk.OsType, expectedVMConfig.Parameters.OsType)

	assert.Equal(t, expectedVMConfig.Parameters.AdminUsername, *myVM.OsProfile.AdminUsername)

	testWinMultipleVMs(t, resGroup.Name, &subscriptionID, expectedVMConfig)

	testWinVMImageAndDisk(t, &expectedVMConfig.RsName, resGroup.Name, &subscriptionID, expectedVMConfig)

	testWinNetworkOfVM(t, resGroup.Name, &subscriptionID, expectedVMConfig)

	terraform.Destroy(t, options)
}

func testWinVMImageAndDisk(t *testing.T, vmName *string, resGroupName *string, subscriptionID *string, expectedVMConfig AzVM) {
	vmImage := azure.GetVirtualMachineImage(t, *vmName, *resGroupName, *subscriptionID)
	assert.Equal(t, expectedVMConfig.Parameters.ImageOffer, vmImage.Offer)
	assert.Equal(t, expectedVMConfig.Parameters.ImagePublisher, vmImage.Publisher)
	assert.Equal(t, expectedVMConfig.Parameters.ImageSku, vmImage.SKU)
	// expected VM version was specified as 'latest', the actual VM version is the actual release value of the latest
	//assert.Equal(t, expectedVMConfig.Parameters.ImageVersion, vmImage.Version)
	vmOSDisk := azure.GetVirtualMachineManagedDisks(t, expectedVMConfig.RsName, *resGroupName, *subscriptionID)
	fmt.Println("OSDisks: ", vmOSDisk)
	disk, _ := azure.GetDiskE(vmOSDisk[0], *resGroupName, *subscriptionID)
	actualVMTags := azure.GetVirtualMachineTags(t, *vmName, *resGroupName, *subscriptionID)
	assert.Contains(t, expectedVMConfig.Parameters.OsDiskType, *disk.Sku.Tier)
	// Check the assigned Tags of the VM, assert empty if no tags.
	assert.Equal(t, expectedVMConfig.Parameters.ResourceTags.App, actualVMTags["App"])
	assert.Equal(t, expectedVMConfig.Parameters.ResourceTags.Bu, actualVMTags["BU"])
	assert.Equal(t, expectedVMConfig.Parameters.ResourceTags.PatchGroupID, actualVMTags["patchGroupId"])
	assert.Equal(t, expectedVMConfig.Parameters.ResourceTags.Project, actualVMTags["Project"])
	//assert.Equal(t, expectedVMConfig.Parameters.ResourceTags.RuntimeEnv, actualVMTags["RunTimeEnv"])
}

func testWinMultipleVMs(t *testing.T, resGroupName *string, subscriptionID *string, expectedVMConfig AzVM) {
	// Check against all VM names in a Resource Group.
	vmList := azure.ListVirtualMachinesForResourceGroup(t, *resGroupName, *subscriptionID)
	expectedVMCount := 1
	assert.GreaterOrEqual(t, len(vmList), expectedVMCount)
	assert.Contains(t, vmList, expectedVMConfig.RsName)

	// Check Availability Set for multiple VMs.
	//actualVMsInAvs := azure.GetAvailabilitySetVMNamesInCaps(t, "", resGroupName, subscriptionID)
	//assert.Contains(t, actualVMsInAvs, strings.ToUpper(expectedVMName))
	//fmt.Println("VMs: ", actualVMsInAvs)

	// Get all VMs in a Resource Group, including their properties, therefore avoiding
	// multiple SDK calls. The penalty for this approach is introducing direct references
	// which need to be checked for nil for optional configurations.
	vmsByRef := azure.GetVirtualMachinesForResourceGroup(t, *resGroupName, *subscriptionID)
	thisVM := vmsByRef[expectedVMConfig.RsName]
	assert.Contains(t, thisVM.HardwareProfile.VMSize, expectedVMConfig.Parameters.VMSize)

	// Check for the VM negative test.
	fakeVM := fmt.Sprintf("vm-%s", random.UniqueId())
	assert.Nil(t, vmsByRef[fakeVM].VMID)
}

// These tests check the underlying Virtual Network, Network Interface and associated Public IP Address.
// The following Terratest Azure modules are utilized in addition to the compute module:
// - networkinterface
// - publicaddress
// - virtualnetwork
// See the terraform_azure_network_example_test.go for other related tests.
func testWinNetworkOfVM(t *testing.T, resGroupName *string, subscrID *string, expectedVMConfig AzVM) {
	subnetRegEx := regexp.MustCompile(SUBNET_REGEX)

	match1 := subnetRegEx.FindStringSubmatch(expectedVMConfig.Parameters.SubnetID)
	subscriptionID := match1[1]
	fmt.Println("subscription id: ", subscriptionID)
	resourceGroupName := match1[2]
	fmt.Println("resource group: ", resourceGroupName)
	expectedVNetName := match1[3]
	fmt.Println("vnet accounts: ", expectedVNetName)
	expectedSubnetName := match1[4]
	fmt.Println("subnet: ", expectedSubnetName)

	// VirtualNetwork and Subnet tests
	// Check the Subnet exists in the Virtual Network.
	actualVnetSubnets := azure.GetVirtualNetworkSubnets(t, expectedVNetName, resourceGroupName, subscriptionID)
	assert.NotNil(t, actualVnetSubnets[expectedVNetName])

	// Check the Private IP is in the Subnet Range.
	actualVMNicIPInSubnet := azure.CheckSubnetContainsIP(t, expectedVMConfig.Parameters.IPAddresses[0], expectedSubnetName, expectedVNetName, resourceGroupName, subscriptionID)
	assert.True(t, actualVMNicIPInSubnet)

	// Network Interface Card tests
	// Check the VM Network Interface exists in the list of all VM Network Interfaces.
	actualNics := azure.GetVirtualMachineNics(t, expectedVMConfig.RsName, *resGroupName, *subscrID)
	//assert.Contains(t, actualNics, expectedNicName)
	fmt.Println("actualNICs: ", actualNics)

	// Check the Network Interface count of the VM.
	expectedNICCount := 1
	assert.Equal(t, expectedNICCount, len(actualNics))

	// Check for the Private IP in the NICs IP list.
	actualPrivateIPAddress := azure.GetNetworkInterfacePrivateIPs(t, actualNics[0], *resGroupName, *subscrID)
	fmt.Println("privateIPs: ", actualPrivateIPAddress)
	//assert.Contains(t, actualPrivateIPAddress, expectedPrivateIPAddress)

	// There are other tests that we can do with the VM resource regarding its network
	//
	//azure.GetLoadBalancerFrontendIPConfig()
	// Public IP Address test
	// Check for the Public IP for the NIC. No expected value since it is assigned runtime.
	//actualPublicIP := azure.GetIPOfPublicIPAddressByName(t, expectedPublicAddressName, resourceGroupName, subscriptionID)
	//assert.NotNil(t, actualPublicIP)
}

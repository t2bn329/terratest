package test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestNetworking(t *testing.T) {
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

	yamlConfig, _ := ioutil.ReadFile("az-lz-test_data/networking.yaml")
	expectedNetworking := AzNetworking{}
	_ = yaml.Unmarshal([]byte(yamlConfig), &expectedNetworking)
	expectedVNetName := expectedNetworking.Parameters.VirtualNetworkName
	expectedSubnetName := expectedNetworking.Parameters.SubnetName
	fmt.Println("expected subnet name: ", expectedSubnetName)

	vnetExists := false
	subnetExists := false

	fmt.Println("Looking for VNet: ", expectedVNetName)
	subscriptionID, resGroup, err := LocateResource(expectedNetworking.RsType, expectedNetworking.RsName)
	if err == nil {
		fmt.Printf("Found rsType: %s,  rsName: %s,  in resourceGroup: %s, subscription: %s\n", expectedNetworking.RsType, expectedNetworking.RsName, *resGroup.Name, subscriptionID)
		vnetExists = true
		subnetExists, _ = azure.SubnetExistsE(expectedSubnetName, expectedVNetName, *resGroup.Name, subscriptionID)
	}

	// website::tag::4:: Verify subnet properties and ensure it matches the output.
	assert.True(t, vnetExists, "VNet does not exist")

	// website::tag::4:: Verify subnet properties and ensure it matches the output.
	assert.True(t, subnetExists, "subnet does not exist")

	subnet, _ := azure.GetSubnetE(expectedSubnetName, expectedVNetName, *resGroup.Name, subscriptionID)

	//fmt.Println("subnet prefix: ", subnet.AddressPrefix)
	fmt.Println("subnet prefixes: ", subnet.AddressPrefixes)
	fmt.Println("subnet nsg name: ", subnet.NetworkSecurityGroup.Name)
	fmt.Println("subnet prefix: ", *subnet.AddressPrefix)
	fmt.Println("subnet name: ", *subnet.Name)

	assert.Equal(t, expectedNetworking.Parameters.SubnetAddressPrefix, *subnet.AddressPrefix)
	assert.Equal(t, expectedSubnetName, *subnet.Name)

	terraform.Destroy(t, options)
}

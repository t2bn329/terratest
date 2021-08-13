package test

import (
	"fmt"
	"io/ioutil"
	"testing"

	//"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestStorageAccount(t *testing.T) {
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

	yamlConfig, _ := ioutil.ReadFile("az-lz-test_data/storage-accounts.yaml")
	expectedStorage := AzStorage{}
	_ = yaml.Unmarshal([]byte(yamlConfig), &expectedStorage)
	storageContainers := expectedStorage.Parameters.Containers
	fmt.Println("Storage container: ", storageContainers)

	storageAccountExists := false
	subscriptionID, resGroup, err := LocateResource(expectedStorage.RsType, expectedStorage.RsName)
	if err == nil {
		fmt.Printf("Found rsType: %s,  rsName: %s,  in resourceGroup: %s, subscription: %s\n", expectedStorage.RsType, expectedStorage.RsName, *resGroup.Name, subscriptionID)
		storageAccountExists = true
	}

	// website::tag::4:: Verify storage account properties and ensure it matches the output.
	assert.True(t, storageAccountExists, "storage account does not exist")

	for _, storageContainer := range expectedStorage.Parameters.Containers {
		containerExists := azure.StorageBlobContainerExists(t, storageContainer, expectedStorage.RsName, *resGroup.Name, subscriptionID)
		assert.True(t, containerExists, "storage container does not exist")
		publicAccess := azure.GetStorageBlobContainerPublicAccess(t, storageContainer, expectedStorage.RsName, *resGroup.Name, subscriptionID)
		assert.False(t, publicAccess, "storage container has public access")
	}

	accountKind := azure.GetStorageAccountKind(t, expectedStorage.RsName, *resGroup.Name, subscriptionID)
	fmt.Println("account kind: ", accountKind)
	assert.Equal(t, "StorageV2", accountKind, "storage account kind mismatch")

	skuTier := azure.GetStorageAccountSkuTier(t, expectedStorage.RsName, *resGroup.Name, subscriptionID)
	fmt.Println("sku tier: ", skuTier)
	assert.Equal(t, "Standard", skuTier, "sku tier mismatch")

	actualDNSString := azure.GetStorageDNSString(t, expectedStorage.RsName, *resGroup.Name, subscriptionID)
	storageSuffix, _ := azure.GetStorageURISuffixE()
	expectedDNS := fmt.Sprintf("https://%s.blob.%s/", expectedStorage.RsName, storageSuffix)
	assert.Equal(t, expectedDNS, actualDNSString, "Storage DNS string mismatch")

	terraform.Destroy(t, options)
}

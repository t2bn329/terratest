package test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2016-10-01/keyvault"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestKeyVault(t *testing.T) {
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

	yamlSource, _ := ioutil.ReadFile("az-lz-test_data/keyvault.yaml")
	expectedKeyVault := AzKeyVault{}
	_ = yaml.Unmarshal([]byte(yamlSource), &expectedKeyVault)
	fmt.Println("KeyVault name: ", expectedKeyVault.RsName)
	fmt.Println("KeyVault description: ", expectedKeyVault.RsDescription)
	fmt.Println("KeyVault SKU: ", expectedKeyVault.Parameters.KeyVaultSku)
	fmt.Println("Keyvault wkspcURL: ", expectedKeyVault.Parameters.WorkspaceURL)

	var allSubscriptions []string
	var allResourceGroups map[string][]resources.Group
	var allKeyVaults map[string][]keyvault.Resource
	allSubscriptions, allResourceGroups, allKeyVaults, _ = RetrieveSubscriptionsRGsAndKVs()

	kvExists := false
	kvSubscription := ""
	fmt.Println("Looking for keyvault: ", expectedKeyVault.RsName)
	for _, subscription := range allSubscriptions {
		sKeyVaults := allKeyVaults[subscription]
		for _, skv := range sKeyVaults {
			fmt.Printf("...  keyvault: %s in subscription: %s\n", *skv.Name, subscription)
			if strings.Compare(expectedKeyVault.RsName, *skv.Name) == 0 {
				fmt.Printf("Found keyvault: %s in subscription: %s\n", *skv.Name, subscription)
				kvExists = true
				kvSubscription = subscription
				break
			}
		}
	}

	// website::tag::4:: Verify key vault exists and ensure it matches the output.
	assert.True(t, kvExists, "key vault does not exist")

	var resGroups = allResourceGroups[kvSubscription]
	for _, resGroup := range resGroups {
		// Retrieve the key vault from Azure
		kv, err := azure.GetKeyVaultE(t, *resGroup.Name, expectedKeyVault.RsName, kvSubscription)
		if err == nil {
			fmt.Printf("Successfully retrieved key vault: %s in resource group: %s in subscription: %s\n", expectedKeyVault.RsName, *resGroup.Name, kvSubscription)
			assert.Equal(t, expectedKeyVault.RsName, *kv.Name, "Mismatched key vault name")
			// Verify that the SKU values are the same
			assert.Equal(t, expectedKeyVault.Parameters.KeyVaultSku, (string)(kv.Properties.Sku.Name))
			accessPolicies := *kv.Properties.AccessPolicies
			for _, accPolicy := range accessPolicies {
				if strings.Compare(expectedKeyVault.Parameters.AccessPolicies[0].ObjectID, *accPolicy.ObjectID) == 0 {
					fmt.Println("accpol obj ID: ", *accPolicy.ObjectID)
					permKeys := *(accPolicy.Permissions.Keys)
					// Need to convert into an array of strings before comparing results...
					var permissionKeys = make([]string, len(permKeys))
					for j, _ := range permissionKeys {
						permissionKeys[j] = (string)(permKeys[j])
					}
					fmt.Println("keys: ", permissionKeys)
					// Verify that the policies permission keys are the same
					assert.Equal(t, expectedKeyVault.Parameters.AccessPoliciesKeys, permissionKeys)

					permCerts := *(accPolicy.Permissions.Certificates)
					// Need to convert into an array of strings before comparing results...
					var permissionCertificates = make([]string, len(permCerts))
					for j, _ := range permCerts {
						permissionCertificates[j] = (string)(permCerts[j])
					}
					fmt.Println("certs: ", permissionCertificates)
					// Verify that the policies permission certificates are the same
					assert.Equal(t, expectedKeyVault.Parameters.AccessPoliciesCertificates, permissionCertificates)

					permSecrets := *(accPolicy.Permissions.Secrets)
					// Need to convert into an array of strings before comparing results...
					var permissionSecrets = make([]string, len(permSecrets))
					for j, _ := range permSecrets {
						permissionSecrets[j] = (string)(permSecrets[j])
					}
					fmt.Println("secrets: ", permissionSecrets)
					// Verify that the policies permission secrets are the same
					assert.Equal(t, expectedKeyVault.Parameters.AccessPoliciesSecrets, permissionSecrets)
				}
			}
			// Verify if the secret exists
			// We currently get a "403" error since we do not have access
			//secretExists := azure.KeyVaultSecretExists(t, expectedKeyVault.RsName, expectedKeyVault.Parameters.SecretNameLawKey)
			//assert.True(t, secretExists, "kv-secret does not exist")

			fmt.Println("URL: ", kv.Request.URL)
			break
		}
	}

}

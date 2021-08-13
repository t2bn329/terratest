package test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"gopkg.in/yaml.v2"
)

func TestNetworkSecurityGroup(t *testing.T) {
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

	yamlConfig, _ := ioutil.ReadFile("az-lz-test_data/network-security-group.yaml")
	expectedNSG := AzNetworkSecurityGroup{}
	_ = yaml.Unmarshal([]byte(yamlConfig), &expectedNSG)

	var allSubscriptions []string
	var allResourceGroups map[string][]resources.Group
	allSubscriptions, allResourceGroups, _, _ = RetrieveSubscriptionsRGsAndKVs()
	for _, subscription := range allSubscriptions {
		resGroups := allResourceGroups[subscription]
		for _, resGroup := range resGroups {
			fmt.Printf("Searching for rules ... %s in resourceGroup: %s   in subscription: %s\n", expectedNSG.RsName, *resGroup.Name, subscription)

			// A default NSG has 6 rules, and we have two custom rules for a total of 8
			rules, err := azure.GetAllNSGRulesE(*resGroup.Name, expectedNSG.RsName, subscription)

			if err == nil {
				for _, summarizedRule := range rules.SummarizedRules {
					fmt.Println("Summarized rule: ", summarizedRule.Name)
				}
				break
			} else if strings.Contains(err.Error(), "ResourceNotFound") {
				// Ignore for now
			} else {
				fmt.Println("Unexpected error: ", err.Error())
			}

		}
	}

	terraform.Destroy(t, options)
}

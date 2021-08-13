package test

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2016-10-01/keyvault"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-06-01/subscriptions"
	"github.com/gruntwork-io/terratest/modules/azure"
)

const STORAGE_ACCOUNTS_REGEX = `^/subscriptions/([0-9a-z\-]+)/resourceGroups/([0-9a-z\-]+)/[0-9a-zA-Z-]+/[0-9a-zA-Z-\.]+/storageAccounts/([0-9a-z]+)`
const KEY_VAULT_REGEX = `^/subscriptions/([0-9a-z\-]+)/resourceGroups/([0-9a-z\-]+)/[0-9a-zA-Z-]+/[0-9a-zA-Z-\.]+/vaults/([0-9a-z\-]+)`
const LAWS_REGEX = `^/subscriptions/([0-9a-z\-]+)/resourceGroups/([0-9a-z\-]+)/[0-9a-zA-Z-]+/[0-9a-zA-Z-\.]+/workspaces/([0-9a-z]+)`
const SUBNET_REGEX = `^/subscriptions/([0-9a-z\-]+)/resourceGroups/([0-9a-z\-]+)/[0-9a-zA-Z-]+/[0-9a-zA-Z-\.]+/virtualNetworks/([0-9a-zA-Z\-]+)/subnets/([0-9a-zA-Z\-]+)`
const NSG_REGEX = `^/subscriptions/([0-9a-z\-]+)/resourceGroups/([0-9a-z\-]+)/[0-9a-zA-Z-]+/[0-9a-zA-Z-\.]+/networkSecurityGroups/([0-9a-z]+)`

type AzAdminPassword struct {
	Reference struct {
		KeyVault struct {
			ID string `yaml:"id"`
		} `yaml:"keyVault"`
		SecretName string `yaml:"secretName"`
	} `yaml:"reference"`
}

type AzResourceTags struct {
	App          string `yaml:"App"`
	Bu           string `yaml:"BU"`
	Project      string `yaml:"Project"`
	RuntimeEnv   string `yaml:"RuntimeEnv"`
	PatchGroupID string `yaml:"patchGroupId"`
}

type AzWorkspaceKey struct {
	Reference struct {
		KeyVault struct {
			ID string `yaml:"id"`
		} `yaml:"keyVault"`
		SecretName string `yaml:"secretName"`
	} `yaml:"reference"`
}

type AzResourceBasic struct {
	RsType                 string `yaml:"rsType"`
	RsName                 string `yaml:"rsName"`
	RsDescription          string `yaml:"rsDescription"`
	ApplicationProjectName string `yaml:"applicationProjectName"`
}

type AzKeyVault struct {
	RsType                 string `yaml:"rsType"`
	RsName                 string `yaml:"rsName"`
	RsDescription          string `yaml:"rsDescription"`
	ApplicationProjectName string `yaml:"applicationProjectName"`
	Parameters             struct {
		AccessPolicies []struct {
			ApplicationID string `yaml:"applicationId"`
			Name          string `yaml:"name"`
			ObjectID      string `yaml:"objectId"`
			TenantID      string `yaml:"tenantId"`
		} `yaml:"accessPolicies"`
		AccessPoliciesCertificates []string       `yaml:"accessPoliciesCertificates"`
		AccessPoliciesKeys         []string       `yaml:"accessPoliciesKeys"`
		AccessPoliciesSecrets      []string       `yaml:"accessPoliciesSecrets"`
		AccessPoliciesStorage      []string       `yaml:"accessPoliciesStorage"`
		AzureTenantID              string         `yaml:"azureTenantId"`
		KeyVaultName               string         `yaml:"keyVaultName"`
		KeyVaultSku                string         `yaml:"keyVaultSku"`
		ResourceTags               AzResourceTags `yaml:"resourceTags"`
		SecretNameLawKey           string         `yaml:"secretNameLawKey"`
		WorkspaceURL               string         `yaml:"workspaceUrl"`
	} `yaml:"parameters"`
}

type AzNetworking struct {
	RsType                 string `yaml:"rsType"`
	RsName                 string `yaml:"rsName"`
	RsDescription          string `yaml:"rsDescription"`
	ApplicationProjectName string `yaml:"applicationProjectName"`
	Parameters             struct {
		BastionSubnetAddressRange       string         `yaml:"bastionSubnetAddressRange"`
		NsgName                         string         `yaml:"nsgName"`
		ResourceTags                    AzResourceTags `yaml:"resourceTags"`
		RouteTableID                    string         `yaml:"routeTableId"`
		SubnetAddressPrefix             string         `yaml:"subnetAddressPrefix"`
		SubnetName                      string         `yaml:"subnetName"`
		VirtualNetworkName              string         `yaml:"virtualNetworkName"`
		VirtualNetworkResourceGroupName string         `yaml:"virtualNetworkResourceGroupName"`
	} `yaml:"parameters"`
}

type AzRecoveryServiceVault struct {
	RsType                 string `yaml:"rsType"`
	RsName                 string `yaml:"rsName"`
	RsDescription          string `yaml:"rsDescription"`
	ApplicationProjectName string `yaml:"applicationProjectName"`
	Parameters             struct {
		DailyRetentionDurationCount    int64          `yaml:"virtualNetworkResourceGroupName"`
		DaysOfTheWeek                  []string       `yaml:"daysOfTheWeek"`
		DiagnosticsLogStorageAccountID string         `yaml:"diagnosticsLogStorageAccountId"`
		EnvironmentName                string         `yaml:"EnvironmentName"`
		LocationAbbreviation           string         `yaml:"locationAbbreviation"`
		MonthlyRetentionDurationCount  int64          `yaml:"monthlyRetentionDurationCount"`
		PolicyName                     string         `yaml:"policyName"`
		RecoveryVaultName              string         `yaml:"recoveryVaultName"`
		ResourceTags                   AzResourceTags `yaml:"resourceTags"`
		VaultStorageType               string         `yaml:"vaultStorageType"`
		WeeklyRetentionDurationCount   int64          `yaml:"weeklyRetentionDurationCount"`
		WorkspaceID                    string         `yaml:"workspaceId"`
	} `yaml:"parameters"`
}

type AzStorage struct {
	RsType                 string `yaml:"rsType"`
	RsName                 string `yaml:"rsName"`
	RsDescription          string `yaml:"rsDescription"`
	ApplicationProjectName string `yaml:"applicationProjectName"`
	Parameters             struct {
		Containers           []string       `yaml:"containers"`
		EnvironmentName      string         `yaml:"environmentName"`
		LocationAbbreviation string         `yaml:"locationAbbreviation"`
		Tags                 AzResourceTags `yaml:"resourceTags"`
	} `yaml:"parameters"`
}

type AzNetworkSecurityGroup struct {
	RsType                 string `yaml:"rsType"`
	RsName                 string `yaml:"rsName"`
	RsDescription          string `yaml:"rsDescription"`
	ApplicationProjectName string `yaml:"applicationProjectName"`
	Parameters             struct {
		Containers           []string       `yaml:"containers"`
		EnvironmentName      string         `yaml:"environmentName"`
		LocationAbbreviation string         `yaml:"locationAbbreviation"`
		Tags                 AzResourceTags `yaml:"resourceTags"`
	} `yaml:"parameters"`
}

type AzVM struct {
	RsType                 string `yaml:"rsType"`
	RsName                 string `yaml:"rsName"`
	RsDescription          string `yaml:"rsDescription"`
	ApplicationProjectName string `yaml:"applicationProjectName"`
	Parameters             struct {
		AdminPassword                        AzAdminPassword `yaml:"adminPassword"`
		AdminUsername                        string          `yaml:"adminUsername"`
		BackupPolicyName                     string          `yaml:"backupPolicyName"`
		ConfigureAddDNSSuffixScript          string          `yaml:"configureAddDNSSuffixScript"`
		ConfigureMMAMultiHomingScript        string          `yaml:"configureMMAMultiHomingScript"`
		ConfigureVMScript                    string          `yaml:"configureVMScript"`
		ConfigureWinRMScript                 string          `yaml:"configureWinRMScript"`
		EnvironmentName                      string          `yaml:"environmentName"`
		ImageOffer                           string          `yaml:"imageOffer"`
		ImagePublisher                       string          `yaml:"imagePublisher"`
		ImageSku                             string          `yaml:"imageSku"`
		ImageVersion                         string          `yaml:"imageVersion"`
		IPAddresses                          []string        `yaml:"ipAddresses"`
		LocationAbbreviation                 string          `yaml:"locationAbbreviation"`
		NsgID                                string          `yaml:"nsgId"`
		OsDiskType                           string          `yaml:"osDiskType"`
		OsType                               string          `yaml:"osType"`
		RecoveryVaultName                    string          `yaml:"recoveryVaultName"`
		ResourceTags                         AzResourceTags  `yaml:"resourceTags"`
		ScriptsStorageAccountName            string          `yaml:"scriptsStorageAccountName"`
		SecondaryWorkspaceID                 string          `yaml:"secondaryWorkspaceId"`
		SecondaryWorkspaceKey                AzWorkspaceKey  `yaml:"secondaryWorkspaceKey"`
		SubnetID                             string          `yaml:"subnetId"`
		VMBootDiagnosticsStorageURI          string          `yaml:"vmBootDiagnosticsStorageUri"`
		VMCount                              int64           `yaml:"vmCount"`
		VMLogStorageAccountEndpoint          string          `yaml:"vmLogStorageAccountEndpoint"`
		VMLogStorageAccountID                string          `yaml:"vmLogStorageAccountId"`
		VMLogStorageAccountName              string          `yaml:"vmLogStorageAccountName"`
		VMName                               string          `yaml:"vmName"`
		VMSize                               string          `yaml:"vmSize"`
		WorkspaceID                          string          `yaml:"workspaceId"`
		WorkspaceKey                         AzWorkspaceKey  `yaml:"workspaceKey"`
		LinuxPerformanceCounterConfiguration []struct {
			Annotation []struct {
				DisplayName string `yaml:"displayName"`
				Locale      string `yaml:"locale"`
			} `yaml:"annotation"`
			Class            string `yaml:"class"`
			Condition        string `yaml:"condition"`
			Counter          string `yaml:"counter"`
			CounterSpecifier string `yaml:"counterSpecifier"`
			SampleRate       string `yaml:"sampleRate"`
			Type             string `yaml:"type"`
			Unit             string `yaml:"unit"`
		} `yaml:"linuxPerformanceCounterConfiguration"`
	} `yaml:"parameters"`
}

type empty struct{}

//================================================================================================================
//  Retrieve from Azure cloud platform, and return the following items:
//
//      1. List of all available subscriptions
//      2. Map of all resource groups keyed by subscription ID
//      3. Map of all key vaults keyed by subscription ID
//
func RetrieveSubscriptionsRGsAndKVs() ([]string, map[string][]resources.Group, map[string][]keyvault.Resource, error) {
	ctx := context.TODO()
	subscriptionClient, err := azure.GetSubscriptionClientE()
	if err != nil {
		fmt.Printf("Failed to retrieve subscription client: %+v\n", err)
	}

	listResult, _ := subscriptionClient.List(ctx)
	var allSubscriptions = make([]string, 0, 10)
	var resGroups = make(map[string][]resources.Group)
	var keyVaults = make(map[string][]keyvault.Resource)
	sem := make(chan empty, len(listResult.Values())) // semaphore pattern
	for _, subscription := range listResult.Values() {
		go func(subscr subscriptions.Subscription) {
			fmt.Printf("subscription id: %s, displayname: %s\n", *subscr.SubscriptionID, *subscr.DisplayName)
			groupsClient, _ := azure.GetResourceGroupClientE(*subscr.SubscriptionID)
			listResult, _ := groupsClient.List(ctx, "", nil)
			allSubscriptions = append(allSubscriptions, *subscr.SubscriptionID)
			resGroups[*subscr.SubscriptionID] = listResult.Values()
			kvMgtClient, _ := azure.GetKeyVaultManagementClientE(*subscr.SubscriptionID)
			kvListResult, _ := kvMgtClient.List(ctx, nil)
			keyVaults[*subscr.SubscriptionID] = kvListResult.Values()
			sem <- empty{}
		}(subscription)
	}
	// wait for goroutines to finish
	for i := 0; i < len(listResult.Values()); i++ {
		<-sem
	}
	return allSubscriptions, resGroups, keyVaults, nil
}

func LocateResource(rsType, rsName string) (string, resources.Group, error) {
	ctx := context.TODO()
	subscriptionClient, err := azure.GetSubscriptionClientE()
	if err != nil {
		fmt.Printf("Failed to retrieve subscription client: %+v\n", err)
		resGroup := resources.Group{}
		return "", resGroup, azure.NotFoundError{}
	}

	fmt.Printf("Looking for resource type: %s with name: %s\n", rsType, rsName)

	var allSubscriptions = make([]string, 0, 10)
	var allResourceGroups = make([]resources.Group, 0, 10)
	listResult, _ := subscriptionClient.List(ctx)
	var checkFunc = azure.VirtualMachineExistsE
	switch rsType {
	case "vm", "vm-linux", "vm-windows":
		checkFunc = azure.VirtualMachineExistsE
	case "storage", "storageaccount", "storageaccounts", "storage-account", "storage-accounts":
		checkFunc = azure.StorageAccountExistsE
	case "networking":
		checkFunc = azure.VirtualNetworkExistsE
		//case "networksecuritygroup", "network-security-group":
		//	checkFunc = azure.GetAllNSGRulesE
	}

	totalCount := 0
	for _, subscription := range listResult.Values() {
		groupsClient, _ := azure.GetResourceGroupClientE(*subscription.SubscriptionID)
		lr, _ := groupsClient.List(ctx, "", nil)
		totalCount += len(lr.Values())
	}

	sem := make(chan empty, totalCount) // semaphore pattern
	fmt.Printf("Searching through %d # of resource groups...\n", totalCount)
	for _, subscription := range listResult.Values() {
		go func(subscr string) {
			groupsClient, _ := azure.GetResourceGroupClientE(subscr)
			lr, _ := groupsClient.List(ctx, "", nil)
			for _, rg := range lr.Values() {
				go func(fResourceExists func(string, string, string) (bool, error), resGrp resources.Group) {
					exists, _ := fResourceExists(rsName, *resGrp.Name, subscr)
					if exists {
						fmt.Printf("\nFOUND %s %s\n", subscr, *resGrp.Name)
						allSubscriptions = append(allSubscriptions, subscr)
						allResourceGroups = append(allResourceGroups, resGrp)
					}
					sem <- empty{}
				}(checkFunc, rg)
			}
		}(*subscription.SubscriptionID)
	}
	// wait for goroutines to finish
	for i := 0; i < totalCount; i++ {
		<-sem
	}
	if len(allSubscriptions) == 0 {
		return "", resources.Group{}, azure.NotFoundError{}
	} else {
		return allSubscriptions[0], allResourceGroups[0], nil
	}
}

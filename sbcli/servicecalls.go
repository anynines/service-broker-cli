package sbcli

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

// Get plan ID from plan name
func getPlanID(planName string, serviceID string) (string, error) {
	sb := NewSBClient()
	catalog, err := sb.Catalog()
	CheckErr(err)

	// find service in catalog and look for plan name
	for _, service := range catalog.Services {
		if service.ID != serviceID {
			continue
		}

		for _, plan := range service.Plans {
			if plan.Name == planName {
				return plan.ID, nil
			}
		}
	}

	return "", fmt.Errorf("Plan %s not found.", planName)
}

// retreieves all service instances from the service broker
func Services(cmd *Commandline) {
	sb := NewSBClient()
	fmt.Printf("Getting services from Servicebroker %s as %s\n", sb.Host, sb.Username)

	catalog, err := sb.Catalog()
	CheckErr(err)

	// make a map for each service which contains the available plans
	plans := make(map[string]map[string]string)
	for _, service := range catalog.Services {
		plans[service.ID] = make(map[string]string)
		for _, plan := range service.Plans {
			plans[service.ID][plan.ID] = plan.Name
		}
	}

	catalogService := make(map[string]CatalogService)
	for _, service := range catalog.Services {
		catalogService[service.ID] = service
	}

	services, err := sb.Instances()
	CheckErr(err)

	fmt.Printf("OK\n\n")

	// start writing service infos to the console
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 7, ' ', tabwriter.FilterHTML)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", "name", "service", "plan", "bound apps", "last operation", "deployment", "organization guid", "space guid")
	for _, service := range services.Resources {
		if service.State == "deleted" && cmd.NoFilter == false {
			continue
		}

		planName := "unknown"
		if _, found := plans[service.ServiceGUID]; found {
			if name, found := plans[service.ServiceGUID][service.PlanGUID]; found {
				planName = name
			}
		}

		deploymentName := "./."
		if service.DeploymentName != "" {
			deploymentName = service.DeploymentName
		}

		orgGUID := service.Context.OrganizationGUID
		spaceGUID := service.Context.SpaceGUID

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", service.GUIDAtTenant, catalogService[service.ServiceGUID].Name, planName, "./.", service.State, deploymentName, orgGUID, spaceGUID)
	}
	w.Flush()
	fmt.Println("")
}

func FindService(cmd *Commandline) {
	if len(cmd.Options) != 1 {
		CheckErr(errors.New("Missing arguments!"), GetHelpText("Deployment name"))
	}

	sb := NewSBClient()

	serviceGuid := ""

	// iterate over all service instances and find the service
	services, err := sb.Instances()
	CheckErr(err)
	for _, service := range services.Resources {
		if service.DeploymentName != "" && service.DeploymentName == cmd.Options[0] {
			serviceGuid = service.GUIDAtTenant
			break
		}
	}

	if len(serviceGuid) == 0 {
		fmt.Printf("Can't find service with deployment name: %s\n", cmd.Options[0])
	}

	serviceImpl(serviceGuid)
}

// retrieves the available service plans from the service broker
func Marketplace(cmd *Commandline) {
	sb := NewSBClient()

	fmt.Printf("Getting services from Servicebroker %s as %s\n", sb.Host, sb.Username)

	catalog, err := sb.Catalog()
	CheckErr(err)

	fmt.Print("OK\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "service\tplans\tdescription\n")
	for _, service := range catalog.Services {
		plans := make([]string, len(service.Plans))

		for i, plan := range service.Plans {
			plans[i] = plan.Name
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", service.Name, strings.Join(plans[:], ", "), service.Description)
	}
	w.Flush()
	fmt.Println("")
}

// pollLastOperation polls the last_operation endpoint until the operation reaches a
// terminal state (succeeded or failed) and prints progress to stdout.
// For async delete operations, HTTP 410 Gone is treated as success.
func pollLastOperation(sb *SBClient, instanceID string, operation string) {
	encodedOp := url.QueryEscape(operation)
	fmt.Printf("Waiting for operation to complete")
	for {
		time.Sleep(5 * time.Second)
		resp, statusCode, err := sb.LastOperation(instanceID, encodedOp)
		if err != nil {
			fmt.Printf("\nError polling last operation: %s\n", err)
			return
		}
		// 410 Gone means the instance was deleted successfully
		if statusCode == 410 {
			fmt.Printf("\nOK\n")
			return
		}
		if resp == nil {
			fmt.Printf("\nError: empty last_operation response\n")
			return
		}
		switch resp.State {
		case "succeeded":
			fmt.Printf("\nOK\n")
			if resp.Description != "" {
				fmt.Printf("%s\n", resp.Description)
			}
			return
		case "failed":
			fmt.Printf("\nFailed!\n")
			if resp.Description != "" {
				fmt.Printf("Error: %s\n", resp.Description)
			}
			os.Exit(1)
		case "in progress":
			if resp.Description != "" {
				fmt.Printf("\n  %s", resp.Description)
			} else {
				fmt.Printf(".")
			}
		default:
			fmt.Printf(".")
		}
	}
}

func getServiceIDPlanID(servicename string) (*ProvisonPayload, error) {
	sb := NewSBClient()

	instance, err := sb.Instance(servicename)
	CheckErr(err)

	orgGUID := TargetedOrganizationGUID(instance.Context.OrganizationGUID)
	spaceGUID := TargetedSpaceGUID(instance.Context.SpaceGUID)

	payload := ProvisonPayload{ServiceID: instance.ServiceGUID, PlanID: instance.PlanGUID, SpaceGUID: spaceGUID, OrganizationGUID: orgGUID}
	payload.Context.Platform = "cloudfoundry"
	payload.Context.OrganizationGUID = orgGUID
	payload.Context.SpaceGUID = spaceGUID
	payload.Context.InstanceName = servicename
	return &payload, nil
}

func Service(cmd *Commandline) {
	if len(cmd.Options) != 1 {
		CheckErr(errors.New("Missing arguments!"), GetHelpText("Service"))
	}
	serviceImpl(cmd.Options[0])
}

func serviceImpl(serviceName string) {

	sb := NewSBClient()

	catalog, err := sb.Catalog()
	CheckErr(err)

	plans := make(map[string]string)

	for _, service := range catalog.Services {
		for _, plan := range service.Plans {
			plans[plan.ID] = plan.Name
		}
	}

	catalogService := make(map[string]CatalogService)
	for _, service := range catalog.Services {
		catalogService[service.ID] = service
	}

	service, err := sb.Instance(serviceName)
	CheckErr(err)

	fmt.Println("")
	planName := "unknown"
	if name, found := plans[service.PlanGUID]; found {
		planName = name
	}

	fmt.Printf("Service instance: %s\n", service.GUIDAtTenant)
	fmt.Printf("Service: %s\n", catalogService[service.ServiceGUID].Name)
	fmt.Printf("Bound apps: %s\n", "./.")
	fmt.Printf("Tags:%s\n", strings.Join(catalogService[service.ServiceGUID].Tags, ", "))
	fmt.Printf("Plan: %s\n", planName)
	fmt.Printf("Description: %s\n", catalogService[service.ServiceGUID].Description)
	fmt.Printf("Documentation url: \n")
	if service.DashboardURL == nil {
		fmt.Printf("Dashboard: \n")
	} else {
		fmt.Printf("Dashboard: %s\n", service.DashboardURL.(string))
	}
	if service.DeploymentName == "" {
		fmt.Printf("Deployment name: \n")
	} else {
		fmt.Printf("Deployment name: %s\n", service.DeploymentName)
	}
	fmt.Printf("\n")
	fmt.Printf("Status: %s\n", service.State)
	fmt.Printf("Started: %s\n", service.CreatedAt)
	fmt.Printf("Updated: %s\n", service.UpdatedAt)
	fmt.Printf("\n")
	fmt.Printf("Organization GUID: %s\n", service.Context.OrganizationGUID)
	fmt.Printf("Space GUID: %s\n", service.Context.SpaceGUID)
	fmt.Printf("Tenant ID: %s\n", service.Metadata.TenantID)
	fmt.Printf("\n")
	if service.Metadata.UserParams == nil {
		fmt.Printf("User params: {}\n")
	} else {
		fmt.Printf("User params:\n%v\n", service.Metadata.UserParams)
	}
	fmt.Printf("\n")
	fmt.Printf("VM details: {%v}\n", service.VMDetails)

	return
}

func CreateService(cmd *Commandline) {
	sb := NewSBClient()
	fmt.Printf("Creating service at Servicebroker %s as %s\n", sb.Host, sb.Username)

	serviceName := ""

	switch len(cmd.Options) {
	case 2:
		// We will create a UUID as a servicename
		serviceName = GetUUID()
		fmt.Printf("\nNo service name give, create uuid instead: %s\n", serviceName)
	case 3:
		serviceName = cmd.Options[2]
	default:
		CheckErr(errors.New("Missing arguments!"), GetHelpText("CreateService"))
	}

	catalog, err := sb.Catalog()
	CheckErr(err)

	serviceID := -1

	for i := 0; i < len(catalog.Services); i++ {
		if catalog.Services[i].Name == cmd.Options[0] {
			serviceID = i
			break
		}
	}

	if serviceID == -1 {
		CheckErr(errors.New("Service offering not found. Check the marketplace."))
	}

	var planID string
	for _, plan := range catalog.Services[serviceID].Plans {
		if cmd.Options[1] == plan.Name {
			planID = plan.ID
		}
	}

	if planID == "" {
		CheckErr(errors.New("Service plan not found. Check the marketplace."))
	}

	orgID, _ := newUUID()
	spaceID, _ := newUUID()
	orgID = TargetedOrganizationGUID(orgID)
	spaceID = TargetedSpaceGUID(spaceID)

	data := ProvisonPayload{
		PlanID:           planID,
		OrganizationGUID: orgID,
		SpaceGUID:        spaceID,
		ServiceID:        catalog.Services[serviceID].ID,
	}

	data.Context.Platform = "cloudfoundry"
	data.Context.OrganizationGUID = orgID
	data.Context.SpaceGUID = spaceID
	data.Context.InstanceName = serviceName

	if cmd.Custom != "" {
		data.Parameters = getJSONFromCustom(cmd.Custom)
	}

	statusCode, operation, err := sb.Provision(&data, serviceName)
	CheckErr(err)

	switch statusCode {
	case 201:
		fmt.Printf("OK\n\n")
		fmt.Printf("Service %s created successfully.\n", serviceName)
	case 202:
		fmt.Printf("OK\n\n")
		fmt.Printf("Create in progress. Use 'sb services' or 'sb service %s' to check operation status.\n", serviceName)

		if cmd.Wait {
			pollLastOperation(sb, serviceName, operation)
		}
	}
}

func DeleteService(cmd *Commandline) {
	if len(cmd.Options) != 1 {
		CheckErr(errors.New("Missing arguments!"), GetHelpText("DeleteService"))
	}

	sb := NewSBClient()

	_, err := sb.Instance(cmd.Options[0])
	CheckErr(err)

	if !cmd.Force {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("\nReally delete service %s> ", cmd.Options[0])
		asked, _ := reader.ReadString('\n')
		asked = strings.TrimSpace(asked)
		if asked != "yes" {
			fmt.Println("Delete cancelled!")
			return
		}
	}

	data, err := getServiceIDPlanID(cmd.Options[0])
	CheckErr(err)

	payload := BindPayload{ServiceID: data.ServiceID, PlanID: data.PlanID}

	statusCode, operation, err := sb.Deprovision(&payload, cmd.Options[0])
	CheckErr(err)

	fmt.Printf("Deleting service %s at %s as %s...\n", cmd.Options[0], sb.Host, sb.Username)
	switch statusCode {
	case 200:
		fmt.Printf("OK\n\n")
		fmt.Printf("Service %s deleted successfully.\n", cmd.Options[0])
	case 202:
		fmt.Printf("OK\n\n")
		fmt.Printf("Delete in progress. Use 'sb services' or 'sb service %s' to check operation status.\n", cmd.Options[0])

		if cmd.Wait {
			pollLastOperation(sb, cmd.Options[0], operation)
		}
	}
}

func UpdateService(cmd *Commandline) {
	if len(cmd.Options) != 1 {
		CheckErr(errors.New("Missing arguments!"), GetHelpText("UpdateService"))
	}

	sb := NewSBClient()

	instance, err := sb.Instance(cmd.Options[0])
	CheckErr(err)

	var payload = UpdatePayload{ServiceID: instance.ServiceGUID, PlanID: instance.PlanGUID}

	orgGUID := TargetedOrganizationGUID(instance.Context.OrganizationGUID)
	spaceGUID := TargetedSpaceGUID(instance.Context.SpaceGUID)

	payload.PreviousValues.ServiceID = instance.ServiceGUID
	payload.PreviousValues.PlanID = instance.PlanGUID
	payload.PreviousValues.OrganizationID = orgGUID
	payload.PreviousValues.SpaceID = spaceGUID
	payload.Context = instance.Context
	payload.Context.OrganizationGUID = orgGUID
	payload.Context.SpaceGUID = spaceGUID

	if cmd.Plan != "" {
		planID, err := getPlanID(cmd.Plan, instance.ServiceGUID)
		CheckErr(err)

		payload.PlanID = planID
	}

	if cmd.Custom != "" {
		payload.Parameters = getJSONFromCustom(cmd.Custom)
	}

	statusCode, operation, err := sb.UpdateService(&payload, cmd.Options[0])
	CheckErr(err)

	switch statusCode {
	case 201:
		fmt.Printf("OK\n\n")
		fmt.Printf("Service %s updated successfully.\n", cmd.Options[0])
	case 202:
		fmt.Printf("OK\n\n")
		fmt.Printf("Update in progress. Use 'sb services' or 'sb service %s' to check operation status.\n", cmd.Options[0])

		if cmd.Wait {
			pollLastOperation(sb, cmd.Options[0], operation)
		}
	}
}

package servicebindings

import (
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	servicebindings "github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"
	"github.com/poy/service-catalog/cmd/svcat/output"

	"github.com/spf13/cobra"
)

// NewBindServiceCommand allows users to bind apps to service instances.
func NewBindServiceCommand(p *config.KfParams, client servicebindings.ClientInterface) *cobra.Command {
	var (
		bindingName  string
		configAsJSON string
	)

	createCmd := &cobra.Command{
		Use:     "bind-service APP_NAME SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [--binding-name BINDING_NAME]",
		Aliases: []string{"bs"},
		Short:   "Bind a service instance to an app",
		Example: `  kf bind-service myapp mydb -c '{"permissions":"read-only"}'`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			instanceName := args[1]

			cmd.SilenceUsage = true

			if bindingName == "" {
				bindingName = instanceName
			}

			params, err := services.ParseJSONOrFile(configAsJSON)
			if err != nil {
				return err
			}

			binding, err := client.Create(
				instanceName,
				appName,
				servicebindings.WithCreateBindingName(bindingName),
				servicebindings.WithCreateNamespace(p.Namespace),
				servicebindings.WithCreateParams(params))
			if err != nil {
				return err
			}

			output.WriteBindingDetails(p.Output, binding)
			return nil
		},
	}

	createCmd.Flags().StringVarP(
		&configAsJSON,
		"config",
		"c",
		"{}",
		"valid JSON object containing service-specific configuration parameters, provided in-line or in a file")

	createCmd.Flags().StringVarP(
		&bindingName,
		"binding-name",
		"b",
		"",
		"name to expose service instance to app process with (default: service instance name)")

	createCmd.SetOutput(p.Output)
	return createCmd
}

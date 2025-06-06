package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/anchore/clio"
	"github.com/anchore/grype/cmd/grype/cli/options"
	"github.com/anchore/grype/grype/db/v6/distribution"
	"github.com/anchore/grype/grype/db/v6/installation"
	"github.com/anchore/grype/grype/vulnerability"
)

type dbStatusOptions struct {
	Output                  string `yaml:"output" json:"output" mapstructure:"output"`
	options.DatabaseCommand `yaml:",inline" mapstructure:",squash"`
}

var _ clio.FlagAdder = (*dbStatusOptions)(nil)

func (d *dbStatusOptions) AddFlags(flags clio.FlagSet) {
	flags.StringVarP(&d.Output, "output", "o", "format to display results (available=[text, json])")
}

func DBStatus(app clio.Application) *cobra.Command {
	opts := &dbStatusOptions{
		Output:          textOutputFormat,
		DatabaseCommand: *options.DefaultDatabaseCommand(app.ID()),
	}

	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Display database status and metadata",
		Args:    cobra.ExactArgs(0),
		PreRunE: disableUI(app),
		RunE: func(_ *cobra.Command, _ []string) error {
			return runDBStatus(*opts)
		},
	}

	// prevent from being shown in the grype config
	type configWrapper struct {
		Hidden                   *dbStatusOptions `json:"-" yaml:"-" mapstructure:"-"`
		*options.DatabaseCommand `yaml:",inline" mapstructure:",squash"`
	}

	return app.SetupCommand(cmd, &configWrapper{Hidden: opts, DatabaseCommand: &opts.DatabaseCommand})
}

func runDBStatus(opts dbStatusOptions) error {
	client, err := distribution.NewClient(opts.ToClientConfig())
	if err != nil {
		return fmt.Errorf("unable to create distribution client: %w", err)
	}
	c, err := installation.NewCurator(opts.ToCuratorConfig(), client)
	if err != nil {
		return fmt.Errorf("unable to create distribution client: %w", err)
	}

	status := c.Status()

	if err := presentDBStatus(opts.Output, os.Stdout, status); err != nil {
		return fmt.Errorf("failed to present db status information: %+v", err)
	}

	return status.Error
}

func presentDBStatus(format string, writer io.Writer, status vulnerability.ProviderStatus) error {
	switch format {
	case textOutputFormat:
		fmt.Fprintln(writer, "Path:     ", status.Path)
		fmt.Fprintln(writer, "Schema:   ", status.SchemaVersion)
		fmt.Fprintln(writer, "Built:    ", status.Built.Format(time.RFC3339))
		if status.From != "" {
			fmt.Fprintln(writer, "From:     ", status.From)
		}
		fmt.Fprintln(writer, "Status:   ", renderStoreValidation(status))
	case jsonOutputFormat:
		enc := json.NewEncoder(writer)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", " ")
		if err := enc.Encode(&status); err != nil {
			return fmt.Errorf("failed to db status information: %+v", err)
		}
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	return nil
}

func renderStoreValidation(status vulnerability.ProviderStatus) string {
	if status.Error != nil {
		return "invalid"
	}
	return "valid"
}

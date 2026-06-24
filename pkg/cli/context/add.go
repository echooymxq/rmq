package contextcmd

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/echooymxq/rmq/pkg/config"
	"github.com/spf13/cobra"
)

func Add(r *config.RocketMQConfig) *cobra.Command {
	var (
		namesrvAddrs string
		accessKey    string
		secretKey    string
	)
	var cmd = &cobra.Command{
		Use:     "add NAME -n NAMESERVER",
		Short:   "Add a configuration context",
		Example: "  rmq context add prod -n 10.0.0.1:9876 -a xxx -s xxx",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, addrs, err := validateAddContextInput(cmd, args, namesrvAddrs)
			if err != nil {
				return err
			}
			exists, err := r.ContextExists(name)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("context %q already exists", name)
			}
			accessKey = strings.TrimSpace(accessKey)
			secretKey = strings.TrimSpace(secretKey)
			if err := confirmMissingCredentials(cmd, accessKey, secretKey); err != nil {
				return err
			}
			if err := r.AddContext(name, addrs, accessKey, secretKey); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added context %q.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&namesrvAddrs, "nameserver", "n", "", "comma separated list of namesrv host:ports")
	cmd.Flags().StringVarP(&accessKey, "accessKey", "a", "", "access key")
	cmd.Flags().StringVarP(&secretKey, "secretKey", "s", "", "secret key")

	return cmd
}

func confirmMissingCredentials(cmd *cobra.Command, accessKey, secretKey string) error {
	var missing []string
	if accessKey == "" {
		missing = append(missing, "-a, --accessKey")
	}
	if secretKey == "" {
		missing = append(missing, "-s, --secretKey")
	}
	if len(missing) == 0 {
		return nil
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "Missing credential(s): %s.\n", strings.Join(missing, ", "))
	fmt.Fprintln(cmd.ErrOrStderr(), "Only continue if the RocketMQ cluster does not enable ACL.")
	fmt.Fprint(cmd.ErrOrStderr(), "Continue without complete credentials? [y/N] ")

	answer, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && answer == "" {
		return fmt.Errorf("confirmation required: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return nil
	default:
		return errors.New("context add aborted")
	}
}

func validateAddContextInput(cmd *cobra.Command, args []string, namesrvAddrs string) (string, []string, error) {
	if len(args) > 1 {
		return "", nil, fmt.Errorf("too many arguments: %s\n\n%s", strings.Join(args[1:], " "), addContextRequiredMessage(cmd))
	}

	var missing []string
	name := ""
	if len(args) == 1 {
		name = strings.TrimSpace(args[0])
	}
	if name == "" {
		missing = append(missing, "NAME")
	}

	addrs := config.SplitAndTrim(namesrvAddrs)
	if len(addrs) == 0 {
		missing = append(missing, "-n, --nameserver")
	}

	if len(missing) > 0 {
		return "", nil, fmt.Errorf("missing required parameter(s): %s\n\n%s", strings.Join(missing, "; "), addContextRequiredMessage(cmd))
	}

	return name, addrs, nil
}

func addContextRequiredMessage(cmd *cobra.Command) string {
	commandPath := cmd.CommandPath()
	return fmt.Sprintf(`Required:
  NAME              context name, such as prod
  -n, --nameserver  comma separated list of namesrv host:ports

Usage:
  %s NAME -n NAMESERVER [-a ACCESS_KEY] [-s SECRET_KEY]

Example:
  %s prod -n 10.0.0.1:9876 -a xxx -s xxx`, commandPath, commandPath)
}

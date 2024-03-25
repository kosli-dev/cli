package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kosli-dev/cli/internal/security"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

type configOptions struct {
	setKeys   map[string]string
	unSetKeys []string
}

const configShortDesc = `Config global Kosli flags values and store them in $HOME/.kosli .  `

const configLongDesc = configShortDesc + `

Flag values are determined in the following order (highest precedence first):
- command line flags on each executed command.
- environment variables.
- custom config file provided with --config-file flag.
- default config file in $HOME/.kosli

You can configure global Kosli flags (the ones that apply to all/most commands) using their dedicated
convenience flags (e.g. --org). 

API tokens are stored in the suitable credentials manager on your machine. 

Other Kosli flags can be configured using the --set flag which takes a comma-separated list of key=value pairs.
Keys correspond to the specific flag name, capitalized. For instance: --flow would be set using --set FLOW=value
`

const configExample = `
# configure global flags in your default config file
kosli config --org=yourOrg \
	--api-token=yourAPIToken \
	--host=https://app.kosli.com \
	--debug=false \
	--max-api-retries=3

# configure non-global flags in your default config file
kosli config --set FLOW=yourFlowName

# remove a key from the default config file
kosli config --unset FLOW
`

func newConfigCmd(out io.Writer) *cobra.Command {
	o := new(configOptions)
	cmd := &cobra.Command{
		Use:     "config",
		Short:   configShortDesc,
		Long:    configLongDesc,
		Example: configExample,
		Hidden:  true,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("config-file") {
				return fmt.Errorf("cannot use --config-file with config command")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run()
		},
	}

	cmd.Flags().StringToStringVar(&o.setKeys, "set", map[string]string{}, setTagsFlag)
	cmd.Flags().StringSliceVar(&o.unSetKeys, "unset", []string{}, unsetTagsFlag)

	return cmd
}

func (o *configOptions) run() error {
	path := defaultConfigFilePathFunc()
	home := filepath.Dir(path)
	configFileName := filepath.Base(path)
	permissions := os.FileMode(0600)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("setting default config failed. Error creating file: %s", err)
		}
		defer file.Close()

		if err := file.Chmod(permissions); err != nil {
			return fmt.Errorf("setting default config failed. Error setting file permissions: %s", err)
		}

		logger.Debug("default config file created successfully with permissions: %s", permissions)
	} else if err != nil {
		return fmt.Errorf("setting default config failed. Error checking file status: %s", err)
	}

	viper.SetConfigName(configFileName)
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("setting default config failed. Error reading config file: %s", err)
	}

	if global.Org != "" {
		viper.Set("org", global.Org)
	}
	if global.Host != defaultHost {
		viper.Set("host", global.Host)
	}
	if global.ApiToken != "" {
		// get encryption key
		key, err := security.GetSecretFromCredentialsStore(credentialsStoreKeySecretName)
		if err == keyring.ErrNotFound {
			// create and save key
			keyBytes, err := security.GenerateRandomAESKey()
			if err != nil {
				return fmt.Errorf("failed to generate a new AES encryption key: %s", err)
			}
			key = string(keyBytes)
			err = security.SetSecretInCredentialsStore(credentialsStoreKeySecretName, key)
			if err != nil {
				return fmt.Errorf("failed to save AES encryption key in credentials store: %s", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to get AES encryption key: %s", err)
		}

		// encrypt the token
		encryptedTokenBytes, err := security.AESEncrypt(global.ApiToken, []byte(key))
		if err != nil {
			return err
		}
		viper.Set("api-token", string(encryptedTokenBytes))
	}
	if global.MaxAPIRetries != defaultMaxAPIRetries {
		viper.Set("max-api-retries", global.MaxAPIRetries)
	}

	viper.Set("debug", global.Debug)
	viper.Set("dry-run", global.DryRun)

	for key, value := range o.setKeys {
		viper.Set(key, value)
	}

	for _, key := range o.unSetKeys {
		viper.Set(key, nil)
	}

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("setting default config failed. Error writing config file: %s", err)
	}

	logger.Info("default config file [%s] updated successfully.", path)
	return nil
}

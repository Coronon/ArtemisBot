package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// The name of our config file, without the file extension because viper
	// supports many different config file languages
	defaultConfigFilename = ".artemisbot"

	// The environment variable prefix of all environment variables bound to our
	// command line flags.
	// For example, --number is bound to ARTEMISBOT_NUMBER.
	envPrefix = "ARTEMISBOT"

	// Replace hyphenated flag names with camelCase in the config file
	replaceHyphenWithCamelCase = false
)

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: "0.1.0",
	Use:     "artemisbot",
	Short:   "Artemis programming exercise automation",
	Long: `ArtemisBot makes it easy to navigate the flakes of the Artemis platform
by automatically handling the submission of programming exercises.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Read config
		initConfig()

		// Bind config and flag values
		bindFlags(cmd)

		// Initialize data directory
		err := os.MkdirAll(viper.GetString("workdir"), 0755)
		if err != nil {
			log.Errorf("Could not create data directory: %s", err.Error())
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	tmpDir := os.TempDir() + "artemisbot"

	// Define flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/"+defaultConfigFilename+".yaml)")

	rootCmd.PersistentFlags().StringP("workdir", "d", tmpDir, "directory to store data")
	rootCmd.PersistentFlags().BoolP("interactive", "i", false, "enter credentials interactively")
	rootCmd.PersistentFlags().StringP("verbose", "v", "", "enable verbose logging")
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Errorf("Could not find home directory: %s", err.Error())
			os.Exit(1)
		}

		// Search config in home directory with name ".artemisbot" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(defaultConfigFilename)
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("Can't read config: %v", err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name

		// If using camelCase in the config file, replace hyphens with a camelCased string.
		// Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
		if replaceHyphenWithCamelCase {
			configName = strings.ReplaceAll(f.Name, "-", "")
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(configName) {
			val := viper.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		} else {
			viper.BindPFlag(configName, f)
		}
	})
}

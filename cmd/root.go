package cmd

import (
	llog "log"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/plumber-cd/chcontainer/k8s"
	"github.com/plumber-cd/chcontainer/log"
	"github.com/plumber-cd/chcontainer/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:                   "chcontainer [args]",
		Short:                 "Like chroot, but changes containers!",
		Long:                  "See https://github.com/plumber-cd/chcontainer/README.md for details",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			viper.Set("quiet", true)
			log.SetupLog()

			log.Debug.Print("Start root command execution")

			executable := filepath.Base(os.Args[0])
			log.Debug.Printf("Executable %s -> %s", os.Args[0], executable)

			pod := viper.GetString("pod")
			if pod == "" {
				log.Normal.Fatal("CHC_POD is not set")
			}
			log.Debug.Printf("Pod: %s", pod)

			thisContainer := viper.GetString("container")
			if thisContainer == "" {
				log.Normal.Fatal("CHC_CONTAINER is not set")
			}
			log.Debug.Printf("This container: %s", thisContainer)

			var container, bin string
			for _, symlinkCfg := range viper.GetStringSlice("symlinks") {
				log.Debug.Printf("Parsing symlinkCfg: %s", symlinkCfg)

				symlinkSlice := strings.SplitN(symlinkCfg, ":", 3)
				if len(symlinkSlice) != 3 {
					log.Normal.Fatalf("Invalid %s", symlinkCfg)
				}
				log.Debug.Printf("Parsing symlinkSlice: %s", symlinkSlice)

				if thisContainer == symlinkSlice[1] {
					log.Debug.Printf("Skip %s because it is for current container %s", symlinkCfg, thisContainer)
					continue
				}

				symlink := filepath.Base(symlinkSlice[0])
				log.Debug.Printf("Symlink: %s", symlink)

				if executable == symlink {
					log.Debug.Printf("Match found: %s -> container[%s]%s", executable, symlinkSlice[1], symlinkSlice[2])
					container = symlinkSlice[1]
					bin = symlinkSlice[2]
					break
				}
			}

			if container == "" || bin == "" {
				log.Normal.Fatalf("Unable to find symlink mapping: %s", executable)
			}

			viper.Set("container", container)

			k8s.Run(append([]string{bin}, args...))
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "chcontainer-config", "", "global config file (default is $HOME/.chcontainer.yaml)")
	rootCmd.PersistentFlags().Lookup("chcontainer-config").Hidden = true

	rootCmd.PersistentFlags().StringSlice("chcontainer-symlinks", []string{}, "List of symlinks")
	if err := viper.BindPFlag("symlinks", rootCmd.PersistentFlags().Lookup("chcontainer-symlinks")); err != nil {
		llog.Panic(err)
	}
	rootCmd.PersistentFlags().Lookup("chcontainer-symlinks").Hidden = true

	rootCmd.DisableFlagParsing = true
	rootCmd.Flags().SetInterspersed(false)
}

func initConfig() {
	log.Debug.Print("Read viper configs")

	if cfgFile != "" {
		log.Debug.Printf("Using custom global config path: %s", cfgFile)

		exists, err := utils.FileExists(cfgFile)
		if err != nil {
			log.Normal.Panic(err)
		}
		if !exists {
			log.Normal.Fatalf("Global config file not found: %s", cfgFile)
		}

		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		log.Debug.Print("Using default global config in user home")

		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Normal.Panic(err)
		}

		// Search config in home directory with name ".chcontainer" (without extension).
		viper.SetConfigName("config")
		viper.AddConfigPath(filepath.Join(home, ".chcontainer"))
	}

	log.Debug.Print("Enabling ENV parsing for viper")
	// This is so we can set any nested viper settings via env variables, replacing every . with _
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// RT short from RunTainer
	viper.SetEnvPrefix("CHC")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Debug.Printf("Global %s, skipping...", err)
		default:
			log.Normal.Panic(err)
		}
	} else {
		log.Debug.Print("Using global config file:", viper.ConfigFileUsed())
	}

	// try to read (if exists) local config file in the cwd
	cwd, err := os.Getwd()
	if err != nil {
		log.Normal.Panic(err)
	}
	readLocalConfig(cwd)

	log.Debug.Print("Viper configs loaded, re-initialize logger in case anything changed...")
	log.SetupLog()
}

// readLocalConfig read viper config in the directory.
// Due to https://github.com/spf13/viper/issues/181,
// seems like there's not really a way to override with multiple config files.
// So we will read local config files into a separate viper instances, and then use MergeConfigMap with AllSettings.
func readLocalConfig(d string) {
	log.Debug.Printf("Reading config file in %s", d)

	v := viper.New()
	v.SetConfigName(".chcontainer")
	v.AddConfigPath(d)
	if err := v.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Debug.Printf("Local %s, skipping...", err)
		default:
			log.Normal.Panic(err)
		}
	} else {
		log.Debug.Print("Using local config file:", v.ConfigFileUsed())
		if err := viper.MergeConfigMap(v.AllSettings()); err != nil {
			log.Normal.Panic(err)
		}
	}
}

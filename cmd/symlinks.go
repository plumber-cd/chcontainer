package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/plumber-cd/chcontainer/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(symlinksCmd)
}

var symlinksCmd = &cobra.Command{
	Use:   "symlinks",
	Short: "Run this once in the container to set up all symlinks",
	Long:  `This will look into the configuration and create all missing symlinks`,
	Run: func(cmd *cobra.Command, args []string) {
		executable, err := os.Executable()
		if err != nil {
			log.Normal.Panic(err)
		}
		log.Debug.Printf("Executable %s", executable)

		container := viper.GetString("container")
		if container == "" {
			log.Normal.Fatal("CHC_CONTAINER is not set")
		}
		log.Debug.Printf("Container: %s", container)

		for _, symlinkCfg := range viper.GetStringSlice("symlinks") {
			log.Debug.Printf("Parsing symlinkCfg: %s", symlinkCfg)

			symlinkSlice := strings.SplitN(symlinkCfg, ":", 3)
			if len(symlinkSlice) != 3 {
				if err != nil {
					log.Normal.Fatalf("Invalid %s", symlinkCfg)
				}
			}
			log.Debug.Printf("Parsing symlinkSlice: %s", symlinkSlice)

			if container == symlinkSlice[1] {
				log.Debug.Printf("Skip %s because it is for current container %s", symlinkCfg, container)
				continue
			}

			path := filepath.Dir(symlinkSlice[0])
			path, err := filepath.Abs(path)
			if err != nil {
				log.Normal.Panic(err)
			}
			os.MkdirAll(path, 0755)

			osSymlink := filepath.Join(path, filepath.Base(symlinkSlice[0]))
			log.Normal.Printf("Creating symlink %s -> container[%s]:%s", osSymlink, symlinkSlice[1], symlinkSlice[2])

			if err := os.Symlink(executable, osSymlink); err != nil {
				log.Normal.Panic(err)
			}
		}
	},
}

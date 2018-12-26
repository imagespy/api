package cmd

import (
	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/scrape"
	"github.com/imagespy/api/store/gorm"
	"github.com/imagespy/api/updater"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	updaterDBConnection     string
	updaterLogLevel         string
	updaterPromPushAddress  string
	updaterRegistryAddress  string
	updaterRegistryInsecure bool
	updaterRegistryPassword string
	updaterRegistryUsername string
	updaterWorkerCount      int
)

var updaterCmd = &cobra.Command{
	Use:   "updater",
	Short: "Updates the latest version of images",
	Run: func(cmd *cobra.Command, args []string) {
		mustInitLogging(updaterLogLevel)
		s, err := gorm.New(updaterDBConnection)
		if err != nil {
			log.Fatal(err)
		}

		defer s.Close()

		reg, err := registry.NewRegistry(
			updaterRegistryAddress,
			registry.Opts{
				Insecure: updaterRegistryInsecure,
				Password: updaterRegistryPassword,
				Username: updaterRegistryUsername,
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		scraper := scrape.NewScraper(s)
		u := updater.NewGroupingUpdater(updaterPromPushAddress, reg, scraper, s, updaterWorkerCount)
		err = u.Run()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	updaterCmd.Flags().StringVar(&updaterDBConnection, "db.connection", "", "connection string to connect to the database")
	updaterCmd.Flags().StringVar(&updaterLogLevel, "log.level", "warn", "log level")
	updaterCmd.Flags().StringVar(&updaterPromPushAddress, "pushgateway.address", "", "address of the Prometheus Pushgateway")
	updaterCmd.Flags().StringVar(&updaterRegistryAddress, "registry.address", "docker.io", "address of the docker registry")
	updaterCmd.Flags().BoolVar(&updaterRegistryInsecure, "registry.insecure", false, "disable certificate validation")
	updaterCmd.Flags().StringVar(&updaterRegistryPassword, "registry.password", "", "password to authenticate against the docker registry")
	updaterCmd.Flags().StringVar(&updaterRegistryUsername, "registry.username", "", "username to authenticate against the docker registry")
	updaterCmd.Flags().IntVar(&updaterWorkerCount, "workers", 1, "number of workers that process updates")
	rootCmd.AddCommand(updaterCmd)
}

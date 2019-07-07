package cmd

import (
	"database/sql"
	"time"

	spylog "github.com/imagespy/api/log"
	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/scrape"
	"github.com/imagespy/api/store/gorm"
	"github.com/imagespy/api/updater"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	updaterDBConnection       string
	updaterLogLevel           string
	updaterPromPushAddress    string
	updaterRegistryAddress    string
	updaterRegistryAuthMethod string
	updaterRegistryInsecure   bool
	updaterRegistryPassword   string
	updaterRegistryUsername   string
	updaterWorkerCount        int
)

var updaterCmd = &cobra.Command{
	Use: "updater",
}

var updaterAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Updates all images",
	Run: func(cmd *cobra.Command, args []string) {
		mustInitLogging(updaterLogLevel)
		s, err := gorm.New(updaterDBConnection)
		if err != nil {
			log.Fatal(spylog.FormatError(err))
		}

		defer s.Close()
		db, err := sql.Open("mysql", updaterDBConnection)
		if err != nil {
			log.Fatal(spylog.FormatError(err))
		}

		builder := &registry.Builder{
			Configs: []*registry.Config{
				{
					Address:        updaterRegistryAddress,
					Authentication: updaterRegistryAuthMethod,
					Insecure:       updaterRegistryInsecure,
					Timeout:        60 * time.Second,
				},
			},
		}
		scraper := scrape.NewScraper(s)
		u := updater.NewAllImagesUpdater(updaterPromPushAddress, db, builder.NewRepositoryFromImage, scraper)
		err = u.Run()
		if err != nil {
			log.Fatal(spylog.FormatError(err))
		}
	},
}

var updaterLatestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Updates the latest version of images",
	Run: func(cmd *cobra.Command, args []string) {
		mustInitLogging(updaterLogLevel)
		s, err := gorm.New(updaterDBConnection)
		if err != nil {
			log.Fatal(spylog.FormatError(err))
		}

		defer s.Close()
		builder := &registry.Builder{
			Configs: []*registry.Config{
				{
					Address:        updaterRegistryAddress,
					Authentication: updaterRegistryAuthMethod,
					Insecure:       updaterRegistryInsecure,
					Timeout:        60 * time.Second,
				},
			},
		}
		scraper := scrape.NewScraper(s)
		u := updater.NewLatestImageUpdater(updaterPromPushAddress, builder.NewRepositoryFromImage, scraper, s, updaterWorkerCount)
		err = u.Run()
		if err != nil {
			log.Fatal(spylog.FormatError(err))
		}
	},
}

func init() {
	updaterCmd.PersistentFlags().StringVar(&updaterDBConnection, "db.connection", "", "connection string to connect to the database")
	updaterCmd.PersistentFlags().StringVar(&updaterLogLevel, "log.level", "warn", "log level")
	updaterCmd.PersistentFlags().StringVar(&updaterPromPushAddress, "pushgateway.address", "", "address of the Prometheus Pushgateway")
	updaterCmd.PersistentFlags().StringVar(&updaterRegistryAddress, "registry.address", "docker.io", "address of the docker registry")
	updaterCmd.PersistentFlags().StringVar(&updaterRegistryAuthMethod, "registry.auth-method", "", "use given method to authenticate against the docker registry (basic or token)")
	updaterCmd.PersistentFlags().BoolVar(&updaterRegistryInsecure, "registry.insecure", false, "disable certificate validation")
	updaterCmd.PersistentFlags().StringVar(&updaterRegistryPassword, "registry.password", "", "password to authenticate against the docker registry")
	updaterCmd.PersistentFlags().StringVar(&updaterRegistryUsername, "registry.username", "", "username to authenticate against the docker registry")
	updaterCmd.PersistentFlags().IntVar(&updaterWorkerCount, "workers", 1, "number of workers that process updates")
	updaterCmd.AddCommand(updaterAllCmd)
	updaterCmd.AddCommand(updaterLatestCmd)
	rootCmd.AddCommand(updaterCmd)
}

package cmd

import (
	"crypto/tls"
	"net/http"

	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/scrape"
	"github.com/imagespy/api/store/gorm"
	"github.com/imagespy/api/web"
	registryC "github.com/imagespy/registry-client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	serverDBConnection      string
	serverHTTPAddress       string
	serverLogLevel          string
	serverMigrationsEnabled bool
	serverMigrationsPath    string
	serverRegistryAddress   string
	serverRegistryInsecure  bool
	serverRegistryPassword  string
	serverRegistryUsername  string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Serves the HTTP API",
	Run: func(cmd *cobra.Command, args []string) {
		mustInitLogging(serverLogLevel)
		if serverMigrationsEnabled {
			log.Info("executing migrations")
			err := gorm.Migrate(serverDBConnection, serverMigrationsPath)
			if err != nil {
				log.Fatalf("error executing migrations: %s", err)
			}
			log.Info("migrations executed")
		}

		s, err := gorm.New(serverDBConnection)
		if err != nil {
			log.Fatalf("unable to connect to database: %s", err)
		}

		defer s.Close()

		reg, err := registry.NewRegistry(
			serverRegistryAddress,
			registry.Opts{
				Insecure: serverRegistryInsecure,
				Password: serverRegistryPassword,
				Username: serverRegistryUsername,
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		regCHttpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: serverRegistryInsecure,
				},
			},
		}
		regC := registryC.New(registryC.Options{
			Client: regCHttpClient,
			Domain: serverRegistryAddress,
		})
		handler := web.Init(regC, reg, scrape.NewScraper(s), s)
		log.Fatal(http.ListenAndServe(serverHTTPAddress, handler))
	},
}

func init() {
	serverCmd.Flags().StringVar(&serverDBConnection, "db.connection", "", "connection string to connect to the database")
	serverCmd.Flags().StringVar(&serverHTTPAddress, "http.address", ":3001", "ip:port combination to bind to")
	serverCmd.Flags().StringVar(&serverLogLevel, "log.level", "warn", "set the log level")
	serverCmd.Flags().BoolVar(&serverMigrationsEnabled, "migrations.enabled", false, "execute migrations on startup")
	serverCmd.Flags().StringVar(&serverMigrationsPath, "migrations.path", "file:///migrations", "path to directory containing migration files")
	serverCmd.Flags().StringVar(&serverRegistryAddress, "registry.address", "docker.io", "the address of the docker registry")
	serverCmd.Flags().BoolVar(&serverRegistryInsecure, "registry.insecure", false, "disable certificate validation")
	serverCmd.Flags().StringVar(&serverRegistryPassword, "registry.password", "", "the password to authenticate against the docker registry")
	serverCmd.Flags().StringVar(&serverRegistryUsername, "registry.username", "", "the username to authenticate against the docker registry")
	rootCmd.AddCommand(serverCmd)
}

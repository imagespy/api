package cmd

import (
	"net/http"

	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/scrape"
	"github.com/imagespy/api/store/gorm"
	"github.com/imagespy/api/web"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	serverDBConnection string
	serverHTTPAddress  string
	serverLogLevel     string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Serves the HTTP API",
	Run: func(cmd *cobra.Command, args []string) {
		mustInitLogging(serverLogLevel)
		s, err := gorm.New(serverDBConnection)
		if err != nil {
			log.Fatalf("unable to connect to database: %s", err)
		}

		defer s.Close()

		reg, err := registry.NewRegistry("", registry.Opts{})
		if err != nil {
			log.Fatal(err)
		}

		handler := web.Init(scrape.NewScraper(reg, s), s)
		log.Fatal(http.ListenAndServe(serverHTTPAddress, handler))
	},
}

func init() {
	serverCmd.Flags().StringVar(&serverDBConnection, "db.connection", "", "connection string to connect to the database")
	serverCmd.Flags().StringVar(&serverHTTPAddress, "http.address", ":3001", "ip/port combination to bind to")
	serverCmd.Flags().StringVar(&serverLogLevel, "log.level", "warn", "set the log level")
	rootCmd.AddCommand(serverCmd)
}

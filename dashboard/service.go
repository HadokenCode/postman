package dashboard

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/rgamba/postman/async"
	"github.com/rgamba/postman/stats"

	"github.com/spf13/viper"
)

var (
	appConfig  *viper.Viper
	appVersion string
	appBuild   string
)

// StartHTTPServer starts the new HTTP Dashboard service.
func StartHTTPServer(port int, config *viper.Viper, version string, build string) *http.Server {
	appConfig = config
	appVersion = version
	appBuild = build
	mux := http.NewServeMux()
	mux.HandleFunc("/settings", settingsHandler)
	mux.HandleFunc("/", defaultHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Dashboard: ListenAndServe() error: %s", err)
		}
	}()

	return srv
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	context := map[string]interface{}{
		"service":   appConfig.GetStringMap("service"),
		"http":      appConfig.GetStringMap("http"),
		"dashboard": appConfig.GetStringMap("dashboard"),
	}
	renderView(w, "settings.html", context)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	context := map[string]interface{}{
		"currentServiceName":      appConfig.GetString("service.name"),
		"currentServiceInstances": async.GetServiceInstances(appConfig.GetString("service.name")),
		"processId":               os.Getpid(),
		"requests":                stats.GetRequestsLastMinutePerService(),
		"appVersion":              appVersion,
		"appBuild":                appBuild,
	}
	renderView(w, "index.html", context)
}

func renderView(w http.ResponseWriter, tpl string, data interface{}) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			http.Error(w, rec.(string), http.StatusInternalServerError)
		}
	}()

	root := "assets/html/"
	t := template.Must(template.ParseFiles(root+tpl, root+"header.html", root+"footer.html"))

	err := t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
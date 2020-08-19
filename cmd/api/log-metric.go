package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"gitlab.moebius.com/mantis/pkg/logmetric/db"
	"gitlab.moebius.com/mantis/pkg/logmetric/server"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	logmonitorv1 "gitlab.moebius.com/mantis/api/v1"
	controllers "gitlab.moebius.com/mantis/pkg/logmetric/controller"
	"gitlab.moebius.com/mantis/pkg/logmetric/model"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = logmonitorv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type Args struct {
	MetricsAddr      string
	Debug            bool
	ElasticsearchUrl string
	DryRun           bool
}

func main() {
	args := Args{
		ElasticsearchUrl: os.Getenv("ELASTICSEARCH-URL"),
		MetricsAddr:      os.Getenv("METRICS-ADDR"),
	}
	// flag.StringVar(&args.MetricsAddr, "listen-address", ":8080", "The address to listen on for HTTP requests.")
	flag.BoolVar(&args.Debug, "debug", true, "debug or not.")
	flag.BoolVar(&args.DryRun, "dryrun", false, "uses a null db driver that writes received webhooks to stdout")

	flag.Parse()

	var driver db.Storer
	if args.DryRun {
		log.Println("dry-run")
		driver = db.NullDB{}
	} else {
		elasticUrls := strings.Split(args.ElasticsearchUrl, ",")
		driver, _ = db.ConnectES(elasticUrls)
	}
	driver.GetVersion()
	elasticMetricMap := &model.ElasticMetricMap{}

	s := server.New(args.Debug, driver, elasticMetricMap)
	go s.Start(args.MetricsAddr)

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		Port:               9443,
		MetricsBindAddress: "0",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.LogMonitorReconciler{
		Client:           mgr.GetClient(),
		Log:              ctrl.Log.WithName("controllers").WithName("LogMonitor"),
		Scheme:           mgr.GetScheme(),
		ElasticMetricMap: elasticMetricMap,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LogMonitor")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

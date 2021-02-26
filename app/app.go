package app

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"github.com/ralfonso-directnic/utron/config"
	"github.com/ralfonso-directnic/utron/controller"
	"github.com/ralfonso-directnic/utron/logger"
	"github.com/ralfonso-directnic/utron/models"
	"github.com/ralfonso-directnic/utron/router"
	"github.com/ralfonso-directnic/utron/view"
	"github.com/ralfonso-directnic/utron/session"
	"github.com/gorilla/sessions"
)

//StaticServerFunc is a function that returns the static assetsfiles server.
//
// The first argument retrued is the path prefix for the static assets. If strp
// is set to true then the prefix is going to be stripped.
type StaticServerFunc func(*config.Config) (prefix string, strip bool,  h http.Handler)

// App is the main utron application.
type App struct {
	Router       *router.Router
	Config       *config.Config
	View         view.View
	Log          logger.Logger
	Model        *models.Model
	ConfigPath   string
	StaticServer StaticServerFunc
	SessionStore sessions.Store
	isInit       bool
}

// NewApp creates a new bare-bone utron application. To use the MVC components, you should call
// the Init method before serving requests.
func NewApp() *App {
	return &App{
		Log:    logger.NewDefaultLogger(os.Stdout),
		Router: router.NewRouter(),
		Model:  models.NewModel(),
	}
}

// NewMVC creates a new MVC utron app. If cfg is passed, it should be a directory to look for
// the configuration files. The App returned is initialized.
func NewMVC(cfg ...string) (*App, error) {
	app := NewApp()
	if len(cfg) > 0 {
		app.SetConfigPath(cfg[0])
	}
	if err := app.Init(); err != nil {
		return nil, err
	}
	return app, nil
}

//StaticServer implements StaticServerFunc.
//
// This uses the http.Filesystem to handle static assets. The routes prefixed
// with /static/ are static asset routes by default.
func StaticServer(cfg *config.Config) (string, bool,  http.Handler) {
	static, _ := GetAbsolutePath(cfg.StaticDir)
	if static != "" {
		return "/static/", true, http.StripPrefix("/static/", http.FileServer(http.Dir(static)))
	}
	return "", false, nil
}

func (a *App) options() *router.Options {
	return &router.Options{
		Model:        a.Model,
		View:         a.View,
		Config:       a.Config,
		Log:          a.Log,
		SessionStore: a.SessionStore,
	}
}

// Init initializes the MVC App.
func (a *App) Init() error {
	if a.ConfigPath == "" {
		a.SetConfigPath("config")
	}
	return a.init()
}

// SetConfigPath sets the directory path to search for the config files.
func (a *App) SetConfigPath(dir string) {
	a.ConfigPath = dir
}

// init initializes values to the app components.
func (a *App) init() error {
	appConfig, err := loadConfig(a.ConfigPath)
	if err != nil {
		return err
	}
	a.Config = appConfig

	// only when mode is allowed
	if !a.Config.NoModel {
		model := models.NewModel()
		err = model.OpenWithConfig(appConfig)
		if err != nil {
			return err
		}
		a.Model = model
	}
	
	//if no session store is set, then load a default
	
	if(a.SessionStore == nil){

    	// The sessionistore s really not critical. The application can just run
    	// without session set
    	store, err := getSessionStore(appConfig)
    	if err == nil {
    		a.SessionStore = store
    	}
    	
	}

	a.Router.Options = a.options()
	a.Router.LoadRoutes(a.ConfigPath) // Load a routes file if available.	
	a.isInit = true

	// In case the StaticDir is specified in the Config file, register
	// a handler serving contents of that directory under the PathPrefix /static/.

	if (a.StaticServer!=nil){
    	
    	dir,success,fs := a.StaticServer(a.Config)
    	
    	if(success == true){
    
    	//we can't use static because if we strip the path prefix in Router.Static then we don't look into the embedded fs correctly
    	//a.Router.PathPrefix(dir).Handler(http.FileServer(fs))
    	a.Router.AddHandler(dir,fs)

    	
    	}
    	
    	
	}else if a.Config.StaticDir != "" {
		           
		           
		            static, _ := GetAbsolutePath(a.Config.StaticDir)
			       //a.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(static))))
			       if(static!=""){
			       a.Router.Static("/static/",http.Dir(static))
		           }

	}
	
	
	
	return nil
}

//sets the view to use

func (a *App) SetView(vw view.View){

a.View = vw

}

func getSessionStore(cfg *config.Config) (sessions.Store, error) {
	
	sess := session.New(cfg)
	
	return sess.LoadStore()
	
}

func keyPairs(src []string) [][]byte {
	var pairs [][]byte
	for _, v := range src {
		pairs = append(pairs, []byte(v))
	}
	return pairs
}

// getAbsolutePath returns the absolute path to dir. If the dir is relative, then we add
// the current working directory. Checks are made to ensure the directory exist.
// In case of any error, an empty string is returned.
func GetAbsolutePath(dir string) (string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("untron: %s is not a directory", dir)
	}

	if filepath.IsAbs(dir) { // If dir is already absolute, return it.
		return dir, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	absDir := filepath.Join(wd, dir)
	_, err = os.Stat(absDir)
	if err != nil {
		return "", err
	}
	return absDir, nil
}

func (a *App) LoadConfig() error { 
	
   	appConfig, err := loadConfig(a.ConfigPath)
	if err != nil {
		return err
	}
	a.Config = appConfig
        return nil
}

// loadConfig loads the configuration file. If cfg is provided, then it is used as the directory
// for searching the configuration files. It defaults to the directory named config in the current
// working directory.
func loadConfig(cfg ...string) (*config.Config, error) {
	cfgDir := "config"
	if len(cfg) > 0 {
		cfgDir = cfg[0]
	}

	// Load configurations.
	cfgFile, err := findConfigFile(cfgDir, "app")
	if err != nil {
		return nil, err
	}
	return config.NewConfig(cfgFile)
}

// findConfigFile finds the configuration file name in the directory dir.
func findConfigFile(dir string, name string) (file string, err error) {
	extensions := []string{".json", ".toml", ".yml", ".hcl"}

	for _, ext := range extensions {
		file = filepath.Join(dir, name)
		if info, serr := os.Stat(file); serr == nil && !info.IsDir() {
			return
		}
		file = file + ext
		if info, serr := os.Stat(file); serr == nil && !info.IsDir() {
			return
		}
	}
	return "", fmt.Errorf("utron: can't find configuration file %s in %s", name, dir)
}

// AddController registers a controller, and middlewares if any is provided.
func (a *App) AddController(ctrlfn func() controller.Controller, middlewares ...interface{}) {
	_ = a.Router.Add(ctrlfn, middlewares...)
}

// ServeHTTP serves http requests. It can be used with other http.Handler implementations.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.Router.ServeHTTP(w, r)
}

//SetNotFoundHandler this sets the hadler that is will execute when the route is
//not found.
func (a *App) SetNotFoundHandler(h http.Handler) error {
	if a.Router != nil {
		a.Router.NotFoundHandler = h
		return nil
	}
	return errors.New("untron: application router is not set")
}

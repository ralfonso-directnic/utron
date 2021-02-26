package main

import (
   "fmt"
    "github.com/example/view"
    "github.com/example/models"
    "log"
    "reflect"
    "net/http"
    "html/template"
    "github.com/ralfonso-directnic/utron"
    "github.com/ralfonso-directnic/utron/config"
    "github.com/gorilla/sessions"
    "strings"
)



var store sessions.Store
var cfg  *config.Config




func main() {

    log.SetFlags(log.LstdFlags | log.Lshortfile)
    // Start the MVC App
    app := utron.NewApp()
    
    app.SetConfigPath("config")
    
    err := app.LoadConfig()

    if(err!=nil){
        log.Fatal("Possible config error:",err)
    }


    app.Model.Register(&models.Example{})
    vw,_ :=  appview.NewBasicView(app.Config.ViewsDir,template.FuncMap{})
    //you must set view before init or it doesn't get passed tot he router
    app.SetView(vw)


    app.StaticServer = myStaticServer //optional, will use built in if not specified
  
    app.Init()

    
    // CReate Models tables if they dont exist yet
    //app.Model.AutoMigrateAll()

    // Register Controller
    app.AddController(c.NewExampleController)
    app.AddController(c.NewAuthController)

    //listen to both http and https
    
    go func() {
    
    port := fmt.Sprintf(":%d",app.Config.Port)
    app.Log.Info(fmt.Sprintf("Listening on: %s:%d", app.Config.BaseURL ,app.Config.Port))
    log.Fatal(http.ListenAndServe(port, app))
    
    }()

	if len(app.Config.SslCertPath) > 0 && len(app.Config.SslKeyPath) > 0 {
    	portssl := fmt.Sprintf(":%d",app.Config.PortSsl)
    	app.Log.Info(fmt.Sprintf("Listening on: %s:%d", app.Config.BaseURL ,app.Config.PortSsl))
		log.Fatal(http.ListenAndServeTLS(portssl, app.Config.SslCertPath, app.Config.SslKeyPath, app))
	}
	
	select{}
}

func myStaticServer(cfg *config.Config) (string, bool,  http.FileSystem) {
	static, _ := app.GetAbsolutePath(cfg.StaticDir)
	if static != "" {
		return "/static/", true, http.Dir(static)
	}
	return "", false, nil
}


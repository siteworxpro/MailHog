package main

import (
	"flag"
	"fmt"
	"os"

	gohttp "net/http"

	"github.com/gorilla/pat"
	"github.com/ian-kent/go-log/log"
	cfgcom "github.com/siteworxpro/MailHog/config"
	"github.com/siteworxpro/MailHog/server/api"
	cfgapi "github.com/siteworxpro/MailHog/server/config"
	"github.com/siteworxpro/MailHog/server/mailhoghttp"
	"github.com/siteworxpro/MailHog/server/smtp"
	"github.com/siteworxpro/MailHog/ui/assets"
	cfgui "github.com/siteworxpro/MailHog/ui/config"
	"github.com/siteworxpro/MailHog/ui/web"
	"github.com/siteworxpro/mhsendmail/cmd"
	"golang.org/x/crypto/bcrypt"
)

var configMap *cfgapi.Config
var uiConf *cfgui.Config
var comConf *cfgcom.Config
var exitCh chan int
var version string

func configure() {
	cfgcom.RegisterFlags()
	cfgapi.RegisterFlags()
	cfgui.RegisterFlags()
	flag.Parse()
	configMap = cfgapi.Configure()
	uiConf = cfgui.Configure()
	comConf = cfgcom.Configure()

	configMap.WebPath = comConf.WebPath
	uiConf.WebPath = comConf.WebPath
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-version" || os.Args[1] == "--version") {
		fmt.Println("MailHog version: " + version)
		os.Exit(0)
	}

	if len(os.Args) > 1 && os.Args[1] == "sendmail" {
		args := os.Args
		os.Args = []string{args[0]}
		if len(args) > 2 {
			os.Args = append(os.Args, args[2:]...)
		}
		cmd.Go()
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "bcrypt" {
		var pw string
		if len(os.Args) > 2 {
			pw = os.Args[2]
		} else {
			// TODO: read from stdin
		}
		b, err := bcrypt.GenerateFromPassword([]byte(pw), 4)
		if err != nil {
			log.Fatalf("error bcrypting password: %s", err)
			os.Exit(1)
		}
		fmt.Println(string(b))
		os.Exit(0)
	}

	configure()

	if comConf.AuthFile != "" {
		http.AuthFile(comConf.AuthFile)
	}

	exitCh = make(chan int)
	if uiConf.UIBindAddr == configMap.APIBindAddr {
		cb := func(r gohttp.Handler) {
			web.CreateWeb(uiConf, r.(*pat.Router), assets.Asset)
			api.CreateAPI(configMap, r.(*pat.Router))
		}
		go http.Listen(uiConf.UIBindAddr, assets.Asset, exitCh, cb)
	} else {
		cb1 := func(r gohttp.Handler) {
			api.CreateAPI(configMap, r.(*pat.Router))
		}
		cb2 := func(r gohttp.Handler) {
			web.CreateWeb(uiConf, r.(*pat.Router), assets.Asset)
		}
		go http.Listen(configMap.APIBindAddr, assets.Asset, exitCh, cb1)
		go http.Listen(uiConf.UIBindAddr, assets.Asset, exitCh, cb2)
	}
	go smtp.Listen(configMap, exitCh)

	<-exitCh
	log.Printf("Received exit signal")
}

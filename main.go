package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/jessevdk/go-flags"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type JSON map[string]interface{}

type Rest struct {
	Version    string
	lock       sync.Mutex
	httpServer *http.Server
}

type Opts struct {
	Port int `long:"port" env:"PORT" default:"8080" description:"port"`
}

const VERSION = "v2"
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!-1234567890+_"

func RandStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Run the lister and request's router, activate rest server
func (s *Rest) Run(hostUrl string, port int) {
	s.lock.Lock()
	router := s.routes()
	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", hostUrl, port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	s.lock.Unlock()
	_ = s.httpServer.ListenAndServe()
}

// Shutdown the rest server
func (s *Rest) Shutdown() {
	s.httpServer.SetKeepAlivesEnabled(false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	s.lock.Lock()
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("[DEBUG] rest shutdown error, %s", err)
		}
	}
	log.Print("[INFO] shutdown rest server completed")
	s.lock.Unlock()
}

func (s *Rest) routes() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Route("/api/v2", func(rapi chi.Router) {
		// Login & Logout
		rapi.Group(func(rlogin chi.Router) {
			rlogin.Post("/Login.json", s.loginCtrl)
			rlogin.Post("/Logout.json", s.logoutCtrl)
		})
		// Actg
		rapi.Group(func(rActg chi.Router) {
			rActg.Post("/List/ActgClass.json", s.actgClassListCtrl)
		})
		// Vendor
		rapi.Group(func(rVendor chi.Router) {
			rVendor.Post("/List/Vendor.json", s.vendorListCtrl)
			rVendor.Post("/Crud/Create/Vendor.json", s.vendorCreateCtrl)
			rVendor.Post("/Crud/Read/Vendor.json", s.vendorReadCtrl)
			rVendor.Post("/Crud/Update/Vendor.json", s.vendorUpdateCtrl)
		})
		// Bill
		rapi.Group(func(rVendor chi.Router) {
			rVendor.Post("/Crud/Read/Bill.json", s.billReadCtrl)
			rVendor.Post("/Crud/Create/Bill.json", s.billCreateCtrl)
			rVendor.Post("/Crud/Delete/Bill.json", s.billReadCtrl)
			rVendor.Post("/Crud/Update/Bill.json", s.billUpdateCtrl)
		})
	})
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		render.HTML(w, r, "pong")
	})
	return router
}

// POST /Login.json returns the login response
func (s *Rest) loginCtrl(w http.ResponseWriter, r *http.Request) {
	loginResponse := JSON{
		"response_status":  0,
		"response_message": "Success",
		"response_data": JSON{
			"apiEndPoint": "https://api-mock.bill.com/api/v2",
			"usersId":     RandStringBytes(20),
			"sessionId":   RandStringBytes(45),
			"orgId":       RandStringBytes(20),
		},
	}
	render.JSON(w, r, loginResponse)
}

// POST /Logout.json returns the logout response
func (s *Rest) logoutCtrl(w http.ResponseWriter, r *http.Request) {
	loginResponse := JSON{
		"response_status": 0, "response_message": "Success", "response_data": JSON{},
	}
	render.JSON(w, r, loginResponse)
}

// POST /List/ActgClass.json search class by name
func (s *Rest) actgClassListCtrl(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println("[ERROR] Can't parse request data")
		return
	}
	var data JSON
	err := json.Unmarshal([]byte(r.FormValue("data")), &data)
	if err != nil {
		log.Println("[ERROR] Can't parse request data")
		return
	}
	campaignName := data["filters"].([]interface{})[0].(map[string]interface{})["value"]
	classList := make([]JSON, 0)
	classList = append(classList, JSON{
		"updatedTime":       "2019-01-30T08:05:20.000+0000",
		"parentActgClassId": fmt.Sprintf("cls%s", RandStringBytes(17)),
		"name":              campaignName,
		"mergedIntoId":      "00000000000000000000",
		"entity":            "ActgClass",
		"createdTime":       "2019-01-30T08:05:20.000+0000",
		"shortName":         "",
		"id":                fmt.Sprintf("cls%s", RandStringBytes(17)),
		"isActive":          "1",
		"description":       "",
	})
	actgClassResponse := JSON{
		"response_status":  0,
		"response_message": "Success",
		"response_data":    classList,
	}
	render.JSON(w, r, actgClassResponse)
}

func main() {
	var opts Opts
	p := flags.NewParser(&opts, flags.Default)

	if _, e := p.ParseArgs(os.Args[1:]); e != nil {
		os.Exit(1)
	}

	restSrv := &Rest{
		Version: VERSION,
	}
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		restSrv.Shutdown()
	}()
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	log.Printf("[INFO] Start the mock.bill.com server http://0.0.0.0:%d", opts.Port)
	restSrv.Run("0.0.0.0", opts.Port)
}

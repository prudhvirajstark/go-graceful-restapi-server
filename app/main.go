package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
)

//  Added resources links
// https://pkg.go.dev/github.com/julienschmidt/httprouter@v1.3.0

func newRouter() *httprouter.Router {
	mux := httprouter.New()

	mux.GET("/users/profile/:username", getUserProfileHandler())

	return mux
}

func getUserProfileHandler() httprouter.Handle {

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		response := fmt.Sprintf("Fetching profile for user: %s...", p.ByName("username"))
		w.Write([]byte(response))
	}
}

func main() {

	srv := &http.Server{
		Addr:    ":10101",
		Handler: newRouter(),
	}

	idleConnesClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)
		<-sigint

		log.Println("service interrupt received")

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("http server shutting down error : %v", err)
		}

		log.Println("shutdown complete")

		close(idleConnesClosed)
	}()

	log.Println("Starting the server on port: 10101")

	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("fatal http server failed to start: %v", err)
		}
	}

	<-idleConnesClosed

	log.Println("service stopped")
}

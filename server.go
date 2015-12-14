package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-zoo/bone"
	"github.com/rs/cors"
	"github.com/rs/xhandler"
	"golang.org/x/net/context"
)

type key int

const requestIDKey key = 0

type requestIdMiddleware struct {
	next xhandler.HandlerC
}

func (h requestIdMiddleware) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx = context.WithValue(ctx, requestIDKey, r.Header.Get("X-Request-ID"))
	h.next.ServeHTTPC(ctx, w, r)
}

func loggingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func recoverMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("panic: %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func account(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	accountId := bone.GetValue(req, "id")
	reqId := ctx.Value(requestIDKey).(string)
	fmt.Fprintf(rw, "accountId: %s, Request-ID: %s", accountId, reqId)
}

func note(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	noteId := bone.GetValue(req, "id")
	reqId := ctx.Value(requestIDKey).(string)
	fmt.Fprintf(rw, "noteId: %s, Request-ID: %s", noteId, reqId)
}

func simple(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Hello, world!!")
}

func main() {
	c := xhandler.Chain{}
	c.Use(recoverMiddleware)
	c.Use(cors.Default().Handler)
	c.UseC(xhandler.CloseHandler)
	c.UseC(xhandler.TimeoutHandler(2 * time.Second))
	c.UseC(func(next xhandler.HandlerC) xhandler.HandlerC {
		return requestIdMiddleware{next: next}
	})
	c.Use(loggingMiddleware)

	simpleHandler := xhandler.HandlerFuncC(simple)
	accountHandler := xhandler.HandlerFuncC(account)
	noteHandler := xhandler.HandlerFuncC(note)

	mux := bone.New()
	mux.Get("/account/:id", c.Handler(accountHandler))
	mux.Get("/note/:id", c.Handler(noteHandler))
	mux.Get("/simple", c.Handler(simpleHandler))
	http.ListenAndServe(":8080", mux)
}

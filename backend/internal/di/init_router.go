package di

import (
	"net/http"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/router"
)

func initRouter(env string) repository.HTTPRouter {
	r := router.NewGinRouter(env)

	r.Handle(http.MethodGet, "/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	return r
}

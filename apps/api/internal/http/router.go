package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/ismobaga/apgn/apps/api/internal/http/handlers"
	"github.com/ismobaga/apgn/internal/domain/asset"
	"github.com/ismobaga/apgn/internal/domain/brief"
	"github.com/ismobaga/apgn/internal/domain/job"
	"github.com/ismobaga/apgn/internal/domain/script"
	"github.com/ismobaga/apgn/internal/domain/show"
	"github.com/ismobaga/apgn/internal/domain/source"
	episodedomain "github.com/ismobaga/apgn/internal/domain/episode"
	"github.com/ismobaga/apgn/internal/orchestrator"
)

// Repos aggregates all repository dependencies for the router.
type Repos struct {
	Shows    show.Repository
	Episodes episodedomain.Repository
	Sources  source.Repository
	Briefs   brief.Repository
	Scripts  script.Repository
	Assets   asset.Repository
	Jobs     job.Repository
}

func NewRouter(repos Repos, orch *orchestrator.Orchestrator) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	showsH := handlers.NewShowsHandler(repos.Shows)
	episodesH := handlers.NewEpisodesHandler(repos.Episodes, orch)
	sourcesH := handlers.NewSourcesHandler(repos.Sources)
	briefsH := handlers.NewBriefsHandler(repos.Briefs)
	scriptsH := handlers.NewScriptsHandler(repos.Scripts)
	assetsH := handlers.NewAssetsHandler(repos.Assets)
	jobsH := handlers.NewJobsHandler(repos.Jobs)
	publishH := handlers.NewPublishingHandler()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Shows
		r.Route("/shows", func(r chi.Router) {
			r.Get("/", showsH.List)
			r.Post("/", showsH.Create)
			r.Route("/{showID}", func(r chi.Router) {
				r.Get("/", showsH.Get)
				r.Put("/", showsH.Update)

				// Host profiles
				r.Route("/hosts", func(r chi.Router) {
					r.Get("/", showsH.ListHostProfiles)
					r.Post("/", showsH.CreateHostProfile)
					r.Get("/{hostID}", showsH.GetHostProfile)
				})

				// Episodes scoped to show
				r.Route("/episodes", func(r chi.Router) {
					r.Get("/", episodesH.List)
					r.Post("/", episodesH.Create)
				})
			})
		})

		// Episodes (top-level for cross-show queries)
		r.Get("/episodes", episodesH.List)

		// Episode detail routes
		r.Route("/episodes/{episodeID}", func(r chi.Router) {
			r.Get("/", episodesH.Get)
			r.Put("/", episodesH.Update)
			r.Post("/queue", episodesH.Queue)
			r.Post("/retry", episodesH.Retry)

			// Sources
			r.Route("/sources", func(r chi.Router) {
				r.Get("/", sourcesH.List)
				r.Post("/import-url", sourcesH.ImportURL)
				r.Post("/import-text", sourcesH.ImportText)
				r.Get("/{sourceID}", sourcesH.Get)
			})

			// Brief
			r.Route("/brief", func(r chi.Router) {
				r.Get("/", briefsH.Get)
				r.Post("/generate", briefsH.Generate)
			})

			// Scripts
			r.Route("/scripts", func(r chi.Router) {
				r.Get("/", scriptsH.List)
				r.Get("/latest", scriptsH.Get)
				r.Post("/generate", scriptsH.Generate)
				r.Post("/{draftID}/approve", scriptsH.Approve)
			})

			// Assets
			r.Route("/assets", func(r chi.Router) {
				r.Get("/", assetsH.List)
				r.Get("/{assetID}", assetsH.Get)
				r.Post("/render-audio", assetsH.RenderAudio)
				r.Post("/render-video", assetsH.RenderVideo)
			})

			// Jobs
			r.Route("/jobs", func(r chi.Router) {
				r.Get("/", jobsH.List)
				r.Get("/{jobID}", jobsH.Get)
			})

			// Publishing
			r.Post("/publish", publishH.Publish)
			r.Get("/publish/status", publishH.GetStatus)
		})
	})

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sheltonFr/bootdev/chirspy/internal/database"
	"github.com/sheltonFr/bootdev/chirspy/internal/handlers"
)

const filePathRoot = "."
const port = "8080"

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	jwtSecret      string
}

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	err = db.Ping()
	if err != nil {
		log.Fatal("Could not connect to database", err)
	}
	dbQueries := database.New(db)

	logger := log.New(os.Stdout, "chirpy-api: ", log.Flags())

	jwtSecret := os.Getenv("JWT_SECRET")

	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
		db:             dbQueries,
		jwtSecret:      jwtSecret,
	}

	userHandler := handlers.NewUserHandler(dbQueries, logger)
	chirpyHandler := handlers.NewChirpyHandler(dbQueries, logger, apiCfg.jwtSecret)
	authHandler := handlers.NewAuthHandler(dbQueries, logger, apiCfg.jwtSecret)

	mux := http.NewServeMux()

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/users", userHandler.CreateUser)

	mux.HandleFunc("POST /api/chirps", chirpyHandler.CreateChirpy)
	mux.HandleFunc("GET /api/chirps", chirpyHandler.GetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", chirpyHandler.GetChirpyById)

	//Auth
	mux.HandleFunc("POST /api/login", authHandler.LoginHandler)
	mux.HandleFunc("POST /api/refresh", authHandler.RefreshTokenHandler)
	mux.HandleFunc("POST /api/revoke", authHandler.RevokeRefreshTokenHandler)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	logger.Printf("Serving files from %s on port: %s\n", filePathRoot, port)
	logger.Fatal(srv.ListenAndServe())
}

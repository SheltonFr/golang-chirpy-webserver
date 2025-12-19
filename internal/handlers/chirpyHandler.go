package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/sheltonFr/bootdev/chirspy/internal/auth"
	"github.com/sheltonFr/bootdev/chirspy/internal/database"
	"github.com/sheltonFr/bootdev/chirspy/internal/utils"
)

type chirpyHandler struct {
	db        *database.Queries
	logger    *log.Logger
	jwtSecret string
}

type createChirpyDto struct {
	Body string `json:"body"`
}

type ChirpResponse struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func mapChirp(dbChirp database.Chirp) ChirpResponse {
	return ChirpResponse{
		ID:        dbChirp.ID,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
	}
}

func mapChirps(dbChirp []database.Chirp) []ChirpResponse {
	chirps := make([]ChirpResponse, len(dbChirp))

	for i, chirp := range dbChirp {
		chirps[i] = mapChirp(chirp)
	}

	return chirps
}

func NewChirpyHandler(
	db *database.Queries,
	logger *log.Logger,
	jwtSecret string) *chirpyHandler {
	return &chirpyHandler{db, logger, jwtSecret}
}

func (c *chirpyHandler) CreateChirpy(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	userID, err := auth.ValidateJWT(token, c.jwtSecret)
	if err != nil {
		c.logger.Printf("Unauthorized: %v", err)
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chirpyDto := createChirpyDto{}
	err = json.NewDecoder(r.Body).Decode(&chirpyDto)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if utf8.RuneCountInString(chirpyDto.Body) > 140 {
		utils.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedChirp := replaceBadWords(chirpyDto.Body)

	user, err := c.db.GetUserByID(r.Context(), userID)
	if err != nil {
		c.logger.Printf("User check failed: %v\n", err)
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	chirpyParams := database.CreateChirpyParams{
		Body:   cleanedChirp,
		UserID: user.ID,
	}

	created, err := c.db.CreateChirpy(r.Context(), chirpyParams)
	if err != nil {
		c.logger.Printf("DB error: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create chirpy")
		return
	}
	utils.RespondWithJSON(w, http.StatusCreated, mapChirp(created))
}

func (c *chirpyHandler) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := c.db.GetChirps(r.Context())
	if err != nil {
		c.logger.Printf("DB error: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not fetch chirps")
		return
	}

	response := mapChirps(chirps)
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (c *chirpyHandler) GetChirpyById(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	if chirpIDString == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Chirp ID is required")
		return
	}

	id, err := uuid.Parse(chirpIDString)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid UUID format")
		return

	}

	chirp, err := c.db.GetChirpyByID(r.Context(), id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, http.StatusNotFound, "Chirp Not Found")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not retrieve chirp")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, chirp)
}

func replaceBadWords(chirp string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	chirp = strings.ToLower(chirp)
	words := strings.Split(chirp, " ")

	for i, word := range words {
		if slices.Contains(badWords, word) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

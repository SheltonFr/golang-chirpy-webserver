package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sheltonFr/bootdev/chirspy/internal/auth"
	"github.com/sheltonFr/bootdev/chirspy/internal/database"
	"github.com/sheltonFr/bootdev/chirspy/internal/mappers"
	"github.com/sheltonFr/bootdev/chirspy/internal/utils"
)

type authHandler struct {
	db        *database.Queries
	logger    *log.Logger
	jwtSecret string
}

func NewAuthHandler(db *database.Queries, logger *log.Logger, jwtSecret string) *authHandler {
	return &authHandler{db, logger, jwtSecret}
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshJWTResponse struct {
	Token string `json:"token"`
}

func (a *authHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var loginDTO LoginDTO
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&loginDTO)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	user, err := a.db.GetUserByEmail(r.Context(), loginDTO.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		a.logger.Printf("DB error: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "An error occured")
		return
	}

	match, _ := auth.CheckPasswordHash(loginDTO.Password, user.HashedPassword)
	if !match {
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	expiry := time.Hour
	token, err := auth.MakeJWT(user.ID, a.jwtSecret, expiry)
	if err != nil {
		a.logger.Fatalf("Jwt error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "An error occured")
		return
	}
	refreshToken, _ := auth.MakeRefreshToken()
	_, err = a.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	})
	if err != nil {
		a.logger.Fatalf("Refresh Token  error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "An error occured")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, mappers.MapUserLogin(&user, token, refreshToken))
}

func (a *authHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Access toke is required")
		return
	}
	userID, err := a.validateRefreshToken(r.Context(), token)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	expiry := time.Hour
	accessToken, err := auth.MakeJWT(userID, a.jwtSecret, expiry)

	if err != nil {
		a.logger.Fatalf("Jwt Token  error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "An error occured")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, RefreshJWTResponse{accessToken})
}

func (a *authHandler) RevokeRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.db.RevokeRefreshToken(r.Context(), token)
	w.WriteHeader(http.StatusNoContent)
}

func (a *authHandler) validateRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	refreshToken, err := a.db.GetRefreshToken(ctx, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, errors.New("Token not found")
		}
		return uuid.Nil, errors.New("Unexpected error")
	}

	if refreshToken.ExpiresAt.Before(time.Now()) || refreshToken.RevokedAt.Valid {
		return uuid.Nil, errors.New("Refresh token is no longer valid")
	}

	return refreshToken.UserID, nil
}

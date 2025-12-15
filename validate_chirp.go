package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type successResponse struct {
		CleanedChirp string `json:"cleaned_chirp"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedChirp := replaceBadWords(params.Body)

	respondWithJSON(w, http.StatusOK, successResponse{CleanedChirp: cleanedChirp})
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

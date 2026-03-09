package api

import (
	"encoding/json"
	"net/http"
)

type Config struct {
	Password string
}

type signinReq struct {
	Password string `json:"password"`
}

type signinRes struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func SigninHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req signinReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, errInvalidJSON)
			return
		}

		if cfg.Password == "" {
			writeError(w, http.StatusBadRequest, "authentication disabled")
			return
		}
		if req.Password != cfg.Password {
			writeError(w, http.StatusUnauthorized, "Неверный пароль")
			return
		}

		token, err := makeJWT(cfg.Password)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "cannot generate token")
			return
		}
		writeJSON(w, http.StatusOK, signinRes{Token: token})
	}
}

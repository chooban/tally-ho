// Package auth implements a middleware providing indieauth.
//
// See the specification https://www.w3.org/TR/indieauth/.
package auth

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"hawx.me/code/indieauth"
)

// Only delegates handling the request to next only if the user specified by me
// has provided authentication as expected by IndieAuth, either:
//
//   - passing a valid token as the 'access_token' form parameter, or
//   - including a valid token in the Authorization header with a prefix of
//     'Bearer'.
func Only(me string, next http.Handler) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Finding indieauth endpoint", slog.String("me", me))
		endpoints, err := indieauth.FindEndpoints(me)
		if err != nil {
			slog.Error("find indieauth endpoints", slog.Any("err", err))
			return
		}

		auth := r.Header.Get("Authorization")
		if auth == "" || strings.TrimSpace(auth) == "Bearer" {
			if r.FormValue("access_token") == "" {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			auth = "Bearer " + r.FormValue("access_token")
		}

		req, err := http.NewRequest("GET", endpoints.Token.String(), nil)
		if err != nil {
			slog.Error("auth make request failed", slog.Any("err", err))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		req.Header.Add("Authorization", auth)
		req.Header.Add("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("auth request failed", slog.Any("err", err))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var tokenData struct {
			Me       string `json:"me"`
			ClientID string `json:"client_id"`
			Scope    string `json:"scope"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
			slog.Error("auth decode token", slog.Any("err", err))
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}

		if tokenData.Me != me {
			slog.Warn("token does not match user", slog.String("me", me), slog.String("token", tokenData.Me))
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r.WithContext(
			context.WithValue(context.WithValue(r.Context(),
				scopesKey, strings.Fields(tokenData.Scope)),
				clientKey, tokenData.ClientID,
			),
		))
	}
}

func BypassAuth(me string, next http.Handler) http.HandlerFunc {
	var tokenData struct {
		Me       string `json:"me"`
		ClientID string `json:"client_id"`
		Scope    string `json:"scope"`
	}
	tokenData.Me = me
	tokenData.ClientID = "sparkles"
	tokenData.Scope = "create"

	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(
			context.WithValue(context.WithValue(r.Context(),
				scopesKey, strings.Fields(tokenData.Scope)),
				clientKey, tokenData.ClientID,
			),
		))
	}
}

const scopesKey = "__hawx.me/code/tally-ho:Scopes__"
const clientKey = "__hawx.me/code/tally-ho:ClientID__"

// HasScope checks that a request, authenticated with Only, contains one of the
// listed valid scopes.
func HasScope(w http.ResponseWriter, r *http.Request, valid ...string) bool {
	rv := r.Context().Value(scopesKey)
	if rv == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"insufficient_scope"}`, http.StatusUnauthorized)
		return false
	}

	scopes := rv.([]string)

	hasScope := intersects(valid, scopes)
	if !hasScope {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"insufficient_scope"}`, http.StatusUnauthorized)
		return false
	}

	return true
}

// ClientID returns the clientId that was issued for the token in a request that
// has been authenticated with Only.
func ClientID(r *http.Request) string {
	rv := r.Context().Value(clientKey)
	if rv == nil {
		return ""
	}

	return rv.(string)
}

func intersects(needles []string, list []string) bool {
	for _, needle := range needles {
		for _, item := range list {
			if item == needle {
				return true
			}
		}
	}

	return false
}

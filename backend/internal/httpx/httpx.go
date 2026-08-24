package httpx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gotosky/gotosky/internal/logger"
)

type Envelope struct {
	Data any `json:"data"`
	Meta any `json:"meta,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: v})
}

func Raw(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func Error(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": APIError{Code: code, Message: msg}})
}

func Decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = strconv.FormatInt(time.Now().UnixNano(), 36)
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r)
	})
}

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.L().Error("panic", "err", rec)
				Error(w, 500, "internal_error", "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func CORS(origins []string) func(http.Handler) http.Handler {
	allow := map[string]bool{}
	for _, o := range origins {
		allow[o] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			o := r.Header.Get("Origin")
			if allow[o] {
				w.Header().Set("Access-Control-Allow-Origin", o)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(204)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type hijackRW struct {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
}

func WrapHijack(w http.ResponseWriter) http.ResponseWriter {
	h, _ := w.(http.Hijacker)
	f, _ := w.(http.Flusher)
	return &hijackRW{ResponseWriter: w, Hijacker: h, Flusher: f}
}

func SignToken(secret, user string, ttl time.Duration) string {
	exp := time.Now().Add(ttl).Unix()
	msg := user + "~" + strconv.FormatInt(exp, 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	return base64.RawURLEncoding.EncodeToString([]byte(msg)) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func VerifyToken(secret, tok string) (string, error) {
	parts := strings.Split(tok, ".")
	if len(parts) != 2 {
		return "", errors.New("bad token")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(raw)
	sig, _ := base64.RawURLEncoding.DecodeString(parts[1])
	if !hmac.Equal(mac.Sum(nil), sig) {
		return "", errors.New("bad sig")
	}
	sp := strings.Split(string(raw), "~")
	if len(sp) != 2 {
		return "", errors.New("bad payload")
	}
	exp, _ := strconv.ParseInt(sp[1], 10, 64)
	if time.Now().Unix() > exp {
		return "", errors.New("expired")
	}
	return sp[0], nil
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ferjoaguilar/rest-ws/models"
	"github.com/ferjoaguilar/rest-ws/repository"
	"github.com/ferjoaguilar/rest-ws/server"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

type UpsertPostRequest struct {
	PostContent string `json:"post_content"`
}

type PostResponse struct {
	Id          int64  `json:"id"`
	PostContent string `json:"post_content"`
}

type PostUpdateResponse struct {
	Message string `json:"message"`
}

func InsertPostHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := strings.TrimSpace(r.Header.Get("Authorization"))

		token, err := jwt.ParseWithClaims(tokenString, &models.AppClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.Config().JWTSecret), nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*models.AppClaims); ok && token.Valid {

			var postRequest = UpsertPostRequest{}
			if err := json.NewDecoder(r.Body).Decode(&postRequest); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			post := models.Post{
				PostContent: postRequest.PostContent,
				UserId:      claims.UserId,
			}

			err := repository.InsertPost(r.Context(), &post)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Websocker implementation
			var postMessage = models.WebsocketMessage{
				Type:    "Post_Created",
				Payload: post,
			}

			s.Hub().Broadcast(postMessage, nil)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(PostResponse{
				Id:          post.Id,
				PostContent: post.PostContent,
			})
		} else {
			http.Error(w, "Failed to read token", http.StatusInternalServerError)
		}
	}
}

func GetPostByIdHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		paramsInt, _ := strconv.Atoi(params["id"])

		post, err := repository.GetPostById(r.Context(), int64(paramsInt))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(post)
	}
}

func UpdatePostHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		paramsInt, _ := strconv.Atoi(params["id"])
		tokenString := strings.TrimSpace(r.Header.Get("Authorization"))

		token, err := jwt.ParseWithClaims(tokenString, &models.AppClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.Config().JWTSecret), nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*models.AppClaims); ok && token.Valid {

			var postRequest = UpsertPostRequest{}
			if err := json.NewDecoder(r.Body).Decode(&postRequest); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			post := models.Post{
				Id:          int64(paramsInt),
				PostContent: postRequest.PostContent,
				UserId:      claims.UserId,
			}

			err := repository.UpdatePost(r.Context(), &post)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(PostUpdateResponse{
				Message: "Post Updated",
			})
		} else {
			http.Error(w, "Failed to read token", http.StatusInternalServerError)
		}
	}
}

func ListPostHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		pageStr := r.URL.Query().Get("page")
		var page = uint64(0)
		if pageStr != "" {
			page, err = strconv.ParseUint(pageStr, 10, 60)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		posts, err := repository.ListPost(r.Context(), page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(posts)

	}
}

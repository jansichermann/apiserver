package apiserver

import (
	"github.com/fzzy/radix/redis"
	"encoding/json"
	"net/http"
	"fmt"
	"runtime/debug"
)

func ToJsonString(object interface{}) string {
	d, err := json.Marshal(object)
	if (err != nil) {
		panic(err)
	}
	return string(d)
}

func InvalidRequest(application Application, message string) HTTPError {
	return HTTPError{400, fmt.Sprintf("Invalid Request: %s", message)}
}

func ServerErrorHandler(application Application,
	 handlerFunc func(application Application) interface{} ) {
	defer func() {
		r := recover()
		if (r != nil) {
			if response, ok := r.(HTTPError); ok {
				http.Error(application.Writer, ToJsonString(response), int(response.Status))
			} else {
				http.Error(application.Writer, ToJsonString(HTTPResponse{500, nil}), 500)
			}
			fmt.Print(r)
			debug.PrintStack()
		}
		}()
	r := handlerFunc(application)
	response, ok := r.(HTTPResponse)
	// errorResponse, ok := response.(HTTPError)
	if ok {
		application.Writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(application.Writer, ToJsonString(response))
		
	} else if errorResponse, eok := r.(HTTPError); eok {
		http.Error(application.Writer, ToJsonString(HTTPResponse{errorResponse.Status, errorResponse.Error}), int(errorResponse.Status))
	}
}

func AuthHandler(application Application,
 authFunction func(application Application, token string) (AuthenticationUser, bool),
 handlerFunc func(application Application) interface{} ) {
	token := application.Request.FormValue("token")
	if len(token) == 0 {
		http.Error(application.Writer, ToJsonString(HTTPResponse{401, "Missing Token"}), 401)
		return
	}

	user, ok := authFunction(application, token)
	if ok {
		application.User = user
		ServerErrorHandler(application, handlerFunc)
	} else {
		http.Error(application.Writer, ToJsonString(HTTPResponse{401, "Not Authorized"}), 401)		
	}
}

type Application struct {
	RedisClient *redis.Client
	Writer http.ResponseWriter
	Request *http.Request
	User AuthenticationUser
}

type AuthenticationUser struct {
	Id string
	Name string
	Token string
}

type HTTPError struct {
	Status int16
	Error string
}

type HTTPResponse struct {
	Status int16
	Response interface {}
}
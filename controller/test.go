package controller

import (
	"evolve/util"
	"log"
	"net/http"
)

// Test is a test API for checking the server status.
func Test(res http.ResponseWriter, req *http.Request) {
	log.Println("[INFO]: Test API called from", req.RemoteAddr)

	switch req.Method {
	case "GET":
		util.JSONResponse(res, http.StatusOK, "It works! 👍🏻", nil)
	case "POST":
		data, err := util.Body(req)
		if err != nil {
			util.JSONResponse(res, http.StatusBadRequest, err.Error(), nil)
			return
		}

		util.JSONResponse(res, http.StatusOK, "It works! 👍🏻", data)
	default:
		util.JSONResponse(res, http.StatusMethodNotAllowed, "Method not allowed", nil)
	}

}

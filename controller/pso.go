package controller

import (
	"evolve/modules"
	"evolve/util"
	"fmt"
	"net/http"
	"os"
)

func CreatePSO(res http.ResponseWriter, req *http.Request) {
	var logger = util.NewLogger()
	logger.Info("CreatePSO API called.")

	// Comment this out to test the API without authentication.
	user, err := modules.Auth(req)
	if err != nil {
		util.JSONResponse(res, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	// User has id, role, userName, email & fullName.
	logger.Info(fmt.Sprintf("User: %s", user))

	data, err := util.Body(req)
	if err != nil {
		util.JSONResponse(res, http.StatusBadRequest, err.Error(), nil)
		return
	}

	pso, err := modules.PSOFromJSON(data)
	if err != nil {
		util.JSONResponse(res, http.StatusBadRequest, err.Error(), nil)
		return
	}

	code, err := pso.Code()
	if err != nil {
		util.JSONResponse(res, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// TODO: Write code for MinIO and remove this.
	os.Mkdir("code", 0755)
	os.WriteFile("code/pso.py", []byte(code), 0644)

	util.JSONResponse(res, http.StatusOK, "It works! 👍🏻", data)
}

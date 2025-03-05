package controller

import (
	"encoding/json"
	"evolve/db/connection"
	"evolve/modules"
	"evolve/util"
	"fmt"
	"net/http"
	"os"
	"time"
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

	db, err := connection.PoolConn(req.Context())
	if err != nil {
		logger.Error(fmt.Sprintf("CreatePSO: %s", err.Error()))
		util.JSONResponse(res, http.StatusInternalServerError, "something went wrong", nil)
		return
	}

	row := db.QueryRow(req.Context(), `
		INSERT INTO run (name, description, type, command, createdBy)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, time.Now().Local().String(), "Particle Swarm Optimization", "pso", "python code.py", user["id"])

	var runID string
	err = row.Scan(&runID)

	if err != nil {
		logger.Error(fmt.Sprintf("CreatePSO.row.Scan: %s", err.Error()))
		util.JSONResponse(res, http.StatusInternalServerError, "something went wrong", nil)
		return
	}

	logger.Info(fmt.Sprintf("RunID: %s", runID))

	_, err = db.Exec(req.Context(), `
		INSERT INTO access (runID, userID, mode)
		VALUES ($1, $2, $3)
	`, runID, user["id"], "write")

	if err != nil {
		logger.Error(fmt.Sprintf("CreatePSO.db.Exec: %s", err.Error()))
		util.JSONResponse(res, http.StatusInternalServerError, "something went wrong", nil)
		return
	}

	inputParams, err := json.Marshal(data)
	if err != nil {
		logger.Error(fmt.Sprintf("CreatePSO.json.Marshal: %s", err.Error()))
		util.JSONResponse(res, http.StatusInternalServerError, "something went wrong", nil)
		return
	}

	// Save code and upload to minIO.
	os.Mkdir("code", 0755)
	if err := os.WriteFile(fmt.Sprintf("code/%v.py", runID), []byte(code), 0644); err != nil {
		logger.Error(fmt.Sprintf("CreatePSO.os.WriteFile: %s", err.Error()))
		util.JSONResponse(res, http.StatusInternalServerError, "something went wrong", nil)
		return
	}
	if err := util.UploadFile(req.Context(), runID, "code", "py"); err != nil {
		util.JSONResponse(res, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Save input and upload to minIO.
	os.Mkdir("input", 0755)
	if err := os.WriteFile(fmt.Sprintf("input/%v.json", runID), inputParams, 0644); err != nil {
		logger.Error(fmt.Sprintf("CreatePSO.os.WriteFile: %s", err.Error()))
		util.JSONResponse(res, http.StatusInternalServerError, "something went wrong", nil)
		return
	}
	if err := util.UploadFile(req.Context(), runID, "input", "json"); err != nil {
		util.JSONResponse(res, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Remove code and input files from local.
	if err := os.Remove(fmt.Sprintf("code/%v.py", runID)); err != nil {
		logger.Error(fmt.Sprintf("CreatePSO.os.Remove: %s", err.Error()))
		util.JSONResponse(res, http.StatusInternalServerError, "something went wrong", nil)
		return
	}
	if err := os.Remove(fmt.Sprintf("input/%v.json", runID)); err != nil {
		logger.Error(fmt.Sprintf("CreatePSO.os.Remove: %s", err.Error()))
		util.JSONResponse(res, http.StatusInternalServerError, "something went wrong", nil)
		return
	}

	if err := util.EnqueueRunRequest(req.Context(), runID, "code", "py"); err != nil {
		util.JSONResponse(res, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	util.JSONResponse(res, http.StatusOK, "It works! 👍🏻", data)
}

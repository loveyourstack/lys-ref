package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procstep"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
)

func (srvApp *httpServerApplication) procGetStepAvailableDependencies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	// get stepId and ensure it is an int
	stepId, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		lys.HandleUserError(lys.ErrIdNotAnInteger, w)
		return
	}

	stepStore := procstep.Store{Db: srvApp.Db}
	items, err := stepStore.SelectAvailableDependencies(ctx, stepId)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("stepStore.SelectAvailableDependencies failed: %w", err), srvApp.Logger, w)
		return
	}

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   items,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) procRunStep(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	// get stepId and ensure it is an int
	stepId, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		lys.HandleUserError(lys.ErrIdNotAnInteger, w)
		return
	}

	type Input struct {
		ParamString      string `json:"param_string" validate:"required"`
		StopOnError      bool   `json:"stop_on_error"`
		WithDependencies bool   `json:"with_dependencies"`
	}

	// get req body
	body, err := lys.ExtractJsonBody(r, srvApp.PostOptions.MaxBodySize)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.ExtractJsonBody failed: %w", err), srvApp.Logger, w)
		return
	}

	// marshal req body into an Input
	input, err := lys.DecodeJsonBody[Input](body)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.DecodeJsonBody failed: %w", err), srvApp.Logger, w)
		return
	}

	// validate Input
	if err = lysmeta.Validate(srvApp.Validate, input); err != nil {
		lys.HandleUserError(lyserr.User{Message: err.Error()}, w)
		return
	}

	// create new run
	runId, err := srvApp.ProcSvc.CreateRunFromStep(ctx, srvApp.Db, stepId, strings.Split(input.ParamString, " "), input.WithDependencies)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.ProcSvc.CreateRunFromStep failed: %w", err), srvApp.Logger, w)
		return
	}

	// run step in async process
	go func(runId int64) {

		// detach context so svc call is not affected by server context timeout, but retains req context values
		bgCtx := context.WithoutCancel(ctx)

		// evaluate which run func to use
		runFunc := srvApp.ProcSvc.RunOnly
		fName := "srvApp.ProcSvc.RunOnly"

		if input.WithDependencies {
			runFunc = srvApp.ProcSvc.RunWithDeps
			fName = "srvApp.ProcSvc.RunWithDeps"

			if input.StopOnError {
				runFunc = srvApp.ProcSvc.MustRunWithDeps
				fName = "srvApp.ProcSvc.MustRunWithDeps"
			}
		}

		err := runFunc(bgCtx, srvApp.Db, runId)
		if err != nil {

			// log error for now (don't use lys.HandleError since it needs the ResponseWriter, which is unavailable in this async context)
			// but note that it is also written into the point's err_msg column
			srvApp.Logger.Error(fmt.Sprintf("%s failed: %v", fName, err), "runId", runId)
			return
		}
	}(runId)

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   runId,
	}
	lys.JsonResponse(resp, http.StatusAccepted, w)
}

func (srvApp *httpServerApplication) procSwapDisplayOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type Input struct {
		StepId1 int64 `json:"step_id1" validate:"required"`
		StepId2 int64 `json:"step_id2" validate:"required"`
	}

	// get req body
	body, err := lys.ExtractJsonBody(r, srvApp.PostOptions.MaxBodySize)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.ExtractJsonBody failed: %w", err), srvApp.Logger, w)
		return
	}

	// marshal req body into an Input
	input, err := lys.DecodeJsonBody[Input](body)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.DecodeJsonBody failed: %w", err), srvApp.Logger, w)
		return
	}

	// validate Input
	if err = lysmeta.Validate(srvApp.Validate, input); err != nil {
		lys.HandleUserError(lyserr.User{Message: err.Error()}, w)
		return
	}

	// swap display order of steps
	stepStore := procstep.Store{Db: srvApp.Db}
	err = stepStore.SwapDisplayOrder(ctx, input.StepId1, input.StepId2)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("stepStore.SwapDisplayOrder failed: %w", err), srvApp.Logger, w)
		return
	}

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   "success",
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

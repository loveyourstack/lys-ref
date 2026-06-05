package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaign"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaignopt"
	"github.com/loveyourstack/lys/lysset"
)

func (srvApp *httpServerApplication) dmGetCampaignOptAggregates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get request modifiers from url params
	campOptStore := dmcampaignopt.Store{Db: srvApp.Db}
	params := lys.ExtractGetRequestModifierParams{
		AdditionalFilterParamNames: nil,
		DbNames:                    lysset.FromSlice(campOptStore.GetPlan().DbNames()),
		GetOptions:                 srvApp.GetOptions,
		JsonKeyDbNameMap:           campOptStore.GetPlan().JsonKeyDbNameMap(),
		SetFuncUrlParamNames:       campOptStore.GetSetFuncUrlParamNames(),
	}
	getReqModifiers, err := lys.ExtractGetRequestModifiers(r, params)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.ExtractGetRequestModifiers failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	aggItems, err := campOptStore.SelectAggregates(ctx, getReqModifiers.SetFuncParamValues, getReqModifiers.Conditions)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("campOptStore.SelectAggregates failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// return success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   aggItems,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) dmPatchActiveByIds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get req body
	body, err := lys.ExtractJsonBody(r, srvApp.PostOptions.MaxBodySize)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.ExtractJsonBody failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	type input struct {
		Ids       []int64 `json:"ids"`
		NewActive bool    `json:"new_active"`
	}

	// marshal req body into inputs
	inp, err := lys.DecodeJsonBody[input](body)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.DecodeJsonBody failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// begin tx
	tx, err := srvApp.Db.Begin(ctx)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("db.Begin failed: %w", err), srvApp.ErrorLog, w)
		return
	}
	defer tx.Rollback(ctx)

	for _, id := range inp.Ids {
		assMap := map[string]any{
			"is_active": inp.NewActive,
		}
		err = dmcampaign.UpdatePartialTx(ctx, tx, assMap, id)
		if err != nil {
			lys.HandleError(ctx, fmt.Errorf("dmcampaign.UpdatePartialTx failed on id: %d: %w", id, err), srvApp.ErrorLog, w)
			return
		}
	}

	// success: commit tx
	err = tx.Commit(ctx)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("tx.Commit failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// return success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   fmt.Sprintf("Updated %d campaign(s)", len(inp.Ids)),
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) dmPatchBudgetPercentByIds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get req body
	body, err := lys.ExtractJsonBody(r, srvApp.PostOptions.MaxBodySize)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.ExtractJsonBody failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	type input struct {
		Ids                 []int64 `json:"ids"`
		BudgetPercentChange float64 `json:"budget_percent_change"`
	}

	// marshal req body into inputs
	inp, err := lys.DecodeJsonBody[input](body)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.DecodeJsonBody failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// select map of k= campId, v = dailyBudgetEur
	campStore := dmcampaign.Store{Db: srvApp.Db}
	campIdBudgetMap, err := campStore.SelectIdBudgetMap(ctx, inp.Ids)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("campStore.SelectIdBudgetMap failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// begin tx
	tx, err := srvApp.Db.Begin(ctx)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("db.Begin failed: %w", err), srvApp.ErrorLog, w)
		return
	}
	defer tx.Rollback(ctx)

	for _, id := range inp.Ids {

		// calculate new budget based on percent change, format to 2 decimal places
		newBudget := campIdBudgetMap[id] * (1 + inp.BudgetPercentChange/100)
		assMap := map[string]any{
			"daily_budget_eur": strconv.FormatFloat(newBudget, 'f', 2, 64),
		}
		err = dmcampaign.UpdatePartialTx(ctx, tx, assMap, id)
		if err != nil {
			lys.HandleError(ctx, fmt.Errorf("dmcampaign.UpdatePartialTx failed on id: %d: %w", id, err), srvApp.ErrorLog, w)
			return
		}
	}

	// success: commit tx
	err = tx.Commit(ctx)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("tx.Commit failed: %w", err), srvApp.ErrorLog, w)
		return
	}

	// return success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   fmt.Sprintf("Updated %d campaign(s)", len(inp.Ids)),
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

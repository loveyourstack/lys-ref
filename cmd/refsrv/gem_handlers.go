package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
)

func (srvApp *httpServerApplication) gemGenerateImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get req body
	body, err := lys.ExtractJsonBody(r, srvApp.PostOptions.MaxBodySize)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.ExtractJsonBody failed: %w", err), srvApp.Logger, w)
		return
	}

	type input struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
	}

	// unmarshal req body
	inp, err := lys.DecodeJsonBody[input](body)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.DecodeJsonBody failed: %w", err), srvApp.Logger, w)
		return
	}

	// only allow model gemini-3.1-flash-lite-image at the moment
	if inp.Model != "gemini-3.1-flash-lite-image" {
		lys.HandleError(ctx, fmt.Errorf("unsupported model: %s", inp.Model), srvApp.Logger, w)
		return
	}

	// call Gemini API to generate image
	fName, err := srvApp.GeminiClient.GenerateImage(ctx, inp.Model, inp.Prompt)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.GeminiClient.GenerateImage failed: %w", err), srvApp.Logger, w)
		return
	}

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   fName,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) gemGenerateMarketingCampaign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get req body
	body, err := lys.ExtractJsonBody(r, srvApp.PostOptions.MaxBodySize)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.ExtractJsonBody failed: %w", err), srvApp.Logger, w)
		return
	}

	type input struct {
		Model   string `json:"model" validate:"required"`
		Product string `json:"product" validate:"required,max=100"`
	}

	// unmarshal req body
	inp, err := lys.DecodeJsonBody[input](body)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.DecodeJsonBody failed: %w", err), srvApp.Logger, w)
		return
	}

	inp.Product = strings.TrimSpace(inp.Product)

	// validate input
	if err = lysmeta.Validate(srvApp.Validate, inp); err != nil {
		lys.HandleUserError(lyserr.User{Message: err.Error()}, w)
		return
	}

	// only allow model gemini-3.1-flash-lite at the moment
	if inp.Model != "gemini-3.1-flash-lite" {
		lys.HandleError(ctx, fmt.Errorf("unsupported model: %s", inp.Model), srvApp.Logger, w)
		return
	}

	// call Gemini API to generate marketing campaign
	camp, err := srvApp.GeminiClient.GenerateMarketingCampaign(ctx, inp.Model, inp.Product)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.GeminiClient.GenerateMarketingCampaign failed: %w", err), srvApp.Logger, w)
		return
	}

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   camp,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) gemGenerateText(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get req body
	body, err := lys.ExtractJsonBody(r, srvApp.PostOptions.MaxBodySize)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.ExtractJsonBody failed: %w", err), srvApp.Logger, w)
		return
	}

	type input struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
	}

	// unmarshal req body
	inp, err := lys.DecodeJsonBody[input](body)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.DecodeJsonBody failed: %w", err), srvApp.Logger, w)
		return
	}

	// only allow model gemini-3.1-flash-lite at the moment
	if inp.Model != "gemini-3.1-flash-lite" {
		lys.HandleError(ctx, fmt.Errorf("unsupported model: %s", inp.Model), srvApp.Logger, w)
		return
	}

	// call Gemini API to generate text
	txt, err := srvApp.GeminiClient.GenerateText(ctx, inp.Model, inp.Prompt)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.GeminiClient.GenerateText failed: %w", err), srvApp.Logger, w)
		return
	}

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   txt,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) gemListModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	models, err := srvApp.GeminiClient.ListModels(ctx)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.GeminiClient.ListModels failed: %w", err), srvApp.Logger, w)
		return
	}

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   models,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

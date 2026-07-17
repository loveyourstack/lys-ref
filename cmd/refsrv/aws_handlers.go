package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/templates"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmail"
	"github.com/loveyourstack/lys/lysmeta"
)

func (srvApp *httpServerApplication) awsBedrockGenerateMarketingImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get req body
	body, err := lys.ExtractJsonBody(r, srvApp.PostOptions.MaxBodySize)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("lys.ExtractJsonBody failed: %w", err), srvApp.Logger, w)
		return
	}

	type input struct {
		Product string `json:"product" validate:"required,max=1000"`
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

	fName, err := srvApp.AwsBedrockClient.GenerateMarketingImage(ctx, inp.Product)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.AwsBedrockClient.GenerateMarketingImage failed: %w", err), srvApp.Logger, w)
		return
	}

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   fName,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) awsBedrockListImageModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	models, err := srvApp.AwsBedrockClient.ListImageModels(ctx)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.AwsBedrockClient.ListImageModels failed: %w", err), srvApp.Logger, w)
		return
	}

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   models,
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func (srvApp *httpServerApplication) awsUpdateUserSecurityGroupRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get session associated with this request, if any, and validate it
	sess, err := srvApp.Sessions.FromRequest(r, srvApp.Logger)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.Sessions.FromRequest failed: %w", err), srvApp.Logger, w)
		return
	}
	if sess.Email == "" {
		lys.HandleUserError(lyserr.User{Message: "email missing in session"}, w)
		return
	}

	// user IP is only correct in prod, so exit if not in prod
	if srvApp.Config.General.Env != appenv.Prod {
		lys.HandleUserError(lyserr.User{Message: "only call this endpoint in prod env"}, w)
		return
	}

	// for dev testing: comment out the prod env check, and comment in the IP hardcoding
	// don't forget to change it back
	//sess.Ip, _ = netip.ParseAddr("79.224.142.120")

	// update firewall rules for this user
	err = srvApp.AwsSvc.UpdateUserSecurityGroupRules(ctx, sess.UserName, sess.Ip)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("srvApp.AwsSvc.UpdateUserSecurityGroupRules failed: %w", err), srvApp.Logger, w)
		return
	}

	// send notification mail to user
	err = sendAwsSgRuleChangeNotificationEmail(srvApp.Config.Smtp, sess.Email)
	if err != nil {
		lys.HandleError(ctx, fmt.Errorf("sendAwsSgRuleChangeNotificationEmail failed: %w", err), srvApp.Logger, w)
		// don't return: update succeeded, just log mail failure
	}

	// success
	resp := lys.StdResponse{
		Status: lys.ReqSucceeded,
		Data:   "updated AWS firewall",
	}
	lys.JsonResponse(resp, http.StatusOK, w)
}

func sendAwsSgRuleChangeNotificationEmail(smtp lysmail.SmtpConfig, email string) (err error) {

	// read email template
	tmplName := "aws_sg_rules_updated.html"
	tmplB, err := fs.ReadFile(templates.Templates, tmplName)
	if err != nil {
		return fmt.Errorf("fs.ReadFile failed for tmpl: %s: %w", tmplName, err)
	}

	emailSubject := "LysRef: AWS Firewall Rules Updated"
	emailBody := string(tmplB)

	// send email
	err = smtp.Send([]string{email}, nil, emailSubject, emailBody)
	if err != nil {
		return fmt.Errorf("smtp.Send failed: %w", err)
	}

	return nil
}

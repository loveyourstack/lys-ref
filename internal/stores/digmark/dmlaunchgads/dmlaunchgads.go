package dmlaunchgads

import (
	"context"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/launchstatus"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunch"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	name           string = "Digmark campaign launchers: Google Ads"
	schemaName     string = "digmark"
	tableName      string = "launcher_gads"
	viewName       string = "v_launcher_gads"
	pkColName      string = "id"
	defaultOrderBy string = "created_at DESC"
)

type Input struct {
	dmlaunch.Input
}

type Computed struct {
	GAdsAccountId  int64 `db:"gads_account_id" json:"gads_account_id,omitempty"`   // set during preparation
	GAdsAdId       int64 `db:"gads_ad_id" json:"gads_ad_id,omitempty"`             // set during processing
	GAdsAdGroupId  int64 `db:"gads_ad_group_id" json:"gads_ad_group_id,omitempty"` // set during processing
	GAdsCampaignId int64 `db:"gads_campaign_id" json:"gads_campaign_id,omitempty"` // set during processing
	dmlaunch.Computed
}

type Model struct {
	Input
	Computed
	dmlaunch.DbManaged
}

var (
	plan, compPlan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.Analyze(Model{})
	if err != nil {
		log.Fatalf("lysmeta.Analyze failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
	compPlan, _ = lysmeta.Analyze(Computed{})
}

type Store struct {
	Db *pgxpool.Pool
}

func (s Store) CancelMany(ctx context.Context, ids []int64) (numDeleted int64, err error) {
	return dmlaunch.CancelMany(ctx, s.Db, schemaName, tableName, ids)
}

// ClaimForPreparation selects a batch of unprepared items and sets their status to Preparing, returning the items.
func (s Store) ClaimForPreparation(ctx context.Context, batchSize int) (items []Model, err error) {
	return dmlaunch.ClaimForPreparation[Model](ctx, s.Db, schemaName, tableName, batchSize)
}

// ClaimNextForProcessing selects the next queued item and sets its status to Processing, returning the item.
func (s Store) ClaimNextForProcessing(ctx context.Context) (item Model, err error) {
	return dmlaunch.ClaimNextForProcessing[Model](ctx, s.Db, schemaName, tableName)
}

func (s Store) Delete(ctx context.Context, id int64) error {
	return dmlaunch.DeleteEditableByID(ctx, s.Db, schemaName, tableName, viewName, pkColName, id)
}

func (s Store) DeleteMany(ctx context.Context, ids []int64) (numDeleted int64, err error) {
	return dmlaunch.DeleteMany(ctx, s.Db, schemaName, tableName, ids)
}

func (s Store) GetName() string {
	return name
}
func (s Store) GetPlan() lysmeta.Plan {
	return plan
}

func (s Store) Insert(ctx context.Context, input Input) (newId int64, err error) {
	return lyspg.Insert[Input, int64](ctx, s.Db, schemaName, tableName, pkColName, input)
}

func (s Store) InsertTx(ctx context.Context, tx pgx.Tx, input Input) (newId int64, err error) {
	return lyspg.Insert[Input, int64](ctx, tx, schemaName, tableName, pkColName, input)
}

func (s Store) QueueMany(ctx context.Context, ids []int64) (numDeleted int64, err error) {
	return dmlaunch.QueueMany(ctx, s.Db, schemaName, tableName, ids)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}

func (s Store) SetAdId(ctx context.Context, adId int64, id int64) (err error) {
	assignmentsMap := map[string]any{
		"gads_ad_id": adId,
		"step":       3,
	}
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, compPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) SetAdGroupId(ctx context.Context, adGroupId int64, id int64) (err error) {
	assignmentsMap := map[string]any{
		"gads_ad_group_id": adGroupId,
		"step":             2,
	}
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, compPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) SetCampaignId(ctx context.Context, campaignId int64, id int64) (err error) {
	assignmentsMap := map[string]any{
		"gads_campaign_id": campaignId,
		"step":             1,
	}
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, compPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) SetFailed(ctx context.Context, msg string, id int64) (err error) {
	assignmentsMap := map[string]any{
		"message": msg,
		"status":  launchstatus.Failed,
	}
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, compPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) SetInvalid(ctx context.Context, msg string, id int64) (err error) {
	assignmentsMap := map[string]any{
		"country_fk":  -1,
		"message":     msg,
		"status":      launchstatus.Invalid,
		"vertical_fk": -1,

		"gads_account_id": 0,
	}
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, compPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) SetReady(ctx context.Context, computed Computed, id int64) (err error) {
	assignmentsMap := map[string]any{
		"country_fk":  computed.CountryFk,
		"message":     "",
		"status":      launchstatus.Ready,
		"vertical_fk": computed.VerticalFk,

		"gads_account_id": computed.GAdsAccountId,
	}
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, compPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) SetStatus(ctx context.Context, status launchstatus.Enum, id int64) (err error) {
	assignmentsMap := map[string]any{
		"status": status,
	}
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, compPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) Update(ctx context.Context, input Input, id int64) (err error) {
	return dmlaunch.Update(ctx, s.Db, schemaName, tableName, viewName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (err error) {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, compPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) Validate(validate *validator.Validate, input Input) error {
	return lysmeta.Validate(validate, input)
}

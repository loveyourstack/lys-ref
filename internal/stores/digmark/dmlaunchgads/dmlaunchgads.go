package dmlaunchgads

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/launchstatus"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunch"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	name           string = "Digmark campaign launchers: Google Ads"
	schemaName     string = "digmark"
	tableName      string = "launcher_gads"
	viewName       string = "v_launcher_gads"
	pkColName      string = "id"
	defaultOrderBy string = "name"
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

func (s Store) Delete(ctx context.Context, id int64) error {

	// select item from id
	item, err := s.SelectById(ctx, id)
	if err != nil {
		return fmt.Errorf("s.SelectById failed: %w", err)
	}

	// reject if status does not allow deletion
	if !slices.Contains(launchstatus.Editable[:], item.Status) {
		return lyserr.User{Message: fmt.Sprintf("deletion not allowed for status: %s", item.Status)}
	}

	return lyspg.DeleteUnique(ctx, s.Db, schemaName, tableName, pkColName, id)
}

func (s Store) DeleteMany(ctx context.Context, ids []int64) (numDeleted int64, err error) {

	// select items from ids
	items, err := lyspg.SelectT[Model](ctx, s.Db, fmt.Sprintf("SELECT id, status FROM %s.%s WHERE id = ANY($1);", schemaName, tableName), ids)
	if err != nil {
		return 0, fmt.Errorf("lyspg.SelectT failed: %w", err)
	}

	var deleteableIds []int64

	// only allow deletion of records where status allows it, ignore others
	for _, item := range items {
		if slices.Contains(launchstatus.Editable[:], item.Status) {
			deleteableIds = append(deleteableIds, item.Id)
		}
	}

	// delete
	stmt := fmt.Sprintf("DELETE FROM %s.%s WHERE id = ANY($1);", schemaName, tableName)
	res, err := s.Db.Exec(ctx, stmt, deleteableIds)
	if err != nil {
		return 0, lyserr.Db{Err: fmt.Errorf("s.Db.Exec failed: %w", err), Stmt: stmt}
	}

	return res.RowsAffected(), nil
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

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

func (s Store) SelectUncheckedTx(ctx context.Context, tx pgx.Tx) (items []Model, err error) {
	stmt := fmt.Sprintf(`SELECT * FROM %s.%s WHERE status = 'Unchecked' ORDER BY id FOR UPDATE SKIP LOCKED;`, schemaName, tableName)
	return lyspg.SelectT[Model](ctx, tx, stmt)
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

func (s Store) SetPreparedTx(ctx context.Context, tx pgx.Tx, computed Computed, id int64) (err error) {
	assignmentsMap := map[string]any{
		"country_fk":  computed.CountryFk,
		"message":     "",
		"status":      launchstatus.Ready,
		"vertical_fk": computed.VerticalFk,

		"gads_account_id": computed.GAdsAccountId,
	}
	return lyspg.UpdatePartial(ctx, tx, schemaName, tableName, pkColName, compPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) SetUnpreparedTx(ctx context.Context, tx pgx.Tx, msg string, id int64) (err error) {
	assignmentsMap := map[string]any{
		"country_fk":  -1,
		"message":     msg,
		"status":      launchstatus.Invalid,
		"vertical_fk": -1,

		"gads_account_id": 0,
	}
	return lyspg.UpdatePartial(ctx, tx, schemaName, tableName, pkColName, compPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) Update(ctx context.Context, input Input, id int64) (err error) {
	return lyspg.Update(ctx, s.Db, schemaName, tableName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (err error) {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, compPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) Validate(validate *validator.Validate, input Input) error {
	return lysmeta.Validate(validate, input)
}

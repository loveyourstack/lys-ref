package sysnotification

import (
	"context"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "System notifications"
	schemaName     string = "system"
	tableName      string = "notification"
	viewName       string = "notification"
	pkColName      string = "id"
	defaultOrderBy string = "created_at DESC"
)

type Input struct {
	IsRead  bool   `db:"is_read" json:"is_read,omitempty"`
	Message string `db:"message" json:"message,omitempty" validate:"required,max=1024"`
	NotType string `db:"not_type" json:"not_type,omitempty" validate:"required,max=64"`
	UserFk  int64  `db:"user_fk" json:"user_fk,omitempty" validate:"required"`
}

type Model struct {
	Id        int64            `db:"id" json:"id,omitempty"`
	CreatedAt lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	UpdatedAt lystype.Datetime `db:"updated_at" json:"updated_at,omitzero"` // assigned by trigger
	Input
}

var (
	plan, inputPlan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.Analyze(Model{})
	if err != nil {
		log.Fatalf("lysmeta.Analyze failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
	inputPlan, _ = lysmeta.Analyze(Input{})
}

type Store struct {
	Db *pgxpool.Pool
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

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

// SelectOnlyUsers is a wrapper for Select that appends a filter to the provided params to only select notifications for the authenticated user.
func (s Store) SelectOnlyUsers(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {

	// get user from ctx
	userId := lys.GetUserIdFromCtx(ctx)
	if userId == 0 {
		return nil, lyspg.TotalCount{}, fmt.Errorf("lys.GetUserIdFromCtx returned 0 - user not authenticated")
	}

	// append mandatory user id filter to params
	params.Conditions = append(params.Conditions, lyspg.Condition{
		Field:    "user_fk",
		Operator: lyspg.OpEquals,
		Value:    fmt.Sprintf("%d", userId),
	})

	return s.Select(ctx, params)
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}

// SelectDetailsById looks up the notification by ID and returns the details needed for broadcasting via WebSocket.
func SelectDetailsById(ctx context.Context, db *pgxpool.Pool, id int64) (userId int64, notType, message string, err error) {

	item, err := lyspg.SelectUnique[Model](ctx, db, schemaName, viewName, pkColName, id)
	if err != nil {
		return 0, "", "", fmt.Errorf("lyspg.SelectUnique failed: %w", err)
	}
	return item.UserFk, item.NotType, item.Message, nil
}

func (s Store) SelectUserUnreadCount(ctx context.Context) (count int64, err error) {

	userId := lys.GetUserIdFromCtx(ctx)
	if userId == 0 {
		return 0, fmt.Errorf("lys.GetUserIdFromCtx returned 0 - user not authenticated")
	}

	stmt := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s WHERE user_fk = $1 AND is_read = false;", schemaName, tableName)
	err = s.Db.QueryRow(ctx, stmt, userId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("s.Db.QueryRow.Scan failed: %w", err)
	}
	return count, nil
}

// SetUsersToRead sets the notifications with the provided ids to read, but only if they belong to the authenticated user.
func (s Store) SetUsersToRead(ctx context.Context, ids []int64) (err error) {

	// get user from ctx
	userId := lys.GetUserIdFromCtx(ctx)
	if userId == 0 {
		return fmt.Errorf("lys.GetUserIdFromCtx returned 0 - user not authenticated")
	}

	// update, but only where the notifications belong to the ctx user
	stmt := fmt.Sprintf("UPDATE %s.%s SET is_read = true WHERE user_fk = $1 AND id = ANY($2);", schemaName, tableName)
	_, err = s.Db.Exec(ctx, stmt, userId, ids)
	if err != nil {
		return fmt.Errorf("s.Db.Exec failed: %w", err)
	}
	return nil
}

// SetAllUsersToRead sets all notifications to read for the authenticated user.
func (s Store) SetAllUsersToRead(ctx context.Context) (err error) {

	// get user from ctx
	userId := lys.GetUserIdFromCtx(ctx)
	if userId == 0 {
		return fmt.Errorf("lys.GetUserIdFromCtx returned 0 - user not authenticated")
	}

	// update, but only where the notifications belong to the ctx user
	stmt := fmt.Sprintf("UPDATE %s.%s SET is_read = true WHERE user_fk = $1 AND is_read = false;", schemaName, tableName)
	_, err = s.Db.Exec(ctx, stmt, userId)
	if err != nil {
		return fmt.Errorf("s.Db.Exec failed: %w", err)
	}
	return nil
}

func (s Store) Update(ctx context.Context, input Input, id int64) (err error) {
	return lyspg.Update(ctx, s.Db, schemaName, tableName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (err error) {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, inputPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) Validate(validate *validator.Validate, input Input) error {
	return lysmeta.Validate(validate, input)
}

package lysinceventsinmonth

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

/*
	uses PG setFunc, not view
*/

const (
	name           string = "Lys Inc employees"
	schemaName     string = "lysinc"
	setFuncName    string = "f_events_in_month"
	defaultOrderBy string = "full_name"
)

type Model struct {
	Department string       `db:"department" json:"department,omitzero"`
	EmployeeId int64        `db:"employee_id" json:"employee_id,omitzero"`
	EventDate  lystype.Date `db:"event_date" json:"event_date,omitzero"`
	EventType  string       `db:"event_type" json:"event_type,omitzero"`
	FullName   string       `db:"full_name" json:"full_name,omitzero"`
	JobTitle   string       `db:"job_title" json:"job_title,omitzero"`
	Message    string       `db:"-" json:"message,omitzero"` // added in Select
	Sex        string       `db:"sex" json:"sex,omitzero"`
	Years      int          `db:"years" json:"years,omitzero"`
}

var (
	plan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.Analyze(Model{})
	if err != nil {
		log.Fatalf("lysmeta.Analyze failed for %s.%s: %s", schemaName, setFuncName, err.Error())
	}
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
func (s Store) GetSetFuncUrlParamNames() []string {
	return []string{"base_date"}
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {

	// validate base date sent via mandatory API param
	baseDateStr := fmt.Sprintf("%s", params.SetFuncParamValues[0])
	if _, err := time.Parse(lystype.DateFormat, baseDateStr); err != nil {
		return nil, lyspg.TotalCount{}, lyserr.User{Message: fmt.Sprintf("invalid base_date value: %s", baseDateStr)}
	}

	// there is no table: don't try to get unpaged count
	params.GetUnpagedCount = false

	items, unpagedCount, err = lyspg.Select[Model](ctx, s.Db, schemaName, "", setFuncName, defaultOrderBy, plan.DbNames(), params)
	if err != nil {
		return nil, lyspg.TotalCount{}, fmt.Errorf("lyspg.Select failed: %w", err)
	}

	// add display message to each item
	for i := range items {
		switch items[i].EventType {
		case "Anniversary":
			pronoun := "his"
			if items[i].Sex == "Female" {
				pronoun = "her"
			}

			if items[i].Years == 1 {
				items[i].Message = fmt.Sprintf("%s celebrates %s first year with the company!", items[i].FullName, pronoun)
			} else {
				items[i].Message = fmt.Sprintf("%s celebrates %d years with the company!", items[i].FullName, items[i].Years)
			}
		case "Birthday":
			items[i].Message = fmt.Sprintf("Happy Birthday, %s!", items[i].FullName)
		default:
			items[i].Message = fmt.Sprintf("Congratulations, %s!", items[i].FullName)
		}
	}

	return items, unpagedCount, nil
}

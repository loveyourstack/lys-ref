package syssvc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmvertical"
	"github.com/loveyourstack/lys-ref/internal/stores/ecb/ecbcurr"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geocountry"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geoocean"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procflow"
	"github.com/loveyourstack/lys-ref/internal/stores/publisher/pubauthor"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppcompany"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppprodcategory"
	"github.com/loveyourstack/lys/lyspg"
)

type UiStoreData struct {
	CoreMandatoryEnums []string `json:"core_mandatory_enums"`
	CoreOptionalEnums  []string `json:"core_optional_enums"`
	CorePeriods        []string `json:"core_periods"`

	DigmarkLauncherStati []string           `json:"digmark_launcher_stati"`
	DigmarkManagers      []string           `json:"digmark_managers"`
	DigmarkVerticals     []dmvertical.Model `json:"digmark_verticals"`

	EcbActiveCurrenciesExEur []ecbcurr.Model `json:"ecb_active_currencies_ex_eur"`

	GeoCountries []geocountry.Model `json:"geo_countries"`
	GeoOceans    []geoocean.Model   `json:"geo_oceans"`

	ProcessFlows []procflow.Model `json:"process_flows"`

	PubAuthors []pubauthor.Model `json:"pub_authors"`

	SuppCompanies         []suppcompany.Model      `json:"supp_companies"`
	SuppProductCategories []suppprodcategory.Model `json:"supp_product_categories"`
}

func (svc Service) SelectUiStoreData(ctx context.Context, db *pgxpool.Pool) (uiStoreData UiStoreData, err error) {

	uiStoreData.CoreMandatoryEnums, err = lyspg.SelectEnum(ctx, db, "core.mandatory_enum", nil, nil, "")
	if err != nil {
		return UiStoreData{}, fmt.Errorf("lyspg.SelectEnum (core.mandatory_enum) failed: %w", err)
	}

	uiStoreData.CoreOptionalEnums, err = lyspg.SelectEnum(ctx, db, "core.optional_enum", nil, nil, "")
	if err != nil {
		return UiStoreData{}, fmt.Errorf("lyspg.SelectEnum (core.optional_enum) failed: %w", err)
	}

	uiStoreData.CorePeriods, err = lyspg.SelectEnum(ctx, db, "core.performance_period", nil, nil, "")
	if err != nil {
		return UiStoreData{}, fmt.Errorf("lyspg.SelectEnum (core.performance_period) failed: %w", err)
	}

	// ----------------------------------------------------------------

	uiStoreData.DigmarkLauncherStati, err = lyspg.SelectEnum(ctx, db, "digmark.launcher_status", nil, nil, "")
	if err != nil {
		return UiStoreData{}, fmt.Errorf("lyspg.SelectEnum (digmark.launcher_status) failed: %w", err)
	}

	uiStoreData.DigmarkManagers, err = lyspg.SelectEnum(ctx, db, "digmark.manager", nil, nil, "")
	if err != nil {
		return UiStoreData{}, fmt.Errorf("lyspg.SelectEnum (digmark.manager) failed: %w", err)
	}

	dmVertStore := dmvertical.Store{Db: db}
	uiStoreData.DigmarkVerticals, _, err = dmVertStore.Select(ctx, lyspg.SelectParams{
		Fields: []string{"id", "name"},
	})
	if err != nil {
		return UiStoreData{}, fmt.Errorf("dmVertStore.Select failed: %w", err)
	}

	// ----------------------------------------------------------------

	ecbCurrStore := ecbcurr.Store{Db: db}
	uiStoreData.EcbActiveCurrenciesExEur, _, err = ecbCurrStore.Select(ctx, lyspg.SelectParams{
		Conditions: []lyspg.Condition{
			{Field: "code", Operator: lyspg.OpNotEquals, Value: "EUR"},
			{Field: "is_active", Operator: lyspg.OpEquals, Value: "true"},
		},
		Fields: []string{"id", "name"},
	})
	if err != nil {
		return UiStoreData{}, fmt.Errorf("ecbCurrStore.Select failed: %w", err)
	}

	// ----------------------------------------------------------------

	geoCountryStore := geocountry.Store{Db: db}
	uiStoreData.GeoCountries, _, err = geoCountryStore.Select(ctx, lyspg.SelectParams{
		Fields: []string{"id", "name"},
	})
	if err != nil {
		return UiStoreData{}, fmt.Errorf("geoCountryStore.Select failed: %w", err)
	}

	geoOceanStore := geoocean.Store{Db: db}
	uiStoreData.GeoOceans, _, err = geoOceanStore.Select(ctx, lyspg.SelectParams{
		Fields: []string{"id", "name"},
	})
	if err != nil {
		return UiStoreData{}, fmt.Errorf("geoOceanStore.Select failed: %w", err)
	}

	// ----------------------------------------------------------------

	procFlowStore := procflow.Store{Db: db}
	uiStoreData.ProcessFlows, _, err = procFlowStore.Select(ctx, lyspg.SelectParams{
		Fields: []string{"id", "name"},
	})
	if err != nil {
		return UiStoreData{}, fmt.Errorf("procFlowStore.Select failed: %w", err)
	}

	// ----------------------------------------------------------------

	pubAuthorStore := pubauthor.Store{Db: db}
	uiStoreData.PubAuthors, _, err = pubAuthorStore.Select(ctx, lyspg.SelectParams{
		Fields: []string{"id", "name"},
	})
	if err != nil {
		return UiStoreData{}, fmt.Errorf("pubAuthorStore.Select failed: %w", err)
	}

	// ----------------------------------------------------------------

	suppCompanyStore := suppcompany.Store{Db: db}
	uiStoreData.SuppCompanies, _, err = suppCompanyStore.Select(ctx, lyspg.SelectParams{
		Fields: []string{"id", "name"},
	})
	if err != nil {
		return UiStoreData{}, fmt.Errorf("suppCompanyStore.Select failed: %w", err)
	}

	suppProdCatStore := suppprodcategory.Store{Db: db}
	uiStoreData.SuppProductCategories, _, err = suppProdCatStore.Select(ctx, lyspg.SelectParams{
		Fields: []string{"id", "name"},
	})
	if err != nil {
		return UiStoreData{}, fmt.Errorf("suppProdCatStore.Select failed: %w", err)
	}

	return uiStoreData, nil
}

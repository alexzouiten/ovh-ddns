package ovh

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"

	"github.com/alexzouiten/ovh-ddns/internal/config"
	"github.com/alexzouiten/ovh-ddns/internal/model"
	"github.com/alexzouiten/ovh-ddns/internal/ovh/dto"
	ovhapi "github.com/ovh/go-ovh/ovh"
)

type Ovh struct {
	client *ovhapi.Client
	dryRun bool
}

func New(auth config.Auth, dryRun bool) (*Ovh, error) {
	client, err := ovhapi.NewClient(
		"ovh-eu",
		auth.AppKey,
		auth.AppSecret,
		auth.ConsumerKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ovh client: %w", err)
	}

	return NewFromClient(client, dryRun), nil
}

func NewFromClient(client *ovhapi.Client, dryRun bool) *Ovh {
	return &Ovh{
		client: client,
		dryRun: dryRun,
	}
}

func (ovh *Ovh) GetRecord(ctx context.Context, zoneName string, recordType model.RecordType, subdomain string) (*model.Record, error) {
	var ids []int
	err := ovh.client.GetWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record?fieldType=%s&subDomain=%s", zoneName, recordType, subdomain),
		&ids,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get id of %s record for zone %s: %w",
			recordType,
			zoneName,
			err,
		)
	}

	if len(ids) == 0 {
		return nil, nil
	}
	if len(ids) > 1 {
		return nil, ErrMultipleRecords
	}

	var dto dto.Record
	err = ovh.client.GetWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record/%d", zoneName, ids[0]),
		&dto,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get %s record for zone %s: %w",
			recordType,
			zoneName,
			err,
		)
	}

	ret := dto.ToModel()
	return &ret, nil
}

func (ovh *Ovh) PostRecord(ctx context.Context, r model.Record) (model.Record, error) {
	d := dto.NewRecordPostDtoFromModel(r)

	if ovh.dryRun {
		json, _ := json.Marshal(d)
		slog.Warn(
			"dry run: skipping post",
			"zone", r.Zone,
			"recordType", r.Type,
			"target", r.Target,
			"dto", json,
		)
		r.Id = rand.Int()
		return r, nil
	}

	var record dto.Record
	err := ovh.client.PostWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record", r.Zone),
		d,
		&record,
	)
	if err != nil {
		return model.Record{}, fmt.Errorf("failed to POST record: %w", err)
	}

	return record.ToModel(), nil
}

func (ovh *Ovh) PutRecord(ctx context.Context, r model.Record) error {
	d := dto.NewRecordPutDtoFromModel(r)

	if ovh.dryRun {
		json, _ := json.Marshal(d)

		slog.Warn(
			"dry run: skipping put",
			"zone", r.Zone,
			"recordType", r.Type,
			"target", r.Target,
			"dto", json,
		)
		return nil
	}

	err := ovh.client.PutWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record/%d", r.Zone, r.Id),
		d,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to update dns record: %w", err)
	}

	return nil
}

func (ovh *Ovh) Refresh(ctx context.Context, zoneName string) error {
	if ovh.dryRun {
		slog.Warn(
			"dry run: skipping refresh",
			"zone", zoneName,
		)
		return nil
	}

	err := ovh.client.PostWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/refresh", zoneName),
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to refresh dns zone: %w", err)
	}

	return nil
}

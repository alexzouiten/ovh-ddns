package updater

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strings"
	"sync"

	"github.com/alexzouiten/ovh-ddns/internal/config"
	"github.com/alexzouiten/ovh-ddns/internal/fetcher"
	"github.com/alexzouiten/ovh-ddns/internal/model"
	"github.com/alexzouiten/ovh-ddns/internal/ovh"
	"golang.org/x/sync/errgroup"
)

type Updater struct {
	cfg     config.Config
	fetcher *fetcher.Fetcher
	ovh     *ovh.Ovh

	previousIps        *model.Ips
	zonesToRefresh     []string
	zonesToRefreshLock sync.Mutex
}

type zone map[model.RecordType]*model.Record

func New(f *fetcher.Fetcher, o *ovh.Ovh, cfg config.Config) *Updater {
	return &Updater{
		fetcher: f,
		ovh:     o,
		cfg:     cfg,
	}
}

func (u *Updater) RunUpdate(ctx context.Context) error {
	slog.Debug("starting dns update")
	slog.Debug("fetching current public IPs")
	currentIps, err := u.fetcher.FetchIps(ctx, u.cfg.IpProvider)
	if err != nil {
		return err
	}
	slog.Debug("current public IPs", "v4", currentIps.V4, "v6", currentIps.V6)

	if u.previousIps != nil && currentIps.Equal(*u.previousIps) {
		slog.Debug("public IPs have not changed, skipping dns update")
		return nil
	}
	u.previousIps = &currentIps

	slog.Debug("updating each record from config")

	eg, egCtx := errgroup.WithContext(ctx)
	for _, recordCfg := range u.cfg.Records {
		eg.Go(func() error {
			return u.updateRecords(egCtx, currentIps, recordCfg)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	if len(u.zonesToRefresh) == 0 {
		slog.Debug("no change was made, no zone to refresh")
		return nil
	}

	// eliminate duplicates from zonesToRefresh
	slices.Sort(u.zonesToRefresh)
	u.zonesToRefresh = slices.Compact(u.zonesToRefresh)

	slog.Debug(fmt.Sprintf("refreshing %d zones", len(u.zonesToRefresh)), "zones", strings.Join(u.zonesToRefresh, ", "))

	eg, egCtx = errgroup.WithContext(ctx)
	for _, zoneToRefresh := range u.zonesToRefresh {
		eg.Go(func() error {
			return u.ovh.Refresh(egCtx, zoneToRefresh)
		})
	}
	return eg.Wait()
}

func (u *Updater) updateRecords(ctx context.Context, ips model.Ips, recordCfg config.Record) error {
	if ips.V4.IsUnspecified() && slices.Contains(recordCfg.Types, model.RecordA) {
		return ErrNoV4Address
	}
	if ips.V6.IsUnspecified() && slices.Contains(recordCfg.Types, model.RecordAAAA) {
		return ErrNoV6Address
	}

	currentRecords, err := u.getCurrentRecord(ctx, recordCfg)
	if err != nil {
		return err
	}

	eg, egCtx := errgroup.WithContext(ctx)
	for recordType, record := range currentRecords {
		eg.Go(func() error {
			var ip net.IP
			switch recordType {
			case model.RecordA:
				ip = ips.V4
			case model.RecordAAAA:
				ip = ips.V6
			default:
				panic("unhandled record type")
			}

			if record == nil {
				return u.createRecord(egCtx, ip, recordType, recordCfg)
			}

			if record.Target.Equal(ip) {
				slog.Debug("up to date record, skipping update", "zone", record.Zone, "type", recordType)
				return nil
			}
			return u.updateRecord(egCtx, ip, *record)
		})
	}

	return eg.Wait()
}

func (u *Updater) getCurrentRecord(ctx context.Context, recordCfg config.Record) (zone, error) {
	ret := make(zone)

	eg, egCtx := errgroup.WithContext(ctx)
	for _, recordType := range recordCfg.Types {
		eg.Go(func() error {
			record, err := u.ovh.GetRecord(egCtx, recordCfg.Zone, recordType, recordCfg.Subdomain)
			if err != nil {
				return err
			}
			ret[recordType] = record
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (u *Updater) createRecord(ctx context.Context, ip net.IP, recordType model.RecordType, recordCfg config.Record) error {
	record := model.Record{
		Type:      recordType,
		SubDomain: recordCfg.Subdomain,
		Target:    ip,
		Ttl:       recordCfg.Ttl,
	}

	if _, err := u.ovh.PostRecord(ctx, record); err != nil {
		return err
	}

	u.addZoneToRefresh(recordCfg.Zone)
	return nil
}

func (u *Updater) updateRecord(ctx context.Context, ip net.IP, record model.Record) error {
	record.Target = ip

	if err := u.ovh.PutRecord(ctx, record); err != nil {
		return err
	}

	u.addZoneToRefresh(record.Zone)
	return nil
}

func (u *Updater) addZoneToRefresh(zone string) {
	u.zonesToRefreshLock.Lock()
	u.zonesToRefresh = append(u.zonesToRefresh, zone)
	u.zonesToRefreshLock.Unlock()
}

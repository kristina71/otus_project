package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kristina71/otus_project/internal/config"
	"github.com/kristina71/otus_project/internal/storage"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zapadapter"
	"go.uber.org/zap"
)

type Storage struct {
	db                    *sqlx.DB
	driverName            string
	dsn                   string
	maxOpenConnections    int
	maxIdleConnections    int
	maxConnectionLifetime time.Duration
}

func NewStorage(driverName string, cnf config.DBConfig) *Storage {
	return &Storage{
		driverName:            driverName,
		dsn:                   cnf.DSN,
		maxOpenConnections:    cnf.MaxOpenConnections,
		maxIdleConnections:    cnf.MaxIdleConnections,
		maxConnectionLifetime: cnf.MaxConnectionLifetime,
	}
}

func NewStorageTest(db *sqlx.DB, driverName string, cnf config.DBConfig) *Storage {
	return &Storage{
		db:                    db,
		driverName:            driverName,
		dsn:                   cnf.DSN,
		maxOpenConnections:    cnf.MaxOpenConnections,
		maxIdleConnections:    cnf.MaxIdleConnections,
		maxConnectionLifetime: cnf.MaxConnectionLifetime,
	}
}

func (s *Storage) Connect(ctx context.Context) (err error) {
	db, err := sql.Open(s.driverName, s.dsn)
	if err != nil {
		return fmt.Errorf("failed to open db connection: %w", err)
	}

	db.SetMaxOpenConns(s.maxOpenConnections)
	db.SetMaxIdleConns(s.maxIdleConnections)
	db.SetConnMaxLifetime(s.maxConnectionLifetime)

	loggerAdapter := zapadapter.New(zap.L())
	db = sqldblogger.OpenDriver(s.dsn, db.Driver(), loggerAdapter)

	s.db = sqlx.NewDb(db, s.driverName)
	if err = s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	return nil
}

func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("error during db connection pool closing: %w", err)
	}
	return nil
}

func (s *Storage) AddSlot(ctx context.Context, description string) (string, error) {
	query := "INSERT INTO slots (slot_description) VALUES (:description) RETURNING slot_id"
	rows, err := s.db.NamedQueryContext(ctx, query, map[string]interface{}{"description": description})
	if err != nil {
		return "", fmt.Errorf("sql execution error: %w", err)
	}
	defer func() {
		err := rows.Close()
		zap.L().Error("error closing sql rows object", zap.Error(err))
	}()

	var id string
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return "", fmt.Errorf("sql AddSlot result parsing error: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("sql AddSlot result parsing error %w", err)
	}
	return id, nil
}

func (s *Storage) GetSlotByID(ctx context.Context, id string) (storage.Slot, error) {
	query := "SELECT slot_id, slot_description FROM slots WHERE slot_id = $1;"
	row := s.db.QueryRowxContext(ctx, query, id)
	if err := row.Err(); err != nil {
		return storage.Slot{}, fmt.Errorf("sql execution error: %w", err)
	}

	slot := new(storage.Slot)
	err := row.StructScan(slot)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return storage.Slot{}, storage.ErrSlotNotFound
	case err != nil:
		return storage.Slot{}, fmt.Errorf("sql GetSlotById result scan error: %w", err)
	default:
		return *slot, nil
	}
}

func (s *Storage) DeleteSlot(ctx context.Context, id string) error {
	query := "DELETE FROM slots WHERE slot_id = :id"
	res, err := s.db.NamedExecContext(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		return fmt.Errorf("sql delete slot delete operation query error: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error during rows affected by delete checking: %w", err)
	}
	if affected == 0 {
		return storage.ErrSlotNotFound
	}
	return nil
}

func (s *Storage) AddBanner(ctx context.Context, description string) (string, error) {
	query := "INSERT INTO banners (banner_description) VALUES (:description) RETURNING banner_id"
	rows, err := s.db.NamedQueryContext(ctx, query, map[string]interface{}{"description": description})
	if err != nil {
		return "", fmt.Errorf("sql execution error: %w", err)
	}
	defer func() {
		err := rows.Close()
		zap.L().Error("error closing sql rows object", zap.Error(err))
	}()

	var id string
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return "", fmt.Errorf("sql AddBanner result parsing error: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("sql AddBanner result parsing error %w", err)
	}
	return id, nil
}

func (s *Storage) GetBannerByID(ctx context.Context, id string) (storage.Banner, error) {
	query := "SELECT banner_id, banner_description FROM banners WHERE banner_id = $1 LIMIT 1;"
	row := s.db.QueryRowxContext(ctx, query, id)
	if err := row.Err(); err != nil {
		return storage.Banner{}, fmt.Errorf("sql execution error: %w", err)
	}

	banner := new(storage.Banner)
	err := row.StructScan(banner)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return storage.Banner{}, storage.ErrSlotNotFound
	case err != nil:
		return storage.Banner{}, fmt.Errorf("sql GetBannerById result scan error: %w", err)
	default:
		return *banner, nil
	}
}

func (s *Storage) DeleteBanner(ctx context.Context, id string) error {
	query := "DELETE FROM banners WHERE banner_id = :id"
	res, err := s.db.NamedExecContext(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		return fmt.Errorf("sql delete slot delete operation query error: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error during rows affected by delete checking: %w", err)
	}
	if affected == 0 {
		return storage.ErrBannerNotFound
	}
	return nil
}

func (s *Storage) AddBannerToSlot(ctx context.Context, slotID, bannerID string) error {
	query := "INSERT INTO slot_banners (slot_id, banner_id) VALUES (:slotId, :bannerId)"
	_, err := s.db.NamedExecContext(ctx, query, map[string]interface{}{
		"slotId":   slotID,
		"bannerId": bannerID,
	})
	if err != nil {
		return fmt.Errorf("error during sql execution: %w", err)
	}
	return nil
}

func (s *Storage) DeleteBannerFromSlot(ctx context.Context, slotID, bannerID string) error {
	query := "DELETE FROM slot_banners WHERE slot_id = :slotId AND banner_id = :bannerId"
	res, err := s.db.NamedExecContext(ctx, query, map[string]interface{}{
		"bannerId": bannerID,
		"slotId":   slotID,
	})
	if err != nil {
		return fmt.Errorf("sql delete slot delete operation query error: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error during rows affected by delete checking: %w", err)
	}
	if affected == 0 {
		return storage.ErrSlotToBannerRelationNotFound
	}
	return nil
}

func (s *Storage) AddGroup(ctx context.Context, description string) (string, error) {
	query := "INSERT INTO social_groups (group_description) VALUES (:description) RETURNING group_id"
	rows, err := s.db.NamedQueryContext(ctx, query, map[string]interface{}{"description": description})
	if err != nil {
		return "", fmt.Errorf("sql execution error: %w", err)
	}
	defer func() {
		err := rows.Close()
		zap.L().Error("error closing sql rows object", zap.Error(err))
	}()

	var id string
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return "", fmt.Errorf("sql AddGroup result parsing error: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("sql AddGroup result parsing error %w", err)
	}
	return id, nil
}

func (s *Storage) GetGroupByID(ctx context.Context, groupID string) (storage.SocialGroup, error) {
	query := "SELECT group_id, group_description FROM social_groups WHERE group_id = $1 LIMIT 1;"
	row := s.db.QueryRowxContext(ctx, query, groupID)
	if err := row.Err(); err != nil {
		return storage.SocialGroup{}, fmt.Errorf("sql execution error: %w", err)
	}

	group := new(storage.SocialGroup)
	err := row.StructScan(group)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return storage.SocialGroup{}, storage.ErrGroupNotFound
	case err != nil:
		return storage.SocialGroup{}, fmt.Errorf("sql GetGroupById result scan error: %w", err)
	default:
		return *group, nil
	}
}

func (s *Storage) DeleteGroup(ctx context.Context, id string) error {
	query := "DELETE FROM social_groups WHERE group_id = :id;"
	res, err := s.db.NamedExecContext(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		return fmt.Errorf("sql delete slot delete operation query error: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error during rows affected by delete checking: %w", err)
	}
	if affected == 0 {
		return storage.ErrGroupNotFound
	}
	return nil
}

func (s *Storage) PersistClick(ctx context.Context, slotID, groupID, bannerID string) error {
	query := `UPDATE banner_stats
			  SET clicks_amount = clicks_amount + 1
			  WHERE slot_id = :slotId AND group_id = :groupId AND banner_id = :bannerId`
	res, err := s.db.NamedExecContext(ctx, query, map[string]interface{}{
		"bannerId": bannerID,
		"slotId":   slotID,
		"groupId":  groupID,
	})
	if err != nil {
		return fmt.Errorf("error during sql execution: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error during sql rows affected checking: %w", err)
	}

	if affected == 0 {
		return storage.ErrBannerNotShown
	}
	return nil
}

func (s *Storage) PersistShow(ctx context.Context, slotID, groupID, bannerID string) error {
	query := `UPDATE banner_stats 
			  SET shows_amount = shows_amount + 1 
			  WHERE slot_id = :slotId AND group_id = :groupId AND banner_id = :bannerId`
	res, err := s.db.NamedExecContext(ctx, query, map[string]interface{}{
		"bannerId": bannerID,
		"slotId":   slotID,
		"groupId":  groupID,
	})
	if err != nil {
		return fmt.Errorf("error during sql execution: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error during sql rows affected checking: %w", err)
	}

	// if no rows affected by query, than this method should to init stats for current banner
	if affected == 0 {
		query := `INSERT INTO banner_stats (slot_id, banner_id, group_id, clicks_amount, shows_amount)
 				  VALUES (:slotId, :bannerId, :groupId, 0, 1)`
		res, err := s.db.NamedExecContext(ctx, query, map[string]interface{}{
			"slotId":   slotID,
			"bannerId": bannerID,
			"groupId":  groupID,
		})
		if err != nil {
			return fmt.Errorf("error during sql execution: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("error during sql rows affected checking: %w", err)
		}
		if rowsAffected == 0 {
			return storage.ErrFailedStatsInit
		}
	}
	return nil
}

func (s *Storage) FindSlotBannerStats(ctx context.Context, slotID, groupID string) ([]storage.SlotBannerStat, error) {
	query := `SELECT sb.banner_id, clicks_amount, shows_amount
			  FROM (select slot_id, banner_id
		  			FROM slot_banners
		  			WHERE slot_id = :slotId
			  ) as sb
			  left join banner_stats bs
			  ON sb.slot_id = bs.slot_id AND sb.banner_id = bs.banner_id AND group_id = :groupId`
	rows, err := s.db.NamedQueryContext(ctx, query, map[string]interface{}{
		"slotId":  slotID,
		"groupId": groupID,
	})
	if err != nil {
		return nil, fmt.Errorf("error during sql execution: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			zap.L().Error("errors during rows closing", zap.Error(err))
		}
	}()

	var stats []storage.SlotBannerStat
	var bannerStat storage.SlotBannerStat
	for rows.Next() {
		if err := rows.StructScan(&bannerStat); err != nil {
			return nil, fmt.Errorf("sql error FindSlotBannerStats result parsing: %w", err)
		}
		stats = append(stats, bannerStat)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("sql error FindSlotBannerStats result parsing: %w", err)
	}
	return stats, nil
}

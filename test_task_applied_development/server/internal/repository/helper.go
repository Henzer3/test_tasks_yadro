package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	"test.task.log.server/internal/entity"
)

type logRow struct {
	ID         int          `db:"id"`
	Path       string       `db:"path"`
	Status     string       `db:"status"`
	NodesTotal int          `db:"nodes_total"`
	PortsTotal int          `db:"ports_total"`
	CreatedAt  time.Time    `db:"created_at"`
	ParsedAt   sql.NullTime `db:"parsed_at"`
	Error      string       `db:"error"`
}

func (r logRow) toEntity() entity.Log {
	var parsedAt *time.Time

	if r.ParsedAt.Valid {
		t := r.ParsedAt.Time
		parsedAt = &t
	}

	return entity.Log{
		ID:         r.ID,
		Path:       r.Path,
		Status:     r.Status,
		NodesTotal: r.NodesTotal,
		PortsTotal: r.PortsTotal,
		CreatedAt:  r.CreatedAt,
		ParsedAt:   parsedAt,
		Error:      r.Error,
	}
}

type nodeRow struct {
	ID              int    `db:"id"`
	LogID           int    `db:"log_id"`
	NodeDesc        string `db:"node_desc"`
	NumPorts        int    `db:"num_ports"`
	NodeType        int    `db:"node_type"`
	ClassVersion    string `db:"class_version"`
	BaseVersion     string `db:"base_version"`
	SystemImageGUID string `db:"system_image_guid"`
	NodeGUID        string `db:"node_guid"`
	PortGUID        string `db:"port_guid"`
}

func (r nodeRow) toEntity() entity.Node {
	return entity.Node{
		ID:              r.ID,
		LogID:           r.LogID,
		NodeDesc:        r.NodeDesc,
		NumPorts:        r.NumPorts,
		NodeType:        r.NodeType,
		ClassVersion:    r.ClassVersion,
		BaseVersion:     r.BaseVersion,
		SystemImageGUID: r.SystemImageGUID,
		NodeGUID:        r.NodeGUID,
		PortGUID:        r.PortGUID,
	}
}

type portRow struct {
	ID                   int    `db:"id"`
	NodeID               int    `db:"node_id"`
	NodeGUID             string `db:"node_guid"`
	PortGUID             string `db:"port_guid"`
	PortNum              int    `db:"port_num"`
	LID                  string `db:"lid"`
	LocalPortNum         string `db:"local_port_num"`
	LinkWidthActive      string `db:"link_width_active"`
	LinkWidthSupported   string `db:"link_width_supported"`
	LinkSpeedActive      string `db:"link_speed_active"`
	LinkSpeedSupported   string `db:"link_speed_supported"`
	PortState            string `db:"port_state"`
	PortPhyState         string `db:"port_phy_state"`
	MTUCap               string `db:"mtu_cap"`
	LinkRoundTripLatency string `db:"link_round_trip_latency"`
}

func (r portRow) toEntity() entity.Port {
	return entity.Port{
		ID:                   r.ID,
		NodeID:               r.NodeID,
		NodeGUID:             r.NodeGUID,
		PortGUID:             r.PortGUID,
		PortNum:              r.PortNum,
		LID:                  r.LID,
		LocalPortNum:         r.LocalPortNum,
		LinkWidthActive:      r.LinkWidthActive,
		LinkWidthSupported:   r.LinkWidthSupported,
		LinkSpeedActive:      r.LinkSpeedActive,
		LinkSpeedSupported:   r.LinkSpeedSupported,
		PortState:            r.PortState,
		PortPhyState:         r.PortPhyState,
		MTUCap:               r.MTUCap,
		LinkRoundTripLatency: r.LinkRoundTripLatency,
	}
}

type nodeInfoRow struct {
	NodeID   int    `db:"node_id"`
	InfoKind string `db:"info_kind"`
	Payload  []byte `db:"payload"`
}

func upsertNodeInfo(
	ctx context.Context,
	tx *sqlx.Tx,
	logID int,
	nodeID int,
	infoKind string,
	payload any,
) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal node info: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		`
		INSERT INTO nodes_info (
			log_id,
			node_id,
			info_kind,
			payload
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (log_id, node_id, info_kind)
		DO UPDATE SET payload = EXCLUDED.payload
		`,
		logID,
		nodeID,
		infoKind,
		payloadBytes,
	)

	if err != nil {
		return mapDBError(err)
	}

	return nil
}

func getNodeInfoRowsByLog(ctx context.Context, db *sqlx.DB, logID int) ([]nodeInfoRow, error) {
	var rows []nodeInfoRow

	err := db.SelectContext(
		ctx,
		&rows,
		`
		SELECT 
			node_id,
			info_kind,
			payload
		FROM nodes_info
		WHERE log_id = $1
		ORDER BY node_id, info_kind
		`,
		logID,
	)
	if err != nil {
		return nil, mapDBError(err)
	}

	return rows, nil
}

func getNodeInfoRowsByNode(ctx context.Context, db *sqlx.DB, nodeID int) ([]nodeInfoRow, error) {
	var rows []nodeInfoRow

	err := db.SelectContext(
		ctx,
		&rows,
		`
		SELECT 
			node_id,
			info_kind,
			payload
		FROM nodes_info
		WHERE node_id = $1
		ORDER BY info_kind
		`,
		nodeID,
	)
	if err != nil {
		return nil, mapDBError(err)
	}

	return rows, nil
}

func applyNodeInfo(node *entity.Node, infoKind string, payload []byte) error {
	switch infoKind {
	case "system":
		var systemInfo entity.SystemInfo

		if len(payload) > 0 && string(payload) != "null" {
			if err := json.Unmarshal(payload, &systemInfo); err != nil {
				return fmt.Errorf("unmarshal system info: %w", err)
			}
		}

		node.SystemInfo = &systemInfo

	case "switch":
		fields := make(map[string]string)

		if len(payload) > 0 && string(payload) != "null" {
			if err := json.Unmarshal(payload, &fields); err != nil {
				return fmt.Errorf("unmarshal switch info: %w", err)
			}
		}

		node.SwitchInfo = fields

	case "node":
		fields := make(map[string]string)

		if len(payload) > 0 && string(payload) != "null" {
			if err := json.Unmarshal(payload, &fields); err != nil {
				return fmt.Errorf("unmarshal node info: %w", err)
			}
		}

		node.NodeInfo = fields

	default:
		return fmt.Errorf("unknown node info kind: %s", infoKind)
	}

	return nil
}

func normalizeGUID(guid string) string {
	guid = strings.TrimSpace(strings.ToLower(guid))
	guid = strings.TrimPrefix(guid, "0x")

	return guid
}

func mapDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return entity.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%w: %s", entity.ErrConflict, pgErr.Message)

		case "23503", "23514", "23502":
			return fmt.Errorf("%w: %s", entity.ErrBadArguments, pgErr.Message)
		}
	}

	return err
}

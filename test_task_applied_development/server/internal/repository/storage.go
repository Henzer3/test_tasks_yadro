package repository

import (
	"context"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"test.task.log.server/internal/entity"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {
	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Close() error {
	if err := db.conn.Close(); err != nil {
		db.log.Error("cant close db conn", "err", err)
		return err
	}
	return nil
}

func (db *DB) CreateLog(path string) (int, error) {
	ctx := context.Background()

	var id int
	err := db.conn.QueryRowxContext(
		ctx,
		`
		INSERT INTO logs (path, status)
		VALUES ($1, $2)
		RETURNING id
		`,
		path,
		"created",
	).Scan(&id)

	if err != nil {
		return 0, mapDBError(err)
	}

	return id, nil
}

func (db *DB) UpdateLogStatus(logID int, status string, errText string) error {
	ctx := context.Background()

	result, err := db.conn.ExecContext(
		ctx,
		`
		UPDATE logs
		SET 
			status = $2,
			error = $3,
			parsed_at = CASE 
				WHEN $2 IN ('parsed', 'failed') THEN NOW()
				ELSE parsed_at
			END
		WHERE id = $1
		`,
		logID,
		status,
		errText,
	)
	if err != nil {
		return mapDBError(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("%w: log id=%d", entity.ErrNotFound, logID)
	}

	return nil
}

func (db *DB) SaveParsedData(logID int, data entity.ParsedData) error {
	ctx := context.Background()

	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return mapDBError(err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			db.log.Error("cant rollback in rep", "err", err)
		}
	}()
	var exists int
	err = tx.QueryRowxContext(
		ctx,
		`SELECT id FROM logs WHERE id = $1 FOR UPDATE`,
		logID,
	).Scan(&exists)
	if err != nil {
		return mapDBError(err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM nodes_info WHERE log_id = $1`, logID); err != nil {
		return mapDBError(err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM ports WHERE log_id = $1`, logID); err != nil {
		return mapDBError(err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM nodes WHERE log_id = $1`, logID); err != nil {
		return mapDBError(err)
	}

	nodeIDByGUID := make(map[string]int, len(data.Nodes))

	for i := range data.Nodes {
		node := data.Nodes[i]

		var nodeID int
		err := tx.QueryRowxContext(
			ctx,
			`
			INSERT INTO nodes (
				log_id,
				node_desc,
				num_ports,
				node_type,
				class_version,
				base_version,
				system_image_guid,
				node_guid,
				port_guid
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
			`,
			logID,
			node.NodeDesc,
			node.NumPorts,
			node.NodeType,
			node.ClassVersion,
			node.BaseVersion,
			node.SystemImageGUID,
			node.NodeGUID,
			node.PortGUID,
		).Scan(&nodeID)

		if err != nil {
			return mapDBError(err)
		}

		data.Nodes[i].ID = nodeID
		data.Nodes[i].LogID = logID

		nodeIDByGUID[normalizeGUID(node.NodeGUID)] = nodeID

		if node.SystemInfo != nil {
			if err := upsertNodeInfo(ctx, tx, logID, nodeID, "system", node.SystemInfo); err != nil {
				return err
			}
		}

		if len(node.SwitchInfo) > 0 {
			if err := upsertNodeInfo(ctx, tx, logID, nodeID, "switch", node.SwitchInfo); err != nil {
				return err
			}
		}

		if len(node.NodeInfo) > 0 {
			if err := upsertNodeInfo(ctx, tx, logID, nodeID, "node", node.NodeInfo); err != nil {
				return err
			}
		}
	}

	for _, info := range data.SystemInfos {
		nodeID, ok := nodeIDByGUID[normalizeGUID(info.NodeGUID)]
		if !ok {
			return fmt.Errorf("%w: system info references unknown node guid %s", entity.ErrParseFailed, info.NodeGUID)
		}

		infoCopy := info
		if err := upsertNodeInfo(ctx, tx, logID, nodeID, "system", &infoCopy); err != nil {
			return err
		}
	}

	for _, info := range data.SwitchInfos {
		nodeID, ok := nodeIDByGUID[normalizeGUID(info.NodeGUID)]
		if !ok {
			return fmt.Errorf("%w: switch info references unknown node guid %s", entity.ErrParseFailed, info.NodeGUID)
		}

		if err := upsertNodeInfo(ctx, tx, logID, nodeID, "switch", info.Fields); err != nil {
			return err
		}
	}

	for _, info := range data.NodeInfos {
		nodeID, ok := nodeIDByGUID[normalizeGUID(info.NodeGUID)]
		if !ok {
			return fmt.Errorf("%w: node info references unknown node guid %s", entity.ErrParseFailed, info.NodeGUID)
		}

		if err := upsertNodeInfo(ctx, tx, logID, nodeID, "node", info.Fields); err != nil {
			return err
		}
	}

	for _, port := range data.Ports {
		nodeID, ok := nodeIDByGUID[normalizeGUID(port.NodeGUID)]
		if !ok {
			return fmt.Errorf("%w: port references unknown node guid %s", entity.ErrParseFailed, port.NodeGUID)
		}

		_, err = tx.ExecContext(
			ctx,
			`
			INSERT INTO ports (
				log_id,
				node_id,
				node_guid,
				port_guid,
				port_num,
				lid,
				local_port_num,
				link_width_active,
				link_width_supported,
				link_speed_active,
				link_speed_supported,
				port_state,
				port_phy_state,
				mtu_cap,
				link_round_trip_latency
			)
			VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9, $10,
				$11, $12, $13, $14, $15
			)
			`,
			logID,
			nodeID,
			port.NodeGUID,
			port.PortGUID,
			port.PortNum,
			port.LID,
			port.LocalPortNum,
			port.LinkWidthActive,
			port.LinkWidthSupported,
			port.LinkSpeedActive,
			port.LinkSpeedSupported,
			port.PortState,
			port.PortPhyState,
			port.MTUCap,
			port.LinkRoundTripLatency,
		)

		if err != nil {
			return mapDBError(err)
		}
	}

	_, err = tx.ExecContext(
		ctx,
		`
		UPDATE logs
		SET 
			nodes_total = $2,
			ports_total = $3
		WHERE id = $1
		`,
		logID,
		len(data.Nodes),
		len(data.Ports),
	)
	if err != nil {
		return mapDBError(err)
	}

	if err := tx.Commit(); err != nil {
		return mapDBError(err)
	}

	return nil
}

func (db *DB) GetLog(logID int) (entity.Log, error) {
	ctx := context.Background()

	var row logRow
	err := db.conn.GetContext(
		ctx,
		&row,
		`
		SELECT 
			id,
			path,
			status,
			nodes_total,
			ports_total,
			created_at,
			parsed_at,
			error
		FROM logs
		WHERE id = $1
		`,
		logID,
	)
	if err != nil {
		return entity.Log{}, mapDBError(err)
	}

	return row.toEntity(), nil
}

func (db *DB) GetTopology(logID int) (entity.Topology, error) {
	ctx := context.Background()

	var exists int
	err := db.conn.GetContext(ctx, &exists, `SELECT id FROM logs WHERE id = $1`, logID)
	if err != nil {
		return entity.Topology{}, mapDBError(err)
	}

	var rows []nodeRow
	err = db.conn.SelectContext(
		ctx,
		&rows,
		`
		SELECT 
			id,
			log_id,
			node_desc,
			num_ports,
			node_type,
			class_version,
			base_version,
			system_image_guid,
			node_guid,
			port_guid
		FROM nodes
		WHERE log_id = $1
		ORDER BY id
		`,
		logID,
	)
	if err != nil {
		return entity.Topology{}, mapDBError(err)
	}

	nodes := make([]entity.Node, 0, len(rows))
	nodeIndexByID := make(map[int]int, len(rows))

	for _, row := range rows {
		node := row.toEntity()
		nodes = append(nodes, node)
		nodeIndexByID[node.ID] = len(nodes) - 1
	}

	infoRows, err := getNodeInfoRowsByLog(ctx, db.conn, logID)
	if err != nil {
		return entity.Topology{}, err
	}

	for _, infoRow := range infoRows {
		index, ok := nodeIndexByID[infoRow.NodeID]
		if !ok {
			continue
		}

		if err := applyNodeInfo(&nodes[index], infoRow.InfoKind, infoRow.Payload); err != nil {
			return entity.Topology{}, err
		}
	}

	return entity.Topology{
		LogID: logID,
		Nodes: nodes,
	}, nil
}

func (db *DB) GetNode(nodeID int) (entity.Node, error) {
	ctx := context.Background()

	var row nodeRow
	err := db.conn.GetContext(
		ctx,
		&row,
		`
		SELECT 
			id,
			log_id,
			node_desc,
			num_ports,
			node_type,
			class_version,
			base_version,
			system_image_guid,
			node_guid,
			port_guid
		FROM nodes
		WHERE id = $1
		`,
		nodeID,
	)
	if err != nil {
		return entity.Node{}, mapDBError(err)
	}

	node := row.toEntity()

	infoRows, err := getNodeInfoRowsByNode(ctx, db.conn, nodeID)
	if err != nil {
		return entity.Node{}, err
	}

	for _, infoRow := range infoRows {
		if err := applyNodeInfo(&node, infoRow.InfoKind, infoRow.Payload); err != nil {
			return entity.Node{}, err
		}
	}

	return node, nil
}

func (db *DB) GetPortsByNode(nodeID int) ([]entity.Port, error) {
	ctx := context.Background()

	var exists int
	err := db.conn.GetContext(ctx, &exists, `SELECT id FROM nodes WHERE id = $1`, nodeID)
	if err != nil {
		return nil, mapDBError(err)
	}

	var rows []portRow
	err = db.conn.SelectContext(
		ctx,
		&rows,
		`
		SELECT 
			id,
			node_id,
			node_guid,
			port_guid,
			port_num,
			lid,
			local_port_num,
			link_width_active,
			link_width_supported,
			link_speed_active,
			link_speed_supported,
			port_state,
			port_phy_state,
			mtu_cap,
			link_round_trip_latency
		FROM ports
		WHERE node_id = $1
		ORDER BY port_num
		`,
		nodeID,
	)
	if err != nil {
		return nil, mapDBError(err)
	}

	ports := make([]entity.Port, 0, len(rows))

	for _, row := range rows {
		ports = append(ports, row.toEntity())
	}

	return ports, nil
}

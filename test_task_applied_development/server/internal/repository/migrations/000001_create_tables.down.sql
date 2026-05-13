DROP INDEX IF EXISTS nodes_info_payload_gin_idx;
DROP INDEX IF EXISTS nodes_info_info_kind_idx;
DROP INDEX IF EXISTS nodes_info_log_id_node_id_idx;
DROP INDEX IF EXISTS nodes_info_node_id_idx;
DROP INDEX IF EXISTS nodes_info_log_id_idx;

DROP INDEX IF EXISTS ports_node_guid_idx;
DROP INDEX IF EXISTS ports_log_id_node_id_idx;
DROP INDEX IF EXISTS ports_node_id_idx;
DROP INDEX IF EXISTS ports_log_id_idx;

DROP INDEX IF EXISTS nodes_log_id_node_guid_idx;
DROP INDEX IF EXISTS nodes_node_guid_idx;
DROP INDEX IF EXISTS nodes_log_id_idx;

DROP INDEX IF EXISTS logs_status_idx;

DROP TABLE IF EXISTS nodes_info;
DROP TABLE IF EXISTS ports;
DROP TABLE IF EXISTS nodes;
DROP TABLE IF EXISTS logs;
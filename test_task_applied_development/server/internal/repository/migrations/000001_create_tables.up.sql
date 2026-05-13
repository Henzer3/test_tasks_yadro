CREATE TABLE logs (
    id SERIAL PRIMARY KEY,

    path TEXT NOT NULL,
    status TEXT NOT NULL,

    nodes_total INTEGER NOT NULL DEFAULT 0,
    ports_total INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    parsed_at TIMESTAMPTZ NULL,

    error TEXT NOT NULL DEFAULT '',

    CONSTRAINT logs_status_check CHECK (
        status IN ('created', 'parsing', 'parsed', 'failed')
    )
);

CREATE TABLE nodes (
    id SERIAL PRIMARY KEY,

    log_id INTEGER NOT NULL REFERENCES logs(id) ON DELETE CASCADE,

    node_desc TEXT NOT NULL,
    num_ports INTEGER NOT NULL,
    node_type INTEGER NOT NULL,

    class_version TEXT NOT NULL DEFAULT '',
    base_version TEXT NOT NULL DEFAULT '',
    system_image_guid TEXT NOT NULL DEFAULT '',
    node_guid TEXT NOT NULL,
    port_guid TEXT NOT NULL DEFAULT '',

    CONSTRAINT nodes_num_ports_check CHECK (num_ports >= 0),
    CONSTRAINT nodes_log_id_node_guid_unique UNIQUE (log_id, node_guid),
    CONSTRAINT nodes_log_id_id_unique UNIQUE (log_id, id)
);

CREATE TABLE ports (
    id SERIAL PRIMARY KEY,

    log_id INTEGER NOT NULL,
    node_id INTEGER NOT NULL,

    node_guid TEXT NOT NULL,
    port_guid TEXT NOT NULL,
    port_num INTEGER NOT NULL,

    lid TEXT NOT NULL DEFAULT '',
    local_port_num TEXT NOT NULL DEFAULT '',

    link_width_active TEXT NOT NULL DEFAULT '',
    link_width_supported TEXT NOT NULL DEFAULT '',

    link_speed_active TEXT NOT NULL DEFAULT '',
    link_speed_supported TEXT NOT NULL DEFAULT '',

    port_state TEXT NOT NULL DEFAULT '',
    port_phy_state TEXT NOT NULL DEFAULT '',

    mtu_cap TEXT NOT NULL DEFAULT '',
    link_round_trip_latency TEXT NOT NULL DEFAULT '',

    CONSTRAINT ports_log_fk
        FOREIGN KEY (log_id)
        REFERENCES logs(id)
        ON DELETE CASCADE,

    CONSTRAINT ports_node_fk
        FOREIGN KEY (log_id, node_id)
        REFERENCES nodes(log_id, id)
        ON DELETE CASCADE,

    CONSTRAINT ports_port_num_check CHECK (port_num >= 0),
    CONSTRAINT ports_log_id_node_id_port_num_unique UNIQUE (log_id, node_id, port_num)
);

CREATE TABLE nodes_info (
    id SERIAL PRIMARY KEY,

    log_id INTEGER NOT NULL,
    node_id INTEGER NOT NULL,

    info_kind TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,

    CONSTRAINT nodes_info_log_fk
        FOREIGN KEY (log_id)
        REFERENCES logs(id)
        ON DELETE CASCADE,

    CONSTRAINT nodes_info_node_fk
        FOREIGN KEY (log_id, node_id)
        REFERENCES nodes(log_id, id)
        ON DELETE CASCADE,

    CONSTRAINT nodes_info_kind_check CHECK (
        info_kind IN ('system', 'switch', 'node')
    ),

    CONSTRAINT nodes_info_log_id_node_id_kind_unique UNIQUE (log_id, node_id, info_kind)
);

CREATE INDEX logs_status_idx ON logs(status);

CREATE INDEX nodes_log_id_idx ON nodes(log_id);
CREATE INDEX nodes_node_guid_idx ON nodes(node_guid);
CREATE INDEX nodes_log_id_node_guid_idx ON nodes(log_id, node_guid);

CREATE INDEX ports_log_id_idx ON ports(log_id);
CREATE INDEX ports_node_id_idx ON ports(node_id);
CREATE INDEX ports_log_id_node_id_idx ON ports(log_id, node_id);
CREATE INDEX ports_node_guid_idx ON ports(node_guid);

CREATE INDEX nodes_info_log_id_idx ON nodes_info(log_id);
CREATE INDEX nodes_info_node_id_idx ON nodes_info(node_id);
CREATE INDEX nodes_info_log_id_node_id_idx ON nodes_info(log_id, node_id);
CREATE INDEX nodes_info_info_kind_idx ON nodes_info(info_kind);
CREATE INDEX nodes_info_payload_gin_idx ON nodes_info USING GIN (payload);
package rest

type ParseRequest struct {
	Path string `json:"path"`
}

type ParseResponse struct {
	LogID int `json:"log_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type LogResponse struct {
	LogID      int    `json:"log_id"`
	Path       string `json:"path"`
	Status     string `json:"status"`
	NodesTotal int    `json:"nodes_total"`
	PortsTotal int    `json:"ports_total"`
	CreatedAt  string `json:"created_at,omitempty"`
	ParsedAt   string `json:"parsed_at,omitempty"`
	Error      string `json:"error,omitempty"`
}

type TopologyResponse struct {
	LogID int                    `json:"log_id"`
	Nodes []TopologyNodeResponse `json:"nodes"`
}

type TopologyNodeResponse struct {
	NodeID      int    `json:"node_id"`
	NodeGUID    string `json:"node_guid"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	TypeCode    int    `json:"type_code"`
	NumPorts    int    `json:"num_ports"`
	ProductName string `json:"product_name,omitempty"`
	Serial      string `json:"serial,omitempty"`
}

type NodeResponse struct {
	NodeID int `json:"node_id"`
	LogID  int `json:"log_id"`

	NodeDesc        string `json:"node_desc"`
	NodeGUID        string `json:"node_guid"`
	PortGUID        string `json:"port_guid"`
	SystemImageGUID string `json:"system_image_guid"`

	Type     string `json:"type"`
	TypeCode int    `json:"type_code"`
	NumPorts int    `json:"num_ports"`

	ClassVersion string `json:"class_version,omitempty"`
	BaseVersion  string `json:"base_version,omitempty"`

	SystemInfo *SystemInfoResponse `json:"system_info,omitempty"`
	SwitchInfo map[string]string   `json:"switch_info,omitempty"`
	NodeInfo   map[string]string   `json:"node_info,omitempty"`
}

type SystemInfoResponse struct {
	SerialNumber string `json:"serial_number,omitempty"`
	PartNumber   string `json:"part_number,omitempty"`
	Revision     string `json:"revision,omitempty"`
	ProductName  string `json:"product_name,omitempty"`
}

type PortsResponse struct {
	NodeID int            `json:"node_id"`
	Ports  []PortResponse `json:"ports"`
	Total  int            `json:"total"`
}

type PortResponse struct {
	PortID int `json:"port_id"`
	NodeID int `json:"node_id"`

	NodeGUID string `json:"node_guid"`
	PortGUID string `json:"port_guid"`
	PortNum  int    `json:"port_num"`

	LID                  string `json:"lid,omitempty"`
	LocalPortNum         string `json:"local_port_num,omitempty"`
	LinkWidthActive      string `json:"link_width_active,omitempty"`
	LinkWidthSupported   string `json:"link_width_supported,omitempty"`
	LinkSpeedActive      string `json:"link_speed_active,omitempty"`
	LinkSpeedSupported   string `json:"link_speed_supported,omitempty"`
	PortState            string `json:"port_state,omitempty"`
	PortPhyState         string `json:"port_phy_state,omitempty"`
	MTUCap               string `json:"mtu_cap,omitempty"`
	LinkRoundTripLatency string `json:"link_round_trip_latency,omitempty"`
}

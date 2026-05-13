package entity

import (
	"time"
)

type Log struct {
	ID         int
	Path       string
	Status     string
	NodesTotal int
	PortsTotal int
	CreatedAt  time.Time
	ParsedAt   *time.Time
	Error      string
}

type Topology struct {
	LogID int
	Nodes []Node
}

type Node struct {
	ID    int
	LogID int

	NodeDesc        string
	NumPorts        int
	NodeType        int
	ClassVersion    string
	BaseVersion     string
	SystemImageGUID string
	NodeGUID        string
	PortGUID        string

	SystemInfo *SystemInfo
	SwitchInfo map[string]string
	NodeInfo   map[string]string
}

type SystemInfo struct {
	NodeGUID string

	SerialNumber string
	PartNumber   string
	Revision     string
	ProductName  string
}

type SwitchInfo struct {
	NodeGUID string
	Fields   map[string]string
}

type NodeInfo struct {
	NodeGUID string
	Fields   map[string]string
}

type Port struct {
	ID     int
	NodeID int

	NodeGUID string
	PortGUID string
	PortNum  int

	LID                  string
	LocalPortNum         string
	LinkWidthActive      string
	LinkWidthSupported   string
	LinkSpeedActive      string
	LinkSpeedSupported   string
	PortState            string
	PortPhyState         string
	MTUCap               string
	LinkRoundTripLatency string
}

type ParsedData struct {
	Nodes       []Node
	Ports       []Port
	SwitchInfos []SwitchInfo
	SystemInfos []SystemInfo
	NodeInfos   []NodeInfo
}

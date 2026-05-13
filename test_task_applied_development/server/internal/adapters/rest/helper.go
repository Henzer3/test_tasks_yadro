package rest

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"test.task.log.server/internal/entity"
)

func toLogResponse(logEntity entity.Log) LogResponse {
	return LogResponse{
		LogID:      logEntity.ID,
		Path:       logEntity.Path,
		Status:     logEntity.Status,
		NodesTotal: logEntity.NodesTotal,
		PortsTotal: logEntity.PortsTotal,
		CreatedAt:  formatTime(logEntity.CreatedAt),
		ParsedAt:   formatTimePtr(logEntity.ParsedAt),
		Error:      logEntity.Error,
	}
}

func toTopologyResponse(topology entity.Topology) TopologyResponse {
	nodes := make([]TopologyNodeResponse, 0, len(topology.Nodes))

	for _, node := range topology.Nodes {
		nodes = append(nodes, toTopologyNodeResponse(node))
	}

	return TopologyResponse{
		LogID: topology.LogID,
		Nodes: nodes,
	}
}

func toTopologyNodeResponse(node entity.Node) TopologyNodeResponse {
	resp := TopologyNodeResponse{
		NodeID:   node.ID,
		NodeGUID: node.NodeGUID,
		Name:     node.NodeDesc,
		Type:     nodeTypeName(node.NodeType),
		TypeCode: node.NodeType,
		NumPorts: node.NumPorts,
	}

	if node.SystemInfo != nil {
		resp.ProductName = node.SystemInfo.ProductName
		resp.Serial = node.SystemInfo.SerialNumber
	}

	return resp
}

func toNodeResponse(node entity.Node) NodeResponse {
	resp := NodeResponse{
		NodeID:          node.ID,
		LogID:           node.LogID,
		NodeDesc:        node.NodeDesc,
		NodeGUID:        node.NodeGUID,
		PortGUID:        node.PortGUID,
		SystemImageGUID: node.SystemImageGUID,
		Type:            nodeTypeName(node.NodeType),
		TypeCode:        node.NodeType,
		NumPorts:        node.NumPorts,
		ClassVersion:    node.ClassVersion,
		BaseVersion:     node.BaseVersion,
		SwitchInfo:      node.SwitchInfo,
		NodeInfo:        node.NodeInfo,
	}

	if node.SystemInfo != nil {
		resp.SystemInfo = &SystemInfoResponse{
			SerialNumber: node.SystemInfo.SerialNumber,
			PartNumber:   node.SystemInfo.PartNumber,
			Revision:     node.SystemInfo.Revision,
			ProductName:  node.SystemInfo.ProductName,
		}
	}

	return resp
}

func toPortResponses(ports []entity.Port) []PortResponse {
	result := make([]PortResponse, 0, len(ports))

	for _, port := range ports {
		result = append(result, PortResponse{
			PortID:               port.ID,
			NodeID:               port.NodeID,
			NodeGUID:             port.NodeGUID,
			PortGUID:             port.PortGUID,
			PortNum:              port.PortNum,
			LID:                  port.LID,
			LocalPortNum:         port.LocalPortNum,
			LinkWidthActive:      port.LinkWidthActive,
			LinkWidthSupported:   port.LinkWidthSupported,
			LinkSpeedActive:      port.LinkSpeedActive,
			LinkSpeedSupported:   port.LinkSpeedSupported,
			PortState:            port.PortState,
			PortPhyState:         port.PortPhyState,
			MTUCap:               port.MTUCap,
			LinkRoundTripLatency: port.LinkRoundTripLatency,
		})
	}

	return result
}

func nodeTypeName(nodeType int) string {
	switch nodeType {
	case 1:
		return "host"
	case 2:
		return "switch"
	default:
		return "unknown"
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}

	return formatTime(*t)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "application/json" {
			return entity.ErrBadArguments
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	defer func() {
		_ = r.Body.Close()
	}()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return entity.ErrBadArguments
		}

		return entity.ErrBadArguments
	}

	var extra struct{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return entity.ErrBadArguments
	}

	return nil
}

func writeJSON(log *slog.Logger, w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil && log != nil {
		log.Error("cannot encode json response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: msg,
	})
}

func handleServiceError(log *slog.Logger, w http.ResponseWriter, err error, msg string) {
	switch {
	case errors.Is(err, entity.ErrBadArguments):
		writeError(w, http.StatusBadRequest, entity.ErrBadArguments.Error())

	case errors.Is(err, entity.ErrNotFound):
		writeError(w, http.StatusNotFound, entity.ErrNotFound.Error())

	case errors.Is(err, entity.ErrConflict):
		writeError(w, http.StatusConflict, entity.ErrConflict.Error())

	case errors.Is(err, entity.ErrParseFailed):
		writeError(w, http.StatusUnprocessableEntity, entity.ErrParseFailed.Error())

	default:
		if log != nil {
			log.Error(msg, "err", err)
		}

		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func positiveIntPathValue(r *http.Request, name string) (int, error) {
	raw := strings.TrimSpace(r.PathValue(name))
	if raw == "" {
		return 0, entity.ErrBadArguments
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, entity.ErrBadArguments
	}

	return value, nil
}

func validateDataPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", entity.ErrBadArguments
	}

	path = strings.ReplaceAll(path, "\\", "/")

	if filepath.IsAbs(path) {
		return "", entity.ErrBadArguments
	}

	cleanPath := filepath.Clean(path)
	dataDir := filepath.Clean("data")

	rel, err := filepath.Rel(dataDir, cleanPath)
	if err != nil {
		return "", entity.ErrBadArguments
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", entity.ErrBadArguments
	}

	return cleanPath, nil
}

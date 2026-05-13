package rest

import (
	"log/slog"
	"net/http"

	"test.task.log.server/internal/entity"
)

const maxRequestBodyBytes = 1 << 20

type logParser interface {
	Parse(path string) (int, error)
	GetLog(logID int) (entity.Log, error)
	GetTopology(logID int) (entity.Topology, error)
	GetNode(nodeID int) (entity.Node, error)
	GetPorts(nodeID int) ([]entity.Port, error)
}

func NewParseHandler(log *slog.Logger, parse logParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req ParseRequest
		if err := decodeJSON(w, r, &req); err != nil {
			writeError(w, http.StatusBadRequest, entity.ErrBadArguments.Error())
			return
		}

		path, err := validateDataPath(req.Path)
		if err != nil {
			writeError(w, http.StatusBadRequest, entity.ErrBadArguments.Error())
			return
		}

		logID, err := parse.Parse(path)
		if err != nil {
			handleServiceError(log, w, err, "cannot parse log")
			return
		}

		writeJSON(log, w, http.StatusOK, ParseResponse{
			LogID: logID,
		})
	}
}

func NewTopologyHandler(log *slog.Logger, parse logParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		logID, err := positiveIntPathValue(r, "log_id")
		if err != nil {
			writeError(w, http.StatusBadRequest, entity.ErrBadArguments.Error())
			return
		}

		topology, err := parse.GetTopology(logID)
		if err != nil {
			handleServiceError(log, w, err, "cannot get topology")
			return
		}

		writeJSON(log, w, http.StatusOK, toTopologyResponse(topology))
	}
}

func NewNodeHandler(log *slog.Logger, parse logParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		nodeID, err := positiveIntPathValue(r, "node_id")
		if err != nil {
			writeError(w, http.StatusBadRequest, entity.ErrBadArguments.Error())
			return
		}

		node, err := parse.GetNode(nodeID)
		if err != nil {
			handleServiceError(log, w, err, "cannot get node")
			return
		}

		writeJSON(log, w, http.StatusOK, toNodeResponse(node))
	}
}

func NewPortsHandler(log *slog.Logger, parse logParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		nodeID, err := positiveIntPathValue(r, "node_id")
		if err != nil {
			writeError(w, http.StatusBadRequest, entity.ErrBadArguments.Error())
			return
		}

		ports, err := parse.GetPorts(nodeID)
		if err != nil {
			handleServiceError(log, w, err, "cannot get ports")
			return
		}

		writeJSON(log, w, http.StatusOK, PortsResponse{
			NodeID: nodeID,
			Ports:  toPortResponses(ports),
			Total:  len(ports),
		})
	}
}

func NewLogHandler(log *slog.Logger, parse logParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		logID, err := positiveIntPathValue(r, "log_id")
		if err != nil {
			writeError(w, http.StatusBadRequest, entity.ErrBadArguments.Error())
			return
		}

		logEntity, err := parse.GetLog(logID)
		if err != nil {
			handleServiceError(log, w, err, "cannot get log")
			return
		}

		writeJSON(log, w, http.StatusOK, toLogResponse(logEntity))
	}
}

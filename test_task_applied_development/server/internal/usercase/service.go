package usercase

import (
	"log/slog"

	"test.task.log.server/internal/entity"
)

type logRepository interface {
	CreateLog(path string) (int, error)
	UpdateLogStatus(logID int, status string, errText string) error

	SaveParsedData(logID int, data entity.ParsedData) error

	GetLog(logID int) (entity.Log, error)
	GetTopology(logID int) (entity.Topology, error)
	GetNode(nodeID int) (entity.Node, error)
	GetPortsByNode(nodeID int) ([]entity.Port, error)
}

type service struct {
	log *slog.Logger
	rep logRepository
}

func New(log *slog.Logger, rep logRepository) *service {
	return &service{
		log: log,
		rep: rep,
	}
}

func (s *service) Parse(path string) (int, error) {
	logID, err := s.rep.CreateLog(path)
	if err != nil {
		return 0, err
	}

	if err := s.rep.UpdateLogStatus(logID, "parsing", ""); err != nil {
		return logID, err
	}

	data, err := parseDirectory(path)
	if err != nil {
		s.markFailed(logID, err)
		return logID, err
	}

	if err := s.rep.SaveParsedData(logID, data); err != nil {
		s.markFailed(logID, err)
		return logID, err
	}

	if err := s.rep.UpdateLogStatus(logID, "parsed", ""); err != nil {
		return logID, err
	}

	return logID, nil
}

func (s *service) GetTopology(logID int) (entity.Topology, error) {
	return s.rep.GetTopology(logID)
}

func (s *service) GetNode(nodeID int) (entity.Node, error) {
	return s.rep.GetNode(nodeID)
}

func (s *service) GetPorts(nodeID int) ([]entity.Port, error) {
	return s.rep.GetPortsByNode(nodeID)
}

func (s *service) GetPortsByNode(nodeID int) ([]entity.Port, error) {
	return s.GetPorts(nodeID)
}

func (s *service) GetLog(logID int) (entity.Log, error) {
	return s.rep.GetLog(logID)
}

func (s *service) markFailed(logID int, parseErr error) {
	if err := s.rep.UpdateLogStatus(logID, "failed", parseErr.Error()); err != nil {
		s.log.Error("cannot update log status to failed", "log_id", logID, "err", err)
	}
}

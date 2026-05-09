package dns

import (
	"log/slog"
	"os"
	"strings"

	"test.task.dns/internal/entity"
)

type dnsManager struct {
	path string
	log  *slog.Logger
}

func New(log *slog.Logger, path string) *dnsManager {
	return &dnsManager{
		log:  log,
		path: path,
	}
}

func (d *dnsManager) AddDNS(ip string) error {
	data, err := os.ReadFile(d.path)
	if err != nil {
		d.log.Error("cant read file in adapter dns", "err", err)
		return err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)

		if len(fields) >= 2 && fields[0] == "nameserver" && fields[1] == ip {
			return entity.ErrAlreadyExist
		}
	}

	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	content += "nameserver " + ip + "\n"

	if err := os.WriteFile(d.path, []byte(content), 0644); err != nil {
		d.log.Error("cant WriteFile in adapter dns", "err", err)
		return err
	}
	return nil
}

func (d *dnsManager) RemoveDNS(ip string) error {
	data, err := os.ReadFile(d.path)
	if err != nil {
		d.log.Error("cant read file in adapter dns", "err", err)
		return err
	}

	lines := strings.Split(string(data), "\n")

	var result []string

	notFound := true

	for _, line := range lines {
		fields := strings.Fields(line)

		if len(fields) >= 2 && fields[0] == "nameserver" && fields[1] == ip {
			notFound = false
			continue
		}

		result = append(result, line)
	}
	if notFound {
		return entity.ErrNotFoundDNS
	}

	newContent := strings.Join(result, "\n")

	if err := os.WriteFile(d.path, []byte(newContent), 0644); err != nil {
		d.log.Error("cant WriteFile in adapter dns", "err", err)
		return err
	}
	return nil

}

func (d *dnsManager) ListDNS() ([]string, error) {
	data, err := os.ReadFile(d.path)
	if err != nil {
		d.log.Error("cant read file in adapter dns", "err", err)
		return nil, err
	}

	lines := strings.Split(string(data), "\n")

	var servers []string

	for _, line := range lines {
		fields := strings.Fields(line)

		if len(fields) >= 2 && fields[0] == "nameserver" {
			servers = append(servers, fields[1])
		}
	}

	return servers, nil
}

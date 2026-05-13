package usercase

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"test.task.log.server/internal/entity"
)

func parseDirectory(dir string) (entity.ParsedData, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return entity.ParsedData{}, fmt.Errorf("%w: empty directory path", entity.ErrBadArguments)
	}

	info, err := os.Stat(dir)
	if err != nil {
		return entity.ParsedData{}, fmt.Errorf("%w: cannot stat directory: %v", entity.ErrBadArguments, err)
	}

	if !info.IsDir() {
		return entity.ParsedData{}, fmt.Errorf("%w: path is not a directory: %s", entity.ErrBadArguments, dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return entity.ParsedData{}, fmt.Errorf("%w: cannot read directory: %v", entity.ErrBadArguments, err)
	}

	var mainFile string
	var mainContent string

	var nodeInfoFile string
	var nodeInfoContent string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())

		contentBytes, err := os.ReadFile(fullPath)
		if err != nil {
			return entity.ParsedData{}, fmt.Errorf("%w: cannot read file %s: %v", entity.ErrBadArguments, fullPath, err)
		}

		content := string(contentBytes)

		isMainFile := strings.Contains(content, "START_NODES") &&
			strings.Contains(content, "START_PORTS")

		isNodeInfoFile := strings.Contains(content, "SW_GUID=")

		if isMainFile && isNodeInfoFile {
			return entity.ParsedData{}, fmt.Errorf("%w: file %s contains both main sections and SW_GUID sections", entity.ErrParseFailed, fullPath)
		}

		if isMainFile {
			if mainFile != "" {
				return entity.ParsedData{}, fmt.Errorf("%w: more than one main file found", entity.ErrParseFailed)
			}

			mainFile = fullPath
			mainContent = content
		}

		if isNodeInfoFile {
			if nodeInfoFile != "" {
				return entity.ParsedData{}, fmt.Errorf("%w: more than one node info file found", entity.ErrParseFailed)
			}

			nodeInfoFile = fullPath
			nodeInfoContent = content
		}
	}

	if mainFile == "" {
		return entity.ParsedData{}, fmt.Errorf("%w: main file with START_NODES and START_PORTS not found", entity.ErrParseFailed)
	}

	if nodeInfoFile == "" {
		return entity.ParsedData{}, fmt.Errorf("%w: node info file with SW_GUID not found", entity.ErrParseFailed)
	}

	data, err := parseMainFile(mainContent)
	if err != nil {
		return entity.ParsedData{}, fmt.Errorf("%w: cannot parse main file %s: %v", entity.ErrParseFailed, mainFile, err)
	}

	nodeInfos, err := parseNodeInfoFile(nodeInfoContent)
	if err != nil {
		return entity.ParsedData{}, fmt.Errorf("%w: cannot parse node info file %s: %v", entity.ErrParseFailed, nodeInfoFile, err)
	}

	data.NodeInfos = nodeInfos

	if err := validateReferences(data); err != nil {
		return entity.ParsedData{}, fmt.Errorf("%w: %v", entity.ErrParseFailed, err)
	}

	mergeParsedData(&data)

	return data, nil
}

func parseMainFile(content string) (entity.ParsedData, error) {
	sections, err := splitSections(content)
	if err != nil {
		return entity.ParsedData{}, err
	}

	nodesLines, ok := sections["NODES"]
	if !ok {
		return entity.ParsedData{}, fmt.Errorf("START_NODES section not found")
	}

	portsLines, ok := sections["PORTS"]
	if !ok {
		return entity.ParsedData{}, fmt.Errorf("START_PORTS section not found")
	}

	nodes, err := parseNodesSection(nodesLines)
	if err != nil {
		return entity.ParsedData{}, err
	}

	ports, err := parsePortsSection(portsLines)
	if err != nil {
		return entity.ParsedData{}, err
	}

	var switchInfos []entity.SwitchInfo
	if switchLines, ok := sections["SWITCHES"]; ok {
		switchInfos, err = parseSwitchesSection(switchLines)
		if err != nil {
			return entity.ParsedData{}, err
		}
	}

	var systemInfos []entity.SystemInfo
	if systemLines, ok := sections["SYSTEM_GENERAL_INFORMATION"]; ok {
		systemInfos, err = parseSystemInfoSection(systemLines)
		if err != nil {
			return entity.ParsedData{}, err
		}
	}

	return entity.ParsedData{
		Nodes:       nodes,
		Ports:       ports,
		SwitchInfos: switchInfos,
		SystemInfos: systemInfos,
	}, nil
}

func splitSections(content string) (map[string][]string, error) {
	sections := make(map[string][]string)

	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 1024), 20*1024*1024)

	var currentSection string
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "START_") {
			if currentSection != "" {
				return nil, fmt.Errorf("line %d: section %s was not closed", lineNumber, currentSection)
			}

			name := strings.TrimPrefix(line, "START_")
			name = strings.TrimSpace(name)
			if name == "" {
				return nil, fmt.Errorf("line %d: empty section name", lineNumber)
			}

			if _, exists := sections[name]; exists {
				return nil, fmt.Errorf("line %d: duplicate section %s", lineNumber, name)
			}

			currentSection = name
			sections[currentSection] = nil
			continue
		}

		if strings.HasPrefix(line, "END_") {
			name := strings.TrimPrefix(line, "END_")
			name = strings.TrimSpace(name)

			if currentSection == "" {
				return nil, fmt.Errorf("line %d: unexpected END_%s", lineNumber, name)
			}

			if name != currentSection {
				return nil, fmt.Errorf("line %d: expected END_%s, got END_%s", lineNumber, currentSection, name)
			}

			currentSection = ""
			continue
		}

		if currentSection != "" {
			sections[currentSection] = append(sections[currentSection], line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if currentSection != "" {
		return nil, fmt.Errorf("section %s was not closed", currentSection)
	}

	return sections, nil
}

func parseNodesSection(lines []string) ([]entity.Node, error) {
	header, rows, err := readCSVSection("NODES", lines)
	if err != nil {
		return nil, err
	}

	idx := headerIndex(header)

	seenGUIDs := make(map[string]struct{})
	nodes := make([]entity.Node, 0, len(rows))

	for rowNumber, row := range rows {
		nodeDesc, err := requiredValue(row, idx, "NodeDesc")
		if err != nil {
			return nil, fmt.Errorf("NODES row %d: %w", rowNumber+1, err)
		}

		numPortsRaw, err := requiredValue(row, idx, "NumPorts")
		if err != nil {
			return nil, fmt.Errorf("NODES row %d: %w", rowNumber+1, err)
		}

		nodeTypeRaw, err := requiredValue(row, idx, "NodeType")
		if err != nil {
			return nil, fmt.Errorf("NODES row %d: %w", rowNumber+1, err)
		}

		numPorts, err := strconv.Atoi(numPortsRaw)
		if err != nil {
			return nil, fmt.Errorf("NODES row %d: invalid NumPorts %q", rowNumber+1, numPortsRaw)
		}

		nodeType, err := strconv.Atoi(nodeTypeRaw)
		if err != nil {
			return nil, fmt.Errorf("NODES row %d: invalid NodeType %q", rowNumber+1, nodeTypeRaw)
		}

		nodeGUIDRaw, err := requiredValue(row, idx, "NodeGUID")
		if err != nil {
			return nil, fmt.Errorf("NODES row %d: %w", rowNumber+1, err)
		}

		nodeGUID := ensure0xGUID(nodeGUIDRaw)
		if nodeGUID == "" {
			return nil, fmt.Errorf("NODES row %d: empty NodeGUID", rowNumber+1)
		}

		normalizedGUID := normalizeGUID(nodeGUID)
		if _, exists := seenGUIDs[normalizedGUID]; exists {
			return nil, fmt.Errorf("NODES row %d: duplicate NodeGUID %s", rowNumber+1, nodeGUID)
		}
		seenGUIDs[normalizedGUID] = struct{}{}

		systemImageGUID, err := requiredValue(row, idx, "SystemImageGUID")
		if err != nil {
			return nil, fmt.Errorf("NODES row %d: %w", rowNumber+1, err)
		}

		portGUID, err := requiredValue(row, idx, "PortGUID")
		if err != nil {
			return nil, fmt.Errorf("NODES row %d: %w", rowNumber+1, err)
		}

		node := entity.Node{
			NodeDesc:        nodeDesc,
			NumPorts:        numPorts,
			NodeType:        nodeType,
			ClassVersion:    optionalValue(row, idx, "ClassVersion"),
			BaseVersion:     optionalValue(row, idx, "BaseVersion"),
			SystemImageGUID: ensure0xGUID(systemImageGUID),
			NodeGUID:        nodeGUID,
			PortGUID:        ensure0xGUID(portGUID),
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func parsePortsSection(lines []string) ([]entity.Port, error) {
	header, rows, err := readCSVSection("PORTS", lines)
	if err != nil {
		return nil, err
	}

	idx := headerIndex(header)

	seenPorts := make(map[string]struct{})
	ports := make([]entity.Port, 0, len(rows))

	for rowNumber, row := range rows {
		nodeGUIDRaw, err := requiredValue(row, idx, "NodeGuid")
		if err != nil {
			return nil, fmt.Errorf("PORTS row %d: %w", rowNumber+1, err)
		}

		portGUIDRaw, err := requiredValue(row, idx, "PortGuid")
		if err != nil {
			return nil, fmt.Errorf("PORTS row %d: %w", rowNumber+1, err)
		}

		portNumRaw, err := requiredValue(row, idx, "PortNum")
		if err != nil {
			return nil, fmt.Errorf("PORTS row %d: %w", rowNumber+1, err)
		}

		portNum, err := strconv.Atoi(portNumRaw)
		if err != nil {
			return nil, fmt.Errorf("PORTS row %d: invalid PortNum %q", rowNumber+1, portNumRaw)
		}

		nodeGUID := ensure0xGUID(nodeGUIDRaw)
		portGUID := ensure0xGUID(portGUIDRaw)

		portKey := normalizeGUID(nodeGUID) + ":" + strconv.Itoa(portNum)
		if _, exists := seenPorts[portKey]; exists {
			return nil, fmt.Errorf(
				"PORTS row %d: duplicate port node_guid=%s port_num=%d",
				rowNumber+1,
				nodeGUID,
				portNum,
			)
		}
		seenPorts[portKey] = struct{}{}

		port := entity.Port{
			NodeGUID:             nodeGUID,
			PortGUID:             portGUID,
			PortNum:              portNum,
			LID:                  optionalValue(row, idx, "LID"),
			LocalPortNum:         optionalValue(row, idx, "LocalPortNum"),
			LinkWidthActive:      optionalValue(row, idx, "LinkWidthActv"),
			LinkWidthSupported:   optionalValue(row, idx, "LinkWidthSup"),
			LinkSpeedActive:      optionalValue(row, idx, "LinkSpeedActv"),
			LinkSpeedSupported:   optionalValue(row, idx, "LinkSpeedSup"),
			PortState:            optionalValue(row, idx, "PortState"),
			PortPhyState:         optionalValue(row, idx, "PortPhyState"),
			MTUCap:               optionalValue(row, idx, "MTUCap"),
			LinkRoundTripLatency: optionalValue(row, idx, "LinkRoundTripLatency"),
		}

		ports = append(ports, port)
	}

	return ports, nil
}

func parseSwitchesSection(lines []string) ([]entity.SwitchInfo, error) {
	header, rows, err := readCSVSection("SWITCHES", lines)
	if err != nil {
		return nil, err
	}

	idx := headerIndex(header)

	seenGUIDs := make(map[string]struct{})
	result := make([]entity.SwitchInfo, 0, len(rows))

	for rowNumber, row := range rows {
		nodeGUIDRaw, err := requiredValue(row, idx, "NodeGUID")
		if err != nil {
			return nil, fmt.Errorf("SWITCHES row %d: %w", rowNumber+1, err)
		}

		nodeGUID := ensure0xGUID(nodeGUIDRaw)
		normalizedGUID := normalizeGUID(nodeGUID)

		if _, exists := seenGUIDs[normalizedGUID]; exists {
			return nil, fmt.Errorf("SWITCHES row %d: duplicate NodeGUID %s", rowNumber+1, nodeGUID)
		}
		seenGUIDs[normalizedGUID] = struct{}{}

		fields := make(map[string]string)

		for i, column := range header {
			if strings.EqualFold(column, "NodeGUID") {
				continue
			}

			if i >= len(row) {
				continue
			}

			fields[column] = row[i]
		}

		result = append(result, entity.SwitchInfo{
			NodeGUID: nodeGUID,
			Fields:   fields,
		})
	}

	return result, nil
}

func parseSystemInfoSection(lines []string) ([]entity.SystemInfo, error) {
	header, rows, err := readCSVSection("SYSTEM_GENERAL_INFORMATION", lines)
	if err != nil {
		return nil, err
	}

	idx := headerIndex(header)

	seenGUIDs := make(map[string]struct{})
	result := make([]entity.SystemInfo, 0, len(rows))

	for rowNumber, row := range rows {
		nodeGUIDRaw, err := requiredValue(row, idx, "NodeGuid")
		if err != nil {
			return nil, fmt.Errorf("SYSTEM_GENERAL_INFORMATION row %d: %w", rowNumber+1, err)
		}

		nodeGUID := ensure0xGUID(nodeGUIDRaw)
		normalizedGUID := normalizeGUID(nodeGUID)

		if _, exists := seenGUIDs[normalizedGUID]; exists {
			return nil, fmt.Errorf("SYSTEM_GENERAL_INFORMATION row %d: duplicate NodeGuid %s", rowNumber+1, nodeGUID)
		}
		seenGUIDs[normalizedGUID] = struct{}{}

		result = append(result, entity.SystemInfo{
			NodeGUID:     nodeGUID,
			SerialNumber: optionalValue(row, idx, "SerialNumber"),
			PartNumber:   optionalValue(row, idx, "PartNumber"),
			Revision:     optionalValue(row, idx, "Revision"),
			ProductName:  optionalValue(row, idx, "ProductName"),
		})
	}

	return result, nil
}

func parseNodeInfoFile(content string) ([]entity.NodeInfo, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 1024), 20*1024*1024)

	var result []entity.NodeInfo

	var currentGUID string
	var currentFields map[string]string

	seenGUIDs := make(map[string]struct{})

	finishCurrent := func() error {
		if currentGUID == "" {
			return nil
		}

		if len(currentFields) == 0 {
			return fmt.Errorf("SW_GUID=%s has no fields", currentGUID)
		}

		result = append(result, entity.NodeInfo{
			NodeGUID: currentGUID,
			Fields:   currentFields,
		})

		currentGUID = ""
		currentFields = nil

		return nil
	}

	lineNumber := 0

	for scanner.Scan() {
		lineNumber++

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if isSeparatorLine(line) {
			continue
		}

		if strings.HasPrefix(line, "SW_GUID=") {
			if err := finishCurrent(); err != nil {
				return nil, err
			}

			rawGUID := strings.TrimSpace(strings.TrimPrefix(line, "SW_GUID="))
			if rawGUID == "" {
				return nil, fmt.Errorf("node info line %d: empty SW_GUID", lineNumber)
			}

			currentGUID = ensure0xGUID(rawGUID)

			normalizedGUID := normalizeGUID(currentGUID)
			if _, exists := seenGUIDs[normalizedGUID]; exists {
				return nil, fmt.Errorf("node info line %d: duplicate SW_GUID %s", lineNumber, currentGUID)
			}
			seenGUIDs[normalizedGUID] = struct{}{}

			currentFields = make(map[string]string)
			continue
		}

		if currentGUID == "" {
			return nil, fmt.Errorf("node info line %d: field before SW_GUID", lineNumber)
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("node info line %d: invalid line %q", lineNumber, line)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" {
			return nil, fmt.Errorf("node info line %d: empty field name", lineNumber)
		}

		if _, exists := currentFields[key]; exists {
			return nil, fmt.Errorf("node info line %d: duplicate field %s for %s", lineNumber, key, currentGUID)
		}

		currentFields[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if err := finishCurrent(); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("node info file has no SW_GUID blocks")
	}

	return result, nil
}

func readCSVSection(sectionName string, lines []string) ([]string, [][]string, error) {
	if len(lines) == 0 {
		return nil, nil, fmt.Errorf("%s section is empty", sectionName)
	}

	reader := csv.NewReader(strings.NewReader(strings.Join(lines, "\n")))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	var records [][]string

	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, nil, fmt.Errorf("%s csv parse error: %w", sectionName, err)
		}

		if len(record) == 1 && strings.TrimSpace(record[0]) == "" {
			continue
		}

		for i := range record {
			record[i] = strings.TrimSpace(record[i])
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return nil, nil, fmt.Errorf("%s section has no header", sectionName)
	}

	header := records[0]
	if len(header) == 0 {
		return nil, nil, fmt.Errorf("%s section has empty header", sectionName)
	}

	rows := records[1:]

	for i, row := range rows {
		if len(row) != len(header) {
			return nil, nil, fmt.Errorf(
				"%s row %d: columns count mismatch: got %d, want %d",
				sectionName,
				i+1,
				len(row),
				len(header),
			)
		}
	}

	return header, rows, nil
}

func headerIndex(header []string) map[string]int {
	result := make(map[string]int, len(header))

	for i, name := range header {
		result[strings.ToLower(strings.TrimSpace(name))] = i
	}

	return result
}

func requiredValue(row []string, idx map[string]int, name string) (string, error) {
	value := optionalValue(row, idx, name)
	if value == "" {
		return "", fmt.Errorf("required field %s is empty or missing", name)
	}

	return value, nil
}

func optionalValue(row []string, idx map[string]int, name string) string {
	i, ok := idx[strings.ToLower(name)]
	if !ok {
		return ""
	}

	if i >= len(row) {
		return ""
	}

	return strings.TrimSpace(row[i])
}

func validateReferences(data entity.ParsedData) error {
	nodeGUIDs := make(map[string]struct{})

	for _, node := range data.Nodes {
		guid := normalizeGUID(node.NodeGUID)
		if guid == "" {
			return fmt.Errorf("node has empty NodeGUID")
		}

		nodeGUIDs[guid] = struct{}{}
	}

	for _, port := range data.Ports {
		if _, exists := nodeGUIDs[normalizeGUID(port.NodeGUID)]; !exists {
			return fmt.Errorf("port references unknown NodeGUID %s", port.NodeGUID)
		}
	}

	for _, switchInfo := range data.SwitchInfos {
		if _, exists := nodeGUIDs[normalizeGUID(switchInfo.NodeGUID)]; !exists {
			return fmt.Errorf("switch info references unknown NodeGUID %s", switchInfo.NodeGUID)
		}
	}

	for _, systemInfo := range data.SystemInfos {
		if _, exists := nodeGUIDs[normalizeGUID(systemInfo.NodeGUID)]; !exists {
			return fmt.Errorf("system info references unknown NodeGUID %s", systemInfo.NodeGUID)
		}
	}

	for _, nodeInfo := range data.NodeInfos {
		if _, exists := nodeGUIDs[normalizeGUID(nodeInfo.NodeGUID)]; !exists {
			return fmt.Errorf("node info references unknown NodeGUID %s", nodeInfo.NodeGUID)
		}
	}

	return nil
}

func mergeParsedData(data *entity.ParsedData) {
	systemInfoByGUID := make(map[string]entity.SystemInfo)
	for _, info := range data.SystemInfos {
		systemInfoByGUID[normalizeGUID(info.NodeGUID)] = info
	}

	switchInfoByGUID := make(map[string]entity.SwitchInfo)
	for _, info := range data.SwitchInfos {
		switchInfoByGUID[normalizeGUID(info.NodeGUID)] = info
	}

	nodeInfoByGUID := make(map[string]entity.NodeInfo)
	for _, info := range data.NodeInfos {
		nodeInfoByGUID[normalizeGUID(info.NodeGUID)] = info
	}

	for i := range data.Nodes {
		nodeGUID := normalizeGUID(data.Nodes[i].NodeGUID)

		if systemInfo, ok := systemInfoByGUID[nodeGUID]; ok {
			systemInfoCopy := systemInfo
			data.Nodes[i].SystemInfo = &systemInfoCopy
		}

		if switchInfo, ok := switchInfoByGUID[nodeGUID]; ok {
			data.Nodes[i].SwitchInfo = cloneMap(switchInfo.Fields)
		}

		if nodeInfo, ok := nodeInfoByGUID[nodeGUID]; ok {
			data.Nodes[i].NodeInfo = cloneMap(nodeInfo.Fields)
		}
	}
}

func ensure0xGUID(guid string) string {
	guid = strings.ToLower(strings.TrimSpace(guid))
	if guid == "" {
		return ""
	}

	guid = strings.TrimPrefix(guid, "0x")

	return "0x" + guid
}

func normalizeGUID(guid string) string {
	guid = strings.ToLower(strings.TrimSpace(guid))
	guid = strings.TrimPrefix(guid, "0x")

	return guid
}

func isSeparatorLine(line string) bool {
	if line == "" {
		return false
	}

	for _, r := range line {
		if r != '-' {
			return false
		}
	}

	return true
}

func cloneMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}

	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}

	return dst
}

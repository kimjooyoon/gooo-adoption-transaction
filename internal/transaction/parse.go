package transaction

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func parseSource(path string) (SourceDecl, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SourceDecl{}, err
	}
	source := SourceDecl{SourceDigest: DigestBytes(data)}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "gooo":
			if len(fields) != 3 || fields[1] != "adoption_transaction" || fields[2] != "v1" {
				return SourceDecl{}, fmt.Errorf("line %d: invalid gooo header", lineNumber)
			}
			source.Schema = SourceSchema
			source.Version = fields[2]
		case "denominator":
			values, err := keyValues(fields[1:])
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.DenominatorID = values["id"]
			source.CellCount, err = integer(values, "cell_count")
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
		case "authority":
			values, err := keyValues(fields[1:])
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.Authority.RepositoryWrites, err = integer(values, "repository_writes")
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.Authority.LocalTestExecutions, err = integer(values, "local_test_executions")
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.Authority.CrossProjectRequiredGates, err = integer(values, "cross_project_required_gates")
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
		case "precedence":
			if len(fields) != 2 {
				return SourceDecl{}, fmt.Errorf("line %d: invalid precedence", lineNumber)
			}
			source.Precedence = strings.Split(fields[1], ">")
		case "unknown_fields":
			if len(fields) != 2 {
				return SourceDecl{}, fmt.Errorf("line %d: invalid unknown_fields", lineNumber)
			}
			source.UnknownFields = strings.Split(fields[1], ",")
		case "activity":
			values, err := keyValues(fields[1:])
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			ordinal, err := integer(values, "ordinal")
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.Activities = append(source.Activities, Activity{
				Ordinal: ordinal, ID: values["id"], SemanticEdge: values["edge"], Phase: values["phase"],
			})
		default:
			return SourceDecl{}, fmt.Errorf("line %d: unknown declaration %q", lineNumber, fields[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return SourceDecl{}, err
	}
	return source, nil
}

func keyValues(fields []string) (map[string]string, error) {
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid key/value %q", field)
		}
		values[parts[0]] = strings.Trim(parts[1], "\"")
	}
	return values, nil
}

func integer(values map[string]string, key string) (int, error) {
	value, ok := values[key]
	if !ok {
		return 0, fmt.Errorf("missing %s", key)
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q", key, value)
	}
	return number, nil
}

func loadContract(path string) (Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, err
	}
	var contract Contract
	if err := json.Unmarshal(data, &contract); err != nil {
		return Contract{}, fmt.Errorf("decode contract: %w", err)
	}
	if contract.Schema != ContractSchema {
		return Contract{}, fmt.Errorf("unexpected contract schema %q", contract.Schema)
	}
	return contract, nil
}

func validateDeclarations(source SourceDecl, contract Contract) error {
	if source.Schema != SourceSchema || source.Version != "v1" || contract.Version != "v1" ||
		source.DenominatorID != contract.ID || source.CellCount != FixedCells ||
		contract.CellCount != FixedCells || !contract.Fixed {
		return fmt.Errorf("fixed denominator declaration mismatch")
	}
	if len(source.Activities) != FixedCells || len(contract.Activities) != FixedCells {
		return fmt.Errorf("expected exactly %d activities", FixedCells)
	}
	if !sameStrings(source.UnknownFields, []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"}) {
		return fmt.Errorf("UNKNOWN six-field contract mismatch")
	}
	if !sameStrings(source.Precedence, []string{"REFUTED", "UNKNOWN", "CLOSED"}) {
		return fmt.Errorf("resolution precedence mismatch")
	}
	if source.Authority.RepositoryWrites != 0 || source.Authority.LocalTestExecutions != 0 || source.Authority.CrossProjectRequiredGates != 0 {
		return fmt.Errorf("authority declaration must be zero")
	}
	for index := 0; index < FixedCells; index++ {
		left, right := source.Activities[index], contract.Activities[index]
		if left.Ordinal != index+1 || right.Ordinal != index+1 || left.ID == "" ||
			left.ID != right.ID || left.SemanticEdge != right.SemanticEdge || left.Phase != right.Phase {
			return fmt.Errorf("activity %d does not match fixed contract", index+1)
		}
	}
	if contract.Phases != (PhaseCounts{Prepare: 4, Authorize: 4, Commit: 4, VerifyOrAbort: 4}) {
		return fmt.Errorf("phase denominator must be four by four")
	}
	return nil
}

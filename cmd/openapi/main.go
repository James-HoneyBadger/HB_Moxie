//go:generate go run .

// Command openapi generates the machine-readable API description and the
// TypeScript models used by the frontend.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/James-HoneyBadger/ratel-ledger/pkg/models"
)

const (
	openAPIFileName    = "docs/openapi.json"
	typescriptFileName = "src/models/generated/api.ts"
)

type apiSchemaTypes struct {
	BackupRequest        models.BackupRequest
	BackupListResponse   models.BackupListResponse
	BackupRestoreRequest models.BackupRestoreRequest
	BackupDeleteRequest  models.BackupDeleteRequest

	AuthResponse          models.AuthResponse
	UserLoginRequest      models.UserLoginRequest
	TwoFactorLoginRequest models.TwoFactorLoginRequest

	TokenGenerateAPIRequest  models.TokenGenerateAPIRequest
	TokenGenerateAPIResponse models.TokenGenerateAPIResponse
	TokenGenerateMCPRequest  models.TokenGenerateMCPRequest
	TokenGenerateMCPResponse models.TokenGenerateMCPResponse

	TransactionStatisticRequest                 models.TransactionStatisticRequest
	TransactionStatisticResponse                models.TransactionStatisticResponse
	TransactionStatisticTrendsRequest           models.TransactionStatisticTrendsRequest
	TransactionStatisticTrendsResponseItem      models.TransactionStatisticTrendsResponseItem
	TransactionStatisticAssetTrendsRequest      models.TransactionStatisticAssetTrendsRequest
	TransactionStatisticAssetTrendsResponseItem models.TransactionStatisticAssetTrendsResponseItem
	TransactionAmountsRequest                   models.TransactionAmountsRequest
	TransactionAmountsResponseItem              models.TransactionAmountsResponseItem

	TransactionSplitUpdateRequest models.TransactionSplitUpdateRequest
	TransactionSplitsResponse     models.TransactionSplitsResponse
	TransactionSplitIndexItem     models.TransactionSplitIndexItem
	SharedExpenseSummaryItem      models.SharedExpenseSummaryItem
	SharedExpenseEntryResponse    models.SharedExpenseEntryResponse
}

func main() {
	check := flag.Bool("check", false, "verify that generated files are up to date")
	flag.Parse()

	repositoryRoot, err := findRepositoryRoot()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	openAPIDocument, typescriptDocument, err := generate(repositoryRoot)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *check {
		if err := checkFile(filepath.Join(repositoryRoot, openAPIFileName), openAPIDocument); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := checkFile(filepath.Join(repositoryRoot, typescriptFileName), typescriptDocument); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Println("OpenAPI and TypeScript models are up to date")
		return
	}

	if err := writeFile(filepath.Join(repositoryRoot, openAPIFileName), openAPIDocument); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := writeFile(filepath.Join(repositoryRoot, typescriptFileName), typescriptDocument); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Generated", openAPIFileName, "and", typescriptFileName)
}

func generate(repositoryRoot string) ([]byte, []byte, error) {
	reflector := &jsonschema.Reflector{
		FieldNameTag: "json",
		Anonymous:    true,
	}

	rootSchema := reflector.Reflect(&apiSchemaTypes{})
	delete(rootSchema.Definitions, "apiSchemaTypes")

	applyJSONStringOptions(rootSchema.Definitions)

	for _, definition := range rootSchema.Definitions {
		rewriteReferences(definition)
	}

	version, err := readPackageVersion(repositoryRoot)

	if err != nil {
		return nil, nil, err
	}

	document := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       "Honey Ledger API",
			"description": "Self-hosted personal finance API exposed by the local Honey Ledger backend.",
			"version":     version,
		},
		"paths": generatedPaths(),
		"components": map[string]any{
			"schemas": rootSchema.Definitions,
		},
	}

	openAPIDocument, err := json.MarshalIndent(document, "", "  ")

	if err != nil {
		return nil, nil, err
	}

	openAPIDocument = append(openAPIDocument, '\n')

	typescriptDocument := []byte(generateTypeScript(rootSchema.Definitions))

	return openAPIDocument, typescriptDocument, nil
}

func generatedPaths() map[string]any {
	return map[string]any{
		"/api/authorize.json": map[string]any{
			"post": jsonOperation("authorize", "Log in with username or email and password", "UserLoginRequest", "AuthResponse"),
		},
		"/api/auto_login.json": map[string]any{
			"post": responseOperation("autoLogin", "Log in the single desktop user without a password", referenceSchema("AuthResponse")),
		},
		"/api/v1/backups/list.json": map[string]any{
			"get": responseOperation("listBackups", "List backup archives", arraySchema("BackupListResponse")),
		},
		"/api/v1/backups/create.json": map[string]any{
			"post": jsonOperation("createBackup", "Create a SQLite snapshot backup", "BackupRequest", "BackupListResponse"),
		},
		"/api/v1/backups/restore.json": map[string]any{
			"post": jsonOperation("restoreBackup", "Restore a SQLite snapshot backup", "BackupRestoreRequest", "boolean"),
		},
		"/api/v1/backups/delete.json": map[string]any{
			"post": jsonOperation("deleteBackup", "Delete a backup archive", "BackupDeleteRequest", "boolean"),
		},
		"/api/v1/tokens/generate/api.json": map[string]any{
			"post": jsonOperation("generateApiToken", "Generate an API token", "TokenGenerateAPIRequest", "TokenGenerateAPIResponse"),
		},
		"/api/v1/tokens/generate/mcp.json": map[string]any{
			"post": jsonOperation("generateMcpToken", "Generate an MCP token", "TokenGenerateMCPRequest", "TokenGenerateMCPResponse"),
		},
		"/api/v1/transactions/splits/index.json": map[string]any{
			"get": withParameters(
				responseOperation("getTransactionSplitIndex", "List the split state of the transactions in a time range", arraySchema("TransactionSplitIndexItem")),
				optionalQueryParameter("min_time", "Minimum transaction time in milliseconds"),
				optionalQueryParameter("max_time", "Maximum transaction time in milliseconds, 0 for no upper bound"),
			),
		},
		"/api/v1/transactions/splits/list.json": map[string]any{
			"get": withParameters(
				responseOperation("listTransactionSplits", "List the shares of a transaction owed by other people", referenceSchema("TransactionSplitsResponse")),
				queryParameter("transactionId", "Id of the transaction"),
			),
		},
		"/api/v1/transactions/splits/update.json": map[string]any{
			"post": jsonOperation("updateTransactionSplits", "Replace the shares of an expense transaction", "TransactionSplitUpdateRequest", "TransactionSplitsResponse"),
		},
		"/api/v1/transactions/splits/summary.json": map[string]any{
			"get": responseOperation("getSharedExpenseSummary", "Summarize what each person owes the user, grouped by currency", arraySchema("SharedExpenseSummaryItem")),
		},
		"/api/v1/transactions/splits/person.json": map[string]any{
			"get": withParameters(
				responseOperation("getSharedExpensesByPerson", "List the shared expenses owed by one person", arraySchema("SharedExpenseEntryResponse")),
				queryParameter("personName", "Name of the person"),
			),
		},
	}
}

func withParameters(operation map[string]any, parameters ...map[string]any) map[string]any {
	operation["parameters"] = parameters
	return operation
}

func queryParameter(name string, description string) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "query",
		"required":    true,
		"description": description,
		"schema":      map[string]any{"type": "string"},
	}
}

func optionalQueryParameter(name string, description string) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "query",
		"required":    false,
		"description": description,
		"schema":      map[string]any{"type": "string"},
	}
}

func jsonOperation(operationID string, summary string, requestSchema string, responseSchema string) map[string]any {
	return map[string]any{
		"operationId": operationID,
		"summary":     summary,
		"requestBody": map[string]any{
			"required": true,
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": referenceSchema(requestSchema),
				},
			},
		},
		"responses": map[string]any{
			"200": map[string]any{
				"description": "Successful response",
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": referenceSchema(responseSchema),
					},
				},
			},
		},
	}
}

func responseOperation(operationID string, summary string, responseSchema any) map[string]any {
	return map[string]any{
		"operationId": operationID,
		"summary":     summary,
		"responses": map[string]any{
			"200": map[string]any{
				"description": "Successful response",
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": responseSchema,
					},
				},
			},
		},
	}
}

func referenceSchema(schemaName string) map[string]any {
	if schemaName == "boolean" {
		return map[string]any{"type": "boolean"}
	}

	return map[string]any{"$ref": "#/components/schemas/" + schemaName}
}

func arraySchema(itemSchemaName string) map[string]any {
	return map[string]any{
		"type":  "array",
		"items": referenceSchema(itemSchemaName),
	}
}

func rewriteReferences(schema *jsonschema.Schema) {
	if schema == nil {
		return
	}

	if strings.HasPrefix(schema.Ref, "#/$defs/") {
		schema.Ref = "#/components/schemas/" + strings.TrimPrefix(schema.Ref, "#/$defs/")
	}

	if schema.Properties != nil {
		for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
			rewriteReferences(pair.Value)
		}
	}

	rewriteReferences(schema.Items)

	for _, item := range schema.AnyOf {
		rewriteReferences(item)
	}

	for _, item := range schema.OneOf {
		rewriteReferences(item)
	}

	for _, item := range schema.AllOf {
		rewriteReferences(item)
	}

	rewriteReferences(schema.AdditionalProperties)
}

func applyJSONStringOptions(definitions jsonschema.Definitions) {
	typeRegistry := make(map[string]reflect.Type)
	collectModelTypes(reflect.TypeOf(apiSchemaTypes{}), typeRegistry)

	for typeName, definition := range definitions {
		goType, exists := typeRegistry[typeName]

		if !exists || definition.Properties == nil || goType.Kind() != reflect.Struct {
			continue
		}

		for i := 0; i < goType.NumField(); i++ {
			field := goType.Field(i)
			jsonTag := field.Tag.Get("json")

			if !strings.Contains(jsonTag, ",string") {
				continue
			}

			propertyName := strings.Split(jsonTag, ",")[0]

			if propertyName == "" {
				propertyName = field.Name
			}

			if propertyName == "-" {
				continue
			}

			property, exists := definition.Properties.Get(propertyName)

			if !exists {
				continue
			}

			property.Type = "string"
			property.Ref = ""
			property.AnyOf = nil
			property.OneOf = nil
			property.AllOf = nil
			property.Items = nil
		}
	}
}

func collectModelTypes(t reflect.Type, registry map[string]reflect.Type) {
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()
	}

	if t.Kind() == reflect.Map {
		collectModelTypes(t.Elem(), registry)
		return
	}

	if t.Kind() != reflect.Struct || t.Name() == "" {
		return
	}

	if strings.HasSuffix(t.PkgPath(), "/pkg/models") {
		if _, exists := registry[t.Name()]; exists {
			return
		}

		registry[t.Name()] = t
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.PkgPath != "" {
			continue
		}

		collectModelTypes(field.Type, registry)
	}
}

func generateTypeScript(definitions jsonschema.Definitions) string {
	var builder strings.Builder
	builder.WriteString("// Code generated by cmd/openapi; DO NOT EDIT.\n")
	builder.WriteString("/* eslint-disable */\n\n")

	names := make([]string, 0, len(definitions))

	for name := range definitions {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		definition := definitions[name]

		if definition.Type == "object" || definition.Properties != nil {
			builder.WriteString("export interface ")
			builder.WriteString(name)
			builder.WriteString(" ")
			builder.WriteString(typeScriptObject(definition, ""))
			builder.WriteString("\n\n")
		} else {
			builder.WriteString("export type ")
			builder.WriteString(name)
			builder.WriteString(" = ")
			builder.WriteString(typeScriptType(definition))
			builder.WriteString(";\n\n")
		}
	}

	return strings.TrimRight(builder.String(), "\n") + "\n"
}

func typeScriptObject(schema *jsonschema.Schema, indent string) string {
	if schema.Properties == nil || schema.Properties.Len() == 0 {
		if schema.AdditionalProperties != nil {
			return "Record<string, " + typeScriptType(schema.AdditionalProperties) + ">"
		}

		return "Record<string, unknown>"
	}

	required := make(map[string]bool, len(schema.Required))

	for _, name := range schema.Required {
		required[name] = true
	}

	var builder strings.Builder
	builder.WriteString("{\n")

	for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
		propertyName := pair.Key
		property := pair.Value

		builder.WriteString(indent)
		builder.WriteString("    readonly ")
		builder.WriteString(typeScriptPropertyName(propertyName))

		if !required[propertyName] {
			builder.WriteString("?")
		}

		builder.WriteString(": ")
		builder.WriteString(typeScriptType(property))
		builder.WriteString(";\n")
	}

	builder.WriteString(indent)
	builder.WriteString("}")

	return builder.String()
}

func typeScriptType(schema *jsonschema.Schema) string {
	if schema == nil {
		return "unknown"
	}

	if schema.Ref != "" {
		return typeScriptReferenceName(schema.Ref)
	}

	if len(schema.Enum) > 0 {
		values := make([]string, len(schema.Enum))

		for i, value := range schema.Enum {
			encoded, err := json.Marshal(value)

			if err != nil {
				values[i] = "unknown"
			} else {
				values[i] = string(encoded)
			}
		}

		return strings.Join(values, " | ")
	}

	if len(schema.AnyOf) > 0 {
		return typeScriptUnion(schema.AnyOf)
	}

	if len(schema.OneOf) > 0 {
		return typeScriptUnion(schema.OneOf)
	}

	if len(schema.AllOf) > 0 {
		parts := make([]string, 0, len(schema.AllOf))

		for _, item := range schema.AllOf {
			parts = append(parts, typeScriptType(item))
		}

		return strings.Join(parts, " & ")
	}

	switch schema.Type {
	case "string":
		return "string"
	case "integer", "number":
		return "number"
	case "boolean":
		return "boolean"
	case "null":
		return "null"
	case "array":
		if schema.Items == nil {
			return "unknown[]"
		}

		return typeScriptType(schema.Items) + "[]"
	case "object", "":
		return typeScriptObject(schema, "")
	default:
		return "unknown"
	}
}

func typeScriptUnion(schemas []*jsonschema.Schema) string {
	parts := make([]string, 0, len(schemas))
	seen := make(map[string]bool)

	for _, schema := range schemas {
		part := typeScriptType(schema)

		if seen[part] {
			continue
		}

		seen[part] = true
		parts = append(parts, part)
	}

	return strings.Join(parts, " | ")
}

func typeScriptReferenceName(reference string) string {
	reference = strings.TrimPrefix(reference, "#/components/schemas/")
	reference = strings.TrimPrefix(reference, "#/$defs/")

	return reference
}

func typeScriptPropertyName(name string) string {
	if name == "" {
		return "unknown"
	}

	for i, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_' || char == '$' || (i > 0 && char >= '0' && char <= '9') {
			continue
		}

		encoded, _ := json.Marshal(name)
		return string(encoded)
	}

	return name
}

func readPackageVersion(repositoryRoot string) (string, error) {
	data, err := os.ReadFile(filepath.Join(repositoryRoot, "package.json"))

	if err != nil {
		return "", err
	}

	var packageFile struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(data, &packageFile); err != nil {
		return "", err
	}

	if packageFile.Version == "" {
		return "0.0.0", nil
	}

	return packageFile.Version, nil
}

func findRepositoryRoot() (string, error) {
	currentDirectory, err := os.Getwd()

	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(currentDirectory, "go.mod")); err == nil {
			return currentDirectory, nil
		}

		parentDirectory := filepath.Dir(currentDirectory)

		if parentDirectory == currentDirectory {
			return "", fmt.Errorf("cannot find repository root")
		}

		currentDirectory = parentDirectory
	}
}

func writeFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, content, 0644)
}

func checkFile(path string, expected []byte) error {
	actual, err := os.ReadFile(path)

	if err != nil {
		return fmt.Errorf("%s is missing; run go run ./cmd/openapi", path)
	}

	if string(actual) != string(expected) {
		return fmt.Errorf("%s is out of date; run go run ./cmd/openapi", path)
	}

	return nil
}

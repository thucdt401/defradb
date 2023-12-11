// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"context"
	"fmt"
	"sort"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/request/graphql/schema/types"

	"github.com/sourcenetwork/graphql-go/language/ast"
	gqlp "github.com/sourcenetwork/graphql-go/language/parser"
	"github.com/sourcenetwork/graphql-go/language/source"
)

// FromString parses a GQL SDL string into a set of collection descriptions.
func FromString(ctx context.Context, schemaString string) (
	[]client.CollectionDefinition,
	error,
) {
	source := source.NewSource(&source.Source{
		Body: []byte(schemaString),
	})

	doc, err := gqlp.Parse(
		gqlp.ParseParams{
			Source: source,
		},
	)
	if err != nil {
		return nil, err
	}

	return fromAst(ctx, doc)
}

// fromAst parses a GQL AST into a set of collection descriptions.
func fromAst(ctx context.Context, doc *ast.Document) (
	[]client.CollectionDefinition,
	error,
) {
	relationManager := NewRelationManager()
	definitions := []client.CollectionDefinition{}

	for _, def := range doc.Definitions {
		switch defType := def.(type) {
		case *ast.ObjectDefinition:
			description, err := fromAstDefinition(ctx, relationManager, defType)
			if err != nil {
				return nil, err
			}

			definitions = append(definitions, description)

		default:
			// Do nothing, ignore it and continue
			continue
		}
	}

	// The details on the relations between objects depend on both sides
	// of the relationship.  The relation manager handles this, and must be applied
	// after all the collections have been processed.
	err := finalizeRelations(relationManager, definitions)
	if err != nil {
		return nil, err
	}

	return definitions, nil
}

// fromAstDefinition parses a AST object definition into a set of collection descriptions.
func fromAstDefinition(
	ctx context.Context,
	relationManager *RelationManager,
	def *ast.ObjectDefinition,
) (client.CollectionDefinition, error) {
	fieldDescriptions := []client.FieldDescription{
		{
			Name: request.KeyFieldName,
			Kind: client.FieldKind_DocKey,
			Typ:  client.NONE_CRDT,
		},
	}

	indexDescriptions := []client.IndexDescription{}
	for _, field := range def.Fields {
		tmpFieldsDescriptions, err := fieldsFromAST(field, relationManager, def)
		if err != nil {
			return client.CollectionDefinition{}, err
		}

		fieldDescriptions = append(fieldDescriptions, tmpFieldsDescriptions...)

		for _, directive := range field.Directives {
			if directive.Name.Value == types.IndexDirectiveLabel {
				index, err := fieldIndexFromAST(field, directive)
				if err != nil {
					return client.CollectionDefinition{}, err
				}
				indexDescriptions = append(indexDescriptions, index)
			}
		}
	}

	// sort the fields lexicographically
	sort.Slice(fieldDescriptions, func(i, j int) bool {
		// make sure that the _key (KeyFieldName) is always at the beginning
		if fieldDescriptions[i].Name == request.KeyFieldName {
			return true
		} else if fieldDescriptions[j].Name == request.KeyFieldName {
			return false
		}
		return fieldDescriptions[i].Name < fieldDescriptions[j].Name
	})

	for _, directive := range def.Directives {
		if directive.Name.Value == types.IndexDirectiveLabel {
			index, err := indexFromAST(directive)
			if err != nil {
				return client.CollectionDefinition{}, err
			}
			indexDescriptions = append(indexDescriptions, index)
		}
	}

	return client.CollectionDefinition{
		Description: client.CollectionDescription{
			Name:    def.Name.Value,
			Indexes: indexDescriptions,
		},
		Schema: client.SchemaDescription{
			Name:   def.Name.Value,
			Fields: fieldDescriptions,
		},
	}, nil
}

// IsValidIndexName returns true if the name is a valid index name.
// Valid index names must start with a letter or underscore, and can
// contain letters, numbers, and underscores.
func IsValidIndexName(name string) bool {
	if len(name) == 0 {
		return false
	}
	if name[0] != '_' && (name[0] < 'a' || name[0] > 'z') && (name[0] < 'A' || name[0] > 'Z') {
		return false
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
			return false
		}
	}
	return true
}

func fieldIndexFromAST(field *ast.FieldDefinition, directive *ast.Directive) (client.IndexDescription, error) {
	desc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{
			{
				Name:      field.Name.Value,
				Direction: client.Ascending,
			},
		},
	}
	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.IndexDirectivePropName:
			nameVal, ok := arg.Value.(*ast.StringValue)
			if !ok {
				return client.IndexDescription{}, ErrIndexWithInvalidArg
			}
			desc.Name = nameVal.Value
			if !IsValidIndexName(desc.Name) {
				return client.IndexDescription{}, NewErrIndexWithInvalidName(desc.Name)
			}
		case types.IndexDirectivePropUnique:
			boolVal, ok := arg.Value.(*ast.BooleanValue)
			if !ok {
				return client.IndexDescription{}, ErrIndexWithInvalidArg
			}
			desc.Unique = boolVal.Value
		default:
			return client.IndexDescription{}, ErrIndexWithUnknownArg
		}
	}
	return desc, nil
}

func indexFromAST(directive *ast.Directive) (client.IndexDescription, error) {
	desc := client.IndexDescription{}
	var directions *ast.ListValue
	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.IndexDirectivePropName:
			nameVal, ok := arg.Value.(*ast.StringValue)
			if !ok {
				return client.IndexDescription{}, ErrIndexWithInvalidArg
			}
			desc.Name = nameVal.Value
			if !IsValidIndexName(desc.Name) {
				return client.IndexDescription{}, ErrIndexWithInvalidArg
			}
		case types.IndexDirectivePropFields:
			fieldsVal, ok := arg.Value.(*ast.ListValue)
			if !ok {
				return client.IndexDescription{}, ErrIndexWithInvalidArg
			}
			for _, field := range fieldsVal.Values {
				fieldVal, ok := field.(*ast.StringValue)
				if !ok {
					return client.IndexDescription{}, ErrIndexWithInvalidArg
				}
				desc.Fields = append(desc.Fields, client.IndexedFieldDescription{
					Name: fieldVal.Value,
				})
			}
		case types.IndexDirectivePropDirections:
			var ok bool
			directions, ok = arg.Value.(*ast.ListValue)
			if !ok {
				return client.IndexDescription{}, ErrIndexWithInvalidArg
			}
		case types.IndexDirectivePropUnique:
			boolVal, ok := arg.Value.(*ast.BooleanValue)
			if !ok {
				return client.IndexDescription{}, ErrIndexWithInvalidArg
			}
			desc.Unique = boolVal.Value
		default:
			return client.IndexDescription{}, ErrIndexWithUnknownArg
		}
	}
	if len(desc.Fields) == 0 {
		return client.IndexDescription{}, ErrIndexMissingFields
	}
	if directions != nil {
		if len(directions.Values) != len(desc.Fields) {
			return client.IndexDescription{}, ErrIndexWithInvalidArg
		}
		for i := range desc.Fields {
			dirVal, ok := directions.Values[i].(*ast.EnumValue)
			if !ok {
				return client.IndexDescription{}, ErrIndexWithInvalidArg
			}
			if dirVal.Value == string(client.Ascending) {
				desc.Fields[i].Direction = client.Ascending
			} else if dirVal.Value == string(client.Descending) {
				desc.Fields[i].Direction = client.Descending
			}
		}
	} else {
		for i := range desc.Fields {
			desc.Fields[i].Direction = client.Ascending
		}
	}
	return desc, nil
}

func fieldsFromAST(field *ast.FieldDefinition,
	relationManager *RelationManager,
	def *ast.ObjectDefinition,
) ([]client.FieldDescription, error) {
	kind, err := astTypeToKind(field.Type)
	if err != nil {
		return nil, err
	}

	schema := ""
	relationName := ""
	relationType := client.RelationType(0)

	fieldDescriptions := []client.FieldDescription{}

	if kind == client.FieldKind_FOREIGN_OBJECT || kind == client.FieldKind_FOREIGN_OBJECT_ARRAY {
		if kind == client.FieldKind_FOREIGN_OBJECT {
			schema = field.Type.(*ast.Named).Name.Value
			relationType = client.Relation_Type_ONE
			if _, exists := findDirective(field, "primary"); exists {
				relationType |= client.Relation_Type_Primary
			}

			// An _id field is added for every 1-N relationship from this object.
			fieldDescriptions = append(fieldDescriptions, client.FieldDescription{
				Name:         fmt.Sprintf("%s_id", field.Name.Value),
				Kind:         client.FieldKind_DocKey,
				Typ:          defaultCRDTForFieldKind[client.FieldKind_DocKey],
				RelationType: client.Relation_Type_INTERNAL_ID,
			})
		} else if kind == client.FieldKind_FOREIGN_OBJECT_ARRAY {
			schema = field.Type.(*ast.List).Type.(*ast.Named).Name.Value
			relationType = client.Relation_Type_MANY
		}

		relationName, err = getRelationshipName(field, def.Name.Value, schema)
		if err != nil {
			return nil, err
		}

		// Register the relationship so that the relationship manager can evaluate
		// relationsip properties dependent on both collections in the relationship.
		_, err := relationManager.RegisterSingle(
			relationName,
			schema,
			field.Name.Value,
			relationType,
		)
		if err != nil {
			return nil, err
		}
	}

	fieldDescription := client.FieldDescription{
		Name:         field.Name.Value,
		Kind:         kind,
		Typ:          defaultCRDTForFieldKind[kind],
		Schema:       schema,
		RelationName: relationName,
		RelationType: relationType,
	}

	fieldDescriptions = append(fieldDescriptions, fieldDescription)
	return fieldDescriptions, nil
}

func astTypeToKind(t ast.Type) (client.FieldKind, error) {
	const (
		typeID       string = "ID"
		typeBoolean  string = "Boolean"
		typeInt      string = "Int"
		typeFloat    string = "Float"
		typeDateTime string = "DateTime"
		typeString   string = "String"
		typeBlob     string = "Blob"
	)

	switch astTypeVal := t.(type) {
	case *ast.List:
		switch innerAstTypeVal := astTypeVal.Type.(type) {
		case *ast.NonNull:
			switch innerAstTypeVal.Type.(*ast.Named).Name.Value {
			case typeBoolean:
				return client.FieldKind_BOOL_ARRAY, nil
			case typeInt:
				return client.FieldKind_INT_ARRAY, nil
			case typeFloat:
				return client.FieldKind_FLOAT_ARRAY, nil
			case typeString:
				return client.FieldKind_STRING_ARRAY, nil
			default:
				return 0, NewErrNonNullForTypeNotSupported(innerAstTypeVal.Type.(*ast.Named).Name.Value)
			}

		default:
			switch astTypeVal.Type.(*ast.Named).Name.Value {
			case typeBoolean:
				return client.FieldKind_NILLABLE_BOOL_ARRAY, nil
			case typeInt:
				return client.FieldKind_NILLABLE_INT_ARRAY, nil
			case typeFloat:
				return client.FieldKind_NILLABLE_FLOAT_ARRAY, nil
			case typeString:
				return client.FieldKind_NILLABLE_STRING_ARRAY, nil
			default:
				return client.FieldKind_FOREIGN_OBJECT_ARRAY, nil
			}
		}

	case *ast.Named:
		switch astTypeVal.Name.Value {
		case typeID:
			return client.FieldKind_DocKey, nil
		case typeBoolean:
			return client.FieldKind_BOOL, nil
		case typeInt:
			return client.FieldKind_INT, nil
		case typeFloat:
			return client.FieldKind_FLOAT, nil
		case typeDateTime:
			return client.FieldKind_DATETIME, nil
		case typeString:
			return client.FieldKind_STRING, nil
		case typeBlob:
			return client.FieldKind_BLOB, nil
		default:
			return client.FieldKind_FOREIGN_OBJECT, nil
		}

	case *ast.NonNull:
		return 0, ErrNonNullNotSupported

	default:
		return 0, NewErrTypeNotFound(t.String())
	}
}

func findDirective(field *ast.FieldDefinition, directiveName string) (*ast.Directive, bool) {
	for _, directive := range field.Directives {
		if directive.Name.Value == directiveName {
			return directive, true
		}
	}
	return nil, false
}

// Gets the name of the relationship. Will return the provided name if one is specified,
// otherwise will generate one
func getRelationshipName(
	field *ast.FieldDefinition,
	hostName string,
	targetName string,
) (string, error) {
	// search for a @relation directive name, and return it if found
	for _, directive := range field.Directives {
		if directive.Name.Value == "relation" {
			for _, argument := range directive.Arguments {
				if argument.Name.Value == "name" {
					name, isString := argument.Value.GetValue().(string)
					if !isString {
						return "", client.NewErrUnexpectedType[string]("Relationship name", argument.Value.GetValue())
					}
					return name, nil
				}
			}
		}
	}

	// if no name is provided, generate one
	return genRelationName(hostName, targetName)
}

func finalizeRelations(relationManager *RelationManager, definitions []client.CollectionDefinition) error {
	for _, definition := range definitions {
		for i, field := range definition.Schema.Fields {
			if field.RelationType == 0 || field.RelationType&client.Relation_Type_INTERNAL_ID != 0 {
				continue
			}

			rel, err := relationManager.GetRelation(field.RelationName)
			if err != nil {
				return err
			}

			_, fieldRelationType, ok := rel.GetField(field.Schema, field.Name)
			if !ok {
				return NewErrRelationMissingField(field.Schema, field.Name)
			}

			// if not finalized then we are missing one side of the relationship
			if !rel.finalized {
				return client.NewErrRelationOneSided(field.Name, field.Schema)
			}

			field.RelationType = rel.Kind() | fieldRelationType
			definition.Schema.Fields[i] = field
		}
	}

	return nil
}

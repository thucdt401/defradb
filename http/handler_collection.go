// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"

	"github.com/sourcenetwork/defradb/client"
)

type collectionHandler struct{}

type CollectionDeleteRequest struct {
	Key    string   `json:"key"`
	Keys   []string `json:"keys"`
	Filter any      `json:"filter"`
}

type CollectionUpdateRequest struct {
	Key     string   `json:"key"`
	Keys    []string `json:"keys"`
	Filter  any      `json:"filter"`
	Updater string   `json:"updater"`
}

func (s *collectionHandler) Create(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var body any
	if err := requestJSON(req, &body); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	switch t := body.(type) {
	case []any:
		var docList []*client.Document
		for _, v := range t {
			docMap, ok := v.(map[string]any)
			if !ok {
				responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidRequestBody})
				return
			}
			doc, err := client.NewDocFromMap(docMap)
			if err != nil {
				responseJSON(rw, http.StatusBadRequest, errorResponse{err})
				return
			}
			docList = append(docList, doc)
		}
		if err := col.CreateMany(req.Context(), docList); err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		rw.WriteHeader(http.StatusOK)
	case map[string]any:
		doc, err := client.NewDocFromMap(t)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		if err := col.Create(req.Context(), doc); err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		rw.WriteHeader(http.StatusOK)
	default:
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidRequestBody})
	}
}

func (s *collectionHandler) DeleteWith(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var request CollectionDeleteRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	switch {
	case request.Filter != nil:
		result, err := col.DeleteWith(req.Context(), request.Filter)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	case request.Key != "":
		docKey, err := client.NewDocKeyFromString(request.Key)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		result, err := col.DeleteWith(req.Context(), docKey)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	case request.Keys != nil:
		var docKeys []client.DocKey
		for _, key := range request.Keys {
			docKey, err := client.NewDocKeyFromString(key)
			if err != nil {
				responseJSON(rw, http.StatusBadRequest, errorResponse{err})
				return
			}
			docKeys = append(docKeys, docKey)
		}
		result, err := col.DeleteWith(req.Context(), docKeys)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	default:
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidRequestBody})
	}
}

func (s *collectionHandler) UpdateWith(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var request CollectionUpdateRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	switch {
	case request.Filter != nil:
		result, err := col.UpdateWith(req.Context(), request.Filter, request.Updater)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	case request.Key != "":
		docKey, err := client.NewDocKeyFromString(request.Key)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		result, err := col.UpdateWith(req.Context(), docKey, request.Updater)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	case request.Keys != nil:
		var docKeys []client.DocKey
		for _, key := range request.Keys {
			docKey, err := client.NewDocKeyFromString(key)
			if err != nil {
				responseJSON(rw, http.StatusBadRequest, errorResponse{err})
				return
			}
			docKeys = append(docKeys, docKey)
		}
		result, err := col.UpdateWith(req.Context(), docKeys, request.Updater)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	default:
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidRequestBody})
	}
}

func (s *collectionHandler) Update(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	docKey, err := client.NewDocKeyFromString(chi.URLParam(req, "key"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	doc, err := col.Get(req.Context(), docKey, true)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	patch, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	if err := doc.SetWithJSON(patch); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err = col.Update(req.Context(), doc)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *collectionHandler) Delete(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	docKey, err := client.NewDocKeyFromString(chi.URLParam(req, "key"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	_, err = col.Delete(req.Context(), docKey)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *collectionHandler) Get(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)
	showDeleted, _ := strconv.ParseBool(req.URL.Query().Get("show_deleted"))

	docKey, err := client.NewDocKeyFromString(chi.URLParam(req, "key"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	doc, err := col.Get(req.Context(), docKey, showDeleted)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	docMap, err := doc.ToMap()
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, docMap)
}

type DocKeyResult struct {
	Key   string `json:"key"`
	Error string `json:"error"`
}

func (s *collectionHandler) GetAllDocKeys(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	flusher, ok := rw.(http.Flusher)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrStreamingNotSupported})
		return
	}

	docKeyCh, err := col.GetAllDocKeys(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	rw.WriteHeader(http.StatusOK)
	flusher.Flush()

	for docKey := range docKeyCh {
		results := &DocKeyResult{
			Key: docKey.Key.String(),
		}
		if docKey.Err != nil {
			results.Error = docKey.Err.Error()
		}
		data, err := json.Marshal(results)
		if err != nil {
			return
		}
		fmt.Fprintf(rw, "data: %s\n\n", data)
		flusher.Flush()
	}
}

func (s *collectionHandler) CreateIndex(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var indexDesc client.IndexDescription
	if err := requestJSON(req, &indexDesc); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	index, err := col.CreateIndex(req.Context(), indexDesc)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, index)
}

func (s *collectionHandler) GetIndexes(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)
	indexesMap, err := store.GetAllIndexes(req.Context())

	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	indexes := make([]client.IndexDescription, 0, len(indexesMap))
	for _, index := range indexesMap {
		indexes = append(indexes, index...)
	}
	responseJSON(rw, http.StatusOK, indexes)
}

func (s *collectionHandler) DropIndex(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	err := col.DropIndex(req.Context(), chi.URLParam(req, "index"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *collectionHandler) bindRoutes(router *Router) {
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}
	successResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/success",
	}
	collectionUpdateSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/collection_update",
	}
	updateResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/update_result",
	}
	collectionDeleteSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/collection_delete",
	}
	deleteResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/delete_result",
	}
	documentSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/document",
	}
	indexSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/index",
	}

	collectionNamePathParam := openapi3.NewPathParameter("name").
		WithDescription("Collection name").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	documentArraySchema := openapi3.NewArraySchema()
	documentArraySchema.Items = documentSchema

	collectionCreateSchema := openapi3.NewOneOfSchema()
	collectionCreateSchema.OneOf = openapi3.SchemaRefs{
		documentSchema,
		openapi3.NewSchemaRef("", documentArraySchema),
	}

	collectionCreateRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(collectionCreateSchema))

	collectionCreate := openapi3.NewOperation()
	collectionCreate.OperationID = "collection_create"
	collectionCreate.Description = "Create document(s) in a collection"
	collectionCreate.Tags = []string{"collection"}
	collectionCreate.AddParameter(collectionNamePathParam)
	collectionCreate.RequestBody = &openapi3.RequestBodyRef{
		Value: collectionCreateRequest,
	}
	collectionCreate.Responses = make(openapi3.Responses)
	collectionCreate.Responses["200"] = successResponse
	collectionCreate.Responses["400"] = errorResponse

	collectionUpdateWithRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(collectionUpdateSchema))

	collectionUpdateWithResponse := openapi3.NewResponse().
		WithDescription("Update results").
		WithJSONSchemaRef(updateResultSchema)

	collectionUpdateWith := openapi3.NewOperation()
	collectionUpdateWith.OperationID = "collection_update_with"
	collectionUpdateWith.Description = "Update document(s) in a collection"
	collectionUpdateWith.Tags = []string{"collection"}
	collectionUpdateWith.AddParameter(collectionNamePathParam)
	collectionUpdateWith.RequestBody = &openapi3.RequestBodyRef{
		Value: collectionUpdateWithRequest,
	}
	collectionUpdateWith.AddResponse(200, collectionUpdateWithResponse)
	collectionUpdateWith.Responses["400"] = errorResponse

	collectionDeleteWithRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(collectionDeleteSchema))

	collectionDeleteWithResponse := openapi3.NewResponse().
		WithDescription("Delete results").
		WithJSONSchemaRef(deleteResultSchema)

	collectionDeleteWith := openapi3.NewOperation()
	collectionDeleteWith.OperationID = "collections_delete_with"
	collectionDeleteWith.Description = "Delete document(s) from a collection"
	collectionDeleteWith.Tags = []string{"collection"}
	collectionDeleteWith.AddParameter(collectionNamePathParam)
	collectionDeleteWith.RequestBody = &openapi3.RequestBodyRef{
		Value: collectionDeleteWithRequest,
	}
	collectionDeleteWith.AddResponse(200, collectionDeleteWithResponse)
	collectionDeleteWith.Responses["400"] = errorResponse

	createIndexRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(indexSchema))
	createIndexResponse := openapi3.NewResponse().
		WithDescription("Index description").
		WithJSONSchemaRef(indexSchema)

	createIndex := openapi3.NewOperation()
	createIndex.OperationID = "index_create"
	createIndex.Description = "Create a secondary index"
	createIndex.Tags = []string{"index"}
	createIndex.AddParameter(collectionNamePathParam)
	createIndex.RequestBody = &openapi3.RequestBodyRef{
		Value: createIndexRequest,
	}
	createIndex.AddResponse(200, createIndexResponse)
	createIndex.Responses["400"] = errorResponse

	indexArraySchema := openapi3.NewArraySchema()
	indexArraySchema.Items = indexSchema

	getIndexesResponse := openapi3.NewResponse().
		WithDescription("List of indexes").
		WithJSONSchema(indexArraySchema)

	getIndexes := openapi3.NewOperation()
	getIndexes.OperationID = "index_list"
	getIndexes.Description = "List secondary indexes"
	getIndexes.Tags = []string{"index"}
	getIndexes.AddParameter(collectionNamePathParam)
	getIndexes.AddResponse(200, getIndexesResponse)
	getIndexes.Responses["400"] = errorResponse

	indexPathParam := openapi3.NewPathParameter("index").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	dropIndex := openapi3.NewOperation()
	dropIndex.OperationID = "index_drop"
	dropIndex.Description = "Delete a secondary index"
	dropIndex.Tags = []string{"index"}
	dropIndex.AddParameter(collectionNamePathParam)
	dropIndex.AddParameter(indexPathParam)
	dropIndex.Responses = make(openapi3.Responses)
	dropIndex.Responses["200"] = successResponse
	dropIndex.Responses["400"] = errorResponse

	documentKeyPathParam := openapi3.NewPathParameter("key").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	collectionGetResponse := openapi3.NewResponse().
		WithDescription("Document value").
		WithJSONSchemaRef(documentSchema)

	collectionGet := openapi3.NewOperation()
	collectionGet.Description = "Get a document by key"
	collectionGet.OperationID = "collection_get"
	collectionGet.Tags = []string{"collection"}
	collectionGet.AddParameter(collectionNamePathParam)
	collectionGet.AddParameter(documentKeyPathParam)
	collectionGet.AddResponse(200, collectionGetResponse)
	collectionGet.Responses["400"] = errorResponse

	collectionUpdate := openapi3.NewOperation()
	collectionUpdate.Description = "Update a document by key"
	collectionUpdate.OperationID = "collection_update"
	collectionUpdate.Tags = []string{"collection"}
	collectionUpdate.AddParameter(collectionNamePathParam)
	collectionUpdate.AddParameter(documentKeyPathParam)
	collectionUpdate.Responses = make(openapi3.Responses)
	collectionUpdate.Responses["200"] = successResponse
	collectionUpdate.Responses["400"] = errorResponse

	collectionDelete := openapi3.NewOperation()
	collectionDelete.Description = "Delete a document by key"
	collectionDelete.OperationID = "collection_delete"
	collectionDelete.Tags = []string{"collection"}
	collectionDelete.AddParameter(collectionNamePathParam)
	collectionDelete.AddParameter(documentKeyPathParam)
	collectionDelete.Responses = make(openapi3.Responses)
	collectionDelete.Responses["200"] = successResponse
	collectionDelete.Responses["400"] = errorResponse

	collectionKeys := openapi3.NewOperation()
	collectionKeys.AddParameter(collectionNamePathParam)
	collectionKeys.Description = "Get all document keys"
	collectionKeys.OperationID = "collection_keys"
	collectionKeys.Tags = []string{"collection"}
	collectionKeys.Responses = make(openapi3.Responses)
	collectionKeys.Responses["200"] = successResponse
	collectionKeys.Responses["400"] = errorResponse

	router.AddRoute("/collections/{name}", http.MethodGet, collectionKeys, h.GetAllDocKeys)
	router.AddRoute("/collections/{name}", http.MethodPost, collectionCreate, h.Create)
	router.AddRoute("/collections/{name}", http.MethodPatch, collectionUpdateWith, h.UpdateWith)
	router.AddRoute("/collections/{name}", http.MethodDelete, collectionDeleteWith, h.DeleteWith)
	router.AddRoute("/collections/{name}/indexes", http.MethodPost, createIndex, h.CreateIndex)
	router.AddRoute("/collections/{name}/indexes", http.MethodGet, getIndexes, h.GetIndexes)
	router.AddRoute("/collections/{name}/indexes/{index}", http.MethodDelete, dropIndex, h.DropIndex)
	router.AddRoute("/collections/{name}/{key}", http.MethodGet, collectionGet, h.Get)
	router.AddRoute("/collections/{name}/{key}", http.MethodPatch, collectionUpdate, h.Update)
	router.AddRoute("/collections/{name}/{key}", http.MethodDelete, collectionDelete, h.Delete)
}

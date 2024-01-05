//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2024 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

// Code generated by go-swagger; DO NOT EDIT.

package schema

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/weaviate/weaviate/entities/models"
)

// NewSchemaObjectsUpdateParams creates a new SchemaObjectsUpdateParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewSchemaObjectsUpdateParams() *SchemaObjectsUpdateParams {
	return &SchemaObjectsUpdateParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewSchemaObjectsUpdateParamsWithTimeout creates a new SchemaObjectsUpdateParams object
// with the ability to set a timeout on a request.
func NewSchemaObjectsUpdateParamsWithTimeout(timeout time.Duration) *SchemaObjectsUpdateParams {
	return &SchemaObjectsUpdateParams{
		timeout: timeout,
	}
}

// NewSchemaObjectsUpdateParamsWithContext creates a new SchemaObjectsUpdateParams object
// with the ability to set a context for a request.
func NewSchemaObjectsUpdateParamsWithContext(ctx context.Context) *SchemaObjectsUpdateParams {
	return &SchemaObjectsUpdateParams{
		Context: ctx,
	}
}

// NewSchemaObjectsUpdateParamsWithHTTPClient creates a new SchemaObjectsUpdateParams object
// with the ability to set a custom HTTPClient for a request.
func NewSchemaObjectsUpdateParamsWithHTTPClient(client *http.Client) *SchemaObjectsUpdateParams {
	return &SchemaObjectsUpdateParams{
		HTTPClient: client,
	}
}

/*
SchemaObjectsUpdateParams contains all the parameters to send to the API endpoint

	for the schema objects update operation.

	Typically these are written to a http.Request.
*/
type SchemaObjectsUpdateParams struct {

	// ClassName.
	ClassName string

	// ObjectClass.
	ObjectClass *models.Class

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the schema objects update params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *SchemaObjectsUpdateParams) WithDefaults() *SchemaObjectsUpdateParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the schema objects update params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *SchemaObjectsUpdateParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the schema objects update params
func (o *SchemaObjectsUpdateParams) WithTimeout(timeout time.Duration) *SchemaObjectsUpdateParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the schema objects update params
func (o *SchemaObjectsUpdateParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the schema objects update params
func (o *SchemaObjectsUpdateParams) WithContext(ctx context.Context) *SchemaObjectsUpdateParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the schema objects update params
func (o *SchemaObjectsUpdateParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the schema objects update params
func (o *SchemaObjectsUpdateParams) WithHTTPClient(client *http.Client) *SchemaObjectsUpdateParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the schema objects update params
func (o *SchemaObjectsUpdateParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithClassName adds the className to the schema objects update params
func (o *SchemaObjectsUpdateParams) WithClassName(className string) *SchemaObjectsUpdateParams {
	o.SetClassName(className)
	return o
}

// SetClassName adds the className to the schema objects update params
func (o *SchemaObjectsUpdateParams) SetClassName(className string) {
	o.ClassName = className
}

// WithObjectClass adds the objectClass to the schema objects update params
func (o *SchemaObjectsUpdateParams) WithObjectClass(objectClass *models.Class) *SchemaObjectsUpdateParams {
	o.SetObjectClass(objectClass)
	return o
}

// SetObjectClass adds the objectClass to the schema objects update params
func (o *SchemaObjectsUpdateParams) SetObjectClass(objectClass *models.Class) {
	o.ObjectClass = objectClass
}

// WriteToRequest writes these params to a swagger request
func (o *SchemaObjectsUpdateParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param className
	if err := r.SetPathParam("className", o.ClassName); err != nil {
		return err
	}
	if o.ObjectClass != nil {
		if err := r.SetBodyParam(o.ObjectClass); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

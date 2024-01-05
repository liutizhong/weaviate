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

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
)

// NewWeaviateWellknownReadinessParams creates a new WeaviateWellknownReadinessParams object
//
// There are no default values defined in the spec.
func NewWeaviateWellknownReadinessParams() WeaviateWellknownReadinessParams {

	return WeaviateWellknownReadinessParams{}
}

// WeaviateWellknownReadinessParams contains all the bound params for the weaviate wellknown readiness operation
// typically these are obtained from a http.Request
//
// swagger:parameters weaviate.wellknown.readiness
type WeaviateWellknownReadinessParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewWeaviateWellknownReadinessParams() beforehand.
func (o *WeaviateWellknownReadinessParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

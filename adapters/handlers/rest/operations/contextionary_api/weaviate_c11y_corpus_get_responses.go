/*                          _       _
 *__      _____  __ ___   ___  __ _| |_ ___
 *\ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
 * \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
 *  \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
 *
 * Copyright © 2016 - 2019 Weaviate. All rights reserved.
 * LICENSE: https://github.com/semi-technologies/weaviate/blob/develop/LICENSE.md
 * DESIGN & CONCEPT: Bob van Luijt (@bobvanluijt)
 * CONTACT: hello@semi.technology
 */ // Code generated by go-swagger; DO NOT EDIT.

package contextionary_api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// WeaviateC11yCorpusGetNotImplementedCode is the HTTP code returned for type WeaviateC11yCorpusGetNotImplemented
const WeaviateC11yCorpusGetNotImplementedCode int = 501

/*WeaviateC11yCorpusGetNotImplemented Not (yet) implemented.

swagger:response weaviateC11yCorpusGetNotImplemented
*/
type WeaviateC11yCorpusGetNotImplemented struct {
}

// NewWeaviateC11yCorpusGetNotImplemented creates WeaviateC11yCorpusGetNotImplemented with default headers values
func NewWeaviateC11yCorpusGetNotImplemented() *WeaviateC11yCorpusGetNotImplemented {

	return &WeaviateC11yCorpusGetNotImplemented{}
}

// WriteResponse to the client
func (o *WeaviateC11yCorpusGetNotImplemented) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(501)
}
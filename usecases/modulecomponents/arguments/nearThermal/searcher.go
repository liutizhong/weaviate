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

package nearThermal

import (
	"context"

	"github.com/pkg/errors"
	"github.com/liutizhong/weaviate/entities/dto"
	"github.com/liutizhong/weaviate/entities/modulecapabilities"
	"github.com/liutizhong/weaviate/entities/moduletools"
)

type Searcher[T dto.Embedding] struct {
	vectorizer bindVectorizer[T]
}

func NewSearcher[T dto.Embedding](vectorizer bindVectorizer[T]) *Searcher[T] {
	return &Searcher[T]{vectorizer}
}

type bindVectorizer[T dto.Embedding] interface {
	VectorizeThermal(ctx context.Context, thermal string, cfg moduletools.ClassConfig) (T, error)
}

func (s *Searcher[T]) VectorSearches() map[string]modulecapabilities.VectorForParams[T] {
	vectorSearches := map[string]modulecapabilities.VectorForParams[T]{}
	vectorSearches["nearThermal"] = &vectorForParams[T]{s.vectorizer}
	return vectorSearches
}

type vectorForParams[T dto.Embedding] struct {
	vectorizer bindVectorizer[T]
}

func (v *vectorForParams[T]) VectorForParams(ctx context.Context, params interface{}, className string,
	findVectorFn modulecapabilities.FindVectorFn[T],
	cfg moduletools.ClassConfig,
) (T, error) {
	// find vector for given search query
	vector, err := v.vectorizer.VectorizeThermal(ctx, params.(*NearThermalParams).Thermal, cfg)
	if err != nil {
		return nil, errors.Errorf("vectorize thermal: %v", err)
	}
	return vector, nil
}

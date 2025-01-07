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

package clients

import (
	"context"
	"time"

	"github.com/liutizhong/weaviate/entities/moduletools"

	"github.com/liutizhong/weaviate/usecases/modulecomponents"

	"github.com/sirupsen/logrus"
	"github.com/liutizhong/weaviate/modules/text2vec-jinaai/ent"
	"github.com/liutizhong/weaviate/usecases/modulecomponents/clients/jinaai"
)

const (
	defaultRPM = 500 // from https://jina.ai/embeddings/
	defaultTPM = 1_000_000
)

type vectorizer struct {
	client *jinaai.Client[[]float32]
	logger logrus.FieldLogger
}

func New(jinaAIApiKey string, timeout time.Duration, logger logrus.FieldLogger) *vectorizer {
	return &vectorizer{
		client: jinaai.New[[]float32](jinaAIApiKey, timeout, defaultRPM, defaultTPM, jinaai.EmbeddingsBuildUrlFn, logger),
		logger: logger,
	}
}

func (v *vectorizer) Vectorize(ctx context.Context, input []string,
	cfg moduletools.ClassConfig,
) (*modulecomponents.VectorizationResult[[]float32], *modulecomponents.RateLimits, int, error) {
	settings := ent.NewClassSettings(cfg)
	res, _, usage, err := v.client.Vectorize(ctx, input, jinaai.Settings{
		BaseURL:    settings.BaseURL(),
		Model:      settings.Model(),
		Dimensions: settings.Dimensions(),
		Task:       jinaai.RetrievalPassage,
	})
	return res, nil, usage, err
}

func (v *vectorizer) VectorizeQuery(ctx context.Context, input []string,
	cfg moduletools.ClassConfig,
) (*modulecomponents.VectorizationResult[[]float32], error) {
	settings := ent.NewClassSettings(cfg)
	res, _, _, err := v.client.Vectorize(ctx, input, jinaai.Settings{
		BaseURL:    settings.BaseURL(),
		Model:      settings.Model(),
		Dimensions: settings.Dimensions(),
		Task:       jinaai.RetrievalQuery,
	})
	return res, err
}

func (v *vectorizer) GetApiKeyHash(ctx context.Context, config moduletools.ClassConfig) [32]byte {
	return v.client.GetApiKeyHash(ctx, config)
}

func (v *vectorizer) GetVectorizerRateLimit(ctx context.Context, cfg moduletools.ClassConfig) *modulecomponents.RateLimits {
	return v.client.GetVectorizerRateLimit(ctx, cfg)
}

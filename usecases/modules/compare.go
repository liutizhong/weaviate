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

package modules

import (
	"context"

	"github.com/liutizhong/weaviate/entities/additional"
	"github.com/liutizhong/weaviate/entities/models"
	"github.com/liutizhong/weaviate/entities/modulecapabilities"
	"github.com/liutizhong/weaviate/entities/moduletools"
	"github.com/liutizhong/weaviate/entities/schema"
	"github.com/liutizhong/weaviate/entities/search"
	"github.com/liutizhong/weaviate/entities/types"
)

func reVectorize(ctx context.Context,
	cfg moduletools.ClassConfig,
	mod modulecapabilities.Vectorizer[[]float32],
	object *models.Object,
	class *models.Class,
	sourceProperties []string,
	targetVector string,
	findObjectFn modulecapabilities.FindObjectFn,
) (bool, models.AdditionalProperties, []float32) {
	shouldReVectorize, oldObject := reVectorizeEmbeddings(ctx, cfg, mod, object, class, sourceProperties, findObjectFn)
	if shouldReVectorize {
		return shouldReVectorize, nil, nil
	}

	if targetVector == "" {
		return false, oldObject.AdditionalProperties, oldObject.Vector
	} else {
		return false, oldObject.AdditionalProperties, oldObject.Vectors[targetVector]
	}
}

func reVectorizeMulti(ctx context.Context,
	cfg moduletools.ClassConfig,
	mod modulecapabilities.Vectorizer[[][]float32],
	object *models.Object,
	class *models.Class,
	sourceProperties []string,
	targetVector string,
	findObjectFn modulecapabilities.FindObjectFn,
) (bool, models.AdditionalProperties, [][]float32) {
	shouldReVectorize, oldObject := reVectorizeEmbeddings(ctx, cfg, mod, object, class, sourceProperties, findObjectFn)
	if shouldReVectorize {
		return shouldReVectorize, nil, nil
	}

	if targetVector == "" {
		return false, oldObject.AdditionalProperties, oldObject.MultiVectors[""]
	} else {
		return false, oldObject.AdditionalProperties, oldObject.MultiVectors[targetVector]
	}
}

func reVectorizeEmbeddings[T types.Embedding](ctx context.Context,
	cfg moduletools.ClassConfig,
	mod modulecapabilities.Vectorizer[T],
	object *models.Object,
	class *models.Class,
	sourceProperties []string,
	findObjectFn modulecapabilities.FindObjectFn,
) (bool, *search.Result) {
	textProps, mediaProps, err := mod.VectorizableProperties(cfg)
	if err != nil {
		return true, nil
	}

	type compareProps struct {
		Name    string
		IsArray bool
	}
	propsToCompare := make([]compareProps, 0)

	var sourcePropsSet map[string]struct{} = nil
	if len(sourceProperties) > 0 {
		sourcePropsSet = make(map[string]struct{}, len(sourceProperties))
		for _, sourceProp := range sourceProperties {
			sourcePropsSet[sourceProp] = struct{}{}
		}
	}
	mediaPropsSet := make(map[string]struct{}, len(mediaProps))
	for _, mediaProp := range mediaProps {
		mediaPropsSet[mediaProp] = struct{}{}
	}

	for _, prop := range class.Properties {
		if len(prop.DataType) > 1 {
			continue // multi cref
		}

		// for named vectors with explicit source properties, skip if not in the list
		if sourcePropsSet != nil {
			if _, ok := sourcePropsSet[prop.Name]; !ok {
				continue
			}
		}

		if prop.ModuleConfig != nil {
			if modConfig, ok := prop.ModuleConfig.(map[string]interface{})[class.Vectorizer]; ok {
				if skip, ok2 := modConfig.(map[string]interface{})["skip"]; ok2 && skip == true {
					continue
				}
			}
		}

		if prop.DataType[0] == schema.DataTypeText.String() && textProps {
			propsToCompare = append(propsToCompare, compareProps{Name: prop.Name, IsArray: false})
			continue
		}

		if prop.DataType[0] == schema.DataTypeTextArray.String() && textProps {
			propsToCompare = append(propsToCompare, compareProps{Name: prop.Name, IsArray: true})
			continue
		}

		if _, ok := mediaPropsSet[prop.Name]; ok {
			propsToCompare = append(propsToCompare, compareProps{Name: prop.Name, IsArray: schema.IsArrayDataType(prop.DataType)})
			continue
		}
	}

	// if no properties to compare, we can skip the comparison. Return vectors of old object if present
	if len(propsToCompare) == 0 {
		oldObject, err := findObjectFn(ctx, class.Class, object.ID, nil, additional.Properties{}, object.Tenant)
		if err != nil || oldObject == nil {
			return true, nil
		}
		return false, oldObject
	}

	returnProps := make(search.SelectProperties, 0, len(propsToCompare))
	for _, prop := range propsToCompare {
		returnProps = append(returnProps, search.SelectProperty{Name: prop.Name, IsPrimitive: true, IsObject: false})
	}
	oldObject, err := findObjectFn(ctx, class.Class, object.ID, returnProps, additional.Properties{}, object.Tenant)
	if err != nil || oldObject == nil {
		return true, nil
	}
	oldProps := oldObject.Schema.(map[string]interface{})
	var newProps map[string]interface{}
	if object.Properties == nil {
		newProps = make(map[string]interface{})
	} else {
		newProps = object.Properties.(map[string]interface{})
	}
	for _, propStruct := range propsToCompare {
		valNew, isPresentNew := newProps[propStruct.Name]
		valOld, isPresentOld := oldProps[propStruct.Name]

		if isPresentNew != isPresentOld {
			return true, nil
		}

		if !isPresentNew {
			continue
		}

		if propStruct.IsArray {
			// empty strings do not have type information saved with them - the new value can also come from disk if
			// an update happens
			if _, ok := valOld.([]interface{}); ok && len(valOld.([]interface{})) == 0 {
				valOld = []string{}
			}
			if _, ok := valNew.([]interface{}); ok && len(valNew.([]interface{})) == 0 {
				valNew = []string{}
			}

			if len(valOld.([]string)) != len(valNew.([]string)) {
				return true, nil
			}
			for i, val := range valOld.([]string) {
				if val != valNew.([]string)[i] {
					return true, nil
				}
			}
		} else {
			if valOld != valNew {
				return true, nil
			}
		}
	}
	return false, oldObject
}

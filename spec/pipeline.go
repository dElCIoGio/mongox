package spec

import "go.mongodb.org/mongo-driver/bson"

// Pipeline represents a MongoDB aggregation pipeline.
type Pipeline struct {
	stages []bson.M
}

// NewPipeline creates a new empty aggregation pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{stages: make([]bson.M, 0)}
}

// ToPipeline returns the pipeline as []bson.M for use with MongoDB driver.
func (p *Pipeline) ToPipeline() []bson.M {
	return p.stages
}

// Match adds a $match stage to filter documents.
//
// Example:
//
//	pipeline.Match(spec.Eq("status", "active"))
func (p *Pipeline) Match(filter Filter) *Pipeline {
	if filter != nil {
		p.stages = append(p.stages, bson.M{"$match": filter.ToMongo()})
	}
	return p
}

// MatchRaw adds a $match stage with a raw bson.M filter.
func (p *Pipeline) MatchRaw(filter bson.M) *Pipeline {
	if filter != nil {
		p.stages = append(p.stages, bson.M{"$match": filter})
	}
	return p
}

// Project adds a $project stage to reshape documents.
//
// Example:
//
//	pipeline.Project(bson.M{"name": 1, "total": 1, "_id": 0})
func (p *Pipeline) Project(projection bson.M) *Pipeline {
	p.stages = append(p.stages, bson.M{"$project": projection})
	return p
}

// Group adds a $group stage to group documents.
//
// Example:
//
//	pipeline.Group(bson.M{
//	    "_id": "$category",
//	    "total": bson.M{"$sum": "$amount"},
//	    "count": bson.M{"$sum": 1},
//	})
func (p *Pipeline) Group(groupSpec bson.M) *Pipeline {
	p.stages = append(p.stages, bson.M{"$group": groupSpec})
	return p
}

// GroupBy is a convenience method for simple grouping by a field.
//
// Example:
//
//	pipeline.GroupBy("$category", bson.M{
//	    "total": bson.M{"$sum": "$amount"},
//	})
func (p *Pipeline) GroupBy(idExpr any, accumulators bson.M) *Pipeline {
	groupSpec := bson.M{"_id": idExpr}
	for k, v := range accumulators {
		groupSpec[k] = v
	}
	p.stages = append(p.stages, bson.M{"$group": groupSpec})
	return p
}

// Sort adds a $sort stage to order documents.
//
// Example:
//
//	pipeline.Sort(bson.D{{"total", -1}, {"name", 1}})
func (p *Pipeline) Sort(sort any) *Pipeline {
	p.stages = append(p.stages, bson.M{"$sort": sort})
	return p
}

// SortBy is a convenience method for sorting by a single field.
// Use order 1 for ascending, -1 for descending.
//
// Example:
//
//	pipeline.SortBy("created_at", -1)  // Most recent first
func (p *Pipeline) SortBy(field string, order int) *Pipeline {
	p.stages = append(p.stages, bson.M{"$sort": bson.M{field: order}})
	return p
}

// Limit adds a $limit stage to restrict the number of documents.
func (p *Pipeline) Limit(n int64) *Pipeline {
	p.stages = append(p.stages, bson.M{"$limit": n})
	return p
}

// Skip adds a $skip stage to skip a number of documents.
func (p *Pipeline) Skip(n int64) *Pipeline {
	p.stages = append(p.stages, bson.M{"$skip": n})
	return p
}

// Unwind adds an $unwind stage to deconstruct an array field.
//
// Example:
//
//	pipeline.Unwind("$items")
func (p *Pipeline) Unwind(path string) *Pipeline {
	p.stages = append(p.stages, bson.M{"$unwind": path})
	return p
}

// UnwindWithOptions adds an $unwind stage with additional options.
//
// Example:
//
//	pipeline.UnwindWithOptions("$items", true, "itemIndex")
func (p *Pipeline) UnwindWithOptions(path string, preserveNullAndEmpty bool, includeArrayIndex string) *Pipeline {
	unwindSpec := bson.M{"path": path}
	if preserveNullAndEmpty {
		unwindSpec["preserveNullAndEmptyArrays"] = true
	}
	if includeArrayIndex != "" {
		unwindSpec["includeArrayIndex"] = includeArrayIndex
	}
	p.stages = append(p.stages, bson.M{"$unwind": unwindSpec})
	return p
}

// Lookup adds a $lookup stage for left outer join with another collection.
//
// Example:
//
//	pipeline.Lookup("orders", "customer_id", "_id", "customerOrders")
func (p *Pipeline) Lookup(from, localField, foreignField, as string) *Pipeline {
	p.stages = append(p.stages, bson.M{
		"$lookup": bson.M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	})
	return p
}

// LookupWithPipeline adds a $lookup stage with a sub-pipeline.
func (p *Pipeline) LookupWithPipeline(from string, let bson.M, pipeline []bson.M, as string) *Pipeline {
	lookupSpec := bson.M{
		"from":     from,
		"pipeline": pipeline,
		"as":       as,
	}
	if let != nil {
		lookupSpec["let"] = let
	}
	p.stages = append(p.stages, bson.M{"$lookup": lookupSpec})
	return p
}

// AddFields adds an $addFields stage to add new fields to documents.
//
// Example:
//
//	pipeline.AddFields(bson.M{"fullName": bson.M{"$concat": []string{"$firstName", " ", "$lastName"}}})
func (p *Pipeline) AddFields(fields bson.M) *Pipeline {
	p.stages = append(p.stages, bson.M{"$addFields": fields})
	return p
}

// Set is an alias for AddFields (MongoDB 4.2+).
func (p *Pipeline) Set(fields bson.M) *Pipeline {
	p.stages = append(p.stages, bson.M{"$set": fields})
	return p
}

// Unset adds an $unset stage to remove fields from documents.
//
// Example:
//
//	pipeline.Unset("password", "internalField")
func (p *Pipeline) Unset(fields ...string) *Pipeline {
	if len(fields) == 1 {
		p.stages = append(p.stages, bson.M{"$unset": fields[0]})
	} else {
		p.stages = append(p.stages, bson.M{"$unset": fields})
	}
	return p
}

// ReplaceRoot adds a $replaceRoot stage to replace the document with a subdocument.
//
// Example:
//
//	pipeline.ReplaceRoot("$embedded")
func (p *Pipeline) ReplaceRoot(newRoot any) *Pipeline {
	p.stages = append(p.stages, bson.M{"$replaceRoot": bson.M{"newRoot": newRoot}})
	return p
}

// Count adds a $count stage to count the number of documents.
//
// Example:
//
//	pipeline.Count("total")
func (p *Pipeline) Count(field string) *Pipeline {
	p.stages = append(p.stages, bson.M{"$count": field})
	return p
}

// Facet adds a $facet stage to process multiple aggregation pipelines.
//
// Example:
//
//	pipeline.Facet(bson.M{
//	    "byCategory": []bson.M{{"$group": bson.M{"_id": "$category"}}},
//	    "byStatus": []bson.M{{"$group": bson.M{"_id": "$status"}}},
//	})
func (p *Pipeline) Facet(facets bson.M) *Pipeline {
	p.stages = append(p.stages, bson.M{"$facet": facets})
	return p
}

// Bucket adds a $bucket stage to categorize documents into groups.
func (p *Pipeline) Bucket(groupBy any, boundaries []any, defaultBucket any, output bson.M) *Pipeline {
	bucketSpec := bson.M{
		"groupBy":    groupBy,
		"boundaries": boundaries,
	}
	if defaultBucket != nil {
		bucketSpec["default"] = defaultBucket
	}
	if output != nil {
		bucketSpec["output"] = output
	}
	p.stages = append(p.stages, bson.M{"$bucket": bucketSpec})
	return p
}

// Sample adds a $sample stage to randomly select documents.
func (p *Pipeline) Sample(size int64) *Pipeline {
	p.stages = append(p.stages, bson.M{"$sample": bson.M{"size": size}})
	return p
}

// Out adds an $out stage to write results to a collection.
// Note: This must be the last stage in the pipeline.
func (p *Pipeline) Out(collection string) *Pipeline {
	p.stages = append(p.stages, bson.M{"$out": collection})
	return p
}

// Merge adds a $merge stage to write results to a collection (MongoDB 4.2+).
// Note: This must be the last stage in the pipeline.
func (p *Pipeline) Merge(into string, on []string, whenMatched, whenNotMatched string) *Pipeline {
	mergeSpec := bson.M{"into": into}
	if len(on) > 0 {
		mergeSpec["on"] = on
	}
	if whenMatched != "" {
		mergeSpec["whenMatched"] = whenMatched
	}
	if whenNotMatched != "" {
		mergeSpec["whenNotMatched"] = whenNotMatched
	}
	p.stages = append(p.stages, bson.M{"$merge": mergeSpec})
	return p
}

// Raw adds a raw stage to the pipeline.
// Use this for stages not covered by the builder.
func (p *Pipeline) Raw(stage bson.M) *Pipeline {
	p.stages = append(p.stages, stage)
	return p
}

// ---- Accumulator helpers for use with Group/GroupBy ----

// Sum creates a $sum accumulator expression.
//
// Example:
//
//	spec.Sum("$amount")      // Sum of amount field
//	spec.Sum(1)              // Count documents
func Sum(expr any) bson.M {
	return bson.M{"$sum": expr}
}

// Avg creates an $avg accumulator expression.
func Avg(expr any) bson.M {
	return bson.M{"$avg": expr}
}

// MinAcc creates a $min accumulator expression.
func MinAcc(expr any) bson.M {
	return bson.M{"$min": expr}
}

// MaxAcc creates a $max accumulator expression.
func MaxAcc(expr any) bson.M {
	return bson.M{"$max": expr}
}

// First creates a $first accumulator expression.
func First(expr any) bson.M {
	return bson.M{"$first": expr}
}

// Last creates a $last accumulator expression.
func Last(expr any) bson.M {
	return bson.M{"$last": expr}
}

// Push creates a $push accumulator expression.
func PushAcc(expr any) bson.M {
	return bson.M{"$push": expr}
}

// AddToSetAcc creates an $addToSet accumulator expression.
func AddToSetAcc(expr any) bson.M {
	return bson.M{"$addToSet": expr}
}

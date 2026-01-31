package spec_test

import (
	"reflect"
	"testing"

	"github.com/dElCIoGio/mongox/spec"

	"go.mongodb.org/mongo-driver/bson"
)

func TestPipelineMatch(t *testing.T) {
	pipeline := spec.NewPipeline().
		Match(spec.Eq("status", "active"))

	got := pipeline.ToPipeline()
	want := []bson.M{
		{"$match": bson.M{"status": "active"}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pipeline Match mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPipelineMatchWithNil(t *testing.T) {
	pipeline := spec.NewPipeline().
		Match(nil)

	got := pipeline.ToPipeline()
	if len(got) != 0 {
		t.Fatalf("Pipeline Match(nil) should add no stages, got: %#v", got)
	}
}

func TestPipelineProject(t *testing.T) {
	pipeline := spec.NewPipeline().
		Project(bson.M{"name": 1, "_id": 0})

	got := pipeline.ToPipeline()
	want := []bson.M{
		{"$project": bson.M{"name": 1, "_id": 0}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pipeline Project mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPipelineGroup(t *testing.T) {
	pipeline := spec.NewPipeline().
		Group(bson.M{
			"_id":   "$category",
			"total": bson.M{"$sum": "$amount"},
		})

	got := pipeline.ToPipeline()
	want := []bson.M{
		{"$group": bson.M{
			"_id":   "$category",
			"total": bson.M{"$sum": "$amount"},
		}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pipeline Group mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPipelineGroupBy(t *testing.T) {
	pipeline := spec.NewPipeline().
		GroupBy("$status", bson.M{
			"count": spec.Sum(1),
			"total": spec.Sum("$amount"),
		})

	got := pipeline.ToPipeline()

	// Check that the stage has $group with _id and accumulators
	if len(got) != 1 {
		t.Fatalf("expected 1 stage, got %d", len(got))
	}

	groupStage, ok := got[0]["$group"].(bson.M)
	if !ok {
		t.Fatal("expected $group stage")
	}

	if groupStage["_id"] != "$status" {
		t.Fatalf("expected _id=$status, got %v", groupStage["_id"])
	}

	if !reflect.DeepEqual(groupStage["count"], bson.M{"$sum": 1}) {
		t.Fatalf("expected count=$sum:1, got %v", groupStage["count"])
	}
}

func TestPipelineSort(t *testing.T) {
	pipeline := spec.NewPipeline().
		Sort(bson.D{{"total", -1}, {"name", 1}})

	got := pipeline.ToPipeline()
	want := []bson.M{
		{"$sort": bson.D{{"total", -1}, {"name", 1}}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pipeline Sort mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPipelineSortBy(t *testing.T) {
	pipeline := spec.NewPipeline().
		SortBy("created_at", -1)

	got := pipeline.ToPipeline()
	want := []bson.M{
		{"$sort": bson.M{"created_at": -1}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pipeline SortBy mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPipelineLimitSkip(t *testing.T) {
	pipeline := spec.NewPipeline().
		Skip(10).
		Limit(5)

	got := pipeline.ToPipeline()
	want := []bson.M{
		{"$skip": int64(10)},
		{"$limit": int64(5)},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pipeline Skip/Limit mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPipelineUnwind(t *testing.T) {
	pipeline := spec.NewPipeline().
		Unwind("$items")

	got := pipeline.ToPipeline()
	want := []bson.M{
		{"$unwind": "$items"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pipeline Unwind mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPipelineLookup(t *testing.T) {
	pipeline := spec.NewPipeline().
		Lookup("orders", "customer_id", "_id", "customerOrders")

	got := pipeline.ToPipeline()
	want := []bson.M{
		{"$lookup": bson.M{
			"from":         "orders",
			"localField":   "customer_id",
			"foreignField": "_id",
			"as":           "customerOrders",
		}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pipeline Lookup mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPipelineAddFields(t *testing.T) {
	pipeline := spec.NewPipeline().
		AddFields(bson.M{"fullName": bson.M{"$concat": []string{"$first", " ", "$last"}}})

	got := pipeline.ToPipeline()
	want := []bson.M{
		{"$addFields": bson.M{"fullName": bson.M{"$concat": []string{"$first", " ", "$last"}}}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pipeline AddFields mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPipelineCount(t *testing.T) {
	pipeline := spec.NewPipeline().
		Count("total")

	got := pipeline.ToPipeline()
	want := []bson.M{
		{"$count": "total"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pipeline Count mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPipelineSample(t *testing.T) {
	pipeline := spec.NewPipeline().
		Sample(10)

	got := pipeline.ToPipeline()
	want := []bson.M{
		{"$sample": bson.M{"size": int64(10)}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pipeline Sample mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPipelineChaining(t *testing.T) {
	pipeline := spec.NewPipeline().
		Match(spec.Eq("status", "active")).
		GroupBy("$category", bson.M{
			"total": spec.Sum("$amount"),
			"count": spec.Sum(1),
		}).
		SortBy("total", -1).
		Limit(10)

	got := pipeline.ToPipeline()

	if len(got) != 4 {
		t.Fatalf("expected 4 stages, got %d", len(got))
	}

	// Verify stage types
	if _, ok := got[0]["$match"]; !ok {
		t.Fatal("expected $match as first stage")
	}
	if _, ok := got[1]["$group"]; !ok {
		t.Fatal("expected $group as second stage")
	}
	if _, ok := got[2]["$sort"]; !ok {
		t.Fatal("expected $sort as third stage")
	}
	if _, ok := got[3]["$limit"]; !ok {
		t.Fatal("expected $limit as fourth stage")
	}
}

func TestAccumulatorHelpers(t *testing.T) {
	tests := []struct {
		name string
		got  bson.M
		want bson.M
	}{
		{"Sum", spec.Sum("$amount"), bson.M{"$sum": "$amount"}},
		{"Sum count", spec.Sum(1), bson.M{"$sum": 1}},
		{"Avg", spec.Avg("$score"), bson.M{"$avg": "$score"}},
		{"MinAcc", spec.MinAcc("$price"), bson.M{"$min": "$price"}},
		{"MaxAcc", spec.MaxAcc("$price"), bson.M{"$max": "$price"}},
		{"First", spec.First("$name"), bson.M{"$first": "$name"}},
		{"Last", spec.Last("$name"), bson.M{"$last": "$name"}},
		{"PushAcc", spec.PushAcc("$item"), bson.M{"$push": "$item"}},
		{"AddToSetAcc", spec.AddToSetAcc("$tag"), bson.M{"$addToSet": "$tag"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Fatalf("%s mismatch.\n got: %#v\nwant: %#v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestPipelineRaw(t *testing.T) {
	pipeline := spec.NewPipeline().
		Raw(bson.M{"$customStage": bson.M{"option": true}})

	got := pipeline.ToPipeline()
	want := []bson.M{
		{"$customStage": bson.M{"option": true}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pipeline Raw mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

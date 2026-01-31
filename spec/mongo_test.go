package spec_test

import (
	"reflect"
	"testing"

	mongospec "github.com/dElCIoGio/mongox/spec"

	"go.mongodb.org/mongo-driver/bson"
)

func TestEq(t *testing.T) {
	got := mongospec.Eq("paid", true).ToMongo()
	want := bson.M{"paid": true}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Eq mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestGte(t *testing.T) {
	got := mongospec.Gte("total", 100).ToMongo()
	want := bson.M{"total": bson.M{"$gte": 100}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Gte mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestAndTwoFilters(t *testing.T) {
	got := mongospec.And(
		mongospec.Eq("tenant_id", "t1"),
		mongospec.Eq("paid", true),
	).ToMongo()

	want := bson.M{
		"$and": []bson.M{
			{"tenant_id": "t1"},
			{"paid": true},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("And mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestAndSingleFilterReturnsFilterDirectly(t *testing.T) {
	got := mongospec.And(mongospec.Eq("paid", true)).ToMongo()
	want := bson.M{"paid": true}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("And(single) mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestOrTwoFilters(t *testing.T) {
	got := mongospec.Or(
		mongospec.Eq("region", "EU"),
		mongospec.Gt("total", 500),
	).ToMongo()

	want := bson.M{
		"$or": []bson.M{
			{"region": "EU"},
			{"total": bson.M{"$gt": 500}},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Or mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestNotUsesNor(t *testing.T) {
	got := mongospec.Not(mongospec.Eq("paid", true)).ToMongo()
	want := bson.M{"$nor": []bson.M{{"paid": true}}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Not mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestAndFlattensNestedAnd(t *testing.T) {
	got := mongospec.And(
		mongospec.And(
			mongospec.Eq("a", 1),
			mongospec.Eq("b", 2),
		),
		mongospec.Eq("c", 3),
	).ToMongo()

	want := bson.M{
		"$and": []bson.M{
			{"a": 1},
			{"b": 2},
			{"c": 3},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("And(flatten) mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestOrFlattensNestedOr(t *testing.T) {
	got := mongospec.Or(
		mongospec.Or(
			mongospec.Eq("a", 1),
			mongospec.Eq("b", 2),
		),
		mongospec.Eq("c", 3),
	).ToMongo()

	want := bson.M{
		"$or": []bson.M{
			{"a": 1},
			{"b": 2},
			{"c": 3},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Or(flatten) mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

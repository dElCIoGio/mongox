package spec_test

import (
	"reflect"
	"testing"

	"github.com/dElCIoGio/mongox/spec"

	"go.mongodb.org/mongo-driver/bson"
)

func TestNe(t *testing.T) {
	got := spec.Ne("status", "deleted").ToMongo()
	want := bson.M{"status": bson.M{"$ne": "deleted"}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Ne mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestGt(t *testing.T) {
	got := spec.Gt("age", 18).ToMongo()
	want := bson.M{"age": bson.M{"$gt": 18}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Gt mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestLt(t *testing.T) {
	got := spec.Lt("price", 100.0).ToMongo()
	want := bson.M{"price": bson.M{"$lt": 100.0}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Lt mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestLte(t *testing.T) {
	got := spec.Lte("quantity", 10).ToMongo()
	want := bson.M{"quantity": bson.M{"$lte": 10}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Lte mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestIn(t *testing.T) {
	got := spec.In("status", []string{"active", "pending"}).ToMongo()
	want := bson.M{"status": bson.M{"$in": []string{"active", "pending"}}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("In mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestNotIn(t *testing.T) {
	got := spec.NotIn("role", []string{"banned", "suspended"}).ToMongo()
	want := bson.M{"role": bson.M{"$nin": []string{"banned", "suspended"}}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NotIn mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestExists(t *testing.T) {
	t.Run("exists true", func(t *testing.T) {
		got := spec.Exists("email", true).ToMongo()
		want := bson.M{"email": bson.M{"$exists": true}}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Exists(true) mismatch.\n got: %#v\nwant: %#v", got, want)
		}
	})

	t.Run("exists false", func(t *testing.T) {
		got := spec.Exists("deleted_at", false).ToMongo()
		want := bson.M{"deleted_at": bson.M{"$exists": false}}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Exists(false) mismatch.\n got: %#v\nwant: %#v", got, want)
		}
	})
}

func TestAndWithNilFilters(t *testing.T) {
	got := spec.And(
		spec.Eq("a", 1),
		nil,
		spec.Eq("b", 2),
		nil,
	).ToMongo()

	want := bson.M{
		"$and": []bson.M{
			{"a": 1},
			{"b": 2},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("And with nils mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestOrWithNilFilters(t *testing.T) {
	got := spec.Or(
		nil,
		spec.Eq("x", 1),
		nil,
	).ToMongo()

	// Single non-nil filter should be returned directly
	want := bson.M{"x": 1}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Or with nils mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestAndAllNilsReturnsNil(t *testing.T) {
	got := spec.And(nil, nil, nil)

	if got != nil {
		t.Fatalf("And(nil, nil, nil) should return nil, got: %#v", got)
	}
}

func TestOrAllNilsReturnsNil(t *testing.T) {
	got := spec.Or(nil, nil)

	if got != nil {
		t.Fatalf("Or(nil, nil) should return nil, got: %#v", got)
	}
}

func TestNotNilReturnsNil(t *testing.T) {
	got := spec.Not(nil)

	if got != nil {
		t.Fatalf("Not(nil) should return nil, got: %#v", got)
	}
}

func TestComplexComposition(t *testing.T) {
	// Build a complex filter: (status == "active" AND age >= 18) OR (role == "admin")
	filter := spec.Or(
		spec.And(
			spec.Eq("status", "active"),
			spec.Gte("age", 18),
		),
		spec.Eq("role", "admin"),
	)

	got := filter.ToMongo()
	want := bson.M{
		"$or": []bson.M{
			{
				"$and": []bson.M{
					{"status": "active"},
					{"age": bson.M{"$gte": 18}},
				},
			},
			{"role": "admin"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Complex composition mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestNin(t *testing.T) {
	got := spec.Nin("status", []string{"deleted", "archived"}).ToMongo()
	want := bson.M{"status": bson.M{"$nin": []string{"deleted", "archived"}}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Nin mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestRegex(t *testing.T) {
	t.Run("without options", func(t *testing.T) {
		got := spec.Regex("name", "^john").ToMongo()
		want := bson.M{"name": bson.M{"$regex": "^john"}}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Regex mismatch.\n got: %#v\nwant: %#v", got, want)
		}
	})

	t.Run("with case-insensitive option", func(t *testing.T) {
		got := spec.Regex("email", "@example\\.com$", "i").ToMongo()
		want := bson.M{"email": bson.M{"$regex": "@example\\.com$", "$options": "i"}}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Regex with options mismatch.\n got: %#v\nwant: %#v", got, want)
		}
	})
}

func TestAll(t *testing.T) {
	got := spec.All("tags", []string{"mongodb", "database", "nosql"}).ToMongo()
	want := bson.M{"tags": bson.M{"$all": []string{"mongodb", "database", "nosql"}}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("All mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestSize(t *testing.T) {
	got := spec.Size("items", 5).ToMongo()
	want := bson.M{"items": bson.M{"$size": 5}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Size mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestElemMatch(t *testing.T) {
	t.Run("with filter", func(t *testing.T) {
		got := spec.ElemMatch("results", spec.And(
			spec.Gte("score", 80),
			spec.Lt("score", 90),
		)).ToMongo()
		want := bson.M{
			"results": bson.M{
				"$elemMatch": bson.M{
					"$and": []bson.M{
						{"score": bson.M{"$gte": 80}},
						{"score": bson.M{"$lt": 90}},
					},
				},
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ElemMatch mismatch.\n got: %#v\nwant: %#v", got, want)
		}
	})

	t.Run("with nil filter", func(t *testing.T) {
		got := spec.ElemMatch("results", nil).ToMongo()
		want := bson.M{"results": bson.M{"$elemMatch": bson.M{}}}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ElemMatch with nil mismatch.\n got: %#v\nwant: %#v", got, want)
		}
	})
}

func TestBetween(t *testing.T) {
	got := spec.Between("age", 18, 65).ToMongo()
	want := bson.M{
		"$and": []bson.M{
			{"age": bson.M{"$gte": 18}},
			{"age": bson.M{"$lte": 65}},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Between mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

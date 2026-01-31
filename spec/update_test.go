package spec_test

import (
	"reflect"
	"testing"

	"github.com/dElCIoGio/mongox/spec"

	"go.mongodb.org/mongo-driver/bson"
)

func TestSet(t *testing.T) {
	got := spec.Set("status", "active").ToBsonUpdate()
	want := bson.M{"$set": bson.M{"status": "active"}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Set mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestInc(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		got := spec.Inc("counter", 1).ToBsonUpdate()
		want := bson.M{"$inc": bson.M{"counter": 1}}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Inc mismatch.\n got: %#v\nwant: %#v", got, want)
		}
	})

	t.Run("negative", func(t *testing.T) {
		got := spec.Inc("stock", -5).ToBsonUpdate()
		want := bson.M{"$inc": bson.M{"stock": -5}}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Inc mismatch.\n got: %#v\nwant: %#v", got, want)
		}
	})
}

func TestPush(t *testing.T) {
	got := spec.Push("tags", "new-tag").ToBsonUpdate()
	want := bson.M{"$push": bson.M{"tags": "new-tag"}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Push mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPull(t *testing.T) {
	got := spec.Pull("tags", "old-tag").ToBsonUpdate()
	want := bson.M{"$pull": bson.M{"tags": "old-tag"}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pull mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestUnset(t *testing.T) {
	got := spec.Unset("obsolete_field").ToBsonUpdate()
	want := bson.M{"$unset": bson.M{"obsolete_field": ""}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Unset mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestSetFields(t *testing.T) {
	got := spec.SetFields(bson.M{"name": "John", "age": 30}).ToBsonUpdate()
	want := bson.M{"$set": bson.M{"name": "John", "age": 30}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SetFields mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestCombine(t *testing.T) {
	t.Run("multiple set operations", func(t *testing.T) {
		got := spec.Combine(
			spec.Set("name", "John"),
			spec.Set("age", 30),
		).ToBsonUpdate()
		want := bson.M{"$set": bson.M{"name": "John", "age": 30}}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Combine mismatch.\n got: %#v\nwant: %#v", got, want)
		}
	})

	t.Run("mixed operations", func(t *testing.T) {
		got := spec.Combine(
			spec.Set("status", "active"),
			spec.Inc("counter", 1),
		).ToBsonUpdate()
		want := bson.M{
			"$set": bson.M{"status": "active"},
			"$inc": bson.M{"counter": 1},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Combine mismatch.\n got: %#v\nwant: %#v", got, want)
		}
	})

	t.Run("with nils", func(t *testing.T) {
		got := spec.Combine(
			nil,
			spec.Set("name", "Jane"),
			nil,
		).ToBsonUpdate()
		want := bson.M{"$set": bson.M{"name": "Jane"}}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Combine mismatch.\n got: %#v\nwant: %#v", got, want)
		}
	})

	t.Run("single update returns unwrapped", func(t *testing.T) {
		original := spec.Set("field", "value")
		combined := spec.Combine(original)

		if combined != original {
			t.Fatal("Combine of single update should return the original")
		}
	})

	t.Run("all nils returns nil", func(t *testing.T) {
		got := spec.Combine(nil, nil)
		if got != nil {
			t.Fatalf("Combine of all nils should return nil, got: %#v", got)
		}
	})
}

func TestAddToSet(t *testing.T) {
	got := spec.AddToSet("tags", "unique-tag").ToBsonUpdate()
	want := bson.M{"$addToSet": bson.M{"tags": "unique-tag"}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AddToSet mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPopFirst(t *testing.T) {
	got := spec.PopFirst("queue").ToBsonUpdate()
	want := bson.M{"$pop": bson.M{"queue": -1}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PopFirst mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPopLast(t *testing.T) {
	got := spec.PopLast("stack").ToBsonUpdate()
	want := bson.M{"$pop": bson.M{"stack": 1}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PopLast mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestMul(t *testing.T) {
	got := spec.Mul("price", 1.1).ToBsonUpdate()
	want := bson.M{"$mul": bson.M{"price": 1.1}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Mul mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestMin(t *testing.T) {
	got := spec.Min("low_score", 50).ToBsonUpdate()
	want := bson.M{"$min": bson.M{"low_score": 50}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Min mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestMax(t *testing.T) {
	got := spec.Max("high_score", 100).ToBsonUpdate()
	want := bson.M{"$max": bson.M{"high_score": 100}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Max mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestRename(t *testing.T) {
	got := spec.Rename("old_name", "new_name").ToBsonUpdate()
	want := bson.M{"$rename": bson.M{"old_name": "new_name"}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Rename mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

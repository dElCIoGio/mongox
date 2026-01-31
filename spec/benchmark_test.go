package spec_test

import (
	"testing"

	"github.com/dElCIoGio/mongox/spec"

	"go.mongodb.org/mongo-driver/bson"
)

// ========== FILTER BENCHMARKS ==========

func BenchmarkEq(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filter := spec.Eq("status", "active")
		_ = filter.ToMongo()
	}
}

func BenchmarkComparison(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filter := spec.Gte("age", 18)
		_ = filter.ToMongo()
	}
}

func BenchmarkIn(b *testing.B) {
	values := []string{"pending", "active", "completed"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filter := spec.In("status", values)
		_ = filter.ToMongo()
	}
}

func BenchmarkRegex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filter := spec.Regex("email", "@example\\.com$", "i")
		_ = filter.ToMongo()
	}
}

func BenchmarkAndTwoFilters(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filter := spec.And(
			spec.Eq("status", "active"),
			spec.Gte("age", 18),
		)
		_ = filter.ToMongo()
	}
}

func BenchmarkAndFiveFilters(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filter := spec.And(
			spec.Eq("status", "active"),
			spec.Gte("age", 18),
			spec.Lte("age", 65),
			spec.Exists("email", true),
			spec.Ne("role", "banned"),
		)
		_ = filter.ToMongo()
	}
}

func BenchmarkOrThreeFilters(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filter := spec.Or(
			spec.Eq("role", "admin"),
			spec.Eq("role", "moderator"),
			spec.Eq("premium", true),
		)
		_ = filter.ToMongo()
	}
}

func BenchmarkComplexFilter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filter := spec.And(
			spec.Eq("status", "active"),
			spec.Or(
				spec.And(
					spec.Gte("age", 18),
					spec.Lte("age", 65),
				),
				spec.Eq("verified", true),
			),
			spec.Not(spec.Eq("role", "banned")),
		)
		_ = filter.ToMongo()
	}
}

func BenchmarkNestedAndFlattening(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filter := spec.And(
			spec.And(
				spec.Eq("a", 1),
				spec.Eq("b", 2),
			),
			spec.And(
				spec.Eq("c", 3),
				spec.Eq("d", 4),
			),
			spec.Eq("e", 5),
		)
		_ = filter.ToMongo()
	}
}

func BenchmarkBetween(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filter := spec.Between("price", 100, 500)
		_ = filter.ToMongo()
	}
}

func BenchmarkElemMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filter := spec.ElemMatch("results", spec.And(
			spec.Gte("score", 80),
			spec.Lt("score", 90),
		))
		_ = filter.ToMongo()
	}
}

// ========== UPDATE BENCHMARKS ==========

func BenchmarkSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		update := spec.Set("status", "active")
		_ = update.ToBsonUpdate()
	}
}

func BenchmarkInc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		update := spec.Inc("counter", 1)
		_ = update.ToBsonUpdate()
	}
}

func BenchmarkPush(b *testing.B) {
	for i := 0; i < b.N; i++ {
		update := spec.Push("tags", "new-tag")
		_ = update.ToBsonUpdate()
	}
}

func BenchmarkSetFields(b *testing.B) {
	fields := bson.M{
		"name":   "John",
		"age":    30,
		"email":  "john@example.com",
		"active": true,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		update := spec.SetFields(fields)
		_ = update.ToBsonUpdate()
	}
}

func BenchmarkCombineTwoUpdates(b *testing.B) {
	for i := 0; i < b.N; i++ {
		update := spec.Combine(
			spec.Set("status", "active"),
			spec.Inc("counter", 1),
		)
		_ = update.ToBsonUpdate()
	}
}

func BenchmarkCombineFiveUpdates(b *testing.B) {
	for i := 0; i < b.N; i++ {
		update := spec.Combine(
			spec.Set("name", "John"),
			spec.Set("age", 30),
			spec.Inc("visits", 1),
			spec.Push("history", "action"),
			spec.Max("high_score", 100),
		)
		_ = update.ToBsonUpdate()
	}
}

func BenchmarkCombineSameType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		update := spec.Combine(
			spec.Set("field1", "value1"),
			spec.Set("field2", "value2"),
			spec.Set("field3", "value3"),
			spec.Set("field4", "value4"),
			spec.Set("field5", "value5"),
		)
		_ = update.ToBsonUpdate()
	}
}

// ========== PIPELINE BENCHMARKS ==========

func BenchmarkSimplePipeline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pipeline := spec.NewPipeline().
			Match(spec.Eq("status", "active")).
			Limit(10)
		_ = pipeline.ToPipeline()
	}
}

func BenchmarkComplexPipeline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pipeline := spec.NewPipeline().
			Match(spec.And(
				spec.Eq("status", "active"),
				spec.Gte("age", 18),
			)).
			GroupBy("$category", bson.M{
				"total": spec.Sum("$amount"),
				"count": spec.Sum(1),
				"avg":   spec.Avg("$amount"),
			}).
			SortBy("total", -1).
			Limit(10)
		_ = pipeline.ToPipeline()
	}
}

func BenchmarkPipelineWithLookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pipeline := spec.NewPipeline().
			Match(spec.Eq("status", "active")).
			Lookup("orders", "customer_id", "_id", "orders").
			Unwind("$orders").
			Group(bson.M{
				"_id":        "$_id",
				"name":       spec.First("$name"),
				"totalSpent": spec.Sum("$orders.total"),
			}).
			SortBy("totalSpent", -1).
			Limit(10)
		_ = pipeline.ToPipeline()
	}
}

// ========== MEMORY ALLOCATION BENCHMARKS ==========

func BenchmarkFilterAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		filter := spec.And(
			spec.Eq("status", "active"),
			spec.Gte("age", 18),
			spec.Lte("age", 65),
		)
		_ = filter.ToMongo()
	}
}

func BenchmarkUpdateAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		update := spec.Combine(
			spec.Set("name", "John"),
			spec.Inc("counter", 1),
			spec.Push("history", "action"),
		)
		_ = update.ToBsonUpdate()
	}
}

func BenchmarkPipelineAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		pipeline := spec.NewPipeline().
			Match(spec.Eq("status", "active")).
			GroupBy("$category", bson.M{"count": spec.Sum(1)}).
			SortBy("count", -1).
			Limit(10)
		_ = pipeline.ToPipeline()
	}
}

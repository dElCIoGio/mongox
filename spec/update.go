package spec

import "go.mongodb.org/mongo-driver/bson"

// Update represents a MongoDB update operation that can be translated to bson.M.
// Updates are composable building blocks for modifying documents in a type-safe manner.
//
// Use the provided update functions (Set, Inc, Push, etc.) to create updates,
// and combine them with Combine() to merge multiple operations.
//
// Example:
//
//	update := Combine(
//	    Set("status", "active"),
//	    Inc("login_count", 1),
//	)
//	// MongoDB: {"$set": {"status": "active"}, "$inc": {"login_count": 1}}
type Update interface {
	// ToBsonUpdate converts the update to a MongoDB bson.M update document.
	ToBsonUpdate() bson.M
}

// ---- Single-field update operations ----

type setUpdate struct {
	field string
	value any
}

func (u setUpdate) ToBsonUpdate() bson.M {
	return bson.M{"$set": bson.M{u.field: u.value}}
}

// Set creates an update that sets a field to the specified value.
// If the field doesn't exist, it will be created.
//
// MongoDB equivalent: {$set: {field: value}}
//
// Example:
//
//	Set("name", "John")        // {"$set": {"name": "John"}}
//	Set("active", true)        // {"$set": {"active": true}}
//	Set("nested.field", 42)    // {"$set": {"nested.field": 42}}
func Set(field string, value any) Update {
	return setUpdate{field: field, value: value}
}

type incUpdate struct {
	field string
	value any
}

func (u incUpdate) ToBsonUpdate() bson.M {
	return bson.M{"$inc": bson.M{u.field: u.value}}
}

// Inc creates an update that increments a numeric field by the specified value.
// Use positive values to increment, negative values to decrement.
// If the field doesn't exist, it will be created with the increment value.
//
// MongoDB equivalent: {$inc: {field: value}}
//
// Example:
//
//	Inc("view_count", 1)       // Increment by 1
//	Inc("stock", -5)           // Decrement by 5
//	Inc("balance", 100.50)     // Works with floats too
func Inc(field string, value any) Update {
	return incUpdate{field: field, value: value}
}

type pushUpdate struct {
	field string
	value any
}

func (u pushUpdate) ToBsonUpdate() bson.M {
	return bson.M{"$push": bson.M{u.field: u.value}}
}

// Push creates an update that appends a value to an array field.
// If the field doesn't exist, it will be created as an array with the value.
//
// MongoDB equivalent: {$push: {field: value}}
//
// Example:
//
//	Push("tags", "featured")           // Add single tag
//	Push("comments", commentObject)    // Add embedded document
func Push(field string, value any) Update {
	return pushUpdate{field: field, value: value}
}

type pullUpdate struct {
	field string
	value any
}

func (u pullUpdate) ToBsonUpdate() bson.M {
	return bson.M{"$pull": bson.M{u.field: u.value}}
}

// Pull creates an update that removes all instances of a value from an array field.
// If the value appears multiple times, all occurrences are removed.
//
// MongoDB equivalent: {$pull: {field: value}}
//
// Example:
//
//	Pull("tags", "deprecated")         // Remove all "deprecated" tags
//	Pull("scores", 0)                  // Remove all zero scores
func Pull(field string, value any) Update {
	return pullUpdate{field: field, value: value}
}

type unsetUpdate struct {
	field string
}

func (u unsetUpdate) ToBsonUpdate() bson.M {
	return bson.M{"$unset": bson.M{u.field: ""}}
}

// Unset creates an update that removes a field from the document.
// Has no effect if the field doesn't exist.
//
// MongoDB equivalent: {$unset: {field: ""}}
//
// Example:
//
//	Unset("temporary_data")           // Remove the field entirely
//	Unset("user.old_password")        // Remove nested field
func Unset(field string) Update {
	return unsetUpdate{field: field}
}

// ---- Multi-field update operations ----

type setFieldsUpdate struct {
	fields bson.M
}

func (u setFieldsUpdate) ToBsonUpdate() bson.M {
	return bson.M{"$set": u.fields}
}

// SetFields creates an update that sets multiple fields at once.
// More efficient than multiple Set() calls when updating many fields.
//
// MongoDB equivalent: {$set: {field1: value1, field2: value2, ...}}
//
// Example:
//
//	SetFields(bson.M{
//	    "name": "John Doe",
//	    "email": "john@example.com",
//	    "updated_at": time.Now(),
//	})
func SetFields(fields bson.M) Update {
	return setFieldsUpdate{fields: fields}
}

// ---- Combined updates ----

type combinedUpdate struct {
	updates []Update
}

func (u combinedUpdate) ToBsonUpdate() bson.M {
	result := bson.M{}
	for _, update := range u.updates {
		for k, v := range update.ToBsonUpdate() {
			// If the key already exists, merge the values
			if existing, ok := result[k]; ok {
				if existingMap, ok := existing.(bson.M); ok {
					if newMap, ok := v.(bson.M); ok {
						for field, val := range newMap {
							existingMap[field] = val
						}
						continue
					}
				}
			}
			result[k] = v
		}
	}
	return result
}

// Combine merges multiple updates into a single update operation.
// Updates of the same type (e.g., multiple $set operations) are intelligently merged.
//
// Behavior:
//   - Nil updates are automatically ignored
//   - If only one non-nil update is provided, it returns that update directly
//   - Same-type operations are merged (e.g., two Set calls become one $set)
//   - Returns nil if all updates are nil
//
// MongoDB equivalent: Merged update document
//
// Example:
//
//	Combine(
//	    Set("name", "John"),
//	    Set("age", 30),
//	    Inc("visits", 1),
//	    Push("history", "login"),
//	)
//	// Result: {
//	//   "$set": {"name": "John", "age": 30},
//	//   "$inc": {"visits": 1},
//	//   "$push": {"history": "login"}
//	// }
func Combine(updates ...Update) Update {
	nonNil := make([]Update, 0, len(updates))
	for _, u := range updates {
		if u != nil {
			nonNil = append(nonNil, u)
		}
	}
	if len(nonNil) == 0 {
		return nil
	}
	if len(nonNil) == 1 {
		return nonNil[0]
	}
	return combinedUpdate{updates: nonNil}
}

// ---- Additional array operations ----

type addToSetUpdate struct {
	field string
	value any
}

func (u addToSetUpdate) ToBsonUpdate() bson.M {
	return bson.M{"$addToSet": bson.M{u.field: u.value}}
}

// AddToSet creates an update that adds a value to an array only if it doesn't already exist.
// Unlike Push, this prevents duplicate values in the array.
//
// MongoDB equivalent: {$addToSet: {field: value}}
//
// Example:
//
//	AddToSet("roles", "admin")         // Add only if "admin" not already present
//	AddToSet("visited_pages", "/home") // Track unique page visits
func AddToSet(field string, value any) Update {
	return addToSetUpdate{field: field, value: value}
}

type popUpdate struct {
	field    string
	position int
}

func (u popUpdate) ToBsonUpdate() bson.M {
	return bson.M{"$pop": bson.M{u.field: u.position}}
}

// PopFirst creates an update that removes the first element from an array field.
// Useful for implementing queue-like behavior (FIFO).
//
// MongoDB equivalent: {$pop: {field: -1}}
//
// Example:
//
//	PopFirst("message_queue")          // Remove oldest message
func PopFirst(field string) Update {
	return popUpdate{field: field, position: -1}
}

// PopLast creates an update that removes the last element from an array field.
// Useful for implementing stack-like behavior (LIFO).
//
// MongoDB equivalent: {$pop: {field: 1}}
//
// Example:
//
//	PopLast("undo_stack")              // Remove most recent action
func PopLast(field string) Update {
	return popUpdate{field: field, position: 1}
}

// ---- Numeric operations ----

type mulUpdate struct {
	field string
	value any
}

func (u mulUpdate) ToBsonUpdate() bson.M {
	return bson.M{"$mul": bson.M{u.field: u.value}}
}

// Mul creates an update that multiplies a numeric field by the specified value.
// Useful for percentage-based adjustments.
//
// MongoDB equivalent: {$mul: {field: value}}
//
// Example:
//
//	Mul("price", 1.1)                  // Increase price by 10%
//	Mul("score", 0.9)                  // Decrease score by 10%
//	Mul("quantity", 2)                 // Double the quantity
func Mul(field string, value any) Update {
	return mulUpdate{field: field, value: value}
}

type minUpdate struct {
	field string
	value any
}

func (u minUpdate) ToBsonUpdate() bson.M {
	return bson.M{"$min": bson.M{u.field: u.value}}
}

// Min creates an update that sets field to value only if value is less than the current value.
// Useful for tracking minimum values (e.g., lowest score, earliest date).
//
// MongoDB equivalent: {$min: {field: value}}
//
// Example:
//
//	Min("low_price", 50)               // Update only if new price is lower
//	Min("first_seen", time.Now())      // Track earliest occurrence
func Min(field string, value any) Update {
	return minUpdate{field: field, value: value}
}

type maxUpdate struct {
	field string
	value any
}

func (u maxUpdate) ToBsonUpdate() bson.M {
	return bson.M{"$max": bson.M{u.field: u.value}}
}

// Max creates an update that sets field to value only if value is greater than the current value.
// Useful for tracking maximum values (e.g., high scores, latest dates).
//
// MongoDB equivalent: {$max: {field: value}}
//
// Example:
//
//	Max("high_score", 100)             // Update only if new score is higher
//	Max("last_active", time.Now())     // Track most recent activity
func Max(field string, value any) Update {
	return maxUpdate{field: field, value: value}
}

type renameUpdate struct {
	oldField string
	newField string
}

func (u renameUpdate) ToBsonUpdate() bson.M {
	return bson.M{"$rename": bson.M{u.oldField: u.newField}}
}

// Rename creates an update that renames a field in the document.
// The field value is preserved, only the key changes.
//
// MongoDB equivalent: {$rename: {oldField: "newField"}}
//
// Example:
//
//	Rename("user_name", "username")    // Rename for consistency
//	Rename("old.path", "new.path")     // Works with nested fields
func Rename(oldField, newField string) Update {
	return renameUpdate{oldField: oldField, newField: newField}
}

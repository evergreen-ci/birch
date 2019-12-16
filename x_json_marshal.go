package birch

import (
	"time"

	"github.com/evergreen-ci/birch/bsontype"
	"github.com/evergreen-ci/birch/juniper"
)

// MarshalJSON produces a JSON representation of the Document,
// preserving the order of the keys, and type information for types
// that have no JSON equivlent using MongoDB's extended JSON format
// where needed.
func (d *Document) MarshalJSON() ([]byte, error) { return d.toJSON().MarshalJSON() }

func (d *Document) toJSON() *juniper.Document {
	iter := d.Iterator()
	out := juniper.DC.Make(d.Len())
	for iter.Next() {
		elem := iter.Element()
		out.Append(juniper.EC.Value(elem.Key(), elem.Value().toJSON()))
	}
	if iter.Err() != nil {
		return nil
	}

	return out
}

// MarshalJSON produces a JSON representation of an Array preserving
// the type information for the types that have no JSON equivalent
// using MongoDB's extended JSON format where needed.
func (a *Array) MarshalJSON() ([]byte, error) { return a.toJSON().MarshalJSON() }

func (a *Array) toJSON() *juniper.Array {
	iter := a.Iterator()
	out := juniper.AC.Make(a.Len())
	for iter.Next() {
		out.Append(iter.Value().toJSON())
	}
	if iter.Err() != nil {
		panic(iter.Err())
		return nil
	}

	return out
}

func (v *Value) MarshalJSON() ([]byte, error) { return v.toJSON().MarshalJSON() }

func (v *Value) toJSON() *juniper.Value {
	switch v.Type() {
	case bsontype.Double:
		return juniper.VC.Float64(v.Double())
	case bsontype.String:
		return juniper.VC.String(v.StringValue())
	case bsontype.EmbeddedDocument:
		return juniper.VC.Object(v.MutableDocument().toJSON())
	case bsontype.Array:
		return juniper.VC.Array(v.MutableArray().toJSON())
	case bsontype.Binary:
		t, d := v.Binary()

		return juniper.VC.ObjectFromElements(
			juniper.EC.ObjectFromElements("$binary",
				juniper.EC.String("base64", string(t)),
				juniper.EC.String("subType", string(d)),
			),
		)
	case bsontype.Undefined:
		return juniper.VC.ObjectFromElements(juniper.EC.Boolean("$undefined", true))
	case bsontype.ObjectID:
		return juniper.VC.ObjectFromElements(juniper.EC.String("$oid", v.ObjectID().Hex()))
	case bsontype.Boolean:
		return juniper.VC.Boolean(v.Boolean())
	case bsontype.DateTime:
		return juniper.VC.ObjectFromElements(juniper.EC.String("$date", v.Time().Format(time.RFC3339)))
	case bsontype.Null:
		return juniper.VC.Nil()
	case bsontype.Regex:
		pattern, opts := v.Regex()

		return juniper.VC.ObjectFromElements(
			juniper.EC.ObjectFromElements("$regularExpression",
				juniper.EC.String("pattern", pattern),
				juniper.EC.String("options", opts),
			),
		)
	case bsontype.DBPointer:
		ns, oid := v.DBPointer()

		return juniper.VC.ObjectFromElements(
			juniper.EC.ObjectFromElements("$dbPointer",
				juniper.EC.String("$ref", ns),
				juniper.EC.String("$id", oid.Hex()),
			),
		)
	case bsontype.JavaScript:
		return juniper.VC.ObjectFromElements(juniper.EC.String("$code", v.JavaScript()))
	case bsontype.Symbol:
		return juniper.VC.ObjectFromElements(juniper.EC.String("$symbol", v.Symbol()))
	case bsontype.CodeWithScope:
		code, scope := v.MutableJavaScriptWithScope()

		return juniper.VC.ObjectFromElements(
			juniper.EC.String("$code", code),
			juniper.EC.Object("$scope", scope.toJSON()),
		)
	case bsontype.Int32:
		return juniper.VC.Int32(v.Int32())
	case bsontype.Timestamp:
		t, i := v.Timestamp()

		return juniper.VC.ObjectFromElements(
			juniper.EC.ObjectFromElements("$timestamp",
				juniper.EC.Int64("t", int64(t)),
				juniper.EC.Int64("i", int64(i)),
			),
		)
	case bsontype.Int64:
		return juniper.VC.Int64(v.Int64())
	case bsontype.Decimal128:
		return juniper.VC.ObjectFromElements(juniper.EC.String("$numberDecimal", v.Decimal128().String()))
	case bsontype.MinKey:
		return juniper.VC.ObjectFromElements(juniper.EC.Int("$minKey", 1))
	case bsontype.MaxKey:
		return juniper.VC.ObjectFromElements(juniper.EC.Int("$maxKey", 1))
	default:
		return nil
	}
}

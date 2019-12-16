package birch

import (
	"github.com/evergreen-ci/birch/juniper"
	"github.com/pkg/errors"
)

// UnmarshalJSON converts the contents of a document to JSON
// recursively, preserving the order of keys and the rich types from
// bson using MongoDB's extended JSON format for BSON types that have
// no equivalent in JSON.
//
// The underlying document is not emptied before this operation, which
// for non-empty documents could result in duplicate keys.
func (d *Document) UnmarshalJSON(in []byte) error {
	res, err := juniper.ParseBytes(in)
	if err != nil {
		return errors.WithStack(err)
	}
	if !res.IsObject() {
		return errors.New("cannot unmarshal values or arrays into Documents")
	}

	res.ForEach(func(key, value juniper.Result) bool {
		switch {
		case value.Type == juniper.String:
			d.Append(EC.String(key.Str, value.Str))
		case value.Type == juniper.Null:
			d.Append(EC.Null(key.Str))
		case value.Type == juniper.True:
			d.Append(EC.Boolean(key.Str, true))
		case value.Type == juniper.False:
			d.Append(EC.Boolean(key.Str, false))
		case value.Type == juniper.Number:
		case value.IsArray():
		case value.IsObject():
		}
		return true
	})

	return nil
}

// func toBSON() (*Element, error) {
// 	var elem *Element
// 	switch elem.Type {
// 	case jsoniter.InvalidValue:
// 		return nil, errors.New("encountered invalid type while parsing json")
// 	case jsoniter.StringValue:
// 		return EC.String(elem.Key, elem.Value().ToString()), nil
// 	case jsoniter.NumberValue:
// 		val := elem.Iter.Read()
// 		num, ok := val.(json.Number)
// 		if !ok {
// 			return nil, errors.Errorf("problem parsing number for %s [%T]", elem.Key, val)
// 		}

// 		if itg64, ok := num.Int64(); ok == nil {
// 			return EC.Int(elem.Key, int(itg64)), nil
// 		} else if flt64, ok := num.Float64(); ok == nil {
// 			return EC.Double(elem.Key, flt64), nil
// 		} else {
// 			return nil, errors.Errorf("problem parsing number '%s' [%T]", num.String(), val)
// 		}
// 	case jsoniter.NilValue:
// 		return EC.Null(elem.Key), nil
// 	case jsoniter.BoolValue:
// 		return EC.Boolean(elem.Key, elem.Value().ToBool()), nil
// 	case jsoniter.ArrayValue:
// 		vals := []*Value{}
// 		for elem.Iter.ReadArray() {
// 			aelm, err := (jsonElement{
// 				Key:  "",
// 				Type: elem.Iter.WhatIsNext(),
// 				Iter: elem.Iter,
// 			}).toBSON()

// 			if err != nil {
// 				return nil, errors.WithStack(err)
// 			}

// 			vals = append(vals, aelm.value)
// 		}

// 		return EC.ArrayFromElements(elem.Key, vals...), nil
// 	case jsoniter.ObjectValue:
// 		doc := DC.New()
// 	SUB_DOCUMENT_LOOP:
// 		for key := elem.Iter.ReadObject(); key != ""; {
// 			switch key {
// 			case "$minKey":
// 				return EC.MinKey(elem.Key), nil
// 			case "$maxKey":
// 				return EC.MaxKey(elem.Key), nil
// 			case "$numberDecimal":
// 				val, err := decimal.ParseDecimal128(elem.Iter.ReadString())
// 				if err != nil {
// 					return nil, errors.WithStack(err)
// 				}

// 				return EC.Decimal128(elem.Key, val), nil
// 			case "$timestamp":
// 				var (
// 					t   int64
// 					i   int64
// 					err error
// 				)

// 				for i := 0; i < 3; i++ {
// 					k := elem.Iter.ReadObject()
// 					switch k {
// 					case "t":
// 						val := elem.Iter.Read()
// 						num, ok := val.(jsoniter.Number)
// 						if !ok {
// 							return nil, errors.Errorf("problem decoding number for timestamp at %s [%T]", elem.Key, val)
// 						}
// 						if t, err = num.Int64(); err != nil {
// 							return nil, errors.Wrapf(err, "problem parsing for timestamp at %s", elem.Key)
// 						}
// 					case "i":
// 						val := elem.Iter.Read()
// 						num, ok := val.(jsoniter.Number)
// 						if !ok {
// 							return nil, errors.Errorf("problem decoding increment for timestamp at %s [%T]", elem.Key, val)
// 						}
// 						if t, err = num.Int64(); err != nil {
// 							return nil, errors.Wrapf(err, "problem parsing for timestamp at %s", elem.Key)
// 						}
// 					case "":
// 						break
// 					default:
// 						return nil, errors.Errorf("problem decoding timestamp for %s, found '%s' ", elem.Key, k)
// 					}
// 				}

// 				return EC.Timestamp(elem.Key, uint32(t), uint32(i)), nil
// 			case "$code":
// 				var scope *Document
// 				js := elem.Iter.ReadString()
// 				if second := elem.Iter.ReadObject(); second == "" {
// 					return EC.JavaScript(elem.Key, js), nil
// 				} else if second == "$scope" {
// 					scopeElem, err := (jsonElement{
// 						Key:  "",
// 						Type: elem.Iter.WhatIsNext(),
// 						Iter: elem.Iter,
// 					}).toBSON()
// 					if err != nil {
// 						return nil, errors.WithStack(err)
// 					}
// 					var ok bool
// 					scope, ok = scopeElem.Value().MutableDocumentOK()
// 					if !ok {
// 						return nil, errors.New("failed to properly decode a code with scope object")
// 					}
// 				} else {
// 					return nil, errors.Errorf("invalid key '%s' in code with scope for %s", second, elem.Key)
// 				}

// 				return EC.CodeWithScope(elem.Key, js, scope), nil
// 			case "$dbPointer":
// 				var (
// 					ns  string
// 					oid string
// 				)

// 				for i := 0; i < 3; i++ {
// 					k := elem.Iter.ReadObject()
// 					switch k {
// 					case "$ref":
// 						ns = elem.Iter.ReadString()
// 					case "$id":
// 						oid = elem.Iter.ReadString()
// 					case "":
// 						break
// 					default:
// 						return nil, errors.Errorf("problem decoding regex for %s, found '%s' ", elem.Key, k)
// 					}
// 				}

// 				oidp, err := types.ObjectIDFromHex(oid)
// 				if err != nil {
// 					return nil, errors.Wrapf(err, "problem parsing oid from dbref at %s", elem.Key)
// 				}

// 				return EC.DBPointer(elem.Key, ns, oidp), nil
// 			case "$regularExpression":
// 				var (
// 					pattern string
// 					options string
// 				)

// 				for i := 0; i < 3; i++ {
// 					k := elem.Iter.ReadObject()
// 					switch k {
// 					case "pattern":
// 						pattern = elem.Iter.ReadString()
// 					case "options":
// 						options = elem.Iter.ReadString()
// 					case "":
// 						break
// 					default:
// 						return nil, errors.Errorf("problem decoding regex for %s, found '%s' ", elem.Key, k)
// 					}
// 				}

// 				return EC.Regex(elem.Key, pattern, options), nil
// 			case "$date":
// 				date, err := time.Parse(time.RFC3339, elem.Iter.ReadString())
// 				if err != nil {
// 					return nil, errors.WithStack(err)
// 				}
// 				return EC.Time(elem.Key, date), nil
// 			case "$oid":
// 				oid, err := types.ObjectIDFromHex(elem.Iter.ReadString())
// 				if err != nil {
// 					return nil, errors.WithStack(err)
// 				}

// 				return EC.ObjectID(elem.Key, oid), nil
// 			case "$undefined":
// 				return EC.Undefined(elem.Key), nil
// 			case "$binary":
// 				return EC.Binary(elem.Key, []byte(elem.Iter.ReadString())), nil
// 			default:
// 				tpe := elem.Iter.WhatIsNext()

// 				if tpe == jsoniter.InvalidValue {
// 					break SUB_DOCUMENT_LOOP
// 				}

// 				oelm, err := (jsonElement{
// 					Key:  key,
// 					Type: tpe,
// 					Iter: elem.Iter,
// 				}).toBSON()
// 				if err != nil {
// 					return nil, errors.WithStack(err)
// 				}
// 				doc.Append(oelm)
// 			}
// 		}

// 		return EC.SubDocument(elem.Key, doc), nil
// 	default:
// 		return nil, errors.Errorf("unknown json: key=%s type=%v", elem.Key, elem.Type)
// 	}
// }

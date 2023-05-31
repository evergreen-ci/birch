package internal

import (
	"bytes"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestRandomData is a fuzzing test that throws random data at the Parse
// function looking for panics.
func TestRandomData(t *testing.T) {
	var lstr string
	defer func() {
		if v := recover(); v != nil {
			println("'" + hex.EncodeToString([]byte(lstr)) + "'")
			println("'" + lstr + "'")
			panic(v)
		}
	}()
	b := make([]byte, 200)
	for i := 0; i < 2000000; i++ {
		n, err := cryptorand.Read(b[:rand.Int()%len(b)])
		if err != nil {
			t.Fatal(err)
		}
		lstr = string(b[:n])
		require.NotPanics(t, func() {
			_, _ = Parse(lstr)
		})
	}
}

func must(res Result, err error) Result {
	if err != nil {
		panic(err)
	}

	return res
}

func TestParseAny(t *testing.T) {
	assert(t, must(Parse("100")).Float() == 100)
	assert(t, must(Parse("true")).Bool())
	assert(t, must(Parse("false")).Bool() == false)
}

func TestTypes(t *testing.T) {
	assert(t, (Result{Type: String}).Type.String() == "String")
	assert(t, (Result{Type: Number}).Type.String() == "Number")
	assert(t, (Result{Type: Null}).Type.String() == "Null")
	assert(t, (Result{Type: False}).Type.String() == "False")
	assert(t, (Result{Type: True}).Type.String() == "True")
	assert(t, (Result{Type: JSON}).Type.String() == "JSON")
	assert(t, (Result{Type: 100}).Type.String() == "")
	// bool
	assert(t, (Result{Type: String, Str: "true"}).Bool())
	assert(t, (Result{Type: True}).Bool())
	assert(t, (Result{Type: False}).Bool() == false)
	assert(t, (Result{Type: Number, Num: 1}).Bool())
	// int
	assert(t, (Result{Type: String, Str: "1"}).Int() == 1)
	assert(t, (Result{Type: True}).Int() == 1)
	assert(t, (Result{Type: False}).Int() == 0)
	assert(t, (Result{Type: Number, Num: 1}).Int() == 1)
	// uint
	assert(t, (Result{Type: String, Str: "1"}).Uint() == 1)
	assert(t, (Result{Type: True}).Uint() == 1)
	assert(t, (Result{Type: False}).Uint() == 0)
	assert(t, (Result{Type: Number, Num: 1}).Uint() == 1)
	// float
	assert(t, (Result{Type: String, Str: "1"}).Float() == 1)
	assert(t, (Result{Type: True}).Float() == 1)
	assert(t, (Result{Type: False}).Float() == 0)
	assert(t, (Result{Type: Number, Num: 1}).Float() == 1)
}
func TestForEach(t *testing.T) {
	Result{}.ForEach(nil)
	Result{Type: String, Str: "Hello"}.ForEach(func(_, value Result) bool {
		assert(t, value.String() == "Hello")
		return false
	})
	Result{Type: JSON, Raw: "*invalid*"}.ForEach(nil)

	json := ` {"name": {"first": "Janet","last": "Prichard"},
	"asd\nf":"\ud83d\udd13","age": 47}`
	var count int
	must(ParseBytes([]byte(json))).ForEach(func(key, value Result) bool {
		count++
		return true
	})
	assert(t, count == 3)
}
func TestMap(t *testing.T) {
	assert(t, len(must(ParseBytes([]byte(`"asdf"`))).Map()) == 0)
	assert(t, len(Result{Type: JSON, Raw: "**invalid**"}.Map()) == 0)
	assert(t, Result{Type: JSON, Raw: "**invalid**"}.Value() == nil)
	assert(t, Result{Type: JSON, Raw: "{"}.Map() != nil)
}
func TestUnescape(t *testing.T) {
	unescape(string([]byte{'\\', '\\', 0}))
	unescape(string([]byte{'\\', '/', '\\', 'b', '\\', 'f'}))
}
func assert(t testing.TB, cond bool) {
	if !cond {
		panic("assert failed")
	}
}

var exampleJSON = `{
	"widget": {
		"debug": "on",
		"window": {
			"title": "Sample Konfabulator Widget",
			"name": "main_window",
			"width": 500,
			"height": 500
		},
		"image": {
			"src": "Images/Sun.png",
			"hOffset": 250,
			"vOffset": 250,
			"alignment": "center"
		},
		"text": {
			"data": "Click Here",
			"size": 36,
			"style": "bold",
			"vOffset": 100,
			"alignment": "center",
			"onMouseUp": "sun1.opacity = (sun1.opacity / 100) * 90;"
		}
	}
}`

func TestNewParse(t *testing.T) {
	//fmt.Printf("%v\n", parse2(exampleJSON, "widget").String())
}

func TestUnmarshalMap(t *testing.T) {
	var m1 = must(Parse(exampleJSON)).Value().(map[string]interface{})
	var m2 map[string]interface{}
	if err := json.Unmarshal([]byte(exampleJSON), &m2); err != nil {
		t.Fatal(err)
	}
	b1, err := json.Marshal(m1)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := json.Marshal(m2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b1, b2) {
		t.Fatal("b1 != b2")
	}
}

type ComplicatedType struct {
	Tagged    string `json:"tagged"`
	NotTagged bool
	Nested    struct {
		Yellow string `json:"yellow"`
	}
	NestedTagged struct {
		Green string
		Map   map[string]interface{}
		Ints  struct {
			Int   int `json:"int"`
			Int8  int8
			Int16 int16
			Int32 int32
			Int64 int64 `json:"int64"`
		}
		Uints struct {
			Uint   uint
			Uint8  uint8
			Uint16 uint16
			Uint32 uint32
			Uint64 uint64
		}
		Floats struct {
			Float64 float64
			Float32 float32
		}
		Byte byte
		Bool bool
	} `json:"nestedTagged"`
	LeftOut      string `json:"-"`
	SelfPtr      *ComplicatedType
	SelfSlice    []ComplicatedType
	SelfSlicePtr []*ComplicatedType
	SelfPtrSlice *[]ComplicatedType
	Interface    interface{} `json:"interface"`
	Array        [3]int
	Time         time.Time `json:"time"`
	Binary       []byte
	NonBinary    []byte
}

var complicatedJSON = `
{
	"tagged": "OK",
	"Tagged": "KO",
	"NotTagged": true,
	"unsettable": 101,
	"Nested": {
		"Yellow": "Green",
		"yellow": "yellow"
	},
	"nestedTagged": {
		"Green": "Green",
		"Map": {
			"this": "that",
			"and": "the other thing"
		},
		"Ints": {
			"Uint": 99,
			"Uint16": 16,
			"Uint32": 32,
			"Uint64": 65
		},
		"Uints": {
			"int": -99,
			"Int": -98,
			"Int16": -16,
			"Int32": -32,
			"int64": -64,
			"Int64": -65
		},
		"Uints": {
			"Float32": 32.32,
			"Float64": 64.64
		},
		"Byte": 254,
		"Bool": true
	},
	"LeftOut": "you shouldn't be here",
	"SelfPtr": {"tagged":"OK","nestedTagged":{"Ints":{"Uint32":32}}},
	"SelfSlice": [{"tagged":"OK","nestedTagged":{"Ints":{"Uint32":32}}}],
	"SelfSlicePtr": [{"tagged":"OK","nestedTagged":{"Ints":{"Uint32":32}}}],
	"SelfPtrSlice": [{"tagged":"OK","nestedTagged":{"Ints":{"Uint32":32}}}],
	"interface": "Tile38 Rocks!",
	"Interface": "Please Download",
	"Array": [0,2,3,4,5],
	"time": "2017-05-07T13:24:43-07:00",
	"Binary": "R0lGODlhPQBEAPeo",
	"NonBinary": [9,3,100,115]
}
`

func testvalid(t *testing.T, json string, expect bool) {
	t.Helper()
	_, ok := validpayload([]byte(json), 0)
	if ok != expect {
		t.Fatal("mismatch")
	}
}

func TestValidBasic(t *testing.T) {
	testvalid(t, "0", true)
	testvalid(t, "00", false)
	testvalid(t, "-00", false)
	testvalid(t, "-.", false)
	testvalid(t, "0.0", true)
	testvalid(t, "10.0", true)
	testvalid(t, "10e1", true)
	testvalid(t, "10EE", false)
	testvalid(t, "10E-", false)
	testvalid(t, "10E+", false)
	testvalid(t, "10E123", true)
	testvalid(t, "10E-123", true)
	testvalid(t, "10E-0123", true)
	testvalid(t, "", false)
	testvalid(t, " ", false)
	testvalid(t, "{}", true)
	testvalid(t, "{", false)
	testvalid(t, "-", false)
	testvalid(t, "-1", true)
	testvalid(t, "-1.", false)
	testvalid(t, "-1.0", true)
	testvalid(t, " -1.0", true)
	testvalid(t, " -1.0 ", true)
	testvalid(t, "-1.0 ", true)
	testvalid(t, "-1.0 i", false)
	testvalid(t, "-1.0 i", false)
	testvalid(t, "true", true)
	testvalid(t, " true", true)
	testvalid(t, " true ", true)
	testvalid(t, " True ", false)
	testvalid(t, " tru", false)
	testvalid(t, "false", true)
	testvalid(t, " false", true)
	testvalid(t, " false ", true)
	testvalid(t, " False ", false)
	testvalid(t, " fals", false)
	testvalid(t, "null", true)
	testvalid(t, " null", true)
	testvalid(t, " null ", true)
	testvalid(t, " Null ", false)
	testvalid(t, " nul", false)
	testvalid(t, " []", true)
	testvalid(t, " [true]", true)
	testvalid(t, " [ true, null ]", true)
	testvalid(t, " [ true,]", false)
	testvalid(t, `{"hello":"world"}`, true)
	testvalid(t, `{ "hello": "world" }`, true)
	testvalid(t, `{ "hello": "world", }`, false)
	testvalid(t, `{"a":"b",}`, false)
	testvalid(t, `{"a":"b","a"}`, false)
	testvalid(t, `{"a":"b","a":}`, false)
	testvalid(t, `{"a":"b","a":1}`, true)
	testvalid(t, `{"a":"b",2"1":2}`, false)
	testvalid(t, `{"a":"b","a": 1, "c":{"hi":"there"} }`, true)
	testvalid(t, `{"a":"b","a": 1, "c":{"hi":"there", "easy":["going",`+
		`{"mixed":"bag"}]} }`, true)
	testvalid(t, `""`, true)
	testvalid(t, `"`, false)
	testvalid(t, `"\n"`, true)
	testvalid(t, `"\"`, false)
	testvalid(t, `"\\"`, true)
	testvalid(t, `"a\\b"`, true)
	testvalid(t, `"a\\b\\\"a"`, true)
	testvalid(t, `"a\\b\\\uFFAAa"`, true)
	testvalid(t, `"a\\b\\\uFFAZa"`, false)
	testvalid(t, `"a\\b\\\uFFA"`, false)
	testvalid(t, complicatedJSON, true)
	testvalid(t, exampleJSON, true)
}

var jsonchars = []string{"{", "[", ",", ":", "}", "]", "1", "0", "true",
	"false", "null", `""`, `"\""`, `"a"`}

func makeRandomJSONChars(b []byte) {
	var bb []byte
	for len(bb) < len(b) {
		bb = append(bb, jsonchars[rand.Int()%len(jsonchars)]...)
	}
	copy(b, bb[:len(b)])
}

func TestValidRandom(t *testing.T) {
	b := make([]byte, 100000)
	start := time.Now()
	for time.Since(start) < time.Second*3 {
		n := rand.Int() % len(b)
		_, err := cryptorand.Read(b[:n])
		if err != nil {
			t.Fatal(err)
		}
		validpayload(b[:n], 0)
	}

	start = time.Now()
	for time.Since(start) < time.Second*3 {
		n := rand.Int() % len(b)
		makeRandomJSONChars(b[:n])
		validpayload(b[:n], 0)
	}
}

func TestResultRawForLiteral(t *testing.T) {
	for _, lit := range []string{"null", "true", "false"} {
		result := must(Parse(lit))
		if result.Raw != lit {
			t.Fatalf("expected '%v', got '%v'", lit, result.Raw)
		}
	}
}

func randomString() string {
	var key string
	N := 1 + rand.Int()%16
	for i := 0; i < N; i++ {
		r := rand.Int() % 62
		if r < 10 {
			key += string(byte('0' + r))
		} else if r-10 < 26 {
			key += string(byte('a' + r - 10))
		} else {
			key += string(byte('A' + r - 10 - 26))
		}
	}
	return `"` + key + `"`
}
func randomBool() string {
	switch rand.Int() % 2 {
	default:
		return "false"
	case 1:
		return "true"
	}
}
func randomNumber() string {
	return strconv.FormatInt(int64(rand.Int()%1000000), 10)
}

func randomObjectOrArray(keys []string, prefix string, array bool, depth int) (
	string, []string) {
	N := 5 + rand.Int()%5
	var json string
	if array {
		json = "["
	} else {
		json = "{"
	}
	for i := 0; i < N; i++ {
		if i > 0 {
			json += ","
		}
		var pkey string
		if array {
			pkey = prefix + "." + strconv.FormatInt(int64(i), 10)
		} else {
			key := randomString()
			pkey = prefix + "." + key[1:len(key)-1]
			json += key + `:`
		}
		keys = append(keys, pkey[1:])
		var kind int
		if depth == 5 {
			kind = rand.Int() % 4
		} else {
			kind = rand.Int() % 6
		}
		switch kind {
		case 0:
			json += randomString()
		case 1:
			json += randomBool()
		case 2:
			json += "null"
		case 3:
			json += randomNumber()
		case 4:
			var njson string
			njson, keys = randomObjectOrArray(keys, pkey, true, depth+1)
			json += njson
		case 5:
			var njson string
			njson, keys = randomObjectOrArray(keys, pkey, false, depth+1)
			json += njson
		}

	}
	if array {
		json += "]"
	} else {
		json += "}"
	}
	return json, keys
}

func BenchmarkValid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validateJSON([]byte(complicatedJSON))
	}
}

func BenchmarkValidBytes(b *testing.B) {
	complicatedJSON := []byte(complicatedJSON)
	for i := 0; i < b.N; i++ {
		validateJSON(complicatedJSON)
	}
}

func BenchmarkGoStdlibValidBytes(b *testing.B) {
	complicatedJSON := []byte(complicatedJSON)
	for i := 0; i < b.N; i++ {
		json.Valid(complicatedJSON)
	}
}

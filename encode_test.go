package btawel_test

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
	"time"

	"github.com/osamingo/boolconv"
	"github.com/tvlk-data/btawel"
)

type CQ struct {
	ID        string
	Qualifier string `bigtable:"qualifier"`
}

func TestGenerateColumnQualifiersMutation(t *testing.T) {

	// family is empty
	if _, err := btawel.GenerateColumnQualifiersMutation("", time.Now(), ""); err == nil {
		t.Error("error isn't occurred")
	}

	// slice is nil
	if _, err := btawel.GenerateColumnQualifiersMutation("fc", time.Now(), nil); err == nil {
		t.Error("error isn't occurred")
	}

	// slice is empty
	if _, err := btawel.GenerateColumnQualifiersMutation("fc", time.Now(), []interface{}{}); err == nil {
		t.Error("error isn't occurred")
	}

	s := []string{"hoge", "fuga", "foo", "bar"}

	ss := []*CQ{
		&CQ{ID: "1", Qualifier: s[0]},
		&CQ{ID: "2", Qualifier: s[1]},
		&CQ{ID: "3", Qualifier: s[2]},
		&CQ{ID: "4", Qualifier: s[3]},
		&CQ{ID: "5", Qualifier: ""},
	}

	ret, err := btawel.GenerateColumnQualifiersMutation("fc", time.Now(), ss)
	if err != nil {
		t.Error("failed to GenerateColumnQualifiersMutation. msg =", err)
	}

	ops := reflect.ValueOf(ret).Elem().FieldByName("ops")
	if ops.Len() != 4 {
		t.Errorf("expected ops length is %d got %d", 4, ops.Len())
	}

	for i := 0; i < ops.Len(); i++ {

		o := ops.Index(i).Elem().Field(0).Elem().Elem().Field(0).Elem()
		c := o.FieldByName("ColumnQualifier").Bytes()
		v := o.FieldByName("Value").Bytes()

		if v != nil {
			t.Error("value isn't nil")
		}

		notFound := true
		for i := range s {
			if s[i] == string(c) {
				notFound = false
			}
		}

		if notFound {
			t.Error("column qualifier is not found")
		}

	}
}

func TestGenerateColumnsMutation(t *testing.T) {

	// family is empty
	if _, err := btawel.GenerateColumnsMutation("", time.Now(), nil); err == nil {
		t.Error("error isn't occurred")
	}

	// struct is nil
	if _, err := btawel.GenerateColumnsMutation("fc", time.Now(), nil); err == nil {
		t.Error("error isn't occurred")
	}

	// struct hasn't fields
	if _, err := btawel.GenerateColumnsMutation("fc", time.Now(), struct{}{}); err == nil {
		t.Error("error isn't occurred")
	}

	// filed is unsupported type
	if _, err := btawel.GenerateColumnsMutation("fc", time.Now(), struct {
		S map[string]interface{} `bigtable:"wrong"`
	}{
		map[string]interface{}{
			"test": 1,
		},
	}); err == nil {
		t.Error("error isn't occurred")
	}

	s := struct {
		TBytes   []byte  `bigtable:"tbytes"`
		TString  string  `bigtable:"tstr"`
		TInt     int     `bigtable:"tint"`
		TUint    uint    `bigtable:"tuint"`
		TBool    bool    `bigtable:"tbool"`
		TOmitStr string  `bigtable:"tomitstr, omitempty"`
		TOmitInt int     `bigtable:"tomitint, omitempty"`
		TIgnore  int64   `bigtable:"-"`
		TInt8    int8    `bigtable:"tint8"`
		TInt16   int16   `bigtable:"tint16"`
		TInt32   int32   `bigtable:"tint32"`
		TInt64   int64   `bigtable:"tint64"`
		TFloat32 float32 `bigtable:"tfloat32"`
		TFloat64 float64 `bigtable:"tfloat64"`
		TNonTag  string
	}{
		TBytes:   []byte("bytebyte"),
		TString:  "test1",
		TInt:     100,
		TUint:    200,
		TBool:    true,
		TOmitStr: "test2",
		TIgnore:  1000,
		TInt8:    8,
		TInt16:   16,
		TInt32:   32,
		TInt64:   64,
		TFloat32: 3.2,
		TFloat64: 6.4,
		TNonTag:  "test3",
	}

	ret, err := btawel.GenerateColumnsMutation("fc", time.Now(), &s)
	if err != nil {
		t.Error("failed to GenerateColumnsMutation. msg =", err)
	}

	ops := reflect.ValueOf(ret).Elem().FieldByName("ops")
	if ops.Len() != 12 {
		t.Errorf("expected ops length is %d got %d", 12, ops.Len())
	}

	for i := 0; i < ops.Len(); i++ {

		o := ops.Index(i).Elem().Field(0).Elem().Elem().Field(0).Elem()
		c := o.FieldByName("ColumnQualifier").Bytes()
		v := o.FieldByName("Value").Bytes()

		switch string(c) {

		case "tbytes":
			vv := string(v)
			if vv != "bytebyte" {
				t.Errorf("expected %s got %s", "bytebyte", vv)
			}

		case "tstr":
			vv := string(v)
			if vv != "test1" {
				t.Errorf("expected %s got %s", "test1", vv)
			}

		case "tint":
			var vv int64
			binary.Read(bytes.NewBuffer(v), binary.BigEndian, &vv)
			if int(vv) != 100 {
				t.Errorf("expected %d got %d", 100, vv)
			}

		case "tuint":
			var vv uint64
			binary.Read(bytes.NewBuffer(v), binary.BigEndian, &vv)
			if uint(vv) != 200 {
				t.Errorf("expected %d got %d", 200, vv)
			}

		case "tbool":
			vv := boolconv.BtoB(v)
			if vv != boolconv.True {
				t.Errorf("expected %v got %v", boolconv.True, vv)
			}

		case "tomitstr":
			vv := string(v)
			if vv != "test2" {
				t.Errorf("expected %s got %s", "test2", vv)
			}

		case "tint8":
			var vv int8
			binary.Read(bytes.NewBuffer(v), binary.BigEndian, &vv)
			if vv != 8 {
				t.Errorf("expected %d got %d", 8, vv)
			}

		case "tint16":
			var vv int16
			binary.Read(bytes.NewBuffer(v), binary.BigEndian, &vv)
			if vv != 16 {
				t.Errorf("expected %d got %d", 16, vv)
			}

		case "tint32":
			var vv int32
			binary.Read(bytes.NewBuffer(v), binary.BigEndian, &vv)
			if vv != 32 {
				t.Errorf("expected %d got %d", 32, vv)
			}

		case "tint64":
			var vv int64
			binary.Read(bytes.NewBuffer(v), binary.BigEndian, &vv)
			if vv != 64 {
				t.Errorf("expected %d got %d", 64, vv)
			}

		case "tfloat32":
			var vv float32
			binary.Read(bytes.NewBuffer(v), binary.BigEndian, &vv)
			if vv != 3.2 {
				t.Errorf("expected %v got %v", 3.2, vv)
			}

		case "tfloat64":
			var vv float64
			binary.Read(bytes.NewBuffer(v), binary.BigEndian, &vv)
			if vv != 6.4 {
				t.Errorf("expected %v got %v", 6.4, vv)
			}

		default:
			t.Error("undefined ccolumn", c)
		}

	}

}

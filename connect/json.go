package connect

import (
	"encoding/json"
	"github.com/micro/go-micro/v2/codec/bytes"
	"google.golang.org/grpc/encoding"

	"github.com/golang/protobuf/jsonpb"
)

type wrapCodec struct{ encoding.Codec }

var jsonpbMarshaler = &jsonpb.Marshaler{
	EnumsAsInts:  false,
	EmitDefaults: true,
	OrigName:     true,
}

func (w wrapCodec) String() string {
	return w.Codec.Name()
}

func (w wrapCodec) Marshal(v interface{}) ([]byte, error) {
	b, ok := v.(*bytes.Frame)
	if ok {
		return b.Data, nil
	}
	return w.Codec.Marshal(v)
}

func (w wrapCodec) Unmarshal(data []byte, v interface{}) error {
	b, ok := v.(*bytes.Frame)
	if ok {
		b.Data = data
		return nil
	}
	if v == nil {
		return nil
	}
	return w.Codec.Unmarshal(data, v)
}

func RegisterJSONCodec() {
	encoding.RegisterCodec(wrapCodec{jsonCodec{}})
}

type jsonCodec struct{}

func (jsonCodec) Marshal(v interface{}) ([]byte, error) {
	//if pb, ok := v.(proto.Message); ok {
	//	s, err := jsonpbMarshaler.MarshalToString(pb)
	//	return []byte(s), err
	//}

	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v interface{}) error {
	if len(data) == 0 {
		return nil
	}
	//if pb, ok := v.(proto.Message); ok {
	//	return jsonpb.Unmarshal(b.NewReader(data), pb)
	//}
	return json.Unmarshal(data, v)
}

func (jsonCodec) Name() string {
	return "json"
}

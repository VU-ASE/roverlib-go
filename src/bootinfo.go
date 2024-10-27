// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    injectedService, err := UnmarshalInjectedService(bytes)
//    bytes, err = injectedService.Marshal()

package roverlib

import (
	"bytes"
	"encoding/json"
	"errors"
)

func UnmarshalInjectedService(data []byte) (InjectedService, error) {
	var r InjectedService
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *InjectedService) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// The object that injected into a rover process by roverd and then parsed by roverlib to be
// made available for the user process
type InjectedService struct {
	Configuration []Configuration `json:"configuration"`
	// The resolved input dependencies
	Inputs []Input `json:"inputs"`
	// The name of the service (only lowercase letters and hyphens)
	Name    *string  `json:"name,omitempty"`
	Outputs []Output `json:"outputs"`
	Tuning  Tuning   `json:"tuning"`
	// The specific version of the service
	Version *string     `json:"version,omitempty"`
	Service interface{} `json:"service"`
}

type Configuration struct {
	// Unique name of this configuration option
	Name *string `json:"name,omitempty"`
	// Whether or not this value can be tuned (ota)
	Tunable *bool `json:"tunable,omitempty"`
	// The type of this configuration option
	Type *Type `json:"type,omitempty"`
	// The value of this configuration option, which can be a string, integer, or float
	Value *Value `json:"value"`
}

type Input struct {
	// The name of the service for this dependency
	Service *string  `json:"service,omitempty"`
	Streams []Stream `json:"streams,omitempty"`
}

type Stream struct {
	// The (zmq) socket address that input can be read on
	Address *string `json:"address,omitempty"`
	// The name of the stream as outputted by the dependency service
	Name *string `json:"name,omitempty"`
}

type Output struct {
	// The (zmq) socket address that output can be written to
	Address *string `json:"address,omitempty"`
	// Name of the output published by this service
	Name *string `json:"name,omitempty"`
}

type Tuning struct {
	// (If enabled) the (zmq) socket address that tuning data can be read from
	Address *string `json:"address,omitempty"`
	// Whether or not live (ota) tuning is enabled
	Enabled *bool `json:"enabled,omitempty"`
}

// The type of this configuration option
type Type string

const (
	Float  Type = "float"
	Int    Type = "int"
	String Type = "string"
)

// The value of this configuration option, which can be a string, integer, or float
type Value struct {
	Double  *float64
	Integer *int64
	String  *string
}

func (x *Value) UnmarshalJSON(data []byte) error {
	object, err := unmarshalUnion(data, &x.Integer, &x.Double, nil, &x.String, false, nil, false, nil, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		return nil
	}
	return nil
}

func (x *Value) MarshalJSON() ([]byte, error) {
	return marshalUnion(x.Integer, x.Double, nil, x.String, false, nil, false, nil, false, nil, false, nil, false)
}

func unmarshalUnion(data []byte, pi **int64, pf **float64, pb **bool, ps **string, haveArray bool, pa interface{}, haveObject bool, pc interface{}, haveMap bool, pm interface{}, haveEnum bool, pe interface{}, nullable bool) (bool, error) {
	if pi != nil {
		*pi = nil
	}
	if pf != nil {
		*pf = nil
	}
	if pb != nil {
		*pb = nil
	}
	if ps != nil {
		*ps = nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return false, err
	}

	switch v := tok.(type) {
	case json.Number:
		if pi != nil {
			i, err := v.Int64()
			if err == nil {
				*pi = &i
				return false, nil
			}
		}
		if pf != nil {
			f, err := v.Float64()
			if err == nil {
				*pf = &f
				return false, nil
			}
			return false, errors.New("Unparsable number")
		}
		return false, errors.New("Union does not contain number")
	case float64:
		return false, errors.New("Decoder should not return float64")
	case bool:
		if pb != nil {
			*pb = &v
			return false, nil
		}
		return false, errors.New("Union does not contain bool")
	case string:
		if haveEnum {
			return false, json.Unmarshal(data, pe)
		}
		if ps != nil {
			*ps = &v
			return false, nil
		}
		return false, errors.New("Union does not contain string")
	case nil:
		if nullable {
			return false, nil
		}
		return false, errors.New("Union does not contain null")
	case json.Delim:
		if v == '{' {
			if haveObject {
				return true, json.Unmarshal(data, pc)
			}
			if haveMap {
				return false, json.Unmarshal(data, pm)
			}
			return false, errors.New("Union does not contain object")
		}
		if v == '[' {
			if haveArray {
				return false, json.Unmarshal(data, pa)
			}
			return false, errors.New("Union does not contain array")
		}
		return false, errors.New("Cannot handle delimiter")
	}
	return false, errors.New("Cannot unmarshal union")

}

func marshalUnion(pi *int64, pf *float64, pb *bool, ps *string, haveArray bool, pa interface{}, haveObject bool, pc interface{}, haveMap bool, pm interface{}, haveEnum bool, pe interface{}, nullable bool) ([]byte, error) {
	if pi != nil {
		return json.Marshal(*pi)
	}
	if pf != nil {
		return json.Marshal(*pf)
	}
	if pb != nil {
		return json.Marshal(*pb)
	}
	if ps != nil {
		return json.Marshal(*ps)
	}
	if haveArray {
		return json.Marshal(pa)
	}
	if haveObject {
		return json.Marshal(pc)
	}
	if haveMap {
		return json.Marshal(pm)
	}
	if haveEnum {
		return json.Marshal(pe)
	}
	if nullable {
		return json.Marshal(nil)
	}
	return nil, errors.New("Union must not be null")
}

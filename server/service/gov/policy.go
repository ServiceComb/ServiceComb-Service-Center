package gov

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"

	"k8s.io/kube-openapi/pkg/validation/spec"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

type ValueType string

var (
	MsgRequired    = "%s is required"
	MsgTooSmall    = "%s should be bigger than %d"
	MsgTooBig      = "%s should be smaller than %d"
	MsgTypeInt     = "%s must be a number"
	MsgTypeString  = "%s must be a string"
	MsgTypeBool    = "%s must be a bool"
	MsgSkip        = "skip checking %s"
	MsgUnknownType = "%s is unknown type"
)

//policies saves kind and policy schemas
var policies = make(map[string]*spec.Schema)

const (
	ValueTypeInt    ValueType = "int"
	ValueTypeString ValueType = "string"
	ValueTypeMap    ValueType = "map" // TODO not supported
	ValueTypeList   ValueType = "list"
	ValueTypeBool   ValueType = "bool"
)

//RegisterPolicy register a contract of one kind of policy
//this API is not thread safe, only use it during sc init
func RegisterPolicy(kind string, schema *spec.Schema) {
	policies[kind] = schema
}

//ValidateSpec validates spec attributes
func ValidateSpec(kind string, spec interface{}) error {
	schema, ok := policies[kind]
	if !ok {
		log.Warn(fmt.Sprintf("can not recognize %s", kind))
		return nil
	}
	validator := validate.NewSchemaValidator(schema, nil, "", strfmt.Default)
	errs := validator.Validate(spec).Errors
	if len(errs) != 0 {
		var str []string
		for i, err := range errs {
			if i != 0 {
				str = append(str, ";", err.Error())
			} else {
				str = append(str, err.Error())
			}
		}
		return errors.New(strings.Join(str, ";"))
	}
	return nil
}

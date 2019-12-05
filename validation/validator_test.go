package validation

import (
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/System-Glitch/goyave/v2/lang"
	"github.com/stretchr/testify/suite"
)

type ValidatorTestSuite struct {
	suite.Suite
}

func (suite *ValidatorTestSuite) SetupSuite() {
	lang.LoadDefault()
}

func (suite *ValidatorTestSuite) TestIsTypeDependant() {
	suite.True(isTypeDependent("min"))
	suite.False(isTypeDependent("required"))
}

func (suite *ValidatorTestSuite) TestIsRequired() {
	suite.True(isRequired([]string{"string", "required", "min:5"}))
	suite.False(isRequired([]string{"string", "min:5"}))
}

func (suite *ValidatorTestSuite) TestIsNullable() {
	suite.True(isNullable([]string{"string", "required", "nullable", "min:5"}))
	suite.False(isNullable([]string{"string", "min:5", "required"}))
}

func (suite *ValidatorTestSuite) TestIsArray() {
	suite.True(isArray([]string{"array", "required", "nullable", "min:5"}))
	suite.False(isArray([]string{"string", "min:5", "required"}))
}

func (suite *ValidatorTestSuite) TestArrayType() {
	suite.True(isArrayType("integer"))
	suite.False(isArrayType("file"))
}

func (suite *ValidatorTestSuite) TestParseRule() {
	rule, _, params := parseRule("required")
	suite.Equal("required", rule)
	suite.Equal(0, len(params))

	rule, _, params = parseRule("min:5")
	suite.Equal("min", rule)
	suite.Equal(1, len(params))
	suite.Equal("5", params[0])

	suite.Panics(func() {
		parseRule("invalid,rule")
	})

	suite.Panics(func() {
		parseRule("invalidrule")
	})

	rule, validatesArray, params := parseRule(">min:3")
	suite.Equal("min", rule)
	suite.Equal(1, len(params))
	suite.Equal("3", params[0])
	suite.True(validatesArray)
}

func (suite *ValidatorTestSuite) TestGetMessage() {
	suite.Equal("The :field is required.", getMessage("required", reflect.ValueOf("test"), "en-US", false))
	suite.Equal("The :field must be at least :min.", getMessage("min", reflect.ValueOf(42), "en-US", false))
	suite.Equal("The :field values must be at least :min.", getMessage("min", reflect.ValueOf(42), "en-US", true)) // TODO add all validation messages
}

func (suite *ValidatorTestSuite) TestAddRule() {
	suite.Panics(func() {
		AddRule("required", false, func(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
			return false
		})
	})

	AddRule("new_rule", false, func(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
		return true
	})
	_, ok := validationRules["new_rule"]
	suite.True(ok)
	suite.False(isTypeDependent("new_rule"))

	AddRule("new_rule_type_dependent", true, func(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
		return true
	})
	_, ok = validationRules["new_rule_type_dependent"]
	suite.True(ok)
	suite.True(isTypeDependent("new_rule_type_dependent"))
}

func (suite *ValidatorTestSuite) TestValidate() {
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	errors := Validate(rawRequest, nil, RuleSet{}, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("Malformed request", errors["error"][0])

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	errors = Validate(rawRequest, nil, RuleSet{}, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("Malformed JSON", errors["error"][0])

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	errors = Validate(rawRequest, map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}, RuleSet{
		"string": {"required", "string"},
		"number": {"required", "numeric", "min:10"},
	}, "en-US")
	suite.Equal(0, len(errors))

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	data := map[string]interface{}{
		"nullField": nil,
	}
	errors = Validate(rawRequest, data, RuleSet{
		"nullField": {"numeric"},
	}, "en-US")
	_, exists := data["nullField"]
	suite.False(exists)
	suite.Equal(0, len(errors))

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	data = map[string]interface{}{
		"nullField": nil,
	}
	errors = Validate(rawRequest, data, RuleSet{
		"nullField": {"required", "nullable", "numeric"},
	}, "en-US")
	_, exists = data["nullField"]
	suite.True(exists)
	suite.Equal(0, len(errors))

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	data = map[string]interface{}{
		"nullField": "test",
	}
	errors = Validate(rawRequest, data, RuleSet{
		"nullField": {"required", "nullable", "numeric"},
	}, "en-US")
	_, exists = data["nullField"]
	suite.True(exists)
	suite.Equal(1, len(errors))
}

func (suite *ValidatorTestSuite) TestValidateWithArray() {
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello"))
	rawRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	data := map[string]interface{}{
		"string": "hello",
	}
	errors := Validate(rawRequest, data, RuleSet{
		"string": {"required", "array"},
	}, "en-US")
	suite.Equal("array", GetFieldType(data["string"]))
	suite.Equal("hello", data["string"].([]string)[0])
	suite.Equal(0, len(errors))
}

func (suite *ValidatorTestSuite) TestValidateArrayValues() {
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello&string=world"))
	rawRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	data := map[string]interface{}{
		"string": []string{"hello", "world"},
	}
	errors := Validate(rawRequest, data, RuleSet{
		"string": {"required", "array", ">min:3"},
	}, "en-US")
	suite.Equal(0, len(errors))

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hi&string=world"))
	rawRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	data = map[string]interface{}{
		"string": []string{"hi", ",", "there"},
	}
	errors = Validate(rawRequest, data, RuleSet{
		"string": {"required", "array", ">min:3"},
	}, "en-US")
	suite.Equal(1, len(errors))

	suite.Panics(func() {
		validateRuleInArray("required", "string", map[string]interface{}{"string": "hi"}, []string{})
	})
}

func TestValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}

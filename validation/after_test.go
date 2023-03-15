package validation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4/lang"
	"goyave.dev/goyave/v4/util/typeutil"
)

func TestAfterValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		now := time.Now()
		v := After(now)
		assert.NotNil(t, v)
		assert.Equal(t, "after", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":date", now.Format(time.RFC3339)}, v.MessagePlaceholders(&ContextV5{}))
	})

	ref := typeutil.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z"))
	cases := []struct {
		ref   time.Time
		value any
		want  bool
	}{
		{ref: ref, value: typeutil.Must(time.Parse(time.RFC3339, "2023-03-15T11:07:42Z")), want: true},
		{ref: ref, value: typeutil.Must(time.Parse(time.RFC3339, "2023-03-15T10:06:42Z")), want: false},
		{ref: ref, value: ref, want: false},
		{ref: ref, value: "string", want: false},
		{ref: ref, value: 'a', want: false},
		{ref: ref, value: 2, want: false},
		{ref: ref, value: 2.5, want: false},
		{ref: ref, value: []string{"string"}, want: false},
		{ref: ref, value: true, want: false},
		{ref: ref, value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := After(c.ref)
			assert.Equal(t, c.want, v.Validate(&ContextV5{
				Value: c.value,
			}))
		})
	}
}

func TestAfterEqualValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		now := time.Now()
		v := AfterEqual(now)
		assert.NotNil(t, v)
		assert.Equal(t, "after_equal", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":date", now.Format(time.RFC3339)}, v.MessagePlaceholders(&ContextV5{}))
	})

	ref := typeutil.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z"))
	cases := []struct {
		ref   time.Time
		value any
		want  bool
	}{
		{ref: ref, value: typeutil.Must(time.Parse(time.RFC3339, "2023-03-15T11:07:42Z")), want: true},
		{ref: ref, value: typeutil.Must(time.Parse(time.RFC3339, "2023-03-15T10:06:42Z")), want: false},
		{ref: ref, value: ref, want: true},
		{ref: ref, value: "string", want: false},
		{ref: ref, value: 'a', want: false},
		{ref: ref, value: 2, want: false},
		{ref: ref, value: 2.5, want: false},
		{ref: ref, value: []string{"string"}, want: false},
		{ref: ref, value: true, want: false},
		{ref: ref, value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := AfterEqual(c.ref)
			assert.Equal(t, c.want, v.Validate(&ContextV5{
				Value: c.value,
			}))
		})
	}
}

func TestAfterFieldValidator(t *testing.T) {
	path := "object.field[]"
	t.Run("Constructor", func(t *testing.T) {
		v := AfterField(path)
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "after", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":date", "field"}, v.MessagePlaceholders(&ContextV5{}))

		assert.Panics(t, func() {
			AfterField("invalid[path.")
		})
	})

	ref1 := typeutil.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z"))
	ref2 := typeutil.Must(time.Parse(time.RFC3339, "2023-03-16T10:07:42Z"))

	dataSingle := makeAfterFieldData(ref1)
	dataTwo := makeAfterFieldData(ref1, ref2)

	cases := []struct {
		data  map[string]any
		value any
		want  bool
	}{
		{data: dataSingle, value: typeutil.Must(time.Parse(time.RFC3339, "2023-03-15T11:07:42Z")), want: true},
		{data: dataSingle, value: typeutil.Must(time.Parse(time.RFC3339, "2023-03-15T10:06:42Z")), want: false},
		{data: dataSingle, value: ref1, want: false},
		{data: dataTwo, value: typeutil.Must(time.Parse(time.RFC3339, "2023-03-16T10:06:42Z")), want: false},
		{data: dataTwo, value: ref2, want: false},
		{data: dataTwo, value: typeutil.Must(time.Parse(time.RFC3339, "2023-03-16T11:06:42Z")), want: true},
		{data: dataSingle, value: "string", want: false},
		{data: dataSingle, value: 'a', want: false},
		{data: dataSingle, value: 2, want: false},
		{data: dataSingle, value: 2.5, want: false},
		{data: dataSingle, value: []string{"string"}, want: false},
		{data: dataSingle, value: true, want: false},
		{data: dataSingle, value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := AfterField(path)
			assert.Equal(t, c.want, v.Validate(&ContextV5{
				Value: c.value,
				Data:  c.data,
			}))
		})
	}
}

func TestAfterEqualFieldValidator(t *testing.T) {
	path := "object.field[]"
	t.Run("Constructor", func(t *testing.T) {
		v := AfterEqualField(path)
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "after_equal", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":date", "field"}, v.MessagePlaceholders(&ContextV5{}))

		assert.Panics(t, func() {
			AfterEqualField("invalid[path.")
		})
	})

	ref1 := typeutil.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z"))
	ref2 := typeutil.Must(time.Parse(time.RFC3339, "2023-03-16T10:07:42Z"))

	dataSingle := makeAfterFieldData(ref1)
	dataTwo := makeAfterFieldData(ref1, ref2)

	cases := []struct {
		data  map[string]any
		value any
		want  bool
	}{
		{data: dataSingle, value: typeutil.Must(time.Parse(time.RFC3339, "2023-03-15T11:07:42Z")), want: true},
		{data: dataSingle, value: typeutil.Must(time.Parse(time.RFC3339, "2023-03-15T10:06:42Z")), want: false},
		{data: dataTwo, value: typeutil.Must(time.Parse(time.RFC3339, "2023-03-16T10:06:42Z")), want: false},
		{data: dataTwo, value: typeutil.Must(time.Parse(time.RFC3339, "2023-03-16T11:06:42Z")), want: true},

		{data: dataSingle, value: ref1, want: true},
		{data: dataTwo, value: ref1, want: false},
		{data: dataTwo, value: ref2, want: true},

		{data: dataSingle, value: "string", want: false},
		{data: dataSingle, value: 'a', want: false},
		{data: dataSingle, value: 2, want: false},
		{data: dataSingle, value: 2.5, want: false},
		{data: dataSingle, value: []string{"string"}, want: false},
		{data: dataSingle, value: true, want: false},
		{data: dataSingle, value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := AfterEqualField(path)
			assert.Equal(t, c.want, v.Validate(&ContextV5{
				Value: c.value,
				Data:  c.data,
			}))
		})
	}
}

func makeAfterFieldData(ref ...time.Time) map[string]any {
	return map[string]any{
		"object": map[string]any{
			"field": ref,
		},
	}
}

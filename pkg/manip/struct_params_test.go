package manip_test

import (
	"testing"

	"github.com/pantheon-systems/search-secrets/pkg/app/vars"
	. "github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// One level

func TestParam_Nested(t *testing.T) {
	type (
		ChildStruct struct {
			ChildField string
		}
		ParentStruct struct {
			Child ChildStruct
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewParam(parent, &parent.Child.ChildField, "", nil)

	require.NotNil(t, subject)
	assert.Equal(t, "Child.ChildField", subject.PathName())
}

func TestParam_NestedPtr(t *testing.T) {
	type (
		ChildStruct struct {
			ChildField string
		}
		ParentStruct struct {
			Child *ChildStruct
		}
	)
	parent := &ParentStruct{
		Child: &ChildStruct{},
	}

	// Fire
	subject := NewParam(parent, &parent.Child.ChildField, "", nil)

	require.NotNil(t, subject)
	assert.Equal(t, "Child.ChildField", subject.PathName())
}

func TestParam_Embedded(t *testing.T) {
	type (
		ChildStruct struct {
			ChildField string
		}
		ParentStruct struct {
			ChildStruct
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewParam(parent, &parent.ChildField, "", nil)

	require.NotNil(t, subject)
	assert.Equal(t, "ChildStruct.ChildField", subject.PathName())
}

func TestParam_EmbeddedPtr(t *testing.T) {
	type (
		ChildStruct struct {
			ChildField string
		}
		ParentStruct struct {
			*ChildStruct
		}
	)
	parent := &ParentStruct{
		ChildStruct: &ChildStruct{},
	}

	// Fire
	subject := NewParam(parent, &parent.ChildField, "", nil)

	require.NotNil(t, subject)
	assert.Equal(t, "ChildStruct.ChildField", subject.PathName())
}

func TestParam_NestedToNested(t *testing.T) {
	type (
		GrandChildStruct struct {
			GrandChildField string
		}
		ChildStruct struct {
			GrandChild GrandChildStruct
		}
		ParentStruct struct {
			Child ChildStruct
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewParam(parent, &parent.Child.GrandChild.GrandChildField, "", nil)

	require.NotNil(t, subject)
	assert.Equal(t, "Child.GrandChild.GrandChildField", subject.PathName())
}

func TestParam_NestedPtrToNestedPtr(t *testing.T) {
	type (
		GrandChildStruct struct {
			GrandChildField string
		}
		ChildStruct struct {
			GrandChild *GrandChildStruct
		}
		ParentStruct struct {
			Child *ChildStruct
		}
	)
	parent := &ParentStruct{
		Child: &ChildStruct{
			GrandChild: &GrandChildStruct{},
		},
	}

	// Fire
	subject := NewParam(parent, &parent.Child.GrandChild.GrandChildField, "", nil)

	require.NotNil(t, subject)
	assert.Equal(t, "Child.GrandChild.GrandChildField", subject.PathName())
}

func TestParam_NestedToEmbedded(t *testing.T) {
	type (
		GrandChildStruct struct {
			GrandChildField string
		}
		ChildStruct struct {
			GrandChildStruct
		}
		ParentStruct struct {
			Child ChildStruct
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewParam(parent, &parent.Child.GrandChildField, "", nil)

	require.NotNil(t, subject)
	assert.Equal(t, "Child.GrandChildStruct.GrandChildField", subject.PathName())
}

func TestParam_NestedToNestedToNested(t *testing.T) {
	type (
		GreatGrandChildStruct struct {
			GreatGrandChildField string
		}
		GrandChildStruct struct {
			GreatGrandChild GreatGrandChildStruct
		}
		ChildStruct struct {
			GrandChild GrandChildStruct
		}
		ParentStruct struct {
			Child ChildStruct
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewParam(parent, &parent.Child.GrandChild.GreatGrandChild.GreatGrandChildField, "", nil)

	require.NotNil(t, subject)
	assert.Equal(t, "Child.GrandChild.GreatGrandChild.GreatGrandChildField", subject.PathName())
}

func TestParam_Squashed(t *testing.T) {
	type (
		ChildStruct struct {
			ChildField string
		}
		ParentStruct struct {
			ChildStruct `param:",squash"`
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewParam(parent, &parent.ChildField, vars.ConfigParamTag, nil)

	require.NotNil(t, subject)
	assert.Equal(t, "ChildField", subject.PathName())
}

func TestParam_SquashedToSquashed(t *testing.T) {
	type (
		GrandChildStruct struct {
			GrandChildField string
		}
		ChildStruct struct {
			GrandChildStruct `param:",squash"`
		}
		ParentStruct struct {
			ChildStruct `param:",squash"`
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewParam(parent, &parent.GrandChildField, vars.ConfigParamTag, nil)

	require.NotNil(t, subject)
	assert.Equal(t, "GrandChildField", subject.PathName())
}

func TestParam_SquashedToSquashedToSquashed(t *testing.T) {
	type (
		GreatGrandChildStruct struct {
			GreatGrandChildField string
		}
		GrandChildStruct struct {
			GreatGrandChildStruct `param:",squash"`
		}
		ChildStruct struct {
			GrandChildStruct `param:",squash"`
		}
		ParentStruct struct {
			ChildStruct `param:",squash"`
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewParam(parent, &parent.GreatGrandChildField, vars.ConfigParamTag, nil)

	require.NotNil(t, subject)
	assert.Equal(t, "GreatGrandChildField", subject.PathName())
}

func TestParam_Named_Squashed(t *testing.T) {
	type (
		ChildStruct struct {
			ChildField string `param:"child-field"`
		}
		ParentStruct struct {
			ChildStruct `param:",squash"`
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewParam(parent, &parent.ChildField, vars.ConfigParamTag, nil)

	require.NotNil(t, subject)
	assert.Equal(t, "child-field", subject.PathName())
}

func TestParam_Named_SquashedToSquashed(t *testing.T) {
	type (
		GrandChildStruct struct {
			GrandChildField string `param:"grand-child-field"`
		}
		ChildStruct struct {
			GrandChildStruct `param:"grand-child"`
		}
		ParentStruct struct {
			ChildStruct `param:",squash"`
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewParam(parent, &parent.GrandChildField, vars.ConfigParamTag, nil)

	require.NotNil(t, subject)
	assert.Equal(t, "grand-child.grand-child-field", subject.PathName())
}

func TestParam_Named_SquashedToSquashedToSquashed(t *testing.T) {
	type (
		GreatGrandChildStruct struct {
			GreatGrandChildField string
		}
		GrandChildStruct struct {
			GreatGrandChildStruct `param:",squash"`
		}
		ChildStruct struct {
			GrandChildStruct `param:"grand-child"`
		}
		ParentStruct struct {
			ChildStruct `param:"child"`
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewParam(parent, &parent.GreatGrandChildField, vars.ConfigParamTag, nil)

	require.NotNil(t, subject)
	assert.Equal(t, "child.grand-child.GreatGrandChildField", subject.PathName())
}

//
// StructParams

func TestStructParams_Nested(t *testing.T) {
	type (
		ChildStruct struct {
			ChildField string
		}
		ParentStruct struct {
			Child ChildStruct
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewStructParams(parent, "", nil)

	require.NotNil(t, subject)
	assert.Len(t, subject.Params, 2)
	assert.Equal(t, "Child", subject.Params[0].PathName())
	assert.Equal(t, "Child.ChildField", subject.Params[1].PathName())
}

func TestStructParams_NestedPtr(t *testing.T) {
	type (
		ChildStruct struct {
			ChildField string
		}
		ParentStruct struct {
			Child *ChildStruct
		}
	)
	childStruct := &ChildStruct{}
	parent := &ParentStruct{Child: childStruct}

	// Fire
	subject := NewStructParams(parent, "", nil)

	require.NotNil(t, subject)
	assert.Len(t, subject.Params, 2)
	assert.Equal(t, "Child", subject.Params[0].PathName())
	assert.Equal(t, "Child.ChildField", subject.Params[1].PathName())
}

func TestStructParams_Embedded(t *testing.T) {
	type (
		ChildStruct struct {
			ChildField string
		}
		ParentStruct struct {
			ChildStruct
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewStructParams(parent, "", nil)

	require.NotNil(t, subject)
	assert.Len(t, subject.Params, 2)
	assert.Equal(t, "ChildStruct", subject.Params[0].PathName())
	assert.Equal(t, "ChildStruct.ChildField", subject.Params[1].PathName())
}

func TestStructParams_EmbeddedPtr(t *testing.T) {
	type (
		ChildStruct struct {
			ChildField string
		}
		ParentStruct struct {
			*ChildStruct
		}
	)
	parent := &ParentStruct{
		ChildStruct: &ChildStruct{},
	}

	// Fire
	subject := NewStructParams(parent, "", nil)

	require.NotNil(t, subject)
	assert.Len(t, subject.Params, 2)
	assert.Equal(t, "ChildStruct", subject.Params[0].PathName())
	assert.Equal(t, "ChildStruct.ChildField", subject.Params[1].PathName())
}

func TestStructParams_NestedToNested(t *testing.T) {
	type (
		GrandChildStruct struct {
			GrandChildField string
		}
		ChildStruct struct {
			GrandChild GrandChildStruct
		}
		ParentStruct struct {
			Child ChildStruct
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewStructParams(parent, "", nil)

	require.NotNil(t, subject)
	assert.Len(t, subject.Params, 3)
	assert.Equal(t, "Child", subject.Params[0].PathName())
	assert.Equal(t, "Child.GrandChild", subject.Params[1].PathName())
	assert.Equal(t, "Child.GrandChild.GrandChildField", subject.Params[2].PathName())
}

func TestStructParams_NestedPtrToNestedPtr(t *testing.T) {
	type (
		GrandChildStruct struct {
			GrandChildField string
		}
		ChildStruct struct {
			GrandChild *GrandChildStruct
		}
		ParentStruct struct {
			Child *ChildStruct
		}
	)
	parent := &ParentStruct{
		Child: &ChildStruct{
			GrandChild: &GrandChildStruct{},
		},
	}

	// Fire
	subject := NewStructParams(parent, "", nil)

	require.NotNil(t, subject)
	assert.Len(t, subject.Params, 3)
	assert.Equal(t, "Child", subject.Params[0].PathName())
	assert.Equal(t, "Child.GrandChild", subject.Params[1].PathName())
	assert.Equal(t, "Child.GrandChild.GrandChildField", subject.Params[2].PathName())
}

func TestStructParams_NestedToEmbedded(t *testing.T) {
	type (
		GrandChildStruct struct {
			GrandChildField string
		}
		ChildStruct struct {
			GrandChildStruct
		}
		ParentStruct struct {
			Child ChildStruct
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewStructParams(parent, "", nil)

	require.NotNil(t, subject)
	assert.Len(t, subject.Params, 3)
	assert.Equal(t, "Child", subject.Params[0].PathName())
	assert.Equal(t, "Child.GrandChildStruct", subject.Params[1].PathName())
	assert.Equal(t, "Child.GrandChildStruct.GrandChildField", subject.Params[2].PathName())
}

func TestStructParams_NestedToNestedToNested(t *testing.T) {
	type (
		GreatGrandChildStruct struct {
			GreatGrandChildField string
		}
		GrandChildStruct struct {
			GreatGrandChild GreatGrandChildStruct
		}
		ChildStruct struct {
			GrandChild GrandChildStruct
		}
		ParentStruct struct {
			Child ChildStruct
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewStructParams(parent, "", nil)

	require.NotNil(t, subject)
	assert.Len(t, subject.Params, 4)
	assert.Equal(t, "Child", subject.Params[0].PathName())
	assert.Equal(t, "Child.GrandChild", subject.Params[1].PathName())
	assert.Equal(t, "Child.GrandChild.GreatGrandChild", subject.Params[2].PathName())
	assert.Equal(t, "Child.GrandChild.GreatGrandChild.GreatGrandChildField", subject.Params[3].PathName())
}

func TestStructParams_Squashed(t *testing.T) {
	type (
		ChildStruct struct {
			ChildField string
		}
		ParentStruct struct {
			ChildStruct `param:",squash"`
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewStructParams(parent, vars.ConfigParamTag, nil)

	require.NotNil(t, subject)
	assert.Len(t, subject.Params, 2)
	assert.Equal(t, "ChildStruct", subject.Params[0].PathName())
	assert.Equal(t, "ChildField", subject.Params[1].PathName())
}

func TestStructParams_SquashedToSquashed(t *testing.T) {
	type (
		GrandChildStruct struct {
			GrandChildField string
		}
		ChildStruct struct {
			GrandChildStruct `param:",squash"`
		}
		ParentStruct struct {
			ChildStruct `param:",squash"`
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewStructParams(parent, vars.ConfigParamTag, nil)

	require.NotNil(t, subject)
	assert.Len(t, subject.Params, 3)
	assert.Equal(t, "ChildStruct", subject.Params[0].PathName())
	assert.Equal(t, "GrandChildStruct", subject.Params[1].PathName())
	assert.Equal(t, "GrandChildField", subject.Params[2].PathName())
}

func TestStructParams_SquashedToSquashedToSquashed(t *testing.T) {
	type (
		GreatGrandChildStruct struct {
			GreatGrandChildField string
		}
		GrandChildStruct struct {
			GreatGrandChildStruct `param:",squash"`
		}
		ChildStruct struct {
			GrandChildStruct `param:",squash"`
		}
		ParentStruct struct {
			ChildStruct `param:",squash"`
		}
	)
	parent := &ParentStruct{}

	// Fire
	subject := NewStructParams(parent, vars.ConfigParamTag, nil)

	require.NotNil(t, subject)
	assert.Len(t, subject.Params, 4)
	assert.Equal(t, "ChildStruct", subject.Params[0].PathName())
	assert.Equal(t, "GrandChildStruct", subject.Params[1].PathName())
	assert.Equal(t, "GreatGrandChildStruct", subject.Params[2].PathName())
	assert.Equal(t, "GreatGrandChildField", subject.Params[3].PathName())
}

package logic_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

type validateStructParams struct {
	Name string `validate:"required"`
}

func TestValidateStruct(t *testing.T) {
	s := spec.New(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_return_nil_for_valid_struct",
			fn: func(t *testing.T) {
				err := s.Store.ValidateStruct(validateStructParams{
					Name: "valid_name_1",
				})
				require.NoError(t, err)
			},
		},
		{
			name: "should_return_validation_failed_for_invalid_struct",
			fn: func(t *testing.T) {
				err := s.Store.ValidateStruct(validateStructParams{
					Name: "",
				})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.ErrorContains(t, err, "[Name:required]")
			},
		},
		{
			name: "should_return_validation_assertion_for_non_struct_input",
			fn: func(t *testing.T) {
				err := s.Store.ValidateStruct(1)
				require.ErrorIs(t, err, logic.ErrValidationAssertion)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type UtilsTestSuite struct {
	suite.Suite
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *UtilsTestSuite) TestContains() {
	type args struct {
		list []string
		item string
	}
	for _, t := range []struct {
		name string
		args args
		want bool
	}{
		{
			name: "item is not found when the list is empty.",
			args: args{
				list: []string{},
				item: "foo",
			},
			want: false,
		},
		{
			name: "item is found when the list contains it.",
			args: args{
				list: []string{"foo", "bar"},
				item: "foo",
			},
			want: true,
		},
		{
			name: "item is not found when the list does not contain it.",
			args: args{
				list: []string{"foo", "bar"},
				item: "example",
			},
			want: false,
		},
	} {
		suite.Run(t.name, func() {
			actual := Contains(t.args.list, t.args.item)
			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestContains: %s , got: %v -- want: %v", t.name, actual, t.want))
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

package nervo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseAnnounceMessage(t *testing.T) {
	tests := []struct {
		testMessage  string
		line         string
		expectedName string
		expectedok   bool
	}{
		{
			testMessage:  "given a line without spaces",
			line:         "ANNOUNCEsome_name",
			expectedName: "",
			expectedok:   false,
		},
		{
			testMessage:  "given a line with 1 space",
			line:         "ANNOUNCE some_name",
			expectedName: "some_name",
			expectedok:   true,
		},
		{
			testMessage:  "given a line with multiple spaces",
			line:         "ANNOUNCE some name",
			expectedName: "some name",
			expectedok:   true,
		},
		{
			testMessage:  "given a line with a lowercase verb",
			line:         "announce some_name",
			expectedName: "some_name",
			expectedok:   true,
		},
		{
			testMessage:  "given a line with another verb",
			line:         "TEST some_name",
			expectedName: "",
			expectedok:   false,
		},
		{
			testMessage:  "given a line with a newline at the end",
			line:         "announce some_name\n",
			expectedName: "some_name",
			expectedok:   true,
		},
		{
			testMessage:  "given a line with a carriage return and newline at the end",
			line:         "announce some_name\r\n",
			expectedName: "some_name",
			expectedok:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.testMessage, func(t *testing.T) {
			parsedName, parsingOk := ParseAnnounceMessage(test.line)
			assert.Equal(t, test.expectedName, parsedName)
			assert.Equal(t, test.expectedok, parsingOk)
		})
	}
}

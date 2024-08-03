package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestJSONOperations(t *testing.T) {
	conn := getLocalConnection()

	simpleJSON := `{"name":"John","age":30}`
	nestedJSON := `{"name":"Alice","address":{"city":"New York","zip":"10001"},"array":[1,2,3,4,5]}`
	specialCharsJSON := `{"key":"value with spaces","emoji":"😀"}`
	unicodeJSON := `{"unicode":"こんにちは世界"}`
	escapedCharsJSON := `{"escaped":"\"quoted\", \\backslash\\ and /forward/slash"}`

	testCases := []struct {
		name     string
		setCmd   string
		getCmd   string
		expected string
	}{
		{
			name:     "Set and Get Integer",
			setCmd:   `JSON.SET tools $ 2`,
			getCmd:   `JSON.GET tools`,
			expected: "2",
		},
		{
			name:     "Set and Get Boolean True",
			setCmd:   `JSON.SET booleanTrue $ true`,
			getCmd:   `JSON.GET booleanTrue`,
			expected: "true",
		},
		{
			name:     "Set and Get Boolean False",
			setCmd:   `JSON.SET booleanFalse $ false`,
			getCmd:   `JSON.GET booleanFalse`,
			expected: "false",
		},
		{
			name:     "Set and Get Simple JSON",
			setCmd:   `JSON.SET user $ ` + simpleJSON,
			getCmd:   `JSON.GET user`,
			expected: simpleJSON,
		},
		{
			name:     "Set and Get Nested JSON",
			setCmd:   `JSON.SET user:2 $ ` + nestedJSON,
			getCmd:   `JSON.GET user:2`,
			expected: nestedJSON,
		},
		{
			name:     "Set and Get JSON Array",
			setCmd:   `JSON.SET numbers $ [1,2,3,4,5]`,
			getCmd:   `JSON.GET numbers`,
			expected: `[1,2,3,4,5]`,
		},
		{
			name:     "Set and Get JSON with Special Characters",
			setCmd:   `JSON.SET special $ ` + specialCharsJSON,
			getCmd:   `JSON.GET special`,
			expected: specialCharsJSON,
		},
		{
			name:     "Set Invalid JSON",
			setCmd:   `JSON.SET invalid $ {invalid:json}`,
			getCmd:   ``,
			expected: "ERR invalid JSON",
		},
		{
			name:     "Set JSON with Wrong Number of Arguments",
			setCmd:   `JSON.SET`,
			getCmd:   ``,
			expected: "ERR wrong number of arguments for 'JSON.SET' command",
		},
		{
			name:     "Get JSON with Wrong Number of Arguments",
			setCmd:   ``,
			getCmd:   `JSON.GET`,
			expected: "ERR wrong number of arguments for 'JSON.GET' command",
		},
		{
			name:     "Set Non-JSON Value",
			setCmd:   `SET nonJson "not a json"`,
			getCmd:   `JSON.GET nonJson`,
			expected: "WRONGTYPE Operation against a key holding the wrong kind of value",
		},
		{
			name:     "Set Empty JSON Object",
			setCmd:   `JSON.SET empty $ {}`,
			getCmd:   `JSON.GET empty`,
			expected: `{}`,
		},
		{
			name:     "Set Empty JSON Array",
			setCmd:   `JSON.SET emptyArray $ []`,
			getCmd:   `JSON.GET emptyArray`,
			expected: `[]`,
		},
		{
			name:     "Set JSON with Unicode",
			setCmd:   `JSON.SET unicode $ ` + unicodeJSON,
			getCmd:   `JSON.GET unicode`,
			expected: unicodeJSON,
		},
		{
			name:     "Set JSON with Escaped Characters",
			setCmd:   `JSON.SET escaped $ ` + escapedCharsJSON,
			getCmd:   `JSON.GET escaped`,
			expected: escapedCharsJSON,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setCmd != "" {
				result := fireCommand(conn, tc.setCmd)
				if tc.name != "Set Invalid JSON" && tc.name != "Set JSON with Wrong Number of Arguments" {
					assert.Equal(t, "OK", result)
				} else {
					assert.Equal(t, tc.expected, result)
				}
			}

			if tc.getCmd != "" {
				result := fireCommand(conn, tc.getCmd)
				assert.DeepEqual(t, tc.expected, result)
			}
		})
	}
}

func TestJSONSetWithNXAndXX(t *testing.T) {

	conn := getLocalConnection()
	deleteTestKeys(conn, []string{"user"})

	user1 := `{"name":"John","age":30}`
	user2 := `{"name":"Rahul","age":28}`

	for _, tcase := range []DTestCase{
		{
			InCmds: []string{"JSON.SET user $ " + user1 + " XX", "JSON.SET user $ " + user1 + " NX", "JSON.GET user"},
			Out:    []interface{}{"ERR key doesn't exists", "OK", user1},
		},
		{
			InCmds: []string{"DEL user", "JSON.SET user $ " + user1, "JSON.SET user $ " + user1 + " NX"},
			Out:    []interface{}{int64(1), "OK", "ERR key already exists"},
		},
		{
			InCmds: []string{"JSON.SET user $ " + user2 + " XX", "JSON.GET user", "DEL user"},
			Out:    []interface{}{"OK", user2, int64(1)},
		},
		{
			InCmds: []string{"JSON.SET user $ " + user2 + " NX", "JSON.SET user $ " + user1 + " NX", "JSON.GET user"},
			Out:    []interface{}{"OK", "ERR key already exists", user2},
		},
	} {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}

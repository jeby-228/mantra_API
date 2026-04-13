package routes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsIntrospectionQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "standard IntrospectionQuery",
			query:    `query IntrospectionQuery { __schema { queryType { name } } }`,
			expected: true,
		},
		{
			name:     "short __schema query",
			query:    `{ __schema { types { name } } }`,
			expected: true,
		},
		{
			name:     "__type query",
			query:    `{ __type(name: "Member") { name fields { name } } }`,
			expected: true,
		},
		{
			name:     "query keyword with __schema",
			query:    `query { __schema { queryType { name } mutationType { name } } }`,
			expected: true,
		},
		{
			name:     "normal query without introspection",
			query:    `query { members(limit: 10) { id name } }`,
			expected: false,
		},
		{
			name:     "mutation query",
			query:    `mutation { createMantra(input: { content: "test" }) { id } }`,
			expected: false,
		},
		{
			name:     "mutation with __schema in string (edge case)",
			query:    `mutation { createMantra(input: { content: "__schema" }) { id } }`,
			expected: false,
		},
		{
			name:     "empty query string",
			query:    ``,
			expected: false,
		},
		{
			name:     "whitespace only query",
			query:    `   `,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{"query": tt.query})
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, isIntrospectionQuery(body))
		})
	}

	t.Run("invalid JSON", func(t *testing.T) {
		assert.False(t, isIntrospectionQuery([]byte(`not json`)))
	})

	t.Run("empty body", func(t *testing.T) {
		assert.False(t, isIntrospectionQuery([]byte{}))
	})

	t.Run("JSON without query field", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"operationName": "Foo"})
		assert.False(t, isIntrospectionQuery(body))
	})
}

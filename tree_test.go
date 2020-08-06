package mmdbwriter

import (
	"bytes"
	"fmt"
	"net"
	"testing"

	"github.com/oschwald/maxminddb-golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testInsert struct {
	network string
	value   DataType
}

type testGet struct {
	ip                  string
	expectedNetwork     string
	expectedGetValue    *DataType
	expectedLookupValue *interface{}
}

func TestTreeInsertAndGet(t *testing.T) {
	tests := []struct {
		name              string
		inserts           []testInsert
		gets              []testGet
		expectedNodeCount int
	}{
		{
			name: "::/1 insert, IPv4 lookup",
			inserts: []testInsert{
				{
					network: "::/1",
					value:   String("string"),
				},
			},
			gets: []testGet{
				{
					ip:                  "1.1.1.1",
					expectedNetwork:     "::/1",
					expectedGetValue:    s2dtp("string"),
					expectedLookupValue: s2ip("string"),
				},
			},
			expectedNodeCount: 1,
		},
		{
			name: "8000::/1 insert",
			inserts: []testInsert{
				{
					network: "8000::/1",
					value:   String("string"),
				},
			},
			gets: []testGet{
				{
					ip:                  "8000::",
					expectedNetwork:     "8000::/1",
					expectedGetValue:    s2dtp("string"),
					expectedLookupValue: s2ip("string"),
				},
			},
			expectedNodeCount: 1,
		},
		{
			name: "overwriting smaller network with bigger network",
			inserts: []testInsert{
				{
					network: "2002:1000::/32",
					value:   String("string"),
				},
				{
					network: "2002::/16",
					value:   String("new string"),
				},
			},
			gets: []testGet{
				{
					ip:                  "2002::",
					expectedNetwork:     "2002::/16",
					expectedGetValue:    s2dtp("new string"),
					expectedLookupValue: s2ip("new string"),
				},
				{
					ip:                  "2002:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					expectedNetwork:     "2002::/16",
					expectedGetValue:    s2dtp("new string"),
					expectedLookupValue: s2ip("new string"),
				},
			},
			expectedNodeCount: 16,
		},
		{
			name: "insert smaller network into bigger network",
			inserts: []testInsert{
				{
					network: "2002::/16",
					value:   String("string"),
				},
				{
					network: "2002:1000::/32",
					value:   String("new string"),
				},
			},
			gets: []testGet{
				{
					ip:                  "2002::",
					expectedNetwork:     "2002::/20",
					expectedGetValue:    s2dtp("string"),
					expectedLookupValue: s2ip("string"),
				},
				{
					ip:                  "2002:1000::",
					expectedNetwork:     "2002:1000::/32",
					expectedGetValue:    s2dtp("new string"),
					expectedLookupValue: s2ip("new string"),
				},
				{
					ip:                  "2002:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					expectedNetwork:     "2002:8000::/17",
					expectedGetValue:    s2dtp("string"),
					expectedLookupValue: s2ip("string"),
				},
			},
			expectedNodeCount: 32,
		},
		{
			name: "inserting IPv4 address in IPv6 tree",
			inserts: []testInsert{
				{
					network: "1.1.1.1/32",
					value:   String("string"),
				},
			},
			gets: []testGet{
				{
					ip:                  "1.1.1.1",
					expectedNetwork:     "1.1.1.1/32",
					expectedGetValue:    s2dtp("string"),
					expectedLookupValue: s2ip("string"),
				},
				{
					ip:                  "::1.1.1.1",
					expectedNetwork:     "::101:101/128",
					expectedGetValue:    s2dtp("string"),
					expectedLookupValue: s2ip("string"),
				},
			},
			expectedNodeCount: 128,
		},
	}

	for _, recordSize := range []int{24, 28, 32} {
		t.Run(fmt.Sprintf("Record Size: %d", recordSize), func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					tree, err := New(Options{RecordSize: recordSize})
					require.NoError(t, err)
					for _, insert := range test.inserts {
						_, network, err := net.ParseCIDR(insert.network)
						require.NoError(t, err)

						require.NoError(t, tree.Insert(network, insert.value))
					}

					for _, get := range test.gets {
						network, value := tree.Get(net.ParseIP(get.ip))

						assert.Equal(t, get.expectedNetwork, network.String(), "network for %s", get.ip)
						assert.Equal(t, get.expectedGetValue, value, "value for %s", get.ip)
					}

					tree.Finalize()

					assert.Equal(t, test.expectedNodeCount, tree.nodeCount)

					buf := &bytes.Buffer{}
					numBytes, err := tree.WriteTo(buf)
					require.NoError(t, err)

					reader, err := maxminddb.FromBytes(buf.Bytes())
					require.NoError(t, err)

					for _, get := range test.gets {
						var v interface{}
						network, ok, err := reader.LookupNetwork(net.ParseIP(get.ip), &v)
						require.NoError(t, err)

						assert.Equal(t, get.expectedNetwork, network.String(), "network for %s in database", get.ip)

						if get.expectedLookupValue == nil {
							assert.False(t, ok, "%s is not in the database", get.ip)
						} else {
							assert.Equal(t, *get.expectedLookupValue, v, "value for %s in database", get.ip)
						}
					}
					assert.Equal(t, int64(buf.Len()), numBytes, "number of bytes")
				})
			}
		})
	}
}

func s2ip(v string) *interface{} {
	i := interface{}(v)
	return &i
}

func s2dtp(v string) *DataType {
	ts := DataType(String(v))
	return &ts
}
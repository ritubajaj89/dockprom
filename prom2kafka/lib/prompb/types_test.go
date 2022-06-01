// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prompb

import (
	"github.com/gogo/protobuf/proto"
	"testing"
)

func TestRelabel(t *testing.T) {
	tests := []struct {
		input Labels
	}{
		{
			input: FromMap(map[string]string{
				"a": "foo",
				"b": "bar",
				"c": "baz",
			}),
		},
		{
			input: FromMap(map[string]string{
				"a": "foo",
				"b": "bar",
				"c": "baz",
			}),
		},
		{
			input: FromMap(map[string]string{
				"a": "foo",
			}),
		},
		{
			input: FromMap(map[string]string{
				"a": "foo",
				"b": "bar",
			}),
		},
	}

	buf := make([][]byte, 0, 30)
	lbBuf := NewBuilder(nil)
	for _, test := range tests {
		buf = buf[:0]
		lbBuf.Reset(nil)

		proto.Marshal(&test.input)
	}
}

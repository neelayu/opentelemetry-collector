// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package batchprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/internal/testdata"
)

func TestSplitTraces_noop(t *testing.T) {
	td := testdata.GenerateTraceDataManySpansSameResource(20)
	splitSize := 40
	split := splitTraces(splitSize, td)
	assert.Equal(t, td, split)

	td.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().Resize(5)
	assert.EqualValues(t, td, split)
}

func TestSplitTraces(t *testing.T) {
	td := testdata.GenerateTraceDataManySpansSameResource(20)
	spans := td.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans()
	for i := 0; i < spans.Len(); i++ {
		spans.At(i).SetName(getTestSpanName(0, i))
	}
	cp := pdata.NewTraces()
	cpSpans := cp.ResourceSpans().AppendEmpty().InstrumentationLibrarySpans().AppendEmpty().Spans()
	cpSpans.Resize(5)
	td.ResourceSpans().At(0).Resource().CopyTo(
		cp.ResourceSpans().At(0).Resource())
	td.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).InstrumentationLibrary().CopyTo(
		cp.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).InstrumentationLibrary())
	spans.At(0).CopyTo(cpSpans.At(0))
	spans.At(1).CopyTo(cpSpans.At(1))
	spans.At(2).CopyTo(cpSpans.At(2))
	spans.At(3).CopyTo(cpSpans.At(3))
	spans.At(4).CopyTo(cpSpans.At(4))

	splitSize := 5
	split := splitTraces(splitSize, td)
	assert.Equal(t, splitSize, split.SpanCount())
	assert.Equal(t, cp, split)
	assert.Equal(t, 15, td.SpanCount())
	assert.Equal(t, "test-span-0-0", split.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().At(0).Name())
	assert.Equal(t, "test-span-0-4", split.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().At(4).Name())

	split = splitTraces(splitSize, td)
	assert.Equal(t, 10, td.SpanCount())
	assert.Equal(t, "test-span-0-5", split.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().At(0).Name())
	assert.Equal(t, "test-span-0-9", split.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().At(4).Name())

	split = splitTraces(splitSize, td)
	assert.Equal(t, 5, td.SpanCount())
	assert.Equal(t, "test-span-0-10", split.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().At(0).Name())
	assert.Equal(t, "test-span-0-14", split.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().At(4).Name())

	split = splitTraces(splitSize, td)
	assert.Equal(t, 5, td.SpanCount())
	assert.Equal(t, "test-span-0-15", split.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().At(0).Name())
	assert.Equal(t, "test-span-0-19", split.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().At(4).Name())
}

func TestSplitTracesMultipleResourceSpans(t *testing.T) {
	td := testdata.GenerateTraceDataManySpansSameResource(20)
	spans := td.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans()
	for i := 0; i < spans.Len(); i++ {
		spans.At(i).SetName(getTestSpanName(0, i))
	}
	// add second index to resource spans
	testdata.GenerateTraceDataManySpansSameResource(20).
		ResourceSpans().At(0).CopyTo(td.ResourceSpans().AppendEmpty())
	spans = td.ResourceSpans().At(1).InstrumentationLibrarySpans().At(0).Spans()
	for i := 0; i < spans.Len(); i++ {
		spans.At(i).SetName(getTestSpanName(1, i))
	}

	splitSize := 5
	split := splitTraces(splitSize, td)
	assert.Equal(t, splitSize, split.SpanCount())
	assert.Equal(t, 35, td.SpanCount())
	assert.Equal(t, "test-span-0-0", split.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().At(0).Name())
	assert.Equal(t, "test-span-0-4", split.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().At(4).Name())
}

func TestSplitTracesMultipleResourceSpans_SplitSizeGreaterThanSpanSize(t *testing.T) {
	td := testdata.GenerateTraceDataManySpansSameResource(20)
	spans := td.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans()
	for i := 0; i < spans.Len(); i++ {
		spans.At(i).SetName(getTestSpanName(0, i))
	}
	td.ResourceSpans().Resize(2)
	// add second index to resource spans
	testdata.GenerateTraceDataManySpansSameResource(20).
		ResourceSpans().At(0).CopyTo(td.ResourceSpans().At(1))
	spans = td.ResourceSpans().At(1).InstrumentationLibrarySpans().At(0).Spans()
	for i := 0; i < spans.Len(); i++ {
		spans.At(i).SetName(getTestSpanName(1, i))
	}

	splitSize := 25
	split := splitTraces(splitSize, td)
	assert.Equal(t, splitSize, split.SpanCount())
	assert.Equal(t, 40-splitSize, td.SpanCount())
	assert.Equal(t, 1, td.ResourceSpans().Len())
	assert.Equal(t, "test-span-0-0", split.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().At(0).Name())
	assert.Equal(t, "test-span-0-19", split.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans().At(19).Name())
	assert.Equal(t, "test-span-1-0", split.ResourceSpans().At(1).InstrumentationLibrarySpans().At(0).Spans().At(0).Name())
	assert.Equal(t, "test-span-1-4", split.ResourceSpans().At(1).InstrumentationLibrarySpans().At(0).Spans().At(4).Name())
}

func BenchmarkSplitTraces(b *testing.B) {
	td := pdata.NewTraces()
	rms := td.ResourceSpans()
	for i := 0; i < 20; i++ {
		testdata.GenerateTracesManySpansSameResource(20).ResourceSpans().MoveAndAppendTo(td.ResourceSpans())
		ms := rms.At(rms.Len() - 1).InstrumentationLibrarySpans().At(0).Spans()
		for i := 0; i < ms.Len(); i++ {
			ms.At(i).SetName(getTestMetricName(1, i))
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cloneReq := td.Clone()
		split := splitTraces(128, cloneReq)
		if split.SpanCount() != 128 || cloneReq.SpanCount() != 400-128 {
			b.Fail()
		}
	}
}

func BenchmarkCloneSpans(b *testing.B) {
	td := pdata.NewTraces()
	rms := td.ResourceSpans()
	for i := 0; i < 20; i++ {
		testdata.GenerateTracesManySpansSameResource(20).ResourceSpans().MoveAndAppendTo(td.ResourceSpans())
		ms := rms.At(rms.Len() - 1).InstrumentationLibrarySpans().At(0).Spans()
		for i := 0; i < ms.Len(); i++ {
			ms.At(i).SetName(getTestMetricName(1, i))
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cloneReq := td.Clone()
		if cloneReq.SpanCount() != 400 {
			b.Fail()
		}
	}
}

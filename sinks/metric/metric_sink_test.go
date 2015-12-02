// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metricsink

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"k8s.io/heapster/core"
)

func TestGetMetrics(t *testing.T) {
	now := time.Now()
	key := core.PodKey("ns1", "pod1")
	otherKey := core.PodKey("ns1", "other")

	batch1 := core.DataBatch{
		Timestamp: now.Add(-180 * time.Second),
		MetricSets: map[string]*core.MetricSet{
			key: &core.MetricSet{
				Labels: map[string]string{
					core.LabelMetricSetType.Key: core.MetricSetTypePod,
					core.LabelPodNamespace.Key:  "ns1",
				},
				MetricValues: map[string]core.MetricValue{
					"m1": core.MetricValue{
						ValueType:  core.ValueInt64,
						MetricType: core.MetricGauge,
						IntValue:   60,
					},
					"m2": core.MetricValue{
						ValueType:  core.ValueInt64,
						MetricType: core.MetricGauge,
						IntValue:   666,
					},
				},
			},
		},
	}

	batch2 := core.DataBatch{
		Timestamp: now.Add(-60 * time.Second),
		MetricSets: map[string]*core.MetricSet{
			key: &core.MetricSet{
				Labels: map[string]string{
					core.LabelMetricSetType.Key: core.MetricSetTypePod,
					core.LabelPodNamespace.Key:  "ns1",
				},
				MetricValues: map[string]core.MetricValue{
					"m1": core.MetricValue{
						ValueType:  core.ValueInt64,
						MetricType: core.MetricGauge,
						IntValue:   40,
					},
					"m2": core.MetricValue{
						ValueType:  core.ValueInt64,
						MetricType: core.MetricGauge,
						IntValue:   444,
					},
				},
			},
		},
	}

	batch3 := core.DataBatch{
		Timestamp: now.Add(-20 * time.Second),
		MetricSets: map[string]*core.MetricSet{
			key: &core.MetricSet{
				Labels: map[string]string{
					core.LabelMetricSetType.Key: core.MetricSetTypePod,
					core.LabelPodNamespace.Key:  "ns1",
				},
				MetricValues: map[string]core.MetricValue{
					"m1": core.MetricValue{
						ValueType:  core.ValueInt64,
						MetricType: core.MetricGauge,
						IntValue:   20,
					},
					"m2": core.MetricValue{
						ValueType:  core.ValueInt64,
						MetricType: core.MetricGauge,
						IntValue:   222,
					},
				},
			},
			otherKey: &core.MetricSet{
				Labels: map[string]string{
					core.LabelMetricSetType.Key: core.MetricSetTypePod,
					core.LabelPodNamespace.Key:  "ns1",
				},
				MetricValues: map[string]core.MetricValue{
					"m1": core.MetricValue{
						ValueType:  core.ValueInt64,
						MetricType: core.MetricGauge,
						IntValue:   123,
					},
				},
			},
		},
	}

	metrics := NewMetricSink(45*time.Second, 120*time.Second, []string{"m1"})
	metrics.ExportData(&batch1)
	metrics.ExportData(&batch2)
	metrics.ExportData(&batch3)

	//batch1 is discarded by long store
	result1 := metrics.GetMetric("m1", []string{key}, now.Add(-120*time.Second), now)
	assert.Equal(t, 2, len(result1[key]))
	assert.Equal(t, 40, result1[key][0].MetricValue.IntValue)
	assert.Equal(t, 20, result1[key][1].MetricValue.IntValue)
	assert.Equal(t, 1, len(metrics.GetMetric("m1", []string{otherKey}, now.Add(-120*time.Second), now)[otherKey]))

	//batch1 is discarded by long store and batch2 doesn't belong to time window
	assert.Equal(t, 1, len(metrics.GetMetric("m1", []string{key}, now.Add(-30*time.Second), now)[key]))

	//batch1 and batch1 are discarded by short store
	assert.Equal(t, 1, len(metrics.GetMetric("m2", []string{key}, now.Add(-120*time.Second), now)[key]))

	//nothing is in time window
	assert.Equal(t, 0, len(metrics.GetMetric("m2", []string{key}, now.Add(-10*time.Second), now)[key]))

	metricNames := metrics.GetMetricNames(key)
	assert.Equal(t, 2, len(metricNames))
	assert.Contains(t, metricNames, "m1")
	assert.Contains(t, metricNames, "m2")
}
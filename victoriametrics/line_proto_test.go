package victoriametrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMetric(t *testing.T) {
	m := &Metric{Name: "test"}

	_, err := m.MarshalJSON()
	require.Error(t, err, ErrNoData)

	m.Data.Values = append(m.Data.Values, 1)

	_, err = m.MarshalJSON()
	require.Error(t, err, ErrDataCountMismatch)

	m.Data.Timestamps = append(m.Data.Timestamps, time.Unix(0, 0))

	data, err := m.MarshalJSON()
	require.NoError(t, err)

	require.Equal(t, `{"metric":{"__name__":"test"},"values":[1],"timestamps":[0]}`, string(data))
}

func TestMetricTags(t *testing.T) {
	m := &Metric{Name: "test"}

	m.Data.Values = append(m.Data.Values, 1)
	m.Data.Timestamps = append(m.Data.Timestamps, time.Unix(0, 0))

	m.Tags = map[string]string{
		"tag": "value",
	}

	data, err := m.MarshalJSON()
	require.NoError(t, err)

	require.Equal(t, `{"metric":{"__name__":"test","tag":"value"},"values":[1],"timestamps":[0]}`, string(data))
}

func TestMetricLegacyTags(t *testing.T) {
	m := &Metric{Name: "test"}

	m.Data.Values = append(m.Data.Values, 1)
	m.Data.Timestamps = append(m.Data.Timestamps, time.Unix(0, 0))

	m.addLegacyTags([]string{"tag:value"})

	data, err := m.MarshalJSON()
	require.NoError(t, err)

	require.Equal(t, `{"metric":{"__name__":"test","tag":"value"},"values":[1],"timestamps":[0]}`, string(data))
}

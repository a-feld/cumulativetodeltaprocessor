package cumulativetodelta

import "sync"

type MetricIdentity struct {
	name string
}

type State struct {
	RunningTotal float64
	LatestValue  float64
	Offset       float64
	LastFlushed  float64
}
type Metric struct {
	Name  string
	Value float64
}

func (m Metric) Identity() MetricIdentity {
	return MetricIdentity{name: m.Name}
}

type MetricTracker struct {
	mu     sync.Mutex
	States map[MetricIdentity]State
}

func (m *MetricTracker) Record(in Metric) {
	var total, lastFlushed, offset float64
	m.mu.Lock()
	defer m.mu.Unlock()

	// Compute updated offset if applicable
	if state, ok := m.States[in.Identity()]; ok {
		offset = state.Offset
		if in.Value < state.LatestValue {
			offset += state.LatestValue
		}

		// input = output for new struct construction -- ignore this
		lastFlushed = state.LastFlushed
	}

	// Total = Add the input metric value with the offset
	total = in.Value + offset

	// Store state
	m.States[in.Identity()] = State{
		RunningTotal: total,
		LatestValue:  in.Value,
		LastFlushed:  lastFlushed,
		Offset:       offset,
	}

	// TODO: persist to disk
}

func (m *MetricTracker) Flush() []Metric {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics := make([]Metric, len(m.States), 0)
	for identity, state := range m.States {
		metrics = append(metrics, Metric{
			Name:  identity.name,
			Value: state.RunningTotal - state.LastFlushed,
		})
		state.LastFlushed = state.RunningTotal
	}
	// TODO: flush m.States to disk via json marshal
	// Once Flush is called, any metric deltas are considered "sent"
	return metrics
}

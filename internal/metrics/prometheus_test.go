package metrics

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/prometheus/common/model"
)

type fakeCollector struct {
	results ingressResults
	err     error
}

func (f fakeCollector) getIngresses(ctx context.Context, query string) (ingresses ingressResults, err error) {
	return f.results, f.err
}

func (f fakeCollector) getMetric(query string) (bool, error) {
	return true, f.err
}

func Test_prometheusClient_ListActiveIngresses(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		ctx       context.Context
		collector collector
	}
	type args struct {
		maxIdle string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]int
		wantErr bool
	}{
		{
			name: "Test Active List",
			fields: fields{
				ctx: ctx,
				collector: fakeCollector{
					results: ingressResults(map[string]int{"foo/bar": 1, "bar/biz": 2}),
					err:     nil,
				},
			},
			args:    args{maxIdle: "ignored"},
			want:    map[string]int{"foo/bar": 1, "bar/biz": 2},
			wantErr: false,
		},
		{
			name: "Test Active List should fail",
			fields: fields{
				ctx: ctx,
				collector: fakeCollector{
					results: make(map[string]int),
					err:     errors.New("could not get metrics"),
				},
			},
			args:    args{maxIdle: "ignored"},
			want:    make(map[string]int),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := prometheusClient{
				ctx:       tt.fields.ctx,
				collector: tt.fields.collector,
			}
			got, err := c.ListActiveIngresses(tt.args.maxIdle)
			if (err != nil) != tt.wantErr {
				t.Errorf("prometheusClient.ListActiveIngresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prometheusClient.ListActiveIngresses() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fakeClient struct {
	samples []sample
	err     error
}

type sample struct {
	namespace string
	app       string
	val       float64
	time      time.Time
}

func generateSamples(samples []sample) model.Value {
	promSamples := make([]*model.Sample, len(samples))

	for i, sample := range samples {
		labelSet := model.LabelSet(make(map[model.LabelName]model.LabelValue))
		labelSet["ingress"] = model.LabelValue(sample.app)
		labelSet["exported_namespace"] = model.LabelValue(sample.namespace)

		promSamples[i] = &model.Sample{
			Metric:    model.Metric(labelSet),
			Value:     model.SampleValue(sample.val),
			Timestamp: model.TimeFromUnix(sample.time.Unix()),
		}
	}
	return model.Vector(promSamples)
}

func (f fakeClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	return generateSamples(f.samples), f.err
}

func Test_prometheusCollector_getIngresses(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		client prometheusQuery
	}
	type args struct {
		ctx   context.Context
		query string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantIngresses ingressResults
		wantErr       bool
	}{
		{
			name: "Test with fake client and some real samples",
			fields: fields{
				client: fakeClient{
					samples: []sample{
						{
							namespace: "default",
							app:       "bar",
							val:       0.0000696,
							time:      time.Now(),
						},
						{ // not bigger than 0, should be ignored
							namespace: "default",
							app:       "biz",
							val:       0,
							time:      time.Now(),
						},
					},
				},
			},
			args:          args{ctx, "ignored"},
			wantIngresses: map[string]int{"default/bar": 1},
			wantErr:       false,
		},
		{
			name: "Test with fake client and query error",
			fields: fields{
				client: fakeClient{
					samples: nil,
					err:     errors.New("some err"),
				},
			},
			args:          args{ctx, "ignored"},
			wantIngresses: make(map[string]int),
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := prometheusCollector{
				client: tt.fields.client,
			}
			gotIngresses, err := p.getIngresses(tt.args.ctx, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("prometheusCollector.getIngresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotIngresses, tt.wantIngresses) {
				t.Errorf("prometheusCollector.getIngresses() = %v, want %v", gotIngresses, tt.wantIngresses)
			}
		})
	}
}

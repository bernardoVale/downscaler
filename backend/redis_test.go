package backend

import (
	"testing"

	redis "github.com/go-redis/redis"
)

type FakeRedisClient struct{}


func TestRedisClient_Retrieve(t *testing.T) {
	type fields struct {
		baseClient redis.Client
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "foo",
			f
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := RedisClient{
				baseClient: tt.fields.baseClient,
			}
			got, err := client.Retrieve(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisClient.Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RedisClient.Retrieve() = %v, want %v", got, tt.want)
			}
		})
	}
}

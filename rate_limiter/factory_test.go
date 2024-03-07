package rate_limiter

import (
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	tests := []struct {
		name string
		typ  string
		want RateLimiter
	}{
		{
			name: "Fixed Window Counter",
			typ:  FixedWindowCounter,
			want: newFixedWindowCounter(redisClient),
		},
		{
			name: "Sliding Window Counter",
			typ:  SlidingWindowCounter,
			want: newSlidingWindowCounter(redisClient),
		},
		{
			name: "Default",
			typ:  "Default",
			want: newSlidingWindowCounter(redisClient),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Get(tt.typ, redisClient)
			assert.ObjectsAreEqual(tt.want, got)
		})
	}
}

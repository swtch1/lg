package redis

import (
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v7"
	"github.com/swtch1/lg/domain"
)

func NewFeed(rc *redis.Client) *Feed {
	return &Feed{
		rc: rc,
	}
}

// Feed knows how to extract RRPairs from redis.
type Feed struct {
	rc     *redis.Client
	offset int
	max    int
}

func (f *Feed) Next() (domain.RRPair, error) {
	max, err := f.getMax()
	if err != nil {
		return domain.RRPair{}, fmt.Errorf("could not get record data: %w", err)
	}

	defer func() {
		if f.offset == max {
			f.offset = 0
			return
		}
		f.offset++
	}()

	return fakeRedis[int(f.offset)], nil
}

func (f *Feed) getMax() (int, error) {
	if f.max != 0 {
		return f.max, nil
	}

	// need to get max from redis
	fakeMax := len(fakeRedis) - 1
	if fakeMax == 0 {
		return 0, fmt.Errorf("no data to process")
	}

	return fakeMax, nil
}

// for now we're not going to mess with putting records into redis... just fake it.
var fakeRedis = []domain.RRPair{
	{
		Req: domain.Request{
			Method: "GET",
			Headers: http.Header{
				"Authentication": []string{"Bearer foo"},
			},
			Path: "/foo",
		},
		Resp: domain.Response{
			StatusCode: 200,
			Body:       []byte(`{"foo": "bar"}`),
		},
	},
}

package mpesa

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

//TODO: rework failing tests

func TestCache_Set(t *testing.T) {
	type fields struct {
		data map[string]*AccessTokenResponse
		lock *sync.RWMutex
	}
	type args struct {
		val *AccessTokenResponse
	}
	d := NewCache()
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "store",
			fields: fields{
				data: d.data,
				lock: d.lock,
			},
			args: args{
				val: &AccessTokenResponse{AccessToken: "123", ExpiresIn: "12", ExpireTime: time.Now().Add(12 * time.Minute)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cache{
				data: tt.fields.data,
				lock: tt.fields.lock,
			}
			c.Set(tt.args.val)
		})
	}
}

func TestCache_Get(t *testing.T) {
	type fields struct {
		data map[string]*AccessTokenResponse
		lock *sync.RWMutex
	}
	d := NewCache()
	tests := []struct {
		name   string
		key    string
		fields fields
		pre    func()
		want   *AccessTokenResponse
		want1  bool
	}{
		{
			name: "fail",
			fields: fields{
				data: d.data,
				lock: d.lock,
			},
			key: "token",
			pre: func() {
				d.Set(&AccessTokenResponse{AccessToken: "AccessToken", ExpiresIn: "12", ExpireTime: time.Now().Add(0 * time.Millisecond)})
			},
			want:  &AccessTokenResponse{AccessToken: "123", ExpiresIn: "1 Sec", ExpireTime: time.Now().Add(0 * time.Second)},
			want1: false,
		}, {
			name: "no key",
			key:  "",
			fields: fields{
				data: d.data,
				lock: d.lock,
			},
			pre: func() {
				d.Set(&AccessTokenResponse{AccessToken: "AccessToken", ExpiresIn: "12", ExpireTime: time.Now().Add(0 * time.Millisecond)})
			},
			want:  &AccessTokenResponse{AccessToken: "123", ExpiresIn: "1 Sec", ExpireTime: time.Now().Add(0 * time.Second)},
			want1: false,
		}, {
			name: "pass",
			key:  "pass",
			fields: fields{
				data: d.data,
				lock: d.lock,
			},
			pre: func() {
				d.Set(&AccessTokenResponse{AccessToken: "AccessToken", ExpiresIn: "12"})
			},
			want:  &AccessTokenResponse{AccessToken: "AccessToken", ExpiresIn: "12"},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.pre()
			c := &Cache{
				data: tt.fields.data,
				lock: tt.fields.lock,
			}

			got, got1 := c.Get(tt.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cache.Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Cache.Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

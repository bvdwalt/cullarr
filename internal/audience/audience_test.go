package audience

import (
	"reflect"
	"testing"
)

var users = []string{"BJ", "Mom Nando"}

func TestNormaliseUser(t *testing.T) {
	for in, want := range map[string]string{"BJ": "bj", "Mom Nando": "mom-nando", " A_b.c ": "a-b-c"} {
		if got := NormaliseUser(in); got != want {
			t.Errorf("NormaliseUser(%q)=%q want %q", in, got, want)
		}
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name        string
		tags        []string
		wantAud     []string
		wantTagged  bool
		wantUnknown []string
	}{
		{"no tags means all users", nil, users, false, nil},
		{"unrelated tags ignored", []string{"4k", "anime"}, users, false, nil},
		{"single user tag", []string{"cullarr-bj"}, []string{"BJ"}, true, nil},
		{"user with space", []string{"cullarr-mom-nando"}, []string{"Mom Nando"}, true, nil},
		{"two tags", []string{"cullarr-bj", "cullarr-mom-nando"}, users, true, nil},
		{"case insensitive", []string{"Cullarr-BJ"}, []string{"BJ"}, true, nil},
		{"unknown user falls back to all", []string{"cullarr-bob"}, users, false, []string{"cullarr-bob"}},
		{"unknown plus known uses known", []string{"cullarr-bob", "cullarr-bj"}, []string{"BJ"}, true, []string{"cullarr-bob"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aud, tagged, unknown := Resolve(tt.tags, users, DefaultTagPrefix)
			if !reflect.DeepEqual(aud, tt.wantAud) || tagged != tt.wantTagged || !reflect.DeepEqual(unknown, tt.wantUnknown) {
				t.Errorf("got (%v,%v,%v) want (%v,%v,%v)", aud, tagged, unknown, tt.wantAud, tt.wantTagged, tt.wantUnknown)
			}
		})
	}
}

func TestEligible(t *testing.T) {
	tests := []struct {
		name      string
		aud       []string
		watchedBy []string
		tagged    bool
		min       int
		want      bool
	}{
		{"untagged, only one watched", users, []string{"BJ"}, false, 0, false},
		{"untagged, all watched", users, []string{"BJ", "Mom Nando"}, false, 0, true},
		{"tagged bj, bj watched", []string{"BJ"}, []string{"BJ"}, true, 0, true},
		{"tagged bj, only mom watched", []string{"BJ"}, []string{"Mom Nando"}, true, 0, false},
		{"tagged bj, both watched", []string{"BJ"}, []string{"BJ", "Mom Nando"}, true, 0, true},
		{"untagged min 1", users, []string{"BJ"}, false, 1, true},
		{"tagged ignores min", []string{"BJ", "Mom Nando"}, []string{"BJ"}, true, 1, false},
		{"empty audience", nil, []string{"BJ"}, false, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Eligible(tt.aud, tt.watchedBy, tt.tagged, tt.min); got != tt.want {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
}

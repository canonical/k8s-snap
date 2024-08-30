package utils_test

import (
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestSecondsToExpirationDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		seconds int
		want    time.Time
	}{
		{
			name:    "Test 2 Year",
			seconds: 2 * 365 * 24 * 60 * 60,
			want:    now.AddDate(2, 0, 0),
		},
		{
			name:    "Test 2 Month",
			seconds: 2 * 30 * 24 * 60 * 60,
			want:    now.AddDate(0, 2, 0),
		},
		{
			name:    "Test 2 Days",
			seconds: 2 * 24 * 60 * 60,
			want:    now.AddDate(0, 0, 2),
		},
		{
			name:    "Test Default",
			seconds: 10 * 60,
			want:    now.Add(time.Duration(10*60) * time.Second),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got := utils.SecondsToExpirationDate(now, tt.seconds)
			g.Expect(got).To(Equal(tt.want))
		})

	}

}

func TestTTLToSeconds(t *testing.T) {
	tests := []struct {
		name    string
		ttl     string
		want    int
		wantErr bool
	}{
		{
			name:    "Test Year",
			ttl:     "1y",
			want:    31536000,
			wantErr: false,
		},
		{
			name:    "Test Month",
			ttl:     "1mo",
			want:    2592000,
			wantErr: false,
		},
		{
			name:    "Test 10 Minutes",
			ttl:     "10m",
			want:    600,
			wantErr: false,
		},
		{
			name:    "Test Day",
			ttl:     "1d",
			want:    86400,
			wantErr: false,
		},
		{
			name:    "Test Default ParseDuration",
			ttl:     "1h",
			want:    3600,
			wantErr: false,
		},
		{
			name:    "Test Invalid Unit",
			ttl:     "1ftn",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Test Invalid Month",
			ttl:     "10mod",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Test Invalid Month Suffix",
			ttl:     "10xmo",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Test Invalid Number",
			ttl:     "ad",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Test Invalid Format",
			ttl:     "1",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got, err := utils.TTLToSeconds(tt.ttl)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

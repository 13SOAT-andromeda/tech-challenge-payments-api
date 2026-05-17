package sqs

import (
	"testing"
)

func TestUnwrapSNS(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "SNS envelope is unwrapped",
			body: `{"Type":"Notification","Message":"{\"event_type\":\"order.created\"}"}`,
			want: `{"event_type":"order.created"}`,
		},
		{
			name: "plain JSON passes through unchanged",
			body: `{"event_type":"order.created"}`,
			want: `{"event_type":"order.created"}`,
		},
		{
			name: "invalid JSON passes through unchanged",
			body: `not json`,
			want: `not json`,
		},
		{
			name: "SNS envelope with empty Message passes through unchanged",
			body: `{"Type":"Notification","Message":""}`,
			want: `{"Type":"Notification","Message":""}`,
		},
		{
			name: "non-Notification type passes through unchanged",
			body: `{"Type":"SubscriptionConfirmation","Message":"confirm-token"}`,
			want: `{"Type":"SubscriptionConfirmation","Message":"confirm-token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unwrapSNS(tt.body)
			if got != tt.want {
				t.Errorf("unwrapSNS() = %q, want %q", got, tt.want)
			}
		})
	}
}

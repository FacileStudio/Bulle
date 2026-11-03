package push

import (
	"net/http"
	"testing"
)

func TestBadTokenOnlyForgetsMalformedTokens(t *testing.T) {
	cases := []struct {
		status int
		reply  string
		want   bool
	}{
		{http.StatusBadRequest, `{"error":{"message":"The registration token is not a valid FCM registration token","status":"INVALID_ARGUMENT"}}`, true},
		{http.StatusBadRequest, `{"error":{"message":"Invalid value at 'message.android.ttl'","status":"INVALID_ARGUMENT"}}`, false},
		{http.StatusForbidden, `{"error":{"message":"SenderId mismatch","status":"PERMISSION_DENIED"}}`, false},
	}
	for _, c := range cases {
		if got := badToken(c.status, []byte(c.reply)); got != c.want {
			t.Errorf("badToken(%d, %s) = %v, want %v", c.status, c.reply, got, c.want)
		}
	}
}

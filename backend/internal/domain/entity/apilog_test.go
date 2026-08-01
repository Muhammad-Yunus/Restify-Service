package entity

import "testing"

func TestAPILogValidate(t *testing.T) {
	tests := []struct {
		name string
		log  APILog
		want bool
	}{
		{
			name: "valid",
			log:  APILog{RequestID: "req-1", Method: "GET", Path: "/users", StatusCode: 200, LogLevel: LevelInfo},
			want: true,
		},
		{
			name: "missing request id",
			log:  APILog{Method: "GET", Path: "/users", StatusCode: 200, LogLevel: LevelInfo},
			want: false,
		},
		{
			name: "invalid status code",
			log:  APILog{RequestID: "req-1", Method: "GET", Path: "/users", StatusCode: 99, LogLevel: LevelInfo},
			want: false,
		},
		{
			name: "missing log level",
			log:  APILog{RequestID: "req-1", Method: "GET", Path: "/users", StatusCode: 200},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.log.Validate()
			if (err == nil) != tc.want {
				t.Errorf("Validate() error = %v, want success = %v", err, tc.want)
			}
		})
	}
}

package models

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUser_UnmarshalJSON(t *testing.T) {
	type fields struct {
		ID          int
		FirstName   string
		LastName    string
		Sex         int
		Photo       string
		Deactivated string
		LastSeen    *time.Time
		Groups      []*Group
	}
	data, _ := json.Marshal(map[string]map[string]interface{}{
		"last_seen": {
			"time":     1597602025,
			"platform": 1,
		},
	})
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test1",
			fields: fields{},
			args: args{
				data: data,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:          tt.fields.ID,
				FirstName:   tt.fields.FirstName,
				LastName:    tt.fields.LastName,
				Sex:         tt.fields.Sex,
				Photo:       tt.fields.Photo,
				Deactivated: tt.fields.Deactivated,
				LastSeen:    tt.fields.LastSeen,
				Groups:      tt.fields.Groups,
			}
			if err := u.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, u.LastSeen.Unix(), int64(1597602025))

		})
	}
}

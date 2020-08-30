package vkore

import (
	"testing"
)

var expResult = `
var members = [];
var offset = 0;

var count = 0;
var i = 0;
while (i < 25 && (offset + 10000) < 25000) {
  var m = API.groups.getMembers({
    "group_id": 66666,
    "v": "5.27",
    "sort": "id_asc",
    "count": "1000",
    "offset": (10000 + offset),
    "fields": "sex,deactivated,last_seen,photo,photo_200,city,status"
  });
  members.push(m.items);
  count = m.count;
  offset = offset + 1000;
  i = i + 1;
};
return { "users": members, "count": count };
`

func Test_groupMembersQuery(t *testing.T) {
	type args struct {
		gmp *groupMembersParams
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				gmp: &groupMembersParams{
					Offset:     10000,
					TotalCount: 25000,
					GroupID:    66666,
					UsersCount: 1000,
				},
			},
			want: expResult,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := groupMembersQuery(tt.args.gmp); got != tt.want {
				t.Errorf("groupMembersQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getGroupMembers(t *testing.T) {
	Init()
	type args struct {
		queryParams *groupMembersParams
		result      *userGroupsResult
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				queryParams: &groupMembersParams{
					Offset:     0,
					TotalCount: 10000,
					GroupID:    61645362,
					UsersCount: 1000,
				},
				result: new(userGroupsResult),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := getGroupMembers(tt.args.queryParams, tt.args.result); (err != nil) != tt.wantErr {
				t.Errorf("getGroupMembers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

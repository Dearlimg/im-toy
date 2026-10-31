package impl

import (
	"errors"
	"testing"

	"im-chat/internal/application/dto"
	"im-chat/internal/domain/entity"
	"im-chat/internal/domain/repository"
)

type socialUsers struct{ repository.UserRepository }

func (socialUsers) FindByUIDs(uids []string) ([]*entity.User, error) {
	users := make([]*entity.User, 0, len(uids))
	for _, uid := range uids {
		users = append(users, &entity.User{UID: uid})
	}
	return users, nil
}

type socialGroups struct {
	repository.GroupRepository
	members []*entity.GroupMember
	err     error
}

func (r *socialGroups) CreateWithMembers(_ *entity.Group, members []*entity.GroupMember) error {
	r.members = members
	return r.err
}

func TestCreateGroupDeduplicatesMembersAndPropagatesFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		r := &socialGroups{}
		if fail {
			r.err = errors.New("member insert failed")
		}
		s := NewGroupServiceImpl(r, nil, socialUsers{})
		group, err := s.CreateGroup("owner", &dto.CreateGroupReq{Name: " test ", MemberUIDs: []string{"owner", "bob", "bob", "", "alice"}})
		if (err != nil) != fail {
			t.Fatalf("failure=%v: %v", fail, err)
		}
		if !fail && group.Name != "test" {
			t.Fatalf("unexpected group: %+v", group)
		}
		if len(r.members) != 3 || r.members[0].UID != "owner" || r.members[1].UID != "bob" || r.members[2].UID != "alice" {
			t.Fatalf("wrong members: %+v", r.members)
		}
	}
}

type socialFriends struct {
	repository.FriendRepository
	accepted bool
	err      error
}

func (r *socialFriends) FindByUIDAndFriendUID(uid, friendUID string) (*entity.Friend, error) {
	return &entity.Friend{UID: uid, FriendUID: friendUID, Status: 2}, nil
}
func (r *socialFriends) AcceptRequest(uid, friendUID string) error {
	r.accepted = uid == "alice" && friendUID == "bob"
	return r.err
}
func TestAcceptFriendUsesAtomicPairWrite(t *testing.T) {
	for _, fail := range []bool{false, true} {
		r := &socialFriends{}
		if fail {
			r.err = errors.New("transaction rolled back")
		}
		s := NewFriendServiceImpl(r, nil)
		err := s.AcceptFriend("alice", &dto.AcceptFriendReq{FriendUID: "bob"})
		if !r.accepted || (err != nil) != fail {
			t.Fatalf("accepted=%v err=%v", r.accepted, err)
		}
	}
}

func TestFriendMapDecodesDatabaseBytes(t *testing.T) {
	if got := getString(map[string]interface{}{"uid": []byte("bob")}, "uid"); got != "bob" {
		t.Fatalf("got %q", got)
	}
}

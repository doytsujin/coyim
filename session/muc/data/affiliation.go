package data

import "fmt"

const (
	// AffiliationOwner represents XMPP muc 'owner' affiliation
	AffiliationOwner = "owner"
	// AffiliationAdmin represents XMPP muc 'admin' affiliation
	AffiliationAdmin = "admin"
	// AffiliationMember represents XMPP muc 'member' affiliation
	AffiliationMember = "member"
	// AffiliationOutcast represents XMPP muc 'outcast' affiliation
	AffiliationOutcast = "outcast"
	// AffiliationNone represents XMPP muc 'none' affiliation
	AffiliationNone = "none"
)

// AffiliationUpdate contains information related to a new and previous affiliation
type AffiliationUpdate struct {
	Nickname string
	Reason   string
	New      Affiliation
	Previous Affiliation
	Actor    *Actor
}

// SelfAffiliationUpdate contains information related to a new and previous affiliation of the self occupant
type SelfAffiliationUpdate struct {
	AffiliationUpdate
}

// Affiliation represents an affiliation as specificed by section 5.2 in XEP-0045
type Affiliation interface {
	// IsAdmin will return true if this specific affiliation can modify persistent information
	IsAdmin() bool
	// IsBanned will return true if this specific affiliation means that the jid is banned from the room
	IsBanned() bool
	// IsMember will return true if this specific affiliation means that the jid is a member of the room
	IsMember() bool
	// IsOwner will return true if this specific affiliation means that the jid is an owner of the room
	IsOwner() bool
	// IsNone will return true if if the jid doesn't have affiliation
	IsNone() bool
	// Name returns the string name of the affiliation type
	Name() string
	// IsLowerThan returns true if the caller affiliation has a lower hierarchy than the affiliation passed as argument
	IsLowerThan(Affiliation) bool
	// IsDifferentFrom returns a Boolean value indicating whether the given affiliation is different from the current one
	IsDifferentFrom(Affiliation) bool
}

// NoneAffiliation is a representation of MUC's "none" affiliation
type NoneAffiliation struct{}

// OutcastAffiliation is a representation of MUC's "banned" affiliation
type OutcastAffiliation struct{}

// MemberAffiliation is a representation of MUC's "member" affiliation
type MemberAffiliation struct{}

// AdminAffiliation is a representation of MUC's "admin" affiliation
type AdminAffiliation struct{}

// OwnerAffiliation is a representation of MUC's "owner" affiliation
type OwnerAffiliation struct{}

// IsAdmin implements Affiliation interface
func (*NoneAffiliation) IsAdmin() bool { return false }

// IsAdmin implements Affiliation interface
func (*OutcastAffiliation) IsAdmin() bool { return false }

// IsAdmin implements Affiliation interface
func (*MemberAffiliation) IsAdmin() bool { return false }

// IsAdmin implements Affiliation interface
func (*AdminAffiliation) IsAdmin() bool { return true }

// IsAdmin implements Affiliation interface
func (*OwnerAffiliation) IsAdmin() bool { return false }

// IsBanned implements Affiliation interface
func (*NoneAffiliation) IsBanned() bool { return false }

// IsBanned implements Affiliation interface
func (*OutcastAffiliation) IsBanned() bool { return true }

// IsBanned implements Affiliation interface
func (*MemberAffiliation) IsBanned() bool { return false }

// IsBanned implements Affiliation interface
func (*AdminAffiliation) IsBanned() bool { return false }

// IsBanned implements Affiliation interface
func (*OwnerAffiliation) IsBanned() bool { return false }

// IsMember implements Affiliation interface
func (*NoneAffiliation) IsMember() bool { return false }

// IsMember implements Affiliation interface
func (*OutcastAffiliation) IsMember() bool { return false }

// IsMember implements Affiliation interface
func (*MemberAffiliation) IsMember() bool { return true }

// IsMember implements Affiliation interface
func (*AdminAffiliation) IsMember() bool { return true }

// IsMember implements Affiliation interface
func (*OwnerAffiliation) IsMember() bool { return true }

// IsModerator implements Affiliation interface
func (*NoneAffiliation) IsModerator() bool { return false }

// IsModerator implements Affiliation interface
func (*OutcastAffiliation) IsModerator() bool { return false }

// IsModerator implements Affiliation interface
func (*MemberAffiliation) IsModerator() bool { return false }

// IsModerator implements Affiliation interface
func (*AdminAffiliation) IsModerator() bool { return true }

// IsModerator implements Affiliation interface
func (*OwnerAffiliation) IsModerator() bool { return true }

// IsOwner implements Affiliation interface
func (*NoneAffiliation) IsOwner() bool { return false }

// IsOwner implements Affiliation interface
func (*OutcastAffiliation) IsOwner() bool { return false }

// IsOwner implements Affiliation interface
func (*MemberAffiliation) IsOwner() bool { return false }

// IsOwner implements Affiliation interface
func (*AdminAffiliation) IsOwner() bool { return false }

// IsOwner implements Affiliation interface
func (*OwnerAffiliation) IsOwner() bool { return true }

// IsOutcast implements Affiliation interface
func (*NoneAffiliation) IsOutcast() bool { return false }

// IsOutcast implements Affiliation interface
func (*OutcastAffiliation) IsOutcast() bool { return true }

// IsOutcast implements Affiliation interface
func (*MemberAffiliation) IsOutcast() bool { return false }

// IsOutcast implements Affiliation interface
func (*AdminAffiliation) IsOutcast() bool { return false }

// IsOutcast implements Affiliation interface
func (*OwnerAffiliation) IsOutcast() bool { return false }

// IsNone implements Affiliation interface
func (*NoneAffiliation) IsNone() bool { return true }

// IsNone implements Affiliation interface
func (*OutcastAffiliation) IsNone() bool { return false }

// IsNone implements Affiliation interface
func (*MemberAffiliation) IsNone() bool { return false }

// IsNone implements Affiliation interface
func (*AdminAffiliation) IsNone() bool { return false }

// IsNone implements Affiliation interface
func (*OwnerAffiliation) IsNone() bool { return false }

// Name implements Affiliation interface
func (*NoneAffiliation) Name() string { return AffiliationNone }

// Name implements Affiliation interface
func (*OutcastAffiliation) Name() string { return AffiliationOutcast }

// Name implements Affiliation interface
func (*MemberAffiliation) Name() string { return AffiliationMember }

// Name implements Affiliation interface
func (*AdminAffiliation) Name() string { return AffiliationAdmin }

// Name implements Affiliation interface
func (*OwnerAffiliation) Name() string { return AffiliationOwner }

// IsLowerThan implements Affiliation interface
func (*NoneAffiliation) IsLowerThan(a Affiliation) bool {
	return !a.IsNone()
}

// IsLowerThan implements Affiliation interface
func (*OutcastAffiliation) IsLowerThan(Affiliation) bool {
	return true
}

// IsLowerThan implements Affiliation interface
func (*MemberAffiliation) IsLowerThan(a Affiliation) bool {
	return a.IsAdmin() || a.IsOwner()
}

// IsLowerThan implements Affiliation interface
func (*AdminAffiliation) IsLowerThan(a Affiliation) bool {
	return a.IsOwner()
}

// IsLowerThan implements Affiliation interface
func (*OwnerAffiliation) IsLowerThan(a Affiliation) bool {
	return false
}

// IsDifferentFrom implements Affiliation interface
func (*NoneAffiliation) IsDifferentFrom(a Affiliation) bool {
	return !a.IsNone()
}

// IsDifferentFrom implements Affiliation interface
func (*OutcastAffiliation) IsDifferentFrom(a Affiliation) bool {
	return !a.IsBanned()
}

// IsDifferentFrom implements Affiliation interface
// An administrator and an owner are members too
// For that reason, it validates if the affiliation given is an admin or owner, it should be different to member
func (*MemberAffiliation) IsDifferentFrom(a Affiliation) bool {
	return !a.IsMember() || a.IsAdmin() || a.IsOwner()
}

// IsDifferentFrom implements Affiliation interface
func (*AdminAffiliation) IsDifferentFrom(a Affiliation) bool {
	return !a.IsAdmin()
}

// IsDifferentFrom implements Affiliation interface
func (*OwnerAffiliation) IsDifferentFrom(a Affiliation) bool {
	return !a.IsOwner()
}

// AffiliationFromString returns an Affiliation from the given string, or an error if the string doesn't match a known affiliation type
func AffiliationFromString(s string) (Affiliation, error) {
	switch s {
	case AffiliationNone:
		return &NoneAffiliation{}, nil
	case AffiliationOutcast:
		return &OutcastAffiliation{}, nil
	case AffiliationMember:
		return &MemberAffiliation{}, nil
	case AffiliationAdmin:
		return &AdminAffiliation{}, nil
	case AffiliationOwner:
		return &OwnerAffiliation{}, nil
	default:
		return nil, fmt.Errorf("unknown affiliation string: '%s'", s)
	}
}

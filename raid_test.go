package raid

import "testing"

func TestUnknownRoleHasEmptyRoster(t *testing.T) {
	if got := New(map[string]int{"tank": 1}).Roster("dps"); len(got) != 0 {
		t.Fatalf("没人的角色名单应该是空的，得到 %v", got)
	}
}

func TestSignupThenRoster(t *testing.T) {
	desk := New(map[string]int{"tank": 2})
	if !desk.Signup("a", "tank") {
		t.Fatal("有名额时报名应该入团")
	}
	if got := desk.Roster("tank"); len(got) != 1 || got[0] != "a" {
		t.Fatalf("入团名单应该只有 a，得到 %v", got)
	}
}

func TestWaitingStartsEmpty(t *testing.T) {
	if got := New(map[string]int{"tank": 1}).Waiting(); len(got) != 0 {
		t.Fatalf("还没人排队时候补应该是空的，得到 %v", got)
	}
}

func TestScannedStartsAtZero(t *testing.T) {
	if New(map[string]int{"tank": 1}).Scanned() != 0 {
		t.Fatal("Scanned 初始应该是 0")
	}
}

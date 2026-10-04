package raid

import (
	"testing"
	"time"
)

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

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("名单不一致：想要 %v，得到 %v", want, got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("名单不一致：想要 %v，得到 %v", want, got)
		}
	}
}

func assertInvariants(t *testing.T, desk *Desk) {
	t.Helper()
	seen := map[string]bool{}
	for role := range desk.quota {
		if len(desk.Roster(role)) > desk.quota[role] {
			t.Fatalf("角色 %s 人数 %d 超过配额 %d", role, len(desk.Roster(role)), desk.quota[role])
		}
		for _, player := range desk.Roster(role) {
			if seen[player] {
				t.Fatalf("玩家 %s 同时占了多个位置", player)
			}
			seen[player] = true
		}
	}
	for _, player := range desk.Waiting() {
		if seen[player] {
			t.Fatalf("玩家 %s 同时出现在名单与候补中", player)
		}
		seen[player] = true
	}
}

func TestSignupRejectsUnknownRole(t *testing.T) {
	desk := New(map[string]int{"tank": 1})
	if desk.Signup("a", "healer") {
		t.Fatal("未知角色应该直接拒绝")
	}
	if got := desk.Roster("healer"); len(got) != 0 {
		t.Fatalf("未知角色名单应该为空，得到 %v", got)
	}
	if got := desk.Waiting(); len(got) != 0 {
		t.Fatalf("被拒绝的报名不该进候补，得到 %v", got)
	}
}

func TestRoleQuotaOverflowGoesToWaitingInOrder(t *testing.T) {
	desk := New(map[string]int{"tank": 2, "healer": 1})
	if !desk.Signup("a", "tank") || !desk.Signup("b", "tank") {
		t.Fatal("有名额时报名应该入团")
	}
	if desk.Signup("c", "tank") {
		t.Fatal("名额满了应该进候补而不是入团")
	}
	if !desk.Signup("h", "healer") {
		t.Fatal("不同角色各算各的配额")
	}
	if desk.Signup("hh", "healer") {
		t.Fatal("治疗满员也应该进候补")
	}
	assertStrings(t, desk.Roster("tank"), []string{"a", "b"})
	assertStrings(t, desk.Waiting(), []string{"c", "hh"})
	assertInvariants(t, desk)
}

func TestLeaveUnknownPlayerReturnsFalseAndKeepsState(t *testing.T) {
	desk := New(map[string]int{"tank": 1})
	desk.Signup("a", "tank")
	desk.Signup("b", "tank")
	if desk.Leave("nobody") {
		t.Fatal("没报名过的玩家退出应该返回 false")
	}
	assertStrings(t, desk.Roster("tank"), []string{"a"})
	assertStrings(t, desk.Waiting(), []string{"b"})
}

func TestLeavePromotesFirstWaitingOfSameRole(t *testing.T) {
	desk := New(map[string]int{"tank": 1, "healer": 1})
	desk.Signup("t1", "tank")
	desk.Signup("t2", "tank")
	desk.Signup("t3", "tank")
	desk.Signup("h1", "healer")
	desk.Signup("h2", "healer")

	if !desk.Leave("h1") {
		t.Fatal("已入团玩家退出应该返回 true")
	}
	assertStrings(t, desk.Roster("healer"), []string{"h2"})
	assertStrings(t, desk.Roster("tank"), []string{"t1"})
	assertStrings(t, desk.Waiting(), []string{"t2", "t3"})

	if !desk.Leave("t1") {
		t.Fatal("已入团玩家退出应该返回 true")
	}
	assertStrings(t, desk.Roster("tank"), []string{"t2"})
	assertStrings(t, desk.Waiting(), []string{"t3"})
	assertInvariants(t, desk)
}

func TestLeaveFromWaitingDoesNotPromote(t *testing.T) {
	desk := New(map[string]int{"tank": 1})
	desk.Signup("a", "tank")
	desk.Signup("b", "tank")
	desk.Signup("c", "tank")
	if !desk.Leave("b") {
		t.Fatal("候补玩家退出应该返回 true")
	}
	assertStrings(t, desk.Roster("tank"), []string{"a"})
	assertStrings(t, desk.Waiting(), []string{"c"})
	if desk.Leave("b") {
		t.Fatal("已经退出的玩家再退出应该返回 false")
	}
	desk.Leave("a")
	assertStrings(t, desk.Roster("tank"), []string{"c"})
	assertStrings(t, desk.Waiting(), []string{})
	assertInvariants(t, desk)
}

func TestDuplicateSignupIsIdempotent(t *testing.T) {
	desk := New(map[string]int{"tank": 1})
	if !desk.Signup("a", "tank") {
		t.Fatal("首次报名应该入团")
	}
	for attempt := 0; attempt < 3; attempt++ {
		if desk.Signup("a", "tank") {
			t.Fatal("同一玩家重复报名应该返回 false")
		}
		if desk.Signup("a", "healer") {
			t.Fatal("同一玩家换角色重复报名也应该返回 false")
		}
	}
	assertStrings(t, desk.Roster("tank"), []string{"a"})
	assertStrings(t, desk.Waiting(), []string{})

	desk.Signup("b", "tank")
	if desk.Signup("b", "tank") {
		t.Fatal("候补玩家重复报名也应该返回 false")
	}
	assertStrings(t, desk.Waiting(), []string{"b"})
	assertInvariants(t, desk)
}

func TestSameSequenceReplayedGivesSameResult(t *testing.T) {
	ops := []struct {
		signup      bool
		player      string
		role        string
		wantJoined  bool
		wantRemoved bool
	}{
		{signup: true, player: "a", role: "tank", wantJoined: true},
		{signup: true, player: "b", role: "tank", wantJoined: true},
		{signup: true, player: "c", role: "tank", wantJoined: false},
		{signup: false, player: "a", wantRemoved: true},
		{signup: true, player: "a", role: "tank", wantJoined: false},
		{signup: false, player: "ghost", wantRemoved: false},
		{signup: false, player: "b", wantRemoved: true},
	}
	run := func() (roster []string, waiting []string) {
		desk := New(map[string]int{"tank": 2})
		for _, op := range ops {
			if op.signup {
				if got := desk.Signup(op.player, op.role); got != op.wantJoined {
					t.Fatalf("Signup(%s) 返回 %v，想要 %v", op.player, got, op.wantJoined)
				}
			} else {
				if got := desk.Leave(op.player); got != op.wantRemoved {
					t.Fatalf("Leave(%s) 返回 %v，想要 %v", op.player, got, op.wantRemoved)
				}
			}
		}
		return desk.Roster("tank"), desk.Waiting()
	}
	roster1, waiting1 := run()
	roster2, waiting2 := run()
	assertStrings(t, roster1, roster2)
	assertStrings(t, waiting1, waiting2)
	assertStrings(t, roster1, []string{"c", "a"})
	assertStrings(t, waiting1, []string{})
}

func TestScannedDoesNotGrowWithSignupCount(t *testing.T) {
	const signups = 200000
	desk := New(map[string]int{"tank": signups})
	for index := 0; index < signups; index++ {
		desk.Signup("p-"+itoa(index), "tank")
	}
	const queries = 100000
	start := time.Now()
	for index := 0; index < queries; index++ {
		desk.Roster("tank")
	}
	if desk.Scanned() != queries {
		t.Fatalf("索引查询应该每次只计常数扫描，Scanned=%d 想要 %d", desk.Scanned(), queries)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("%d 次查询耗时 %v，达不到每秒 10 万次查询", queries, elapsed)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

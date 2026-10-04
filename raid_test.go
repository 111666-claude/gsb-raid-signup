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

func TestRoleQuotaFillsWaitingQueue(t *testing.T) {
	desk := New(map[string]int{"tank": 1})
	if !desk.Signup("a", "tank") {
		t.Fatal("第一个 tank 应该入团")
	}
	if desk.Signup("b", "tank") {
		t.Fatal("名额满后第二个 tank 应该进候补而不是入团")
	}
	if desk.Signup("c", "tank") {
		t.Fatal("第三个 tank 也应该进候补")
	}
	if got := desk.Roster("tank"); len(got) != 1 || got[0] != "a" {
		t.Fatalf("入团名单应该只有 a，得到 %v", got)
	}
	if got := desk.Waiting(); len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Fatalf("候补应按报名先后为 b、c，得到 %v", got)
	}
	if desk.Signup("x", "healer") {
		t.Fatal("未知角色必须拒绝")
	}
	if got := desk.Roster("healer"); len(got) != 0 {
		t.Fatalf("未知角色不能留下名单，得到 %v", got)
	}
}

func TestLeavePromotesSameRoleInOrder(t *testing.T) {
	desk := New(map[string]int{"tank": 1, "healer": 1})
	desk.Signup("a", "tank")
	desk.Signup("b", "healer")
	desk.Signup("c", "tank")
	desk.Signup("d", "tank")

	if !desk.Leave("a") {
		t.Fatal("已报名玩家退出应返回 true")
	}
	if got := desk.Roster("tank"); len(got) != 1 || got[0] != "c" {
		t.Fatalf("退出后应补同角色候补队首 c，得到 %v", got)
	}
	if got := desk.Waiting(); len(got) != 1 || got[0] != "d" {
		t.Fatalf("候补应该只剩 d，得到 %v", got)
	}

	if !desk.Leave("d") {
		t.Fatal("候补玩家退出也应返回 true")
	}
	if got := desk.Waiting(); len(got) != 0 {
		t.Fatalf("候补退出后应清空，得到 %v", got)
	}

	if desk.Leave("nobody") {
		t.Fatal("没报名过的玩家退出必须返回 false")
	}
	leaveBefore := append([]string{}, desk.Roster("tank")...)
	if desk.Leave("nobody"); len(desk.Roster("tank")) != len(leaveBefore) {
		t.Fatal("查无此人的退出不能改变名单")
	}
}

func TestDuplicateSignupIsIdempotent(t *testing.T) {
	desk := New(map[string]int{"tank": 1})
	if !desk.Signup("a", "tank") {
		t.Fatal("第一次报名应入团")
	}
	if desk.Signup("a", "tank") {
		t.Fatal("重复报名必须返回 false")
	}
	if got := desk.Roster("tank"); len(got) != 1 {
		t.Fatalf("重复报名不能多占位置，得到 %v", got)
	}

	desk.Signup("b", "tank")
	if desk.Signup("b", "tank") {
		t.Fatal("候补玩家重复报名也必须返回 false")
	}
	if got := desk.Waiting(); len(got) != 1 {
		t.Fatalf("候补不能重复排队，得到 %v", got)
	}
	if desk.Signup("a", "healer") {
		t.Fatal("同一玩家换角色重复报名也必须拒绝")
	}

	desk.Leave("a")
	if got := desk.Roster("tank"); len(got) != 1 || got[0] != "b" {
		t.Fatalf("退出后应补 b，得到 %v", got)
	}
}

func TestLeaveThenRejoinIsDeterministic(t *testing.T) {
	run := func() (roster []string, waiting []string) {
		desk := New(map[string]int{"tank": 1})
		desk.Signup("a", "tank")
		desk.Signup("b", "tank")
		desk.Leave("a")
		desk.Signup("a", "tank")
		desk.Leave("nobody")
		return desk.Roster("tank"), desk.Waiting()
	}
	r1, w1 := run()
	r2, w2 := run()
	if len(r1) != len(r2) || (len(r1) > 0 && (r1[0] != r2[0])) {
		t.Fatalf("重复执行入团结果应一致：%v vs %v", r1, r2)
	}
	if len(w1) != len(w2) || (len(w1) > 0 && (w1[0] != w2[0])) {
		t.Fatalf("重复执行候补结果应一致：%v vs %v", w1, w2)
	}
}

func TestScannedDoesNotGrowWithTableSize(t *testing.T) {
	desk := New(map[string]int{"tank": 1000000})
	for index := 0; index < 3000; index++ {
		desk.Signup("p", "tank")
	}
	_ = desk.Roster("tank")
	_ = desk.Waiting()
	if got := desk.Scanned(); got > 3000+3000+2 {
		t.Fatalf("查询只能按索引探测记账，得到 scanned=%d", got)
	}
}

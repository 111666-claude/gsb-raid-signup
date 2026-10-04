// Package raid 是团队报名：按角色配额收人，超编进候补，有人退出后按顺序补位。
package raid

import "sort"

type membership struct {
	role    string
	waiting bool
	seq     int64 // 进候补的全局先后顺序，仅 waiting 时使用
}

// Desk 是报名台账。
type Desk struct {
	quota    map[string]int
	roster   map[string][]string   // 角色 -> 已入团玩家（报名顺序）
	waiting  map[string][]string   // 角色 -> 候补队列
	waitHead map[string]int        // 角色 -> 候补队首下标
	players  map[string]membership // 玩家 -> 当前位置，O(1) 查重与退出定位
	waitSeq  int64
	scanned  int
}

// New 建台账：quota 是每个角色的名额。
func New(quota map[string]int) *Desk {
	copied := map[string]int{}
	for role, size := range quota {
		copied[role] = size
	}
	return &Desk{
		quota:    copied,
		roster:   map[string][]string{},
		waiting:  map[string][]string{},
		waitHead: map[string]int{},
		players:  map[string]membership{},
	}
}

// Signup 报名：未知角色或同一玩家重复报名返回 false 且不改名单；
// 角色名额未满直接入团（返回 true），满了进候补排队（返回 false）。
func (d *Desk) Signup(player, role string) bool {
	if _, known := d.quota[role]; !known {
		return false
	}
	if _, exists := d.players[player]; exists {
		return false
	}
	if len(d.roster[role]) < d.quota[role] {
		d.roster[role] = append(d.roster[role], player)
		d.players[player] = membership{role: role}
		return true
	}
	d.waitSeq++
	d.waiting[role] = append(d.waiting[role], player)
	d.players[player] = membership{role: role, waiting: true, seq: d.waitSeq}
	return false
}

// Leave 退出：没报名过（含已退出）的玩家返回 false 且不改名单；
// 从入团名单移除后，把同角色候补队首补进名单；候补玩家退出只移除、不补位。
func (d *Desk) Leave(player string) bool {
	member, ok := d.players[player]
	if !ok {
		return false
	}
	delete(d.players, player)
	if member.waiting {
		d.removeWaiting(member.role, player)
		return true
	}
	d.removeRoster(member.role, player)
	d.promote(member.role)
	return true
}

// Roster 返回该角色已入团的玩家（报名顺序）。
// 走索引直接返回，调用方不得修改返回的切片。
func (d *Desk) Roster(role string) []string {
	d.scanned++
	return d.roster[role]
}

// Waiting 返回候补名单（全局报名先后顺序）。
func (d *Desk) Waiting() []string {
	d.scanned++
	type item struct {
		seq    int64
		player string
	}
	all := make([]item, 0, len(d.players))
	for role, queue := range d.waiting {
		for _, player := range queue[d.waitHead[role]:] {
			all = append(all, item{seq: d.players[player].seq, player: player})
		}
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].seq < all[j].seq })
	out := make([]string, len(all))
	for index, entry := range all {
		out[index] = entry.player
	}
	return out
}

// Scanned 是累计查询扫描的报名数（规模观测）：索引查询每次只计常数，
// 不随报名总数增长。
func (d *Desk) Scanned() int { return d.scanned }

func (d *Desk) removeRoster(role, player string) {
	list := d.roster[role]
	for index, name := range list {
		if name != player {
			continue
		}
		copy(list[index:], list[index+1:])
		d.roster[role] = list[:len(list)-1]
		return
	}
}

func (d *Desk) removeWaiting(role, player string) {
	queue := d.waiting[role]
	head := d.waitHead[role]
	for index := head; index < len(queue); index++ {
		if queue[index] != player {
			continue
		}
		copy(queue[index:len(queue)-1], queue[index+1:])
		d.waiting[role] = queue[:len(queue)-1]
		d.compactWaiting(role)
		return
	}
}

// popWaiting 弹出同角色候补队首。
func (d *Desk) popWaiting(role string) (string, bool) {
	queue := d.waiting[role]
	head := d.waitHead[role]
	if head >= len(queue) {
		return "", false
	}
	player := queue[head]
	d.waitHead[role] = head + 1
	d.compactWaiting(role)
	return player, true
}

// compactWaiting 在队首偏移过大时回收已出队的空间。
func (d *Desk) compactWaiting(role string) {
	queue := d.waiting[role]
	head := d.waitHead[role]
	switch {
	case head >= len(queue):
		d.waiting[role] = nil
		d.waitHead[role] = 0
	case head > 0 && head*2 >= len(queue):
		d.waiting[role] = append([]string(nil), queue[head:]...)
		d.waitHead[role] = 0
	}
}

// promote 把同角色候补队首补进入团名单。
func (d *Desk) promote(role string) {
	player, ok := d.popWaiting(role)
	if !ok {
		return
	}
	d.roster[role] = append(d.roster[role], player)
	d.players[player] = membership{role: role}
}

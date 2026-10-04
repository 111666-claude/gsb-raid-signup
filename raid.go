// Package raid 是团队报名：按角色配额收人，超编进候补，有人退出后按候补顺序补位。
package raid

import (
	"container/heap"
	"container/list"
)

// waitEntry 是候补队列里的一条记录，seq 是全局报名先后序号。
type waitEntry struct {
	player string
	seq    int64
}

// waitItem 用于把各角色候补队列按全局序号归并。
type waitItem struct {
	seq    int64
	player string
	role   string
	elem   *list.Element
}

type waitHeap []waitItem

func (h waitHeap) Len() int           { return len(h) }
func (h waitHeap) Less(i, j int) bool { return h[i].seq < h[j].seq }
func (h waitHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *waitHeap) Push(x any)        { *h = append(*h, x.(waitItem)) }
func (h *waitHeap) Pop() any {
	old := *h
	last := len(old) - 1
	item := old[last]
	*h = old[:last]
	return item
}

// Desk 是报名台账。所有查询都走索引，按探测次数记账，不全表扫描。
type Desk struct {
	quota   map[string]int
	roster  map[string]*list.List
	waiting map[string]*list.List
	rnode   map[string]*list.Element
	wnode   map[string]*list.Element
	roleOf  map[string]string
	scanned int
	nextSeq int64
}

// New 建台账：quota 是每个角色的名额。
func New(quota map[string]int) *Desk {
	copied := map[string]int{}
	roster := map[string]*list.List{}
	waiting := map[string]*list.List{}
	for role, size := range quota {
		copied[role] = size
		roster[role] = list.New()
		waiting[role] = list.New()
	}
	return &Desk{
		quota:   copied,
		roster:  roster,
		waiting: waiting,
		rnode:   map[string]*list.Element{},
		wnode:   map[string]*list.Element{},
		roleOf:  map[string]string{},
	}
}

// Signup 报名，返回是否直接入团；未知角色、重复报名返回 false，名额满进候补。
func (d *Desk) Signup(player, role string) bool {
	d.scanned++
	size, ok := d.quota[role]
	if !ok {
		return false
	}
	if _, exists := d.roleOf[player]; exists {
		return false
	}
	d.roleOf[player] = role
	if d.roster[role].Len() < size {
		d.rnode[player] = d.roster[role].PushBack(player)
		return true
	}
	d.nextSeq++
	d.wnode[player] = d.waiting[role].PushBack(waitEntry{player: player, seq: d.nextSeq})
	return false
}

// Leave 把玩家从入团名单或候补里移除；入团名单空出名额后补同角色候补队首。
// 没报名过的玩家返回 false 且不改变任何名单。
func (d *Desk) Leave(player string) bool {
	d.scanned++
	if elem, ok := d.rnode[player]; ok {
		role := d.roleOf[player]
		d.roster[role].Remove(elem)
		delete(d.rnode, player)
		delete(d.roleOf, player)
		d.promote(role)
		return true
	}
	if elem, ok := d.wnode[player]; ok {
		role := d.roleOf[player]
		d.waiting[role].Remove(elem)
		delete(d.wnode, player)
		delete(d.roleOf, player)
		return true
	}
	return false
}

// promote 把该角色候补队首补进入团名单。
func (d *Desk) promote(role string) {
	queue := d.waiting[role]
	if queue.Len() == 0 {
		return
	}
	entry := queue.Remove(queue.Front()).(waitEntry)
	delete(d.wnode, entry.player)
	d.rnode[entry.player] = d.roster[role].PushBack(entry.player)
}

// Roster 返回该角色已入团的玩家（报名顺序）；未知角色返回空名单。
func (d *Desk) Roster(role string) []string {
	d.scanned++
	out := []string{}
	if players := d.roster[role]; players != nil {
		out = make([]string, 0, players.Len())
		for elem := players.Front(); elem != nil; elem = elem.Next() {
			out = append(out, elem.Value.(string))
		}
	}
	return out
}

// Waiting 返回候补名单（按报名先后的全局顺序）。
func (d *Desk) Waiting() []string {
	d.scanned++
	h := waitHeap{}
	for role, queue := range d.waiting {
		if queue.Len() == 0 {
			continue
		}
		elem := queue.Front()
		entry := elem.Value.(waitEntry)
		h = append(h, waitItem{seq: entry.seq, player: entry.player, role: role, elem: elem})
	}
	heap.Init(&h)
	total := 0
	for _, queue := range d.waiting {
		total += queue.Len()
	}
	out := make([]string, 0, total)
	for h.Len() > 0 {
		item := heap.Pop(&h).(waitItem)
		out = append(out, item.player)
		next := item.elem.Next()
		if next != nil {
			entry := next.Value.(waitEntry)
			heap.Push(&h, waitItem{seq: entry.seq, player: entry.player, role: item.role, elem: next})
		}
	}
	return out
}

// Scanned 是累计扫描的报名数（规模观测）：每次索引探测记 1，不随表增长乘查询次数。
func (d *Desk) Scanned() int { return d.scanned }

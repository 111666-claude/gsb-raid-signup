// Package raid 是团队报名：按角色配额收人，超编进候补，有人退出后按候补顺序补位。
// 缺陷：不看角色配额、重复报名占两个位置、退出后不清候补也不补位、每次查询全扫报名表。
package raid

type entry struct {
	player string
	role   string
}

// Desk 是报名台账。
type Desk struct {
	quota   map[string]int
	roster  map[string][]string
	waiting []entry
	scanned int
}

// New 建台账：quota 是每个角色的名额。
func New(quota map[string]int) *Desk {
	copied := map[string]int{}
	for role, size := range quota {
		copied[role] = size
	}
	return &Desk{quota: copied, roster: map[string][]string{}}
}

// Signup 报名，返回是否直接入团。缺陷：不看配额、不查重、满员也不进候补。
func (d *Desk) Signup(player, role string) bool {
	d.roster[role] = append(d.roster[role], player)
	return true
}

// Leave 退出，返回是否真的找到了这个人。缺陷：只从入团名单删，不补位、未知玩家也返回 true。
func (d *Desk) Leave(player string) bool {
	for role, list := range d.roster {
		for index, name := range list {
			if name == player {
				d.roster[role] = append(list[:index], list[index+1:]...)
				return true
			}
		}
	}
	return true
}

// Roster 返回该角色已入团的玩家（报名顺序）。缺陷：每次全扫该角色名单。
func (d *Desk) Roster(role string) []string {
	d.scanned += len(d.roster[role])
	return append([]string{}, d.roster[role]...)
}

// Waiting 返回候补名单（加入顺序）。缺陷：全扫候补表。
func (d *Desk) Waiting() []string {
	d.scanned += len(d.waiting)
	out := []string{}
	for _, item := range d.waiting {
		out = append(out, item.player)
	}
	return out
}

// Scanned 是累计扫描的报名数（规模观测）。
func (d *Desk) Scanned() int { return d.scanned }

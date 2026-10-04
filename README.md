# raid-signup

团队报名：按角色配额收人，超编进候补，有人退出后按候补顺序补位，只用 Go 标准库。

```
go run ./cmd/raid-signup --sample=quota
go run ./cmd/raid-signup --sample=dup
go run ./cmd/raid-signup --sample=unknown
go run ./cmd/raid-signup --sample=scan
go test ./...
go vet ./...
```

## 口径（README 为准）

- **角色配额**：每个角色入团人数不超过 `quota[role]`；未知角色直接拒绝。
- **超编进候补**：名额满了就进候补（`Signup` 返回 false），候补按报名先后排队。
- **退出补位**：`Leave` 把玩家从入团名单移除后，从候补里按顺序补第一个同角色的人进来；
  退出没报名过的玩家返回 false 且不改变任何名单。
- **幂等**：同一个玩家重复报名返回 false，且不会同时出现在名单与候补里。
- **不变量**：每个角色人数不超过配额；同一玩家最多出现在一个位置；同一串报名与退出重复执行结果相同。
- **代价**：查询不许全扫报名表，`Scanned` 不随报名总数乘查询次数增长；20 万报名、每秒 10 万次查询。

## 输出契约（不改格式）

```
roster=.. waiting=..
roster=..
scanned=..
```

## 目录

```
raid.go            角色配额、候补队列与退出补位
raid_test.go       单元测试
cmd/raid-signup/   命令行入口
```

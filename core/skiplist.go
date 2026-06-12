package core

import "math/rand"

const (
	skipListMaxLevel = 32
	skipListP        = 0.25
)

// SkipListLevel 跳表的一个层级，包含前向指针和跨度。
type SkipListLevel struct {
	Forward *SkipListNode
	Span    int64
}

// SkipListNode 跳表节点。
type SkipListNode struct {
	Member   string
	Score    float64
	Backward *SkipListNode
	Levels   []SkipListLevel
}

// SkipList 跳表结构。
type SkipList struct {
	Header *SkipListNode
	Tail   *SkipListNode
	Length int64
	Level  int
}

// NewSkipList 创建一个空跳表。
func NewSkipList() *SkipList {
	sl := &SkipList{
		Level: 1,
	}
	header := &SkipListNode{
		Levels: make([]SkipListLevel, skipListMaxLevel),
	}
	for i := range skipListMaxLevel {
		header.Levels[i].Forward = nil
		header.Levels[i].Span = 0
	}
	sl.Header = header
	return sl
}

// randomLevel 随机生成新节点的层级（1-based，≥1）。
func randomLevel() int {
	level := 1
	for rand.Float64() < skipListP && level < skipListMaxLevel {
		level++
	}
	return level
}

// Find 查找指定 member 的节点。
func (sl *SkipList) Find(score float64, member string) *SkipListNode {
	x := sl.Header
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Levels[i].Forward != nil &&
			(x.Levels[i].Forward.Score < score ||
				(x.Levels[i].Forward.Score == score &&
					x.Levels[i].Forward.Member < member)) {
			x = x.Levels[i].Forward
		}
	}
	x = x.Levels[0].Forward
	if x != nil && x.Score == score && x.Member == member {
		return x
	}
	return nil
}

// Insert 插入或更新 member 的 score。返回新建的节点（仅更新 score 时返回 nil）。
func (sl *SkipList) Insert(score float64, member string) *SkipListNode {
	update := make([]*SkipListNode, skipListMaxLevel)
	rank := make([]int64, skipListMaxLevel)

	x := sl.Header
	for i := sl.Level - 1; i >= 0; i-- {
		if i == sl.Level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		for x.Levels[i].Forward != nil &&
			(x.Levels[i].Forward.Score < score ||
				(x.Levels[i].Forward.Score == score &&
					x.Levels[i].Forward.Member < member)) {
			rank[i] += x.Levels[i].Span
			x = x.Levels[i].Forward
		}
		update[i] = x
	}

	if x.Levels[0].Forward != nil &&
		x.Levels[0].Forward.Score == score &&
		x.Levels[0].Forward.Member == member {
		x.Levels[0].Forward.Score = score
		return nil
	}

	newLevel := randomLevel()
	if newLevel > sl.Level {
		for i := sl.Level; i < newLevel; i++ {
			update[i] = sl.Header
			update[i].Levels[i].Span = sl.Length
			rank[i] = 0
		}
		sl.Level = newLevel
	}

	newNode := &SkipListNode{
		Member: member,
		Score:  score,
		Levels: make([]SkipListLevel, newLevel),
	}

	for i := range newLevel {
		newNode.Levels[i].Forward = update[i].Levels[i].Forward
		update[i].Levels[i].Forward = newNode

		newNode.Levels[i].Span = update[i].Levels[i].Span - (rank[0] - rank[i])
		update[i].Levels[i].Span = (rank[0] - rank[i]) + 1
	}

	for i := newLevel; i < sl.Level; i++ {
		update[i].Levels[i].Span++
	}

	if update[0] == sl.Header {
		newNode.Backward = nil
	} else {
		newNode.Backward = update[0]
	}

	if newNode.Levels[0].Forward != nil {
		newNode.Levels[0].Forward.Backward = newNode
	} else {
		sl.Tail = newNode
	}

	sl.Length++
	return newNode
}

// Delete 删除指定 member。返回是否成功删除。
func (sl *SkipList) Delete(score float64, member string) bool {
	update := make([]*SkipListNode, skipListMaxLevel)

	x := sl.Header
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Levels[i].Forward != nil &&
			(x.Levels[i].Forward.Score < score ||
				(x.Levels[i].Forward.Score == score &&
					x.Levels[i].Forward.Member < member)) {
			x = x.Levels[i].Forward
		}
		update[i] = x
	}

	x = x.Levels[0].Forward
	if x == nil || x.Score != score || x.Member != member {
		return false
	}

	for i := 0; i < sl.Level; i++ {
		if update[i].Levels[i].Forward == x {
			update[i].Levels[i].Span += x.Levels[i].Span - 1
			update[i].Levels[i].Forward = x.Levels[i].Forward
		} else {
			update[i].Levels[i].Span--
		}
	}

	if x.Levels[0].Forward != nil {
		x.Levels[0].Forward.Backward = x.Backward
	} else {
		sl.Tail = x.Backward
	}

	for sl.Level > 1 && sl.Header.Levels[sl.Level-1].Forward == nil {
		sl.Level--
	}

	sl.Length--
	return true
}

// GetRank 返回 member 的 0-based 排名。不存在返回 -1。
func (sl *SkipList) GetRank(score float64, member string) int64 {
	var rank int64 = 0
	x := sl.Header
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Levels[i].Forward != nil &&
			(x.Levels[i].Forward.Score < score ||
				(x.Levels[i].Forward.Score == score &&
					x.Levels[i].Forward.Member < member)) {
			rank += x.Levels[i].Span
			x = x.Levels[i].Forward
		}
	}

	x = x.Levels[0].Forward
	if x != nil && x.Score == score && x.Member == member {
		return rank
	}
	return -1
}

// GetElementByRank 返回指定排名（0-based）的节点。
func (sl *SkipList) GetElementByRank(rank int64) *SkipListNode {
	if rank < 0 || rank >= sl.Length {
		return nil
	}

	x := sl.Header
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Levels[i].Forward != nil && x.Levels[i].Span <= rank {
			rank -= x.Levels[i].Span
			x = x.Levels[i].Forward
		}
	}
	return x.Levels[0].Forward
}

// GetRange 返回 [start, stop] 范围内的 member 列表。支持负索引。
func (sl *SkipList) GetRange(start, stop int64) []string {
	if start < 0 {
		start = sl.Length + start
	}
	if stop < 0 {
		stop = sl.Length + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= sl.Length {
		stop = sl.Length - 1
	}
	if start > stop {
		return nil
	}

	node := sl.GetElementByRank(start)
	result := make([]string, 0, stop-start+1)
	for i := start; i <= stop && node != nil; i++ {
		result = append(result, node.Member)
		node = node.Levels[0].Forward
	}
	return result
}

// GetScore 遍历查找 member 的 score。不存在返回 false。
func (sl *SkipList) GetScore(member string) (float64, bool) {
	for x := sl.Header.Levels[0].Forward; x != nil; x = x.Levels[0].Forward {
		if x.Member == member {
			return x.Score, true
		}
	}
	return 0, false
}

package core

import (
	"container/list"
	"fmt"
)

// --- 内部方法 ---

// getList 获取 list 类型的链表，调用者必须持有锁。
// key 不存在或类型不匹配返回 nil, false。
func (s *Store) getList(key string) (*list.List, bool) {
	entry, ok := s.get(key)
	if !ok || entry.Type != DataList {
		return nil, false
	}
	return entry.List, true
}

// lookupList 获取 list 类型的链表，内置惰性过期删除和类型检查。
// 调用者必须持有锁。
func (s *Store) lookupList(key string) (*list.List, bool) {
	s.expireIfNeeded(key)
	return s.getList(key)
}

// lpush 左端推入一个或多个值。调用者必须持有锁，且已确保类型正确。
// 返回操作后的列表长度。
func (s *Store) lpush(key string, values ...string) int {
	entry, ok := s.get(key)
	if !ok {
		entry = &DataEntry{Type: DataList, List: list.New()}
		s.data[key] = entry
	}
	for _, value := range values {
		entry.List.PushFront(value)
	}
	return entry.List.Len()
}

// rpush 右端推入一个或多个值。调用者必须持有锁，且已确保类型正确。
// 返回操作后的列表长度。
func (s *Store) rpush(key string, values ...string) int {
	entry, ok := s.get(key)
	if !ok {
		entry = &DataEntry{Type: DataList, List: list.New()}
		s.data[key] = entry
	}
	for _, value := range values {
		entry.List.PushBack(value)
	}
	return entry.List.Len()
}

// --- 导出方法 ---

// LPush 将一个或多个值推入列表左端，返回操作后的列表长度。
func (s *Store) LPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeeded(key)
	entry, ok := s.get(key)
	if ok && entry.Type != DataList {
		return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return s.lpush(key, values...), nil
}

// RPush 将一个或多个值推入列表右端，返回操作后的列表长度。
func (s *Store) RPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeeded(key)
	entry, ok := s.get(key)
	if ok && entry.Type != DataList {
		return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return s.rpush(key, values...), nil
}

// LPop 弹出列表左端的值。key 不存在或列表为空返回 "", false。
func (s *Store) LPop(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	lst, ok := s.lookupList(key)
	if !ok {
		return "", false
	}
	e := lst.Front()
	if e == nil {
		return "", false
	}
	lst.Remove(e)
	return e.Value.(string), true
}

// RPop 弹出列表右端的值。
func (s *Store) RPop(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	lst, ok := s.lookupList(key)
	if !ok {
		return "", false
	}
	e := lst.Back()
	if e == nil {
		return "", false
	}
	lst.Remove(e)
	return e.Value.(string), true
}

// LLen 返回列表长度。key 不存在或类型不匹配返回 0。
func (s *Store) LLen(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	lst, ok := s.lookupList(key)
	if !ok {
		return 0
	}
	return lst.Len()
}

// LRange 返回列表中指定范围 [start, stop] 的元素。支持负索引。
func (s *Store) LRange(key string, start, stop int) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	lst, ok := s.lookupList(key)
	if !ok {
		return nil
	}
	length := lst.Len()
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop || start >= length {
		return []string{}
	}
	result := make([]string, 0, stop-start+1)
	i := 0
	for e := lst.Front(); e != nil; e = e.Next() {
		if i >= start && i <= stop {
			result = append(result, e.Value.(string))
		}
		if i >= stop {
			break
		}
		i++
	}
	return result
}

// LIndex 返回列表中指定索引的元素。支持负索引。
func (s *Store) LIndex(key string, index int) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	lst, ok := s.lookupList(key)
	if !ok {
		return "", false
	}
	length := lst.Len()
	if index < 0 {
		index = length + index
	}
	if index < 0 || index >= length {
		return "", false
	}
	e := lst.Front()
	for i := 0; i < index; i++ {
		e = e.Next()
	}
	return e.Value.(string), true
}

// LSet 设置列表中指定索引的值。索引越界返回错误。
func (s *Store) LSet(key string, index int, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	lst, ok := s.lookupList(key)
	if !ok {
		return fmt.Errorf("no such key or wrong type")
	}
	length := lst.Len()
	if index < 0 {
		index = length + index
	}
	if index < 0 || index >= length {
		return fmt.Errorf("index out of range")
	}
	e := lst.Front()
	for i := 0; i < index; i++ {
		e = e.Next()
	}
	e.Value = value
	return nil
}

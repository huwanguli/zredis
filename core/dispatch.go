package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CommandFunc 命令处理函数签名：接收 Store 和转好的字符串参数，返回 RESP Value。
type CommandFunc func(s *Store, args []string) Value

// Dispatcher 命令分发器，根据命令名路由到对应处理函数。
type Dispatcher struct {
	cmds map[string]CommandFunc
}

// NewDispatcher 创建一个注册了所有支持命令的分发器。
func NewDispatcher() *Dispatcher {
	d := &Dispatcher{cmds: make(map[string]CommandFunc)}
	d.registerDefaults()
	return d
}

// Dispatch 接收解析后的命令数组，执行并返回结果。
func (d *Dispatcher) Dispatch(s *Store, cmd *ArrayVal) Value {
	if len(cmd.Items) == 0 {
		return NewErrorVal("ERR empty command")
	}

	cmdName := ""
	if bv, ok := cmd.Items[0].(*BulkVal); ok {
		cmdName = strings.ToUpper(string(bv.Data))
	} else if sv, ok := cmd.Items[0].(*StringVal); ok {
		cmdName = strings.ToUpper(sv.Str)
	} else {
		return NewErrorVal("ERR invalid command format")
	}

	handler, ok := d.cmds[cmdName]
	if !ok {
		return NewErrorVal("ERR unknown command '" + cmdName + "'")
	}

	args := make([]string, 0, len(cmd.Items)-1)
	for _, item := range cmd.Items[1:] {
		switch v := item.(type) {
		case *BulkVal:
			args = append(args, string(v.Data))
		case *StringVal:
			args = append(args, v.Str)
		default:
			return NewErrorVal("ERR invalid argument format")
		}
	}

	return handler(s, args)
}

// registerDefaults 注册所有内置命令。
func (d *Dispatcher) registerDefaults() {
	// --- 通用 ---
	d.cmds["PING"] = cmdPing
	d.cmds["DEL"] = cmdDel
	d.cmds["EXISTS"] = cmdExists
	d.cmds["KEYS"] = cmdKeys
	d.cmds["FLUSH"] = cmdFlush
	d.cmds["EXPIRE"] = cmdExpire
	d.cmds["TTL"] = cmdTTL
	d.cmds["PERSIST"] = cmdPersist

	// --- String ---
	d.cmds["SET"] = cmdSet
	d.cmds["GET"] = cmdGet
	d.cmds["INCR"] = cmdIncr
	d.cmds["INCRBY"] = cmdIncrBy
	d.cmds["DECR"] = cmdDecr
	d.cmds["DECRBY"] = cmdDecrBy
	d.cmds["APPEND"] = cmdAppend
	d.cmds["STRLEN"] = cmdStrLen
	d.cmds["MGET"] = cmdMGet
	d.cmds["MSET"] = cmdMSet
	d.cmds["GETSET"] = cmdGetSet
	d.cmds["SETEX"] = cmdSetEX
	d.cmds["SETNX"] = cmdSetNX

	// --- Hash ---
	d.cmds["HSET"] = cmdHSet
	d.cmds["HGET"] = cmdHGet
	d.cmds["HDEL"] = cmdHDel
	d.cmds["HGETALL"] = cmdHGetAll
	d.cmds["HEXISTS"] = cmdHExists
	d.cmds["HLEN"] = cmdHLen
	d.cmds["HKEYS"] = cmdHKeys
	d.cmds["HVALS"] = cmdHVals

	// --- List ---
	d.cmds["LPUSH"] = cmdLPush
	d.cmds["RPUSH"] = cmdRPush
	d.cmds["LPOP"] = cmdLPop
	d.cmds["RPOP"] = cmdRPop
	d.cmds["LLEN"] = cmdLLen
	d.cmds["LRANGE"] = cmdLRange
	d.cmds["LINDEX"] = cmdLIndex
	d.cmds["LSET"] = cmdLSet

	// --- Set ---
	d.cmds["SADD"] = cmdSAdd
	d.cmds["SREM"] = cmdSRem
	d.cmds["SMEMBERS"] = cmdSMembers
	d.cmds["SISMEMBER"] = cmdSIsMember
	d.cmds["SCARD"] = cmdSCard
	d.cmds["SINTER"] = cmdSInter
	d.cmds["SUNION"] = cmdSUnion
	d.cmds["SDIFF"] = cmdSDiff

	// --- ZSet ---
	d.cmds["ZADD"] = cmdZAdd
	d.cmds["ZRANGE"] = cmdZRange
	d.cmds["ZRANK"] = cmdZRank
	d.cmds["ZSCORE"] = cmdZScore
	d.cmds["ZREM"] = cmdZRem
	d.cmds["ZCARD"] = cmdZCard
}

// --- 通用 ---

func cmdPing(s *Store, args []string) Value {
	if len(args) != 0 {
		if len(args) == 1 {
			return NewBulkVal([]byte(args[0]))
		}
		return NewErrorVal("ERR wrong number of arguments for 'PING'")
	}
	return NewStringVal("PONG")
}

func cmdDel(s *Store, args []string) Value {
	if len(args) < 1 {
		return NewErrorVal("ERR wrong number of arguments for 'DEL'")
	}
	count := 0
	for _, key := range args {
		if s.Del(key) {
			count++
		}
	}
	return NewIntVal(int64(count))
}

func cmdExists(s *Store, args []string) Value {
	if len(args) < 1 {
		return NewErrorVal("ERR wrong number of arguments for 'EXISTS'")
	}
	count := 0
	for _, key := range args {
		if s.Exists(key) {
			count++
		}
	}
	return NewIntVal(int64(count))
}

func cmdKeys(s *Store, args []string) Value {
	keys := s.Keys()
	items := make([]Value, len(keys))
	for i, k := range keys {
		items[i] = NewBulkVal([]byte(k))
	}
	return NewArrayVal(items)
}

func cmdFlush(s *Store, args []string) Value {
	s.Flush()
	return NewStringVal("OK")
}

func cmdExpire(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'EXPIRE'")
	}
	seconds, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return NewErrorVal("ERR value is not an integer")
	}
	ok := s.Expire(args[0], time.Duration(seconds)*time.Second)
	if ok {
		return NewIntVal(1)
	}
	return NewIntVal(0)
}

func cmdTTL(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'TTL'")
	}
	d := s.TTL(args[0])
	if d < 0 {
		return NewIntVal(int64(d / time.Nanosecond))
	}
	return NewIntVal(int64(d / time.Second))
}

func cmdPersist(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'PERSIST'")
	}
	if s.Persist(args[0]) {
		return NewIntVal(1)
	}
	return NewIntVal(0)
}

// --- String ---

func cmdSet(s *Store, args []string) Value {
	if len(args) < 2 {
		return NewErrorVal("ERR wrong number of arguments for 'SET'")
	}
	s.Set(args[0], args[1])
	return NewStringVal("OK")
}

func cmdGet(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'GET'")
	}
	v, ok := s.Get(args[0])
	if !ok {
		return NewNullBulkVal()
	}
	return NewBulkVal([]byte(v))
}

func cmdIncr(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'INCR'")
	}
	n, err := s.Incr(args[0])
	if err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewIntVal(n)
}

func cmdIncrBy(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'INCRBY'")
	}
	delta, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return NewErrorVal("ERR value is not an integer")
	}
	n, err := s.IncrBy(args[0], delta)
	if err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewIntVal(n)
}

func cmdDecr(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'DECR'")
	}
	n, err := s.Decr(args[0])
	if err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewIntVal(n)
}

func cmdDecrBy(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'DECRBY'")
	}
	delta, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return NewErrorVal("ERR value is not an integer")
	}
	n, err := s.DecrBy(args[0], delta)
	if err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewIntVal(n)
}

func cmdAppend(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'APPEND'")
	}
	n, err := s.Append(args[0], args[1])
	if err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewIntVal(int64(n))
}

func cmdStrLen(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'STRLEN'")
	}
	return NewIntVal(int64(s.StrLen(args[0])))
}

func cmdMGet(s *Store, args []string) Value {
	if len(args) < 1 {
		return NewErrorVal("ERR wrong number of arguments for 'MGET'")
	}
	vals := s.MGet(args...)
	items := make([]Value, len(vals))
	for i, v := range vals {
		if v == "" && !s.Exists(args[i]) {
			items[i] = NewNullBulkVal()
		} else {
			items[i] = NewBulkVal([]byte(v))
		}
	}
	return NewArrayVal(items)
}

func cmdMSet(s *Store, args []string) Value {
	if len(args) < 2 || len(args)%2 != 0 {
		return NewErrorVal("ERR wrong number of arguments for 'MSET'")
	}
	kvs := make(map[string]string, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		kvs[args[i]] = args[i+1]
	}
	s.MSet(kvs)
	return NewStringVal("OK")
}

func cmdGetSet(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'GETSET'")
	}
	old, ok := s.GetSet(args[0], args[1])
	if !ok {
		return NewNullBulkVal()
	}
	return NewBulkVal([]byte(old))
}

func cmdSetEX(s *Store, args []string) Value {
	if len(args) != 3 {
		return NewErrorVal("ERR wrong number of arguments for 'SETEX'")
	}
	seconds, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return NewErrorVal("ERR value is not an integer")
	}
	s.SetEX(args[0], args[2], time.Duration(seconds)*time.Second)
	return NewStringVal("OK")
}

func cmdSetNX(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'SETNX'")
	}
	if s.SetNX(args[0], args[1]) {
		return NewIntVal(1)
	}
	return NewIntVal(0)
}

// --- Hash ---

func cmdHSet(s *Store, args []string) Value {
	if len(args) < 2 || (len(args)-1)%2 != 0 {
		return NewErrorVal("ERR wrong number of arguments for 'HSET'")
	}
	pairs := args[1:]
	n, err := s.HSet(args[0], pairs...)
	if err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewIntVal(int64(n))
}

func cmdHGet(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'HGET'")
	}
	v, ok := s.HGet(args[0], args[1])
	if !ok {
		return NewNullBulkVal()
	}
	return NewBulkVal([]byte(v))
}

func cmdHDel(s *Store, args []string) Value {
	if len(args) < 2 {
		return NewErrorVal("ERR wrong number of arguments for 'HDEL'")
	}
	n := s.HDel(args[0], args[1:]...)
	return NewIntVal(int64(n))
}

func cmdHGetAll(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'HGETALL'")
	}
	m := s.HGetAll(args[0])
	items := make([]Value, 0, len(m)*2)
	for f, v := range m {
		items = append(items, NewBulkVal([]byte(f)), NewBulkVal([]byte(v)))
	}
	return NewArrayVal(items)
}

func cmdHExists(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'HEXISTS'")
	}
	if s.HExists(args[0], args[1]) {
		return NewIntVal(1)
	}
	return NewIntVal(0)
}

func cmdHLen(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'HLEN'")
	}
	return NewIntVal(int64(s.HLen(args[0])))
}

func cmdHKeys(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'HKEYS'")
	}
	keys := s.HKeys(args[0])
	items := make([]Value, len(keys))
	for i, k := range keys {
		items[i] = NewBulkVal([]byte(k))
	}
	return NewArrayVal(items)
}

func cmdHVals(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'HVALS'")
	}
	vals := s.HVals(args[0])
	items := make([]Value, len(vals))
	for i, v := range vals {
		items[i] = NewBulkVal([]byte(v))
	}
	return NewArrayVal(items)
}

// --- List ---

func cmdLPush(s *Store, args []string) Value {
	if len(args) < 2 {
		return NewErrorVal("ERR wrong number of arguments for 'LPUSH'")
	}
	n, err := s.LPush(args[0], args[1:]...)
	if err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewIntVal(int64(n))
}

func cmdRPush(s *Store, args []string) Value {
	if len(args) < 2 {
		return NewErrorVal("ERR wrong number of arguments for 'RPUSH'")
	}
	n, err := s.RPush(args[0], args[1:]...)
	if err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewIntVal(int64(n))
}

func cmdLPop(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'LPOP'")
	}
	v, ok := s.LPop(args[0])
	if !ok {
		return NewNullBulkVal()
	}
	return NewBulkVal([]byte(v))
}

func cmdRPop(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'RPOP'")
	}
	v, ok := s.RPop(args[0])
	if !ok {
		return NewNullBulkVal()
	}
	return NewBulkVal([]byte(v))
}

func cmdLLen(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'LLEN'")
	}
	return NewIntVal(int64(s.LLen(args[0])))
}

func cmdLRange(s *Store, args []string) Value {
	if len(args) != 3 {
		return NewErrorVal("ERR wrong number of arguments for 'LRANGE'")
	}
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return NewErrorVal("ERR value is not an integer")
	}
	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return NewErrorVal("ERR value is not an integer")
	}
	vals := s.LRange(args[0], start, stop)
	items := make([]Value, len(vals))
	for i, v := range vals {
		items[i] = NewBulkVal([]byte(v))
	}
	return NewArrayVal(items)
}

func cmdLIndex(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'LINDEX'")
	}
	idx, err := strconv.Atoi(args[1])
	if err != nil {
		return NewErrorVal("ERR value is not an integer")
	}
	v, ok := s.LIndex(args[0], idx)
	if !ok {
		return NewNullBulkVal()
	}
	return NewBulkVal([]byte(v))
}

func cmdLSet(s *Store, args []string) Value {
	if len(args) != 3 {
		return NewErrorVal("ERR wrong number of arguments for 'LSET'")
	}
	idx, err := strconv.Atoi(args[1])
	if err != nil {
		return NewErrorVal("ERR value is not an integer")
	}
	if err := s.LSet(args[0], idx, args[2]); err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewStringVal("OK")
}

// --- Set ---

func cmdSAdd(s *Store, args []string) Value {
	if len(args) < 2 {
		return NewErrorVal("ERR wrong number of arguments for 'SADD'")
	}
	n, err := s.SAdd(args[0], args[1:]...)
	if err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewIntVal(int64(n))
}

func cmdSRem(s *Store, args []string) Value {
	if len(args) < 2 {
		return NewErrorVal("ERR wrong number of arguments for 'SREM'")
	}
	n, err := s.SRem(args[0], args[1:]...)
	if err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewIntVal(int64(n))
}

func cmdSMembers(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'SMEMBERS'")
	}
	members := s.SMembers(args[0])
	items := make([]Value, len(members))
	for i, m := range members {
		items[i] = NewBulkVal([]byte(m))
	}
	return NewArrayVal(items)
}

func cmdSIsMember(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'SISMEMBER'")
	}
	if s.SIsMember(args[0], args[1]) {
		return NewIntVal(1)
	}
	return NewIntVal(0)
}

func cmdSCard(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'SCARD'")
	}
	return NewIntVal(int64(s.SCard(args[0])))
}

func cmdSInter(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'SINTER'")
	}
	vals := s.SInter(args[0], args[1])
	items := make([]Value, len(vals))
	for i, v := range vals {
		items[i] = NewBulkVal([]byte(v))
	}
	return NewArrayVal(items)
}

func cmdSUnion(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'SUNION'")
	}
	vals := s.SUnion(args[0], args[1])
	items := make([]Value, len(vals))
	for i, v := range vals {
		items[i] = NewBulkVal([]byte(v))
	}
	return NewArrayVal(items)
}

func cmdSDiff(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'SDIFF'")
	}
	vals := s.SDiff(args[0], args[1])
	items := make([]Value, len(vals))
	for i, v := range vals {
		items[i] = NewBulkVal([]byte(v))
	}
	return NewArrayVal(items)
}

// --- ZSet ---

func cmdZAdd(s *Store, args []string) Value {
	if len(args) < 3 || (len(args)-1)%2 != 0 {
		return NewErrorVal("ERR wrong number of arguments for 'ZADD'")
	}
	n, err := s.ZAdd(args[0], args[1:]...)
	if err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewIntVal(int64(n))
}

func cmdZRange(s *Store, args []string) Value {
	if len(args) != 3 {
		return NewErrorVal("ERR wrong number of arguments for 'ZRANGE'")
	}
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return NewErrorVal("ERR value is not an integer")
	}
	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return NewErrorVal("ERR value is not an integer")
	}
	vals := s.ZRange(args[0], start, stop)
	items := make([]Value, len(vals))
	for i, v := range vals {
		items[i] = NewBulkVal([]byte(v))
	}
	return NewArrayVal(items)
}

func cmdZRank(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'ZRANK'")
	}
	rank, ok := s.ZRank(args[0], args[1])
	if !ok {
		return NewNullBulkVal()
	}
	return NewIntVal(rank)
}

func cmdZScore(s *Store, args []string) Value {
	if len(args) != 2 {
		return NewErrorVal("ERR wrong number of arguments for 'ZSCORE'")
	}
	score, ok := s.ZScore(args[0], args[1])
	if !ok {
		return NewNullBulkVal()
	}
	return NewBulkVal([]byte(strconv.FormatFloat(score, 'f', -1, 64)))
}

func cmdZRem(s *Store, args []string) Value {
	if len(args) < 2 {
		return NewErrorVal("ERR wrong number of arguments for 'ZREM'")
	}
	n, err := s.ZRem(args[0], args[1:]...)
	if err != nil {
		return NewErrorVal("ERR " + err.Error())
	}
	return NewIntVal(int64(n))
}

func cmdZCard(s *Store, args []string) Value {
	if len(args) != 1 {
		return NewErrorVal("ERR wrong number of arguments for 'ZCARD'")
	}
	return NewIntVal(int64(s.ZCard(args[0])))
}

var _ = fmt.Sprintf
var _ = strings.ToUpper

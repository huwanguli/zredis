package core

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

// Server 是一个 RESP-TCP 服务器，将 Store 暴露为网络服务。
type Server struct {
	store *Store
	disp  *Dispatcher
	aof   *AOF
}

// NewServer 创建一个新的 Server 实例。
func NewServer() *Server {
	store := NewStore()
	disp := NewDispatcher()
	aof, err := NewAOF("aof_file.aof")
	if err != nil {
		log.Fatalf("failed to create AOF: %v", err)
	}
	if err := aof.Load(store, disp); err != nil {
		log.Printf("AOF load error: %v", err)
	}
	return &Server{
		store: store,
		disp:  disp,
		aof:   aof,
	}
}

// Listen 在指定地址上监听 TCP 连接，阻塞直到出错。
func (srv *Server) Listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	log.Printf("server listening on %s", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go srv.handleConn(conn)
	}
}

// handleConn 处理单个客户端连接。
func (srv *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	log.Printf("[%s] connected", remote)
	r := NewReader(conn)
	w := NewWriter(conn)
	for {
		val, err := r.ReadValue()
		if err != nil {
			if err == io.EOF {
				log.Printf("[%s] disconnected", remote)
				break
			}
			log.Printf("[%s] read error: %v", remote, err)
			return
		}
		cmd, ok := val.(*ArrayVal)
		if !ok {
			return
		}
		log.Printf("[%s] %s", remote, srv.cmdName(cmd))

		resp := srv.disp.Dispatch(srv.store, cmd)
		if IsWriteCmd(srv.cmdName(cmd)) {
			if err := srv.aof.Write(cmd); err != nil {
				log.Printf("[%s] AOF write error: %v", remote, err)
			}
		}
		if err := w.WriteValue(resp); err != nil {
			return
		}
		if err := w.Flush(); err != nil {
			return
		}
	}
}

// cmdName 从命令数组中提取命令名用于日志。
func (srv *Server) cmdName(cmd *ArrayVal) string {
	if len(cmd.Items) == 0 {
		return "(empty)"
	}
	if bv, ok := cmd.Items[0].(*BulkVal); ok {
		return strings.ToUpper(string(bv.Data))
	}
	return "(unknown)"
}

var _ = fmt.Errorf
var _ = io.EOF

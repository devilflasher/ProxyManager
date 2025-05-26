package server

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"proxy-manager-desktop/internal/config"
)

type HTTPProxy struct {
	config      *config.ProxyConfig
	server      *http.Server
	isRunning   bool
	stopChannel chan struct{}
}

func NewHTTPProxy(proxyConfig *config.ProxyConfig) (*HTTPProxy, error) {
	if proxyConfig.Local.Protocol != "http" {
		return nil, fmt.Errorf("本地协议必须是HTTP，当前为: %s", proxyConfig.Local.Protocol)
	}

	proxy := &HTTPProxy{
		config:      proxyConfig,
		stopChannel: make(chan struct{}),
	}

	return proxy, nil
}

func (p *HTTPProxy) Start() error {
	if p.isRunning {
		return fmt.Errorf("代理已在运行")
	}

	listenAddr := fmt.Sprintf("%s:%d", p.config.Local.ListenIP, p.config.Local.ListenPort)
	p.server = &http.Server{
		Addr:    listenAddr,
		Handler: http.HandlerFunc(p.handleHTTPRequest),
	}

	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP代理服务器错误: %v\n", err)
		}
	}()

	p.isRunning = true
	fmt.Printf("HTTP代理开始监听 %s\n", listenAddr)
	return nil
}

func (p *HTTPProxy) Stop() error {
	if !p.isRunning {
		return fmt.Errorf("代理未运行")
	}

	p.isRunning = false

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := p.server.Shutdown(ctx); err != nil {
			fmt.Printf("HTTP代理优雅关闭失败，强制关闭: %v\n", err)
			p.server.Close()
		}
		fmt.Printf("HTTP代理已完全停止\n")
	}()

	fmt.Printf("HTTP代理正在停止\n")
	return nil
}

func (p *HTTPProxy) IsRunning() bool {
	return p.isRunning
}

func (p *HTTPProxy) GetConfig() *config.ProxyConfig {
	return p.config
}

func (p *HTTPProxy) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		p.handleHTTPSConnect(w, r)
	} else {
		p.handleHTTPForward(w, r)
	}
}

func (p *HTTPProxy) handleHTTPSConnect(w http.ResponseWriter, r *http.Request) {
	upstreamConn, err := p.connectUpstream(r.Host)
	if err != nil {
		http.Error(w, fmt.Sprintf("连接上游代理失败: %v", err), http.StatusBadGateway)
		return
	}
	defer upstreamConn.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "不支持连接劫持", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, fmt.Sprintf("劫持连接失败: %v", err), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		return
	}

	p.relay(clientConn, upstreamConn)
}

func (p *HTTPProxy) handleHTTPForward(w http.ResponseWriter, r *http.Request) {
	upstreamConn, err := p.connectUpstream(r.Host)
	if err != nil {
		http.Error(w, fmt.Sprintf("连接上游代理失败: %v", err), http.StatusBadGateway)
		return
	}
	defer upstreamConn.Close()
	if err := r.Write(upstreamConn); err != nil {
		http.Error(w, fmt.Sprintf("发送请求失败: %v", err), http.StatusBadGateway)
		return
	}

	resp, err := http.ReadResponse(bufio.NewReader(upstreamConn), r)
	if err != nil {
		http.Error(w, fmt.Sprintf("读取响应失败: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (p *HTTPProxy) connectUpstream(targetAddr string) (net.Conn, error) {
	upstreamAddr := p.config.Upstream.Address
	upstreamConn, err := net.Dial("tcp", upstreamAddr)
	if err != nil {
		return nil, fmt.Errorf("无法连接到上游代理 %s: %w", upstreamAddr, err)
	}

	switch p.config.Upstream.Protocol {
	case "http":
		return p.setupHTTPTunnel(upstreamConn, targetAddr)
	case "socks5":
		return p.setupSOCKS5Tunnel(upstreamConn, targetAddr)
	default:
		upstreamConn.Close()
		return nil, fmt.Errorf("不支持的上游代理类型: %s", p.config.Upstream.Protocol)
	}
}

func (p *HTTPProxy) setupHTTPTunnel(conn net.Conn, targetAddr string) (net.Conn, error) {
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", targetAddr, targetAddr)
	if p.config.Upstream.Username != "" && p.config.Upstream.Password != "" {
		auth := fmt.Sprintf("%s:%s", p.config.Upstream.Username, p.config.Upstream.Password)
		encoded := base64.StdEncoding.EncodeToString([]byte(auth))
		connectReq += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", encoded)
	}

	connectReq += "\r\n"

	if _, err := conn.Write([]byte(connectReq)); err != nil {
		return nil, err
	}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	response := string(buf[:n])
	if !strings.Contains(response, "200") {
		return nil, fmt.Errorf("HTTP CONNECT失败: %s", response)
	}

	return conn, nil
}
func (p *HTTPProxy) setupSOCKS5Tunnel(conn net.Conn, targetAddr string) (net.Conn, error) {
	return p.doSOCKS5Handshake(conn, targetAddr)
}
func (p *HTTPProxy) doSOCKS5Handshake(conn net.Conn, targetAddr string) (net.Conn, error) {
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer conn.SetDeadline(time.Time{})
	authMethod := byte(0x00)
	if p.config.Upstream.Username != "" && p.config.Upstream.Password != "" {
		authMethod = byte(0x02)
	}

	if authMethod == 0x02 {
		_, err := conn.Write([]byte{0x05, 0x02, 0x00, 0x02})
		if err != nil {
			return nil, fmt.Errorf("发送认证方法失败: %w", err)
		}
	} else {
		_, err := conn.Write([]byte{0x05, 0x01, 0x00})
		if err != nil {
			return nil, fmt.Errorf("发送认证方法失败: %w", err)
		}
	}

	authResp := make([]byte, 2)
	_, err := io.ReadFull(conn, authResp)
	if err != nil {
		return nil, fmt.Errorf("读取认证响应失败: %w", err)
	}

	if authResp[0] != 0x05 {
		return nil, fmt.Errorf("无效的SOCKS5版本响应: %d", authResp[0])
	}

	if authResp[1] == 0xFF {
		return nil, fmt.Errorf("服务器拒绝所有认证方法")
	}

	if authResp[1] == 0x02 {
		if err := p.doUserPassAuth(conn); err != nil {
			return nil, fmt.Errorf("用户名密码认证失败: %w", err)
		}
	}

	if err := p.sendSOCKS5ConnectRequest(conn, targetAddr); err != nil {
		return nil, fmt.Errorf("发送连接请求失败: %w", err)
	}

	return conn, nil
}
func (p *HTTPProxy) doUserPassAuth(conn net.Conn) error {
	username := p.config.Upstream.Username
	password := p.config.Upstream.Password

	authReq := []byte{0x01}
	authReq = append(authReq, byte(len(username)))
	authReq = append(authReq, []byte(username)...)
	authReq = append(authReq, byte(len(password)))
	authReq = append(authReq, []byte(password)...)
	_, err := conn.Write(authReq)
	if err != nil {
		return err
	}
	authResp := make([]byte, 2)
	_, err = io.ReadFull(conn, authResp)
	if err != nil {
		return err
	}
	if authResp[1] != 0x00 {
		return fmt.Errorf("认证失败，状态码: %d", authResp[1])
	}
	return nil
}

func (p *HTTPProxy) sendSOCKS5ConnectRequest(conn net.Conn, targetAddr string) error {
	host, portStr, err := net.SplitHostPort(targetAddr)
	if err != nil {
		return fmt.Errorf("解析目标地址失败: %w", err)
	}

	port, err := net.LookupPort("tcp", portStr)
	if err != nil {
		return fmt.Errorf("解析端口失败: %w", err)
	}

	req := []byte{0x05, 0x01, 0x00}
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			req = append(req, 0x01)
			req = append(req, ip4...)
		} else {
			req = append(req, 0x04)
			req = append(req, ip.To16()...)
		}
	} else {
		req = append(req, 0x03)
		req = append(req, byte(len(host)))
		req = append(req, []byte(host)...)
	}
	req = append(req, byte(port>>8), byte(port&0xFF))

	_, err = conn.Write(req)
	if err != nil {
		return err
	}
	resp := make([]byte, 4)
	_, err = io.ReadFull(conn, resp)
	if err != nil {
		return err
	}

	if resp[0] != 0x05 {
		return fmt.Errorf("无效的响应版本: %d", resp[0])
	}

	if resp[1] != 0x00 {
		return fmt.Errorf("连接失败，错误码: %d", resp[1])
	}
	addrType := resp[3]
	switch addrType {
	case 0x01:
		_, err = io.ReadFull(conn, make([]byte, 6))
	case 0x04:
		_, err = io.ReadFull(conn, make([]byte, 18))
	case 0x03:
		lenBuf := make([]byte, 1)
		_, err = io.ReadFull(conn, lenBuf)
		if err == nil {
			_, err = io.ReadFull(conn, make([]byte, int(lenBuf[0])+2))
		}
	default:
		return fmt.Errorf("不支持的地址类型: %d", addrType)
	}

	if err != nil {
		return fmt.Errorf("读取绑定地址失败: %w", err)
	}

	return nil
}

// 双向数据转发
func (p *HTTPProxy) relay(client, upstream net.Conn) error {
	errc := make(chan error, 2)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("客户端到上游转发时panic: %v\n", r)
			}
		}()
		_, err := io.Copy(upstream, client)
		errc <- err
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("上游到客户端转发时panic: %v\n", r)
			}
		}()
		_, err := io.Copy(client, upstream)
		errc <- err
	}()
	err := <-errc
	client.Close()
	upstream.Close()
	<-errc
	return err
}

package server

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"proxy-manager-desktop/internal/config"
)

const (
	socks5Version  = byte(0x05)
	cmdConnect     = byte(0x01)
	authNone       = byte(0x00)
	authUserPass   = byte(0x02)
	authFailed     = byte(0xFF)
	addrTypeIPv4   = byte(0x01)
	addrTypeDomain = byte(0x03)
	addrTypeIPv6   = byte(0x04)
)

type SOCKS5Proxy struct {
	config      *config.ProxyConfig
	listener    net.Listener
	isRunning   bool
	wg          sync.WaitGroup
	stopChannel chan struct{}
}

func NewSOCKS5Proxy(proxyConfig *config.ProxyConfig) (*SOCKS5Proxy, error) {
	if proxyConfig.Local.Protocol != "socks5" {
		return nil, fmt.Errorf("本地协议必须是SOCKS5，当前为: %s", proxyConfig.Local.Protocol)
	}

	proxy := &SOCKS5Proxy{
		config:      proxyConfig,
		stopChannel: make(chan struct{}),
	}

	return proxy, nil
}
func (p *SOCKS5Proxy) Start() error {
	if p.isRunning {
		return fmt.Errorf("代理已在运行")
	}

	listenAddr := fmt.Sprintf("%s:%d", p.config.Local.ListenIP, p.config.Local.ListenPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("无法创建SOCKS5监听器: %v", err)
	}
	p.listener = listener

	p.isRunning = true
	p.wg.Add(1)
	go p.serve()

	fmt.Printf("SOCKS5代理开始监听 %s\n", listenAddr)
	return nil
}

func (p *SOCKS5Proxy) Stop() error {
	if !p.isRunning {
		return fmt.Errorf("代理未运行")
	}

	p.isRunning = false

	close(p.stopChannel)
	if err := p.listener.Close(); err != nil {
		return fmt.Errorf("无法关闭SOCKS5监听器: %v", err)
	}

	go func() {
		p.wg.Wait()
		fmt.Printf("SOCKS5代理已完全停止\n")
	}()

	fmt.Printf("SOCKS5代理正在停止\n")
	return nil
}

func (p *SOCKS5Proxy) IsRunning() bool {
	return p.isRunning
}

func (p *SOCKS5Proxy) GetConfig() *config.ProxyConfig {
	return p.config
}

func (p *SOCKS5Proxy) serve() {
	defer p.wg.Done()

	for {
		select {
		case <-p.stopChannel:
			return
		default:
			conn, err := p.listener.Accept()
			if err != nil {
				select {
				case <-p.stopChannel:
					return
				default:
					fmt.Printf("接受SOCKS5连接时出错: %v\n", err)
					continue
				}
			}

			p.wg.Add(1)
			go func(c net.Conn) {
				defer p.wg.Done()
				defer c.Close()

				if err := p.handleConnection(c); err != nil {
					fmt.Printf("处理SOCKS5连接时出错: %v\n", err)
				}
			}(conn)
		}
	}
}

func (p *SOCKS5Proxy) handleConnection(conn net.Conn) error {
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer conn.Close()

	if err := p.handleAuth(conn); err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	targetAddr, err := p.handleRequest(conn)
	if err != nil {
		return fmt.Errorf("处理请求失败: %w", err)
	}

	conn.SetDeadline(time.Time{})

	upstreamConn, err := p.connectUpstream(targetAddr)
	if err != nil {
		return fmt.Errorf("连接上游失败: %w", err)
	}
	defer upstreamConn.Close()

	return p.relay(conn, upstreamConn)
}

func (p *SOCKS5Proxy) handleAuth(conn net.Conn) error {
	buf := make([]byte, 257)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	if n < 3 || buf[0] != socks5Version {
		return fmt.Errorf("无效的SOCKS5版本")
	}

	_, err = conn.Write([]byte{socks5Version, authNone})
	return err
}

func (p *SOCKS5Proxy) handleRequest(conn net.Conn) (string, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	if n < 7 || buf[0] != socks5Version || buf[1] != cmdConnect {
		conn.Write([]byte{socks5Version, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return "", fmt.Errorf("不支持的命令")
	}

	var targetAddr string
	switch buf[3] {
	case addrTypeIPv4:
		if n < 10 {
			return "", fmt.Errorf("IPv4地址长度不足")
		}
		ip := net.IP(buf[4:8]).String()
		port := binary.BigEndian.Uint16(buf[8:10])
		targetAddr = fmt.Sprintf("%s:%d", ip, port)
	case addrTypeIPv6:
		if n < 22 {
			return "", fmt.Errorf("IPv6地址长度不足")
		}
		ip := net.IP(buf[4:20]).String()
		port := binary.BigEndian.Uint16(buf[20:22])
		targetAddr = fmt.Sprintf("[%s]:%d", ip, port)
	case addrTypeDomain:
		if n < 7 {
			return "", fmt.Errorf("域名长度不足")
		}
		domainLen := int(buf[4])
		if n < 7+domainLen {
			return "", fmt.Errorf("域名数据不完整")
		}
		domain := string(buf[5 : 5+domainLen])
		port := binary.BigEndian.Uint16(buf[5+domainLen : 7+domainLen])
		targetAddr = fmt.Sprintf("%s:%d", domain, port)
	default:
		conn.Write([]byte{socks5Version, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return "", fmt.Errorf("不支持的地址类型")
	}

	response := []byte{socks5Version, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
	_, err = conn.Write(response)
	if err != nil {
		return "", err
	}

	return targetAddr, nil
}

func (p *SOCKS5Proxy) connectUpstream(targetAddr string) (net.Conn, error) {
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

func (p *SOCKS5Proxy) setupHTTPTunnel(conn net.Conn, targetAddr string) (net.Conn, error) {
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", targetAddr, targetAddr)

	if p.config.Upstream.Username != "" && p.config.Upstream.Password != "" {
		auth := fmt.Sprintf("%s:%s", p.config.Upstream.Username, p.config.Upstream.Password)
		encoded := base64Encode(auth)
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
	if !contains(response, "200") {
		return nil, fmt.Errorf("HTTP CONNECT失败: %s", response)
	}

	return conn, nil
}

func (p *SOCKS5Proxy) setupSOCKS5Tunnel(conn net.Conn, targetAddr string) (net.Conn, error) {
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer conn.SetDeadline(time.Time{})

	authMethod := authNone
	if p.config.Upstream.Username != "" && p.config.Upstream.Password != "" {
		authMethod = authUserPass
	}

	if authMethod == authUserPass {
		_, err := conn.Write([]byte{socks5Version, 0x02, authNone, authUserPass})
		if err != nil {
			return nil, fmt.Errorf("发送认证方法失败: %w", err)
		}
	} else {
		_, err := conn.Write([]byte{socks5Version, 0x01, authNone})
		if err != nil {
			return nil, fmt.Errorf("发送认证方法失败: %w", err)
		}
	}

	authResp := make([]byte, 2)
	_, err := io.ReadFull(conn, authResp)
	if err != nil {
		return nil, fmt.Errorf("读取认证响应失败: %w", err)
	}

	if authResp[0] != socks5Version {
		return nil, fmt.Errorf("无效的SOCKS5版本响应: %d", authResp[0])
	}

	if authResp[1] == authFailed {
		return nil, fmt.Errorf("服务器拒绝所有认证方法")
	}

	if authResp[1] == authUserPass {
		if err := p.doUserPassAuth(conn); err != nil {
			return nil, fmt.Errorf("用户名密码认证失败: %w", err)
		}
	}

	if err := p.sendConnectRequest(conn, targetAddr); err != nil {
		return nil, fmt.Errorf("发送连接请求失败: %w", err)
	}

	return conn, nil
}

func (p *SOCKS5Proxy) doUserPassAuth(conn net.Conn) error {
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

func (p *SOCKS5Proxy) sendConnectRequest(conn net.Conn, targetAddr string) error {
	host, portStr, err := net.SplitHostPort(targetAddr)
	if err != nil {
		return fmt.Errorf("解析目标地址失败: %w", err)
	}

	port, err := net.LookupPort("tcp", portStr)
	if err != nil {
		return fmt.Errorf("解析端口失败: %w", err)
	}

	req := []byte{socks5Version, cmdConnect, 0x00}

	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			req = append(req, addrTypeIPv4)
			req = append(req, ip4...)
		} else {
			req = append(req, addrTypeIPv6)
			req = append(req, ip.To16()...)
		}
	} else {
		req = append(req, addrTypeDomain)
		req = append(req, byte(len(host)))
		req = append(req, []byte(host)...)
	}

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	req = append(req, portBytes...)

	_, err = conn.Write(req)
	if err != nil {
		return err
	}

	resp := make([]byte, 4)
	_, err = io.ReadFull(conn, resp)
	if err != nil {
		return err
	}

	if resp[0] != socks5Version {
		return fmt.Errorf("无效的响应版本: %d", resp[0])
	}

	if resp[1] != 0x00 {
		return fmt.Errorf("连接失败，错误码: %d", resp[1])
	}

	addrType := resp[3]
	switch addrType {
	case addrTypeIPv4:
		_, err = io.ReadFull(conn, make([]byte, 6))
	case addrTypeIPv6:
		_, err = io.ReadFull(conn, make([]byte, 18))
	case addrTypeDomain:
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

func (p *SOCKS5Proxy) relay(client, upstream net.Conn) error {
	errc := make(chan error, 2)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("客户端到上游转发时panic: %v\n", r)
			}
		}()
		_, err := io.Copy(upstream, client)
		if err != nil {
			fmt.Printf("客户端到上游转发错误: %v\n", err)
		}
		errc <- err
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("上游到客户端转发时panic: %v\n", r)
			}
		}()
		_, err := io.Copy(client, upstream)
		if err != nil {
			fmt.Printf("上游到客户端转发错误: %v\n", err)
		}
		errc <- err
	}()

	err := <-errc
	client.Close()
	upstream.Close()
	<-errc
	return err
}

func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

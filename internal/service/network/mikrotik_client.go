package network

import (
	//"bytes"
	"fmt"
	"net"
	//"strconv"
	"time"

	"Extreme-Solutions/internal/config"
)

type MikrotikClient struct {
	conn    net.Conn
	timeout time.Duration
}

// NewMikrotikClient instantiates a raw TCP connection to the RouterOS API
func NewMikrotikClient(cfg *config.Config) (*MikrotikClient, error) {
	address := fmt.Sprintf("%s:%d", cfg.Mikrotik.Host, cfg.Mikrotik.Port)
	conn, err := net.DialTimeout("tcp", address, cfg.Mikrotik.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MikroTik router: %w", err)
	}

	client := &MikrotikClient{
		conn:    conn,
		timeout: cfg.Mikrotik.Timeout,
	}

	// Authenticate immediately upon connection
	if err := client.login(cfg.Mikrotik.Username, cfg.Mikrotik.Password); err != nil {
		conn.Close()
		return nil, fmt.Errorf("mikrotik authentication failed: %w", err)
	}

	return client, nil
}

// Close gracefully tears down the network socket
func (m *MikrotikClient) Close() error {
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}

// SetSubscriberQueue adds or updates an active rate-limiting simple queue on the router
func (m *MikrotikClient) SetSubscriberQueue(name, targetIP string, maxUpKbps, maxDownKbps int) error {
	// Format limits into RouterOS convention (e.g., "1024k/5120k" for 1M Up / 5M Down)
	limitAttr := fmt.Sprintf("%dk/%dk", maxUpKbps, maxDownKbps)

	// Build an atomic command sentence to upsert a simple queue rule
	sentence := []string{
		"/queue/simple/add",
		fmt.Sprintf("=name=%s", name),
		fmt.Sprintf("=target=%s", targetIP),
		fmt.Sprintf("=max-limit=%s", limitAttr),
		"=comment=Managed by Extreme Solutions ISP Core",
	}

	_, err := m.sendSentence(sentence)
	if err != nil {
		return fmt.Errorf("failed to provision queue rules: %w", err)
	}
	return nil
}

// RemoveSubscriberQueue cuts off network routing access by stripping the queue name
func (m *MikrotikClient) RemoveSubscriberQueue(name string) error {
	sentence := []string{
		"/queue/simple/remove",
		fmt.Sprintf("=.id=%s", name),
	}
	_, err := m.sendSentence(sentence)
	return err
}

// Internal connection helpers for RouterOS word packing length formatting
func (m *MikrotikClient) login(username, password string) error {
	sentence := []string{
		"/login",
		fmt.Sprintf("=name=%s", username),
		fmt.Sprintf("=password=%s", password),
	}
	res, err := m.sendSentence(sentence)
	if err != nil {
		return err
	}
	if len(res) > 0 && res[0] == "!trap" {
		return fmt.Errorf("access denied or bad credentials")
	}
	return nil
}

func (m *MikrotikClient) sendSentence(words []string) ([]string, error) {
	_ = m.conn.SetDeadline(time.Now().Add(m.timeout))

	// Write data out encoded with RouterOS length prefixes
	for _, word := range words {
		if err := m.writeWord(word); err != nil {
			return nil, err
		}
	}
	if err := m.writeWord(""); err != nil { // Zero byte terminates sentence
		return nil, err
	}

	return m.readResponse()
}

func (m *MikrotikClient) writeWord(word string) error {
	b := []byte(word)
	l := len(b)
	var prefix []byte

	if l < 0x80 {
		prefix = []byte{byte(l)}
	} else if l < 0x4000 {
		l |= 0x8000
		prefix = []byte{byte(l >> 8), byte(l)}
	} else {
		return fmt.Errorf("command length too long")
	}

	if _, err := m.conn.Write(prefix); err != nil {
		return err
	}
	if _, err := m.conn.Write(b); err != nil {
		return err
	}
	return nil
}

func (m *MikrotikClient) readResponse() ([]string, error) {
	var reply []string
	for {
		word, err := m.readWord()
		if err != nil {
			return nil, err
		}
		if word == "" {
			break
		}
		reply = append(reply, word)
	}
	return reply, nil
}

func (m *MikrotikClient) readWord() (string, error) {
	var prefix [1]byte
	if _, err := m.conn.Read(prefix[:]); err != nil {
		return "", err
	}

	var length int
	if prefix[0]&0x80 == 0 {
		length = int(prefix[0])
	} else if prefix[0]&0xC0 == 0x80 {
		var remainder [1]byte
		if _, err := m.conn.Read(remainder[:]); err != nil {
			return "", err
		}
		length = (int(prefix[0]&0x3F) << 8) + int(remainder[0])
	} else {
		return "", fmt.Errorf("unsupported chunking signature")
	}

	if length == 0 {
		return "", nil
	}

	buf := make([]byte, length)
	_, err := m.conn.Read(buf)
	return string(buf), err
}

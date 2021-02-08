package stats

import (
	"bytes"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSvr struct {
	Port int
	sock *net.UDPConn
	Count int64
}

func newTestSvr(t *testing.T) *testSvr {
	s := testSvr{}
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	require.NoError(t, err)

	s.sock, err = net.ListenUDP("udp", addr)
	require.NoError(t, err)

	laddr := s.sock.LocalAddr().(*net.UDPAddr)
	s.Port = laddr.Port
	return &s
}

func (s *testSvr) ReadMsg(t *testing.T) {
	buffer := make([]byte, 8193)
	s.sock.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		n, err := s.sock.Read(buffer)
		require.NoError(t, err, "read failed on message %d", atomic.LoadInt64(&s.Count))
		buffer = buffer[:n]
		if bytes.HasPrefix(buffer, []byte("test_namespace.exit:9999")) {
			break
		}
		atomic.AddInt64(&s.Count, 1)
	}
}

func (s *testSvr) Close() {
	s.sock.Close()
}

func TestCount(t *testing.T) {
	s := newTestSvr(t)
	defer s.Close()
	os.Setenv("STATSD_NAMESPACE", "test_namespace")
	os.Setenv("STATSD_PORT", strconv.Itoa(s.Port))
	go s.ReadMsg(t)

	d, err := NewDogstatsD("127.0.0.1", s.Port, "test_namespace", "test:tag")
	require.NoError(t, err)
	d.Count("test", 1, "a", "b")

	d.Count("exit", 9999)
	time.Sleep(time.Millisecond)
	assert.Equal(t, int64(1), atomic.LoadInt64(&s.Count))
}

func TestCount100000(t *testing.T) {
	s := newTestSvr(t)
	defer s.Close()
	os.Setenv("STATSD_NAMESPACE", "test_namespace")
	os.Setenv("STATSD_PORT", strconv.Itoa(s.Port))
	go s.ReadMsg(t)

	d, err := NewDogstatsD("127.0.0.1", s.Port, "test_namespace", "test:tag")
	require.NoError(t, err)
	start := time.Now()
	n := 10000
	for i := 0; i < n; i++ {
		d.Count("test", 1, "a", "b")
	}
	delta := time.Since(start)
	t.Logf("sending %d metrics took %s - %s per message", n, delta, delta/time.Duration(n))
	d.Count("exit", 9999)
	time.Sleep(time.Millisecond)
	assert.Equal(t, int64(n), atomic.LoadInt64(&s.Count))
}
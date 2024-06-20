package sentinel

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/Bose/minisentinel"
	"github.com/alicebob/miniredis/v2"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/suite"
)

type SentinelTestSuite struct {
	suite.Suite
	masterRedis *miniredis.Miniredis
	slaveRedis  *miniredis.Miniredis
	sentinel1   *minisentinel.Sentinel
}

func TestSentinelSuite(t *testing.T) {
	suite.Run(t, new(SentinelTestSuite))
}

func (s *SentinelTestSuite) SetupTest() {
	s.masterRedis = miniredis.NewMiniRedis()
	err := s.masterRedis.StartAddr(":6379")
	s.NoError(err)

	s.slaveRedis = miniredis.NewMiniRedis()
	err = s.slaveRedis.StartAddr(":6378")
	s.NoError(err)
	s.sentinel1 = minisentinel.NewSentinel(s.masterRedis, minisentinel.WithReplica(s.slaveRedis))
	err = s.sentinel1.StartAddr(":26379")
	s.NoError(err)
}

func (s *SentinelTestSuite) TearDownTest() {
	s.masterRedis.Close()
	s.slaveRedis.Close()
	s.sentinel1.Close()
}

func (s *SentinelTestSuite) Test_MasterAddr_Success() {
	sntnl := &Sentinel{
		Addrs:      []string{s.sentinel1.Addr()},
		MasterName: "mymaster",
		Dial:       func(addr string) (redigo.Conn, error) { return redigo.Dial("tcp", addr) },
	}

	addr, err := sntnl.MasterAddr()
	s.NoError(err)
	s.Equal("[::]:6379", addr)
}

func (s *SentinelTestSuite) Test_MasterAddr_PostSwitch_Success() {
	sntnl := &Sentinel{
		Addrs:      []string{s.sentinel1.Addr()},
		MasterName: "mymaster",
		Dial:       func(addr string) (redigo.Conn, error) { return redigo.Dial("tcp", addr) },
	}

	addr, err := sntnl.MasterAddr()
	s.NoError(err)
	s.Equal("[::]:6379", addr)

	s.simulateFailover()

	addr, err = sntnl.MasterAddr()
	s.NoError(err)
	s.Equal("[::]:6378", addr)
}

func (s *SentinelTestSuite) Test_MasterAddr_WithFailoverCallback_Success() {
	sntnl := &Sentinel{
		Addrs:      []string{s.sentinel1.Addr()},
		MasterName: "mymaster",
		Dial:       func(addr string) (redigo.Conn, error) { return redigo.Dial("tcp", addr) },
	}

	var buf1, buf2 bytes.Buffer
	callbackFunc1 := func(oldMasterAddr, newMasterAddr string) {
		_, _ = fmt.Fprintf(&buf1, "Executing callback1 func %s %s", oldMasterAddr, newMasterAddr)
	}
	callbackFunc2 := func(oldMasterAddr, newMasterAddr string) {
		_, _ = fmt.Fprintf(&buf2, "Executing callback2 func %s %s", oldMasterAddr, newMasterAddr)
	}

	// register callback
	sntnl.RegisterFailoverCallback(callbackFunc1)
	sntnl.RegisterFailoverCallback(callbackFunc2)

	addr, err := sntnl.MasterAddr()
	s.NoError(err)
	s.Equal("[::]:6379", addr)

	s.simulateFailover()

	addr, err = sntnl.MasterAddr()
	time.Sleep(2 * time.Second)
	s.NoError(err)
	s.Equal("[::]:6378", addr)
	s.Equal("Executing callback1 func [::]:6379 [::]:6378", buf1.String())
	s.Equal("Executing callback2 func [::]:6379 [::]:6378", buf2.String())
}

func (s *SentinelTestSuite) Test_MasterAddr_WithFailoverCallback_RecoveredFromPanic() {
	sntnl := &Sentinel{
		Addrs:      []string{s.sentinel1.Addr()},
		MasterName: "mymaster",
		Dial:       func(addr string) (redigo.Conn, error) { return redigo.Dial("tcp", addr) },
	}

	callbackFunc := func(oldMasterAddr, newMasterAddr string) {
		panic("simulated panic error\n")
	}
	// register callback
	sntnl.RegisterFailoverCallback(callbackFunc)

	addr, err := sntnl.MasterAddr()
	s.NoError(err)
	s.Equal("[::]:6379", addr)

	s.simulateFailover()

	addr, err = sntnl.MasterAddr()
	time.Sleep(1 * time.Second)
	s.NoError(err)
	s.Equal("[::]:6378", addr)
}

func (s *SentinelTestSuite) simulateFailover() {
	// simulate fail-over by switching the master
	master := s.sentinel1.Master()
	s.sentinel1.WithMaster(s.sentinel1.Replica())
	s.sentinel1.SetReplica(master)
}

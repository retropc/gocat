package main

import (
	"bufio"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
)

type SocketTable struct {
  f *os.File
  i *bufio.Scanner
}

type SocketTableEntry struct {
  Local  string
  Remote string
  Uid    int
}

func NewSocketTable() (*SocketTable, error) {
  f, err := os.Open("/proc/net/tcp")
  if err != nil {
    return nil, err
  }

  b := bufio.NewScanner(f)
  b.Scan() // skip first line
  return &SocketTable{f, b}, nil
}

func (s *SocketTable) Next() bool {
  return s.i.Scan()
}

func (s *SocketTable) Value() (*SocketTableEntry, error) {
  // "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode"
  //     0  1             2             3  4                 5           6            7        8 9 10 11
  // "  10: 0100007F:AABB 0100007F:CCDD 06 00000000:00000000 03:00000E80 00000000     0        0 0 3 0000000000000000"
  v := strings.Fields(s.i.Text())

  uid, err := strconv.ParseInt(v[7], 10, 32)
  if err != nil {
    return nil, err
  }

  return &SocketTableEntry{v[1], v[2], int(uid)}, nil
}

func (s *SocketTable) Close() error {
  return s.f.Close()
}

const hextable = "0123456789ABCDEF"

// yeah it would probably be nicer to parse the addrs and return them but oh well
func AddrToHex(a net.Addr) (string, error) {
  a2, ok := a.(*net.TCPAddr)
  if !ok {
    return "", errors.New("unknown address type")
  }

  i4 := a2.IP.To4()
  p2 := a2.Port
  out := make([]byte, 8+1+4)
  out[0] = hextable[i4[3]>>4]
  out[1] = hextable[i4[3]&0x0f]
  out[2] = hextable[i4[2]>>4]
  out[3] = hextable[i4[2]&0x0f]
  out[4] = hextable[i4[1]>>4]
  out[5] = hextable[i4[1]&0x0f]
  out[6] = hextable[i4[0]>>4]
  out[7] = hextable[i4[0]&0x0f]
  out[8] = ':'
  out[9] = hextable[(p2>>12)&0x0f]
  out[10] = hextable[(p2>>8)&0x0f]
  out[11] = hextable[(p2>>4)&0x0f]
  out[12] = hextable[p2&0x0f]

  return string(out), nil
}
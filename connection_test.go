// The connection layer is tested using real UDP, at the expense of some speed, to find issues with connection.go's
// usage of the UDP library.
package main

import (
    "testing"
    "net"
    "reflect"
    "fmt"
    "time"
)

// Very simple UDP client for interacting with the server.
type TestClient struct {
    conn *net.UDPConn
    sessionAddr *net.UDPAddr
    serverAddr *net.UDPAddr
}

func (t *TestClient) VerifyReceived(expected []byte) {
    reply, err := t.AwaitReceive()
    if err != nil {
        panic(err)
    }

    if !reflect.DeepEqual(expected, reply) {
        panic(fmt.Errorf("Read unexpected reply: %v", reply))
    }
}

func (t *TestClient) AwaitReceive() ([]byte, error) {
    buf := make([]byte, MaxPacketSize)
    t.conn.SetReadDeadline(time.Now().Add(time.Second))
    bytesRead, replyAddr, err := t.conn.ReadFromUDP(buf)
    if err != nil {
        return nil, err
    }
    if t.sessionAddr == nil {
        t.sessionAddr = replyAddr
    }

    return buf[:bytesRead], nil
}

func (t *TestClient) SendSession(data []byte) {
    _, err := t.conn.WriteToUDP(data, t.sessionAddr)
    if err != nil {
        panic(err)
    }
}

func (t *TestClient) SendServer(data []byte) {
    _, err := t.conn.WriteToUDP(data, t.serverAddr)
    if err != nil {
        panic(err)
    }
}

func MakeTestClient(raddr *net.UDPAddr) *TestClient {
    clientAddr := net.UDPAddr{
        IP: net.ParseIP("127.0.0.1"),
        Port: 0,
    }
    conn, _ := net.ListenUDP("udp", &clientAddr)

    return &TestClient{
        conn: conn,
        serverAddr: raddr,
    }
}

func TestListen(t *testing.T) {
    fs := MakeFileSystem()

    options := ConnectionOptions {
        Host: "127.0.0.1",
        IntroductionPort: 11235,
        Timeout: 100 * time.Millisecond,
        MaxRetries: 1,
    }
    // Here we use a port > 1024 so we don't need su for testing.
    go ListenForNewConnections(&options, fs)

    serverAddr := net.UDPAddr{
        IP: net.ParseIP("127.0.0.1"),
        Port: 11235,
    }

    // Unfortunately since ListenForNewConnections is a goroutine, in lieu of
    // synchronisation we make due with sleep for now. Not ideal but could be changed.
    time.Sleep(1 * time.Second)

    BasicRequestReply(MakeTestClient(&serverAddr))
    ResendTimeout(MakeTestClient(&serverAddr))
    FirstPacketIsBad(MakeTestClient(&serverAddr))
    MaxRetries(MakeTestClient(&serverAddr))
}

// This should serve as a basic end-to-end systems test to validate that the layers are wired up correctly.
func BasicRequestReply(client *TestClient) {
    client.SendServer([]byte{0, PKT_WRQ, 'a', 0, 'o', 'c', 't', 'a', 'l', 0})
    client.VerifyReceived([]byte{0, PKT_ACK, 0, 0})
    client.SendSession([]byte{0, PKT_DATA, 0, 1, 'a'})
    client.VerifyReceived([]byte{0, PKT_ACK, 0, 1})
}

func ResendTimeout(client *TestClient) {
    client.SendServer([]byte{0, PKT_WRQ, 'b', 0, 'o', 'c', 't', 'a', 'l', 0})
    client.VerifyReceived([]byte{0, PKT_ACK, 0, 0})
    // Instead of replying, we wait for a retry.
    client.VerifyReceived([]byte{0, PKT_ACK, 0, 0})
}

func FirstPacketIsBad(client *TestClient) {
    client.SendServer([]byte("hi"))
    received, _ := client.AwaitReceive()
    if (ConvertToUInt16(received[:2]) != PKT_ERROR) {
        panic("Did not receive error packet in response!")
    }
}

func MaxRetries(client *TestClient) {
    client.SendServer([]byte{0, PKT_WRQ, 'b', 0, 'o', 'c', 't', 'a', 'l', 0})
    client.VerifyReceived([]byte{0, PKT_ACK, 0, 0})
    client.VerifyReceived([]byte{0, PKT_ACK, 0, 0})
    // If we wait after two replies.
    time.Sleep(100 * time.Millisecond)
    // At this point our read session should time out since the sesion should have been destroyed.
    _, err := client.AwaitReceive()
    if err == nil {
        panic(fmt.Errorf("Should have out awaiting receive, since the session was destroyed."))
    }
    fmt.Println(err)
}

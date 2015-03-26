// TFTP Daemon
// Implements RFC 1350, in octet mode only, over UDP and with files stored in memory only.
// There are three layers:
// * Connection layer - listens for connection requests and communicates with callers.
// * Session layer    - receives request packets and returns reply packets.
// * Filesystem layer - provides a simple in-memory file store.

package main

import (
	"flag"
	"log"
	"os"
)

var Log = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)

func main() {
	listenPort := flag.Int("port", 69, "port to listen on.")
	host := flag.String("host", "127.0.0.1", "host address to listen on.")
	flag.Parse()

	Log.Printf("Listening on host %s, port %d\n", *host, *listenPort)

	fs := MakeFileSystem()
	ListenForNewConnections(*host, *listenPort, fs)
}

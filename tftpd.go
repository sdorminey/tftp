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
    "time"
)

var Log = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)

func main() {
    var options ConnectionOptions

	options.IntroductionPort = *flag.Int("port", 69, "port to listen on.")
	options.Host = *flag.String("host", "127.0.0.1", "host address to listen on.")
    options.MaxRetries = *flag.Int("maxretries", 3, "maximum amount of times to retry a send before terminating the connection.")
    timeoutSeconds := *flag.Int("timeout", 3, "receive timeout in seconds before resending the last packet.")
    options.Timeout = time.Second * time.Duration(timeoutSeconds)
	flag.Parse()

	Log.Printf("Listening on host %s, port %d\n", options.Host, options.IntroductionPort)

	fs := MakeFileSystem()
	ListenForNewConnections(&options, fs)
}
